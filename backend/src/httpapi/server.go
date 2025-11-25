package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"area/server/auth"
)

// NewMux wires HTTP routes that the frontend can call.
func NewMux(service *auth.Service) http.Handler {
	server := &handler{auth: service}

	mux := http.NewServeMux()
	mux.Handle("/login", server.login())
	mux.Handle("/register", server.register())
	mux.Handle("/healthz", server.health())
	return mux
}

type handler struct {
	auth *auth.Service
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
