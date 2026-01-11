package trello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HTTPHandlers exposes Trello actions.
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds Trello HTTP handlers with a default client.
func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// CreateCard handles POST /actions/trello/card
func (h *HTTPHandlers) CreateCard() http.Handler {
	type payload struct {
		ListID string `json:"list_id"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Pos    string `json:"pos"`
		APIKey string `json:"api_key"`
		Token  string `json:"token"`
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
		if strings.TrimSpace(p.ListID) == "" || strings.TrimSpace(p.Name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "list_id and name are required"})
			return
		}
		client, err := clientFromPayload(h.client, p.APIKey, p.Token)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := client.CreateCard(r.Context(), p.ListID, p.Name, p.Desc, p.Pos); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// MoveCard handles POST /actions/trello/card/move
func (h *HTTPHandlers) MoveCard() http.Handler {
	type payload struct {
		CardID string `json:"card_id"`
		ListID string `json:"list_id"`
		Pos    string `json:"pos"`
		APIKey string `json:"api_key"`
		Token  string `json:"token"`
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
		if strings.TrimSpace(p.CardID) == "" || strings.TrimSpace(p.ListID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "card_id and list_id are required"})
			return
		}
		client, err := clientFromPayload(h.client, p.APIKey, p.Token)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := client.MoveCard(r.Context(), p.CardID, p.ListID, p.Pos); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "moved"})
	})
}

// CreateList handles POST /actions/trello/list
func (h *HTTPHandlers) CreateList() http.Handler {
	type payload struct {
		BoardID string `json:"board_id"`
		Name    string `json:"name"`
		Pos     string `json:"pos"`
		APIKey  string `json:"api_key"`
		Token   string `json:"token"`
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
		if strings.TrimSpace(p.BoardID) == "" || strings.TrimSpace(p.Name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "board_id and name are required"})
			return
		}
		client, err := clientFromPayload(h.client, p.APIKey, p.Token)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := client.CreateList(r.Context(), p.BoardID, p.Name, p.Pos); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// clientFromPayload builds a Trello client from payload credentials or returns the default.
func clientFromPayload(defaultClient *Client, apiKey, token string) (*Client, error) {
	key := strings.TrimSpace(apiKey)
	tok := strings.TrimSpace(token)
	if key == "" && tok == "" {
		return defaultClient, nil
	}
	if key == "" || tok == "" {
		return nil, fmt.Errorf("api_key and token are required")
	}
	return NewClientWithCredentials(key, tok), nil
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
