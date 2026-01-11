package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/swaggest/swgui/v5emb"

	"area/src/areas"
	"area/src/auth"
	"area/src/database"
	"area/src/integrations/discord"
	gh "area/src/integrations/github"
	goog "area/src/integrations/google"
	"area/src/integrations/notion"
	"area/src/integrations/slack"
	"area/src/integrations/trello"
	"area/src/workflows"
)

// NewMux wires HTTP routes that the frontend can call.
func NewMux(authService *auth.Service, wfService *workflows.Service) http.Handler {
	server := &Handler{
		Auth:      authService,
		workflows: wfService,
	}

	googleHTTP := goog.NewHTTPHandlers(nil)
	githubHTTP := gh.NewHTTPHandlers(nil)
	githubMobileHTTP := gh.NewHTTPHandlers(gh.NewMobileClient())
	discordHTTP := discord.NewHTTPHandlers(nil)
	slackHTTP := slack.NewHTTPHandlers(nil)
	notionHTTP := notion.NewHTTPHandlers(nil)
	trelloHTTP := trello.NewHTTPHandlers(nil)
	mux := http.NewServeMux()
	mux.Handle("/login", server.Login())
	mux.Handle("/register", server.Register())
	mux.Handle("/healthz", server.Health())
	mux.Handle("/workflows", server.workflowsHandler())
	mux.Handle("/workflows/", server.workflowResource())
	mux.Handle("/hooks/", server.webhook())
	mux.Handle("/oauth/google/start", googleHTTP.Start())
	mux.Handle("/oauth/google/exchange", googleHTTP.Exchange())
	mux.Handle("/oauth/github/exchange", server.exchangeGithubToken())
	mux.Handle("/oauth/google/login", googleHTTP.Login())
	mux.Handle("/oauth/google/callback", googleHTTP.Callback())
	mux.Handle("/oauth/google/mobile/login", googleHTTP.Login())
	mux.Handle("/oauth/google/mobile/callback", googleHTTP.Callback())
	mux.Handle("/oauth/github/login", githubHTTP.Login())
	mux.Handle("/oauth/github/callback", githubHTTP.Callback())
	mux.Handle("/oauth/github/mobile/login", githubMobileHTTP.LoginMobile())
	mux.Handle("/oauth/github/mobile/callback", githubMobileHTTP.CallbackMobile())
	mux.Handle("/actions/github/issue", githubHTTP.Issue())
	mux.Handle("/actions/github/pr", githubHTTP.PullRequest())
	mux.Handle("/actions/google/email", googleHTTP.SendEmail())
	mux.Handle("/actions/google/calendar", googleHTTP.CreateEvent())
	mux.Handle("/actions/discord/message", discordHTTP.Message())
	mux.Handle("/actions/discord/embed", discordHTTP.Embed())
	mux.Handle("/actions/discord/message/edit", discordHTTP.Edit())
	mux.Handle("/actions/discord/message/delete", discordHTTP.Delete())
	mux.Handle("/actions/discord/message/react", discordHTTP.React())
	mux.Handle("/actions/slack/message", slackHTTP.Message())
	mux.Handle("/actions/slack/blocks", slackHTTP.Blocks())
	mux.Handle("/actions/slack/message/update", slackHTTP.Update())
	mux.Handle("/actions/slack/message/delete", slackHTTP.Delete())
	mux.Handle("/actions/slack/message/react", slackHTTP.React())
	mux.Handle("/actions/notion/page", notionHTTP.Page())
	mux.Handle("/actions/notion/blocks", notionHTTP.AppendBlocks())
	mux.Handle("/actions/notion/database", notionHTTP.Database())
	mux.Handle("/actions/notion/page/update", notionHTTP.UpdatePage())
	mux.Handle("/actions/trello/card", trelloHTTP.CreateCard())
	mux.Handle("/actions/trello/card/move", trelloHTTP.MoveCard())
	mux.Handle("/actions/trello/list", trelloHTTP.CreateList())
	mux.Handle("/about.json", server.about())
	mux.Handle("/areas", server.listAreas())
	mux.Handle("/resources/openapi.json", server.openAPISpec())
	mux.Handle("/docs/", v5emb.New(
		"KiKonect API Reference",
		"/resources/openapi.json",
		"/docs/",
	))
	return WithCORS(mux)
}

