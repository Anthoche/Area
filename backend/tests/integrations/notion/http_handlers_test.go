package notion

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"area/src/integrations/notion"
)

func TestPage_BadMethod(t *testing.T) {
	h := notion.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodGet, "/actions/notion/page", nil)
	rr := httptest.NewRecorder()
	h.Page().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestPage_MissingFields(t *testing.T) {
	h := notion.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/notion/page", bytes.NewBufferString(`{"parent_page_id":""}`))
	rr := httptest.NewRecorder()
	h.Page().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestAppendBlocks_MissingFields(t *testing.T) {
	h := notion.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/notion/blocks", bytes.NewBufferString(`{"block_id":""}`))
	rr := httptest.NewRecorder()
	h.AppendBlocks().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestDatabase_MissingFields(t *testing.T) {
	h := notion.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/notion/database", bytes.NewBufferString(`{"database_id":""}`))
	rr := httptest.NewRecorder()
	h.Database().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestUpdatePage_MissingFields(t *testing.T) {
	h := notion.NewHTTPHandlers(nil)
	req := httptest.NewRequest(http.MethodPost, "/actions/notion/page/update", bytes.NewBufferString(`{"page_id":""}`))
	rr := httptest.NewRecorder()
	h.UpdatePage().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
