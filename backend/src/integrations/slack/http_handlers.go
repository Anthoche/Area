package slack

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandlers exposes Slack bot actions.
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds Slack HTTP handlers with a default client.
func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// Message handles POST /actions/slack/message
func (h *HTTPHandlers) Message() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		Text      string `json:"text"`
		BotToken  string `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id and text are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.SendMessage(r.Context(), p.ChannelID, p.Text); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

// Blocks handles POST /actions/slack/blocks
func (h *HTTPHandlers) Blocks() http.Handler {
	type payload struct {
		ChannelID string          `json:"channel_id"`
		Text      string          `json:"text"`
		Blocks    json.RawMessage `json:"blocks"`
		BotToken  string          `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" || len(p.Blocks) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id and blocks are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.SendBlocks(r.Context(), p.ChannelID, p.Text, p.Blocks); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

// Update handles POST /actions/slack/message/update
func (h *HTTPHandlers) Update() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageTS string `json:"message_ts"`
		Text      string `json:"text"`
		BotToken  string `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageTS) == "" || strings.TrimSpace(p.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id, message_ts and text are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.UpdateMessage(r.Context(), p.ChannelID, p.MessageTS, p.Text); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
}

// Delete handles POST /actions/slack/message/delete
func (h *HTTPHandlers) Delete() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageTS string `json:"message_ts"`
		BotToken  string `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageTS) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id and message_ts are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.DeleteMessage(r.Context(), p.ChannelID, p.MessageTS); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	})
}

// React handles POST /actions/slack/message/react
func (h *HTTPHandlers) React() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageTS string `json:"message_ts"`
		Emoji     string `json:"emoji"`
		BotToken  string `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageTS) == "" || strings.TrimSpace(p.Emoji) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id, message_ts and emoji are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.AddReaction(r.Context(), p.ChannelID, p.MessageTS, p.Emoji); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "reacted"})
	})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
