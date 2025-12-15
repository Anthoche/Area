package httpapi

import (
	"area/src/auth"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnsureNoTrailingData_OK(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a":1}`))
	dec.DisallowUnknownFields()
	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("decode first object: %v", err)
	}
	if err := ensureNoTrailingData(dec); err != nil {
		t.Fatalf("ensureNoTrailingData returned error: %v", err)
	}
}

func TestEnsureNoTrailingData_ExtraObject(t *testing.T) {
	buf := bytes.NewBufferString(`{"a":1} {"b":2}`)
	dec := json.NewDecoder(buf)
	dec.DisallowUnknownFields()
	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("decode first object: %v", err)
	}
	if err := ensureNoTrailingData(dec); err == nil {
		t.Fatalf("expected error for trailing object, got nil")
	}
}

type stubStore struct {
	user *auth.User
	hash string
	err  error
}

func (s *stubStore) Create(u *auth.User, passwordHash string) error {
	return s.err
}
func (s *stubStore) GetByEmail(email string) (*auth.User, string, error) {
	return s.user, s.hash, s.err
}

func TestHealth(t *testing.T) {
	h := &handler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.health().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := auth.HashPassword("pw")
	svc := auth.NewService(&stubStore{
		user: &auth.User{ID: 1, Email: "a@b.com"},
		hash: hash,
	})
	h := &handler{auth: svc}
	body := bytes.NewBufferString(`{"email":"a@b.com","password":"pw"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	h.login().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	svc := auth.NewService(&stubStore{})
	h := &handler{auth: svc}
	body := bytes.NewBufferString(`{"email":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.register().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
