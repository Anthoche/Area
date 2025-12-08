package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"area/src/auth"
	goog "area/src/integrations/google"
	"area/src/workflows"
)

// NewMux wires HTTP routes that the frontend can call.
func NewMux(authService *auth.Service, wfService *workflows.Service) http.Handler {
	server := &handler{
		auth:      authService,
		workflows: wfService,
		google:    goog.NewClient(),
	}

	mux := http.NewServeMux()
	mux.Handle("/login", server.login())
	mux.Handle("/register", server.register())
	mux.Handle("/healthz", server.health())
	mux.Handle("/workflows", server.workflowsHandler())
	mux.Handle("/workflows/", server.workflowTrigger())
	mux.Handle("/hooks/", server.webhook())
	mux.Handle("/oauth/google/exchange", server.exchangeGoogleToken())
	mux.Handle("/oauth/github/exchange", server.exchangeGithubToken())
	mux.Handle("/oauth/google/login", server.googleLogin())
	mux.Handle("/oauth/google/callback", server.googleCallback())
	mux.Handle("/actions/google/email", server.googleSendEmail())
	mux.Handle("/actions/google/calendar", server.googleCreateEvent())
	return withCORS(mux)
}

type handler struct {
	auth      *auth.Service
	workflows *workflows.Service
	google    *goog.Client
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type workflowRequest struct {
	Name            string          `json:"name"`
	TriggerType     string          `json:"trigger_type"`
	ActionURL       string          `json:"action_url"`
	TriggerConfig   json.RawMessage `json:"trigger_config"`
	IntervalMinutes *int            `json:"interval_minutes,omitempty"`
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

// login handles POST /login authentication requests.
func (h *handler) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload loginRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		payload.Email = strings.TrimSpace(payload.Email)
		if payload.Email == "" || payload.Password == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
			return
		}

		user, err := h.auth.Authenticate(r.Context(), payload.Email, payload.Password)
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
			return
		case err != nil:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not complete login"})
			return
		}

		writeJSON(w, http.StatusOK, user)
	})
}

// register handles POST /register user creation requests.
func (h *handler) register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload registerRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		payload.Email = strings.TrimSpace(payload.Email)
		payload.FirstName = strings.TrimSpace(payload.FirstName)
		payload.LastName = strings.TrimSpace(payload.LastName)
		if payload.Email == "" || payload.Password == "" || payload.FirstName == "" || payload.LastName == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email, password, firstName and lastName are required"})
			return
		}

		user, err := h.auth.Register(r.Context(), payload.Email, payload.Password, payload.FirstName, payload.LastName)
		switch {
		case errors.Is(err, auth.ErrUserExists):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "user already exists"})
			return
		case err != nil:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not complete registration"})
			return
		}

		writeJSON(w, http.StatusCreated, user)
	})
}

// health serves a simple health-check endpoint.
func (h *handler) health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON serializes a value to JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// withCORS adds permissive CORS headers so the web app (port 80) can call the API (port 8080).
// In production, tighten Allowed-Origin to the actual frontend domain.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ensureNoTrailingData rejects payloads that contain multiple JSON objects.
func ensureNoTrailingData(decoder *json.Decoder) error {
	if decoder.More() {
		return errors.New("unexpected extra data")
	}
	if err := decoder.Decode(new(struct{})); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func randomState() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int63())
}

func currentUserID(r *http.Request) (int64, error) {
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = r.URL.Query().Get("user_id")
	}
	if user == "" {
		return 0, fmt.Errorf("missing user identifier")
	}
	id, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user identifier")
	}
	return id, nil
}

func optionalUserID(r *http.Request) *int64 {
	id, err := currentUserID(r)
	if err != nil || id <= 0 {
		return nil
	}
	return &id
}

// workflowsHandler routes workflow listing and creation requests.
func (h *handler) workflowsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.listWorkflows(w, r)
		case http.MethodPost:
			h.createWorkflow(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// createWorkflow validates input and persists a new workflow.
func (h *handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	if h.workflows == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
		return
	}

	var payload workflowRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
		return
	}
	if err := ensureNoTrailingData(decoder); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
		return
	}

	cfg := payload.TriggerConfig
	if len(cfg) == 0 && payload.IntervalMinutes != nil && payload.TriggerType == "interval" {
		cfg, _ = json.Marshal(map[string]int{"interval_minutes": *payload.IntervalMinutes})
	}
	wf, err := h.workflows.CreateWorkflow(r.Context(), payload.Name, payload.TriggerType, payload.ActionURL, cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, wf)
}

