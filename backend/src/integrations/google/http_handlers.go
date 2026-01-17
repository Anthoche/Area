package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HTTPHandlers groups the HTTP handlers related to Google OAuth and actions.
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds a helper with the given client (or a default one if nil).
func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// Login redirects to Google OAuth consent.
func (h *HTTPHandlers) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := randomState()
		http.SetCookie(w, &http.Cookie{
			Name:     "oauthstate",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
		})
		if uiRedirect := r.URL.Query().Get("ui_redirect"); uiRedirect != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "oauthredirect",
				Value:    url.QueryEscape(uiRedirect),
				Path:     "/",
				HttpOnly: true,
			})
		}
		if userID := r.URL.Query().Get("user_id"); userID != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "oauthuserid",
				Value:    userID,
				Path:     "/",
				HttpOnly: true,
			})
		}
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}
		authURL := h.client.AuthURL(state, redirectURI)
		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// Start returns an auth_url for clients that want a JSON response.
func (h *HTTPHandlers) Start() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		state := randomState()
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}
		authURL := h.client.AuthURL(state, redirectURI)
		writeJSON(w, http.StatusOK, map[string]string{"auth_url": authURL, "state": state})
	})
}

// Callback exchanges code for token, stores it, and returns token_id/email.
func (h *HTTPHandlers) Callback() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateCookie, _ := r.Cookie("oauthstate")
		if stateCookie == nil || r.URL.Query().Get("state") != stateCookie.Value {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oauth state"})
			return
		}
		userID := optionalUserID(r)
		if userID == nil {
			if userCookie, _ := r.Cookie("oauthuserid"); userCookie != nil {
				if id, err := strconv.ParseInt(userCookie.Value, 10, 64); err == nil && id > 0 {
					userID = &id
				}
			}
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
			return
		}
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}
		tokenID, resolvedUserID, email, err := h.client.ExchangeAndStore(r.Context(), code, redirectURI, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if uiCookie, _ := r.Cookie("oauthredirect"); uiCookie != nil && uiCookie.Value != "" {
			if dest, err := url.QueryUnescape(uiCookie.Value); err == nil {
				if redir, err := url.Parse(dest); err == nil && (redir.Scheme == "http" || redir.Scheme == "https") {
					q := redir.Query()
					q.Set("token_id", strconv.FormatInt(tokenID, 10))
					if resolvedUserID > 0 {
						q.Set("user_id", strconv.FormatInt(resolvedUserID, 10))
					}
					if email != "" {
						q.Set("google_email", email)
					}
					redir.RawQuery = q.Encode()
					http.Redirect(w, r, redir.String(), http.StatusFound)
					return
				}
			}
		}
		resp := map[string]any{"token_id": tokenID}
		if resolvedUserID > 0 {
			resp["user_id"] = resolvedUserID
		}
		if email != "" {
			resp["email"] = email
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

// Exchange handles JSON exchanges from mobile clients.
func (h *HTTPHandlers) Exchange() http.Handler {
	type payload struct {
		Code        string `json:"code"`
		State       string `json:"state"`
		RedirectURI string `json:"redirect_uri"`
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
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unexpected data in payload"})
			return
		}
		if p.Code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
			return
		}
		if stateCookie, _ := r.Cookie("oauthstate"); stateCookie != nil && p.State != "" && p.State != stateCookie.Value {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oauth state"})
			return
		}
		redirectURI := p.RedirectURI
		if redirectURI == "" {
			redirectURI = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}
		userID := optionalUserID(r)
		tokenID, resolvedUserID, email, err := h.client.ExchangeAndStore(r.Context(), p.Code, redirectURI, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		resp := map[string]any{
			"token":    tokenID,
			"token_id": tokenID,
		}
		if resolvedUserID > 0 {
			resp["user_id"] = resolvedUserID
		}
		if email != "" {
			resp["email"] = email
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

// SendEmail handles POST /actions/google/email
func (h *HTTPHandlers) SendEmail() http.Handler {
	type payload struct {
		TokenID int64  `json:"token_id"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := optionalUserID(r)
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unexpected data in payload"})
			return
		}
		if p.To == "" || p.Subject == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to and subject are required"})
			return
		}
		if err := h.client.SendEmail(r.Context(), userID, p.TokenID, p.To, p.Subject, p.Body); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

// CreateEvent handles POST /actions/google/calendar
func (h *HTTPHandlers) CreateEvent() http.Handler {
	type payload struct {
		TokenID   int64           `json:"token_id"`
		Summary   string          `json:"summary"`
		Start     string          `json:"start"`
		End       string          `json:"end"`
		Attendees json.RawMessage `json:"attendees"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := optionalUserID(r)
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unexpected data in payload"})
			return
		}
		if p.Summary == "" || p.Start == "" || p.End == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "summary, start, end are required"})
			return
		}
		start, err := time.Parse(time.RFC3339, p.Start)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start datetime"})
			return
		}
		end, err := time.Parse(time.RFC3339, p.End)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid end datetime"})
			return
		}
		attendees, err := parseAttendees(p.Attendees)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := h.client.CreateCalendarEvent(r.Context(), userID, p.TokenID, p.Summary, start, end, attendees); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// parseAttendees parses attendees from a JSON raw message
func parseAttendees(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		single = strings.TrimSpace(single)
		if single == "" {
			return nil, nil
		}
		parts := strings.Split(single, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				list = append(list, p)
			}
		}
		return list, nil
	}
	return nil, fmt.Errorf("invalid attendees format")
}

// Helpers duplicated locally to keep httpapi clean.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// ensureNoTrailingData checks that there is no extra data in the JSON decoder.
func ensureNoTrailingData(decoder *json.Decoder) error {
	if decoder.More() {
		return errors.New("unexpected extra data")
	}
	if err := decoder.Decode(new(struct{})); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// randomState generates a random state string for OAuth.
func randomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// optionalUserID extracts an optional user ID from headers or query parameters.
func optionalUserID(r *http.Request) *int64 {
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = r.URL.Query().Get("user_id")
	}
	if user == "" {
		return nil
	}
	id, err := strconv.ParseInt(user, 10, 64)
	if err != nil || id <= 0 {
		return nil
	}
	return &id
}
