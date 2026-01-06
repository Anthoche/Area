package discord

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"area/src/integrations/discord"
)

func TestMessage_BadMethod(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodGet, "/actions/discord/message", nil)
	rr := httptest.NewRecorder()
	h.Message().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestMessage_MissingFields(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/discord/message", bytes.NewBufferString(`{"channel_id":""}`))
	rr := httptest.NewRecorder()
	h.Message().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestEmbed_InvalidColor(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	body := bytes.NewBufferString(`{"channel_id":"1","title":"t","color":"zz"}`)
	req := httptest.NewRequest(http.MethodPost, "/actions/discord/embed", body)
	rr := httptest.NewRecorder()
	h.Embed().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestEdit_MissingFields(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/discord/message/edit", bytes.NewBufferString(`{"channel_id":"1"}`))
	rr := httptest.NewRecorder()
	h.Edit().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestDelete_MissingFields(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/discord/message/delete", bytes.NewBufferString(`{"channel_id":""}`))
	rr := httptest.NewRecorder()
	h.Delete().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestReact_MissingFields(t *testing.T) {
	h := discord.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/discord/message/react", bytes.NewBufferString(`{"channel_id":"1","message_id":""}`))
	rr := httptest.NewRecorder()
	h.React().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
