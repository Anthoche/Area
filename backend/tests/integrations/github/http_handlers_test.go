package github

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gh "area/src/integrations/github"
)

func TestLogin_MissingRedirectURI(t *testing.T) {
	original, ok := os.LookupEnv("GITHUB_OAUTH_REDIRECT_URI")
	_ = os.Unsetenv("GITHUB_OAUTH_REDIRECT_URI")
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv("GITHUB_OAUTH_REDIRECT_URI", original)
		} else {
			_ = os.Unsetenv("GITHUB_OAUTH_REDIRECT_URI")
		}
	})

	h := gh.NewHTTPHandlers(&gh.Client{})
	req := httptest.NewRequest(http.MethodGet, "/oauth/github/login", nil)
	rr := httptest.NewRecorder()
	h.Login().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestCallback_InvalidState(t *testing.T) {
	h := gh.NewHTTPHandlers(&gh.Client{})
	req := httptest.NewRequest(http.MethodGet, "/oauth/github/callback?state=bad", nil)
	rr := httptest.NewRecorder()
	h.Callback().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestIssue_BadMethod(t *testing.T) {
	h := gh.NewHTTPHandlers(&gh.Client{})
	req := httptest.NewRequest(http.MethodGet, "/actions/github/issue", nil)
	rr := httptest.NewRecorder()
	h.Issue().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestIssue_InvalidRepo(t *testing.T) {
	h := gh.NewHTTPHandlers(&gh.Client{})
	body := bytes.NewBufferString(`{"token_id":1,"repo":"bad","title":"t"}`)
	req := httptest.NewRequest(http.MethodPost, "/actions/github/issue", body)
	rr := httptest.NewRecorder()
	h.Issue().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestIssue_InvalidLabels(t *testing.T) {
	h := gh.NewHTTPHandlers(&gh.Client{})
	body := bytes.NewBufferString(`{"token_id":1,"repo":"o/r","title":"t","labels":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/actions/github/issue", body)
	rr := httptest.NewRecorder()
	h.Issue().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestPullRequest_MissingFields(t *testing.T) {
	h := gh.NewHTTPHandlers(&gh.Client{})
	body := bytes.NewBufferString(`{"token_id":1,"repo":"o/r","title":"t","head":"feature"}`)
	req := httptest.NewRequest(http.MethodPost, "/actions/github/pr", body)
	rr := httptest.NewRecorder()
	h.PullRequest().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}
