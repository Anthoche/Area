package google

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubClient struct {
	authURL  string
	tokenID  int64
	email    string
	sendErr  error
	eventErr error
	exchErr  error
}

func (s *stubClient) AuthURL(state, redirectURI string) string {
	return s.authURL
}
func (s *stubClient) ExchangeAndStore(ctx context.Context, code string, redirectURI string, userID *int64) (int64, string, error) {
	return s.tokenID, s.email, s.exchErr
}
func (s *stubClient) SendEmail(ctx context.Context, userID *int64, tokenID int64, to, subject, body string) error {
	return s.sendErr
}
func (s *stubClient) CreateCalendarEvent(ctx context.Context, userID *int64, tokenID int64, summary string, start time.Time, end time.Time, attendees []string) error {
	return s.eventErr
}

func TestSendEmail_BadMethod(t *testing.T) {
	h := NewHTTPHandlers(&Client{})
	req := httptest.NewRequest(http.MethodGet, "/actions/google/email", nil)
	rr := httptest.NewRecorder()
	h.SendEmail().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestSendEmail_MissingFields(t *testing.T) {
	h := NewHTTPHandlers(&Client{})
	req := httptest.NewRequest(http.MethodPost, "/actions/google/email", bytes.NewBufferString(`{"token_id":0}`))
	rr := httptest.NewRecorder()
	h.SendEmail().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestCallback_InvalidState(t *testing.T) {
	h := NewHTTPHandlers(&Client{})
	req := httptest.NewRequest(http.MethodGet, "/oauth/google/callback?state=bad", nil)
	rr := httptest.NewRecorder()
	h.Callback().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
