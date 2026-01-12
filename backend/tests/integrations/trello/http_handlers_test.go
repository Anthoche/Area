package trello

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"area/src/integrations/trello"
)

func TestCreateCard_BadMethod(t *testing.T) {
	h := trello.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodGet, "/actions/trello/card", nil)
	rr := httptest.NewRecorder()
	h.CreateCard().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestCreateCard_MissingFields(t *testing.T) {
	h := trello.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/trello/card", bytes.NewBufferString(`{"list_id":""}`))
	rr := httptest.NewRecorder()
	h.CreateCard().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestCreateCard_MissingToken(t *testing.T) {
	h := trello.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/trello/card", bytes.NewBufferString(`{"list_id":"1","name":"a","api_key":"x"}`))
	rr := httptest.NewRecorder()
	h.CreateCard().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestMoveCard_MissingFields(t *testing.T) {
	h := trello.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/trello/card/move", bytes.NewBufferString(`{"card_id":"1"}`))
	rr := httptest.NewRecorder()
	h.MoveCard().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestCreateList_MissingFields(t *testing.T) {
	h := trello.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/trello/list", bytes.NewBufferString(`{"board_id":""}`))
	rr := httptest.NewRecorder()
	h.CreateList().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
