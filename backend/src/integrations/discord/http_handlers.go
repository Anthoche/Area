package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// HTTPHandlers exposes Discord bot actions.
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds Discord HTTP handlers with a default client.
func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// Message handles POST /actions/discord/message
func (h *HTTPHandlers) Message() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.Content) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id and content are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.SendMessage(r.Context(), p.ChannelID, p.Content, nil); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

// Embed handles POST /actions/discord/embed
func (h *HTTPHandlers) Embed() http.Handler {
	type payload struct {
		ChannelID   string          `json:"channel_id"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		URL         string          `json:"url"`
		Color       json.RawMessage `json:"color"`
		Content     string          `json:"content"`
		BotToken    string          `json:"bot_token"`
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
		if strings.TrimSpace(p.ChannelID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id is required"})
			return
		}
		if strings.TrimSpace(p.Title) == "" && strings.TrimSpace(p.Description) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title or description is required"})
			return
		}
		color, err := parseColor(p.Color)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		embed := Embed{
			Title:       p.Title,
			Description: p.Description,
			URL:         p.URL,
			Color:       color,
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.SendMessage(r.Context(), p.ChannelID, p.Content, []Embed{embed}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

// Edit handles POST /actions/discord/message/edit
func (h *HTTPHandlers) Edit() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageID string `json:"message_id"`
		Content   string `json:"content"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageID) == "" || strings.TrimSpace(p.Content) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id, message_id and content are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.EditMessage(r.Context(), p.ChannelID, p.MessageID, p.Content); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
}

// Delete handles POST /actions/discord/message/delete
func (h *HTTPHandlers) Delete() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageID string `json:"message_id"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id and message_id are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.DeleteMessage(r.Context(), p.ChannelID, p.MessageID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	})
}

// React handles POST /actions/discord/message/react
func (h *HTTPHandlers) React() http.Handler {
	type payload struct {
		ChannelID string `json:"channel_id"`
		MessageID string `json:"message_id"`
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
		if strings.TrimSpace(p.ChannelID) == "" || strings.TrimSpace(p.MessageID) == "" || strings.TrimSpace(p.Emoji) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel_id, message_id and emoji are required"})
			return
		}
		client := h.client
		if strings.TrimSpace(p.BotToken) != "" {
			client = NewClientWithToken(p.BotToken)
		}
		if err := client.AddReaction(r.Context(), p.ChannelID, p.MessageID, p.Emoji); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "reacted"})
	})
}

// parseColor parses a color from a JSON raw message
func parseColor(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, nil
	}
	var asInt int
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}
	var asStr string
	if err := json.Unmarshal(raw, &asStr); err != nil {
		return 0, fmt.Errorf("color must be a hex string or number")
	}
	asStr = strings.TrimSpace(asStr)
	asStr = strings.TrimPrefix(asStr, "#")
	asStr = strings.TrimPrefix(asStr, "0x")
	if asStr == "" {
		return 0, nil
	}
	val, err := strconv.ParseInt(asStr, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid color value")
	}
	return int(val), nil
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