type Handler struct {
	Auth      *auth.Service
	workflows *workflows.Service
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

// userContext extracts the user id from headers/query and enriches the context.
func userContext(r *http.Request) (context.Context, error) {
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = r.URL.Query().Get("user_id")
	}
	if user == "" {
		return r.Context(), fmt.Errorf("missing user id")
	}
	id, err := strconv.ParseInt(user, 10, 64)
	if err != nil || id <= 0 {
		return r.Context(), fmt.Errorf("invalid user id")
	}
	return workflows.WithUserID(r.Context(), id), nil
}

// Login handles POST /login authentication requests.
func (h *Handler) Login() http.Handler {
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
		if err := EnsureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		payload.Email = strings.TrimSpace(payload.Email)
		if payload.Email == "" || payload.Password == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
			return
		}

		user, err := h.Auth.Authenticate(payload.Email, payload.Password)
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

// Register handles POST /register user creation requests.
func (h *Handler) Register() http.Handler {
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
		if err := EnsureNoTrailingData(decoder); err != nil {
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

		user, err := h.Auth.Register(payload.Email, payload.Password, payload.FirstName, payload.LastName)
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

// Health serves a simple Health-check endpoint.
func (h *Handler) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// listAreas exposes the catalog of available services/triggers/reactions for the clients.
func (h *Handler) listAreas() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var services []areas.Service
		if serviceID := r.URL.Query().Get("service_id"); serviceID != "" {
			svc, err := areas.Get(r.Context(), serviceID)
			if err != nil {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "service not found"})
				return
			}
			services = []areas.Service{*svc}
		} else if serviceID := r.URL.Query().Get("id"); serviceID != "" {
			svc, err := areas.Get(r.Context(), serviceID)
			if err != nil {
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "service not found"})
				return
			}
			services = []areas.Service{*svc}
		} else {
			var err error
			services, err = areas.List(r.Context())
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
				return
			}
		}
		w.Header().Set("Cache-Control", "no-store")
		var userCount int64
		if count, err := database.CountUsers(); err == nil {
			userCount = count
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"services":   services,
			"user_count": userCount,
		})
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

type aboutResponse struct {
	Client struct {
		Host string `json:"host"`
	} `json:"client"`
	Server struct {
		CurrentTime int64          `json:"current_time"`
		Services    []aboutService `json:"services"`
	} `json:"server"`
}

type aboutService struct {
	Name      string       `json:"name"`
	Actions   []aboutEntry `json:"actions"`
	Reactions []aboutEntry `json:"reactions"`
}

type aboutEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// writeJSON serializes a value to JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// about returns the server capabilities for the client.
func (h *Handler) about() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		services, err := areas.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		out := aboutResponse{}
		out.Client.Host = host
		out.Server.CurrentTime = time.Now().Unix()
		out.Server.Services = make([]aboutService, 0, len(services))
		for _, svc := range services {
			entry := aboutService{
				Name:      svc.ID,
				Actions:   make([]aboutEntry, 0, len(svc.Triggers)),
				Reactions: make([]aboutEntry, 0, len(svc.Reactions)),
			}
			for _, act := range svc.Triggers {
				entry.Actions = append(entry.Actions, aboutEntry{
					Name:        act.ID,
					Description: act.Description,
				})
			}
			for _, react := range svc.Reactions {
				entry.Reactions = append(entry.Reactions, aboutEntry{
					Name:        react.ID,
					Description: react.Description,
				})
			}
			out.Server.Services = append(out.Server.Services, entry)
		}
		writeJSON(w, http.StatusOK, out)
	})
}