// listWorkflows responds with all workflows for the current user/session.
func (h *handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	if h.workflows == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
		return
	}
	items, err := h.workflows.ListWorkflows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not list workflows"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// workflowTrigger handles POST /workflows/{id}/trigger to enqueue a run.
func (h *handler) workflowTrigger() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if h.workflows == nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/workflows/") || !strings.HasSuffix(r.URL.Path, "/trigger") {
			http.NotFound(w, r)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 3 {
			http.NotFound(w, r)
			return
		}
		idStr := parts[1]
		workflowID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid workflow id"})
			return
		}

		payload := make(map[string]any)
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		run, err := h.workflows.Trigger(r.Context(), workflowID, payload)
		if err != nil {
			switch {
			case errors.Is(err, workflows.ErrWorkflowNotFound):
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not trigger workflow"})
			}
			return
		}
		writeJSON(w, http.StatusAccepted, run)
	})
}

// webhook handles external POST /hooks/{token} to trigger a webhook workflow.
func (h *handler) webhook() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if h.workflows == nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/hooks/") {
			http.NotFound(w, r)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		token := parts[1]

		payload := make(map[string]any)
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		run, err := h.workflows.TriggerWebhook(r.Context(), token, payload)
		if err != nil {
			if errors.Is(err, workflows.ErrWorkflowNotFound) {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not trigger workflow"})
			return
		}
		writeJSON(w, http.StatusAccepted, run)
	})
}

func (h *handler) exchangeGoogleToken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		oauthState, _ := r.Cookie("oauthstate")

		if r.FormValue("state") != oauthState.Value {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid oauth google state"})
			return
		}

		data, err := auth.GetUserDataFromGoogle(r.FormValue("code"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"data": string(data)})
	})
}

func (h *handler) exchangeGithubToken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		oauthState, _ := r.Cookie("oauthstate")

		if r.FormValue("state") != oauthState.Value {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid oauth github state"})
			return
		}

		data, err := auth.GetUserDataFromGithub(r.FormValue("code"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"data": data.String()})
	})
}

// googleLogin redirects to Google OAuth consent.
func (h *handler) googleLogin() http.Handler {
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
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "redirect_uri is required"})
			return
		}
		url := h.google.AuthURL(state, redirectURI)
		http.Redirect(w, r, url, http.StatusFound)
	})
}

// googleCallback exchanges code for token and stores it; returns token_id.
func (h *handler) googleCallback() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateCookie, _ := r.Cookie("oauthstate")
		if stateCookie == nil || r.URL.Query().Get("state") != stateCookie.Value {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid oauth state"})
			return
		}
		userID := optionalUserID(r)
		code := r.URL.Query().Get("code")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing code"})
			return
		}
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		}
		if redirectURI == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "redirect_uri is required"})
			return
		}
		tokenID, err := h.google.ExchangeAndStore(r.Context(), code, redirectURI, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		if uiCookie, _ := r.Cookie("oauthredirect"); uiCookie != nil && uiCookie.Value != "" {
			if dest, err := url.QueryUnescape(uiCookie.Value); err == nil {
				if redir, err := url.Parse(dest); err == nil && (redir.Scheme == "http" || redir.Scheme == "https") {
					q := redir.Query()
					q.Set("token_id", strconv.FormatInt(tokenID, 10))
					redir.RawQuery = q.Encode()
					http.Redirect(w, r, redir.String(), http.StatusFound)
					return
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"token_id": tokenID})
	})
}

func (h *handler) googleSendEmail() http.Handler {
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
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}
		if p.TokenID == 0 || p.To == "" || p.Subject == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "token_id, to and subject are required"})
			return
		}
		if err := h.google.SendEmail(r.Context(), userID, p.TokenID, p.To, p.Subject, p.Body); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})
}

func (h *handler) googleCreateEvent() http.Handler {
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
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := ensureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}
		if p.TokenID == 0 || p.Summary == "" || p.Start == "" || p.End == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "token_id, summary, start, end are required"})
			return
		}
		start, err := time.Parse(time.RFC3339, p.Start)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid start datetime"})
			return
		}
		end, err := time.Parse(time.RFC3339, p.End)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid end datetime"})
			return
		}
		if err := h.google.CreateCalendarEvent(r.Context(), userID, p.TokenID, p.Summary, start, end, p.Attendee); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "created"})
	})
}
