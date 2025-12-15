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

// Callback exchanges code for token, stores it, and returns token_id/email.
func (h *HTTPHandlers) Callback() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateCookie, _ := r.Cookie("oauthstate")
		if stateCookie == nil || r.URL.Query().Get("state") != stateCookie.Value {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid oauth state"})
			return
		}
		userID := optionalUserID(r)
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
		tokenID, email, err := h.client.ExchangeAndStore(r.Context(), code, redirectURI, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if uiCookie, _ := r.Cookie("oauthredirect"); uiCookie != nil && uiCookie.Value != "" {
			if dest, err := url.QueryUnescape(uiCookie.Value); err == nil {
				if redir, err := url.Parse(dest); err == nil && (redir.Scheme == "http" || redir.Scheme == "https") {
					q := redir.Query()
					q.Set("token_id", strconv.FormatInt(tokenID, 10))
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
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unexpected data in payload"})
			return
		}
		if p.TokenID == 0 || p.To == "" || p.Subject == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_id, to and subject are required"})
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
		TokenID  int64    `json:"token_id"`
		Summary  string   `json:"summary"`
		Start    string   `json:"start"`
		End      string   `json:"end"`
		Attendee []string `json:"attendees"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := optionalUserID(r)
		var p payload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unexpected data in payload"})
			return
		}
		if p.TokenID == 0 || p.Summary == "" || p.Start == "" || p.End == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_id, summary, start, end are required"})
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
		if err := h.client.CreateCalendarEvent(r.Context(), userID, p.TokenID, p.Summary, start, end, p.Attendee); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
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