// WithCORS adds permissive CORS headers for the API.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-User-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnsureNoTrailingData rejects payloads that contain multiple JSON objects.
func EnsureNoTrailingData(decoder *json.Decoder) error {
	if decoder.More() {
		return errors.New("unexpected extra data")
	}
	if err := decoder.Decode(new(struct{})); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// workflowsHandler routes workflow listing and creation requests.
func (h *Handler) workflowsHandler() http.Handler {
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
func (h *Handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	if h.workflows == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
		return
	}
	ctx, err := userContext(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}

	var payload workflowRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
		return
	}
	if err := EnsureNoTrailingData(decoder); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
		return
	}

	cfg := payload.TriggerConfig
	if len(cfg) == 0 && payload.IntervalMinutes != nil && payload.TriggerType == "interval" {
		cfg, _ = json.Marshal(map[string]int{"interval_minutes": *payload.IntervalMinutes})
	}
	wf, err := h.workflows.CreateWorkflow(ctx, payload.Name, payload.TriggerType, payload.ActionURL, cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, wf)
}

// listWorkflows responds with all workflows for the current user/session.
func (h *Handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	if h.workflows == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
		return
	}
	ctx, err := userContext(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}
	items, err := h.workflows.ListWorkflows(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not list workflows"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// workflowResource handles:
// - POST /workflows/{id}/trigger to enqueue a run
// - DELETE /workflows/{id} to delete a workflow
func (h *Handler) workflowResource() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.workflows == nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "workflows not configured"})
			return
		}
		ctx, err := userContext(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: err.Error()})
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/workflows/") {
			http.NotFound(w, r)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		idStr := parts[1]
		workflowID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid workflow id"})
			return
		}

		// POST /workflows/{id}/enabled?action=enable|disable
		if len(parts) == 3 && parts[2] == "enabled" && r.Method == http.MethodPost {
			action := r.URL.Query().Get("action")
			switch action {
			case "enable":
				if err := h.workflows.SetEnabled(ctx, workflowID, true, time.Now()); err != nil {
					if errors.Is(err, workflows.ErrWorkflowNotFound) {
						writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
						return
					}
					writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not enable workflow"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
				return
			case "disable":
				if err := h.workflows.SetEnabled(ctx, workflowID, false, time.Now()); err != nil {
					if errors.Is(err, workflows.ErrWorkflowNotFound) {
						writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
						return
					}
					writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not disable workflow"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
				return
			default:
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid action"})
				return
			}
		}

		// DELETE /workflows/{id}
		if len(parts) == 2 && r.Method == http.MethodDelete {
			if err := h.workflows.DeleteWorkflow(ctx, workflowID); err != nil {
				if errors.Is(err, workflows.ErrWorkflowNotFound) {
					writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
					return
				}
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not delete workflow"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}

		// POST /workflows/{id}/trigger
		if !(len(parts) == 3 && parts[2] == "trigger" && r.Method == http.MethodPost) {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := make(map[string]any)
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
			return
		}
		if err := EnsureNoTrailingData(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unexpected data in payload"})
			return
		}

		run, err := h.workflows.Trigger(ctx, workflowID, payload)
		if err != nil {
			switch {
			case errors.Is(err, workflows.ErrWorkflowNotFound):
				writeJSON(w, http.StatusNotFound, errorResponse{Error: "workflow not found"})
			case errors.Is(err, workflows.ErrWorkflowDisabled):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "workflow disabled"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not trigger workflow"})
			}
			return
		}
		writeJSON(w, http.StatusAccepted, run)
	})
}

// webhook handles external POST /hooks/{token} to trigger a webhook workflow.
func (h *Handler) webhook() http.Handler {
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
		if err := EnsureNoTrailingData(decoder); err != nil {
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

// exchangeGoogleToken handles POST /oauth/google/exchange to exchange an Auth code for user data.
func (h *Handler) exchangeGoogleToken() http.Handler {
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

func (h *Handler) exchangeGithubToken() http.Handler {
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
		writeJSON(w, http.StatusAccepted, map[string]string{"data": string(data)})
	})
}

// openAPISpec serves the OpenAPI specification JSON file
func (h *Handler) openAPISpec() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		http.ServeFile(w, r, "resources/openapi.json")
	})
}
