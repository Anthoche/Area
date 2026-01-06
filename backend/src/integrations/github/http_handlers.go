package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HTTPHandlers exposes minimal GitHub OAuth endpoints (no third-party libs).
type HTTPHandlers struct {
	client *Client
}

// NewHTTPHandlers builds GitHub HTTP handlers with a default client.
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

// LoginMobile redirects to GitHub OAuth consent using the mobile OAuth app.
func (h *HTTPHandlers) LoginMobile() http.Handler {
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
		redirectURI = os.Getenv("GITHUB_MOBILE_OAUTH_REDIRECT_URI")
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

// CallbackMobile exchanges code for token using the mobile OAuth app.
func (h *HTTPHandlers) CallbackMobile() http.Handler {
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
			redirectURI = os.Getenv("GITHUB_MOBILE_OAUTH_REDIRECT_URI")
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

// Issue handles POST /actions/github/issue
func (h *HTTPHandlers) Issue() http.Handler {
	type payload struct {
		TokenID int64           `json:"token_id"`
		Repo    string          `json:"repo"`
		Title   string          `json:"title"`
		Body    string          `json:"body"`
		Labels  json.RawMessage `json:"labels"`
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
		if p.TokenID <= 0 || p.Repo == "" || p.Title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_id, repo and title are required"})
			return
		}
		parts := strings.Split(p.Repo, "/")
		if len(parts) != 2 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "repo must be owner/name"})
			return
		}
		labels, err := parseStringList(p.Labels)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "labels must be an array or comma-separated string"})
			return
		}
		if err := h.client.CreateIssue(r.Context(), userID, p.TokenID, parts[0], parts[1], p.Title, p.Body, labels); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// PullRequest handles POST /actions/github/pr
func (h *HTTPHandlers) PullRequest() http.Handler {
	type payload struct {
		TokenID int64  `json:"token_id"`
		Repo    string `json:"repo"`
		Title   string `json:"title"`
		Head    string `json:"head"`
		Base    string `json:"base"`
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
		if p.TokenID <= 0 || p.Repo == "" || p.Title == "" || p.Head == "" || p.Base == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token_id, repo, title, head and base are required"})
			return
		}
		parts := strings.Split(p.Repo, "/")
		if len(parts) != 2 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "repo must be owner/name"})
			return
		}
		if err := h.client.CreatePullRequest(r.Context(), userID, p.TokenID, parts[0], parts[1], p.Title, p.Head, p.Base, p.Body); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// optionalUserID extracts an optional user ID from the request headers or query parameters.
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

// randomState generates a random string for OAuth state parameter.
func randomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// parseStringList accepts a JSON array of strings or a single comma-separated string.
func parseStringList(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		one = strings.TrimSpace(one)
		if one == "" {
			return nil, nil
		}
		parts := strings.Split(one, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts, nil
	}
	return nil, fmt.Errorf("invalid string list")
}
