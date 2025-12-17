package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// HTTPHandlers exposes minimal GitHub OAuth endpoints (no third-party libs).
type HTTPHandlers struct {
	client *Client
}

func NewHTTPHandlers(client *Client) *HTTPHandlers {
	if client == nil {
		client = NewClient()
	}
	return &HTTPHandlers{client: client}
}

// Login redirects to GitHub OAuth consent.
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
			redirectURI = os.Getenv("GITHUB_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}
		authURL := h.client.AuthURL(state, redirectURI)
		http.Redirect(w, r, authURL, http.StatusFound)
	})
}

// Callback exchanges code for token, stores it, and returns token_id/email/user_id/login.
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
			redirectURI = os.Getenv("GITHUB_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "redirect_uri is required"})
			return
		}

		tokenID, resolvedUserID, email, login, err := h.client.ExchangeAndStore(r.Context(), code, redirectURI, userID)
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
						q.Set("github_email", email)
					}
					if login != "" {
						q.Set("github_login", login)
					}
					redir.RawQuery = q.Encode()
					http.Redirect(w, r, redir.String(), http.StatusFound)
					return
				}
			}
		}
		resp := map[string]any{
			"token_id": tokenID,
			"login":    login,
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

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

func randomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
