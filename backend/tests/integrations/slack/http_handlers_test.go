package slack

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"area/src/integrations/slack"
)

func TestMessage_BadMethod(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodGet, "/actions/slack/message", nil)
	rr := httptest.NewRecorder()
	h.Message().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestMessage_MissingFields(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/slack/message", bytes.NewBufferString(`{"channel_id":""}`))
	rr := httptest.NewRecorder()
	h.Message().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestBlocks_MissingFields(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/slack/blocks", bytes.NewBufferString(`{"channel_id":"c"}`))
	rr := httptest.NewRecorder()
	h.Blocks().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestUpdate_MissingFields(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/slack/message/update", bytes.NewBufferString(`{"channel_id":"c","message_ts":""}`))
	rr := httptest.NewRecorder()
	h.Update().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestDelete_MissingFields(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/slack/message/delete", bytes.NewBufferString(`{"channel_id":""}`))
	rr := httptest.NewRecorder()
	h.Delete().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestReact_MissingFields(t *testing.T) {
	h := slack.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/slack/message/react", bytes.NewBufferString(`{"channel_id":"c","message_ts":"1","emoji":""}`))
	rr := httptest.NewRecorder()
	h.React().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
