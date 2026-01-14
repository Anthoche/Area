package notion

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandlers exposes Notion actions.
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds Notion HTTP handlers with a default client.
func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// Page handles POST /actions/notion/page
func (h *HTTPHandlers) Page() http.Handler {
	type payload struct {
		ParentPageID string          `json:"parent_page_id"`
		Title        string          `json:"title"`
		Content      string          `json:"content"`
		Blocks       json.RawMessage `json:"blocks"`
		BotToken     string          `json:"bot_token"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if strings.TrimSpace(p.ParentPageID) == "" || strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.BotToken) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "parent_page_id, title and bot_token are required"})
			return
		}
		client := NewClientWithToken(p.BotToken)
		if err := client.CreatePage(r.Context(), p.ParentPageID, p.Title, p.Content, p.Blocks); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// AppendBlocks handles POST /actions/notion/blocks
func (h *HTTPHandlers) AppendBlocks() http.Handler {
	type payload struct {
		BlockID  string          `json:"block_id"`
		Blocks   json.RawMessage `json:"blocks"`
		BotToken string          `json:"bot_token"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if strings.TrimSpace(p.BlockID) == "" || len(p.Blocks) == 0 || strings.TrimSpace(p.BotToken) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "block_id, blocks and bot_token are required"})
			return
		}
		client := NewClientWithToken(p.BotToken)
		if err := client.AppendBlocks(r.Context(), p.BlockID, p.Blocks); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
}

// Database handles POST /actions/notion/database
func (h *HTTPHandlers) Database() http.Handler {
	type payload struct {
		DatabaseID string          `json:"database_id"`
		Properties json.RawMessage `json:"properties"`
		Children   json.RawMessage `json:"children"`
		BotToken   string          `json:"bot_token"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if strings.TrimSpace(p.DatabaseID) == "" || len(p.Properties) == 0 || strings.TrimSpace(p.BotToken) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "database_id, properties and bot_token are required"})
			return
		}
		client := NewClientWithToken(p.BotToken)
		if err := client.CreateDatabaseRow(r.Context(), p.DatabaseID, p.Properties, p.Children); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// UpdatePage handles POST /actions/notion/page/update
func (h *HTTPHandlers) UpdatePage() http.Handler {
	type payload struct {
		PageID     string          `json:"page_id"`
		Properties json.RawMessage `json:"properties"`
		BotToken   string          `json:"bot_token"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if strings.TrimSpace(p.PageID) == "" || len(p.Properties) == 0 || strings.TrimSpace(p.BotToken) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "page_id, properties and bot_token are required"})
			return
		}
		client := NewClientWithToken(p.BotToken)
		if err := client.UpdatePage(r.Context(), p.PageID, p.Properties); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
