package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"area/src/database"
)

// Client handles GitHub OAuth without third-party helpers.
type Client struct {
	clientID     string
	clientSecret string
	scopes       []string
	httpClient   *http.Client
}

// NewClient builds a GitHub API client using environment credentials.
func NewClient() *Client {
	return &Client{
		clientID:     mustEnv("GITHUB_OAUTH_CLIENT_ID"),
		clientSecret: mustEnv("GITHUB_OAUTH_CLIENT_SECRET"),
		scopes:       []string{"read:user", "user:email", "repo"},
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthURL builds the GitHub authorization URL.
func (c *Client) AuthURL(state, redirectURI string) string {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("scope", strings.Join(c.scopes, " "))
	v.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + v.Encode()
}

// ExchangeAndStore swaps a code for an access token, stores it, and ensures a user exists.
func (c *Client) ExchangeAndStore(ctx context.Context, code, redirectURI string, userID *int64) (int64, int64, string, string, error) {
	token, scope, tokenType, err := c.exchangeCode(ctx, code, redirectURI)
	if err != nil {
		return -1, 0, "", "", err
	}
	email, login, err := c.fetchIdentity(ctx, token)
	if err != nil {
		return -1, 0, "", "", err
	}

	if userID == nil && email != "" {
		if u, err := database.GetUserByEmail(email); err == nil {
			id := int64(u.ID)
			userID = &id
		} else {
			uid, createErr := database.CreateUser("GitHub", login, email, "github-oauth")
			if createErr == nil {
				userID = &uid
			}
		}
	}

	tokenID, err := database.InsertGithubToken(userID, token, tokenType, scope)
	if err != nil {
		return -1, 0, "", "", err
	}
	var resolved int64
	if userID != nil {
		resolved = *userID
	}
	return tokenID, resolved, email, login, nil
}

// Exchange a code for a token
func (c *Client) exchangeCode(ctx context.Context, code, redirectURI string) (string, string, string, error) {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)
	values.Set("code", code)
	if redirectURI != "" {
		values.Set("redirect_uri", redirectURI)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(values.Encode()))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("github exchange status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", "", err
	}
	if payload.AccessToken == "" {
		return "", "", "", fmt.Errorf("no access_token from github")
	}
	return payload.AccessToken, payload.Scope, payload.TokenType, nil
}

// Fetch user identity
func (c *Client) fetchIdentity(ctx context.Context, token string) (string, string, error) {
	profileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}
	profileReq.Header.Set("Authorization", "Bearer "+token)
	profileReq.Header.Set("Accept", "application/json")

	profileResp, err := c.httpClient.Do(profileReq)
	if err != nil {
		return "", "", err
	}
	defer profileResp.Body.Close()
	var profile struct {
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		return "", "", err
	}

	email := profile.Email
	if email == "" {
		email = c.fetchPrimaryEmail(ctx, token)
	}
	return email, profile.Login, nil
}

// Fetch the primary email address
func (c *Client) fetchPrimaryEmail(ctx context.Context, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}

type Commit struct {
	SHA     string
	Message string
	Author  string
	Date    time.Time
}

type PullRequest struct {
	Number    int
	Title     string
	State     string
	Merged    bool
	Author    string
	Base      string
	Head      string
	UpdatedAt time.Time
	URL       string
}

type Issue struct {
	Number    int
	Title     string
	State     string
	Author    string
	UpdatedAt time.Time
	URL       string
}

// CreateIssue creates a GitHub issue in the given repo.
func (c *Client) CreateIssue(ctx context.Context, userID *int64, tokenID int64, owner, repo, title, body string, labels []string) error {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"title": title,
	}
	if body != "" {
		payload["body"] = body
	}
	if len(labels) > 0 {
		payload["labels"] = labels
	}
	buf, _ := json.Marshal(payload)

	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", url.PathEscape(owner), url.PathEscape(repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("create issue status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// CreatePullRequest opens a new pull request.
func (c *Client) CreatePullRequest(ctx context.Context, userID *int64, tokenID int64, owner, repo, title, head, base, body string) error {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"title": title,
		"head":  head,
		"base":  base,
	}
	if body != "" {
		payload["body"] = body
	}
	buf, _ := json.Marshal(payload)

	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", url.PathEscape(owner), url.PathEscape(repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("create pr status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ListRecentCommits fetches the newest commits for a repo/branch.
func (c *Client) ListRecentCommits(ctx context.Context, userID *int64, tokenID int64, owner, repo, branch string, limit int) ([]Commit, error) {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 5
	}

	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=%d", url.PathEscape(owner), url.PathEscape(repo), url.QueryEscape(branch), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list commits status %d: %s", resp.StatusCode, string(body))
	}

	var payload []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	commits := make([]Commit, 0, len(payload))
	for _, cmt := range payload {
		commits = append(commits, Commit{
			SHA:     cmt.SHA,
			Message: cmt.Commit.Message,
			Author:  cmt.Commit.Author.Name,
			Date:    cmt.Commit.Author.Date,
		})
	}
	return commits, nil
}

// ensureToken retrieves and verifies a GitHub token.
func (c *Client) ensureToken(ctx context.Context, userID *int64, tokenID int64) (*database.GithubToken, error) {
	if tokenID == 0 {
		return nil, fmt.Errorf("missing github token id")
	}
	var t *database.GithubToken
	var err error
	if userID != nil && *userID > 0 {
		t, err = database.GetGithubTokenForUser(tokenID, *userID)
	} else {
		t, err = database.GetGithubToken(tokenID)
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// ListRecentPullRequests fetches recently updated PRs.
func (c *Client) ListRecentPullRequests(ctx context.Context, userID *int64, tokenID int64, owner, repo string, limit int) ([]PullRequest, error) {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&sort=updated&direction=desc&per_page=%d", url.PathEscape(owner), url.PathEscape(repo), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list prs status %d: %s", resp.StatusCode, string(body))
	}
	var payload []struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		HTMLURL   string     `json:"html_url"`
		UpdatedAt time.Time  `json:"updated_at"`
		MergedAt  *time.Time `json:"merged_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		Head struct {
			Ref string `json:"ref"`
		} `json:"head"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	result := make([]PullRequest, 0, len(payload))
	for _, pr := range payload {
		result = append(result, PullRequest{
			Number:    pr.Number,
			Title:     pr.Title,
			State:     pr.State,
			Merged:    pr.MergedAt != nil,
			Author:    pr.User.Login,
			Base:      pr.Base.Ref,
			Head:      pr.Head.Ref,
			UpdatedAt: pr.UpdatedAt,
			URL:       pr.HTMLURL,
		})
	}
	return result, nil
}

// ListRecentIssues fetches recently updated issues (excluding PRs).
func (c *Client) ListRecentIssues(ctx context.Context, userID *int64, tokenID int64, owner, repo string, limit int) ([]Issue, error) {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&sort=updated&direction=desc&per_page=%d", url.PathEscape(owner), url.PathEscape(repo), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list issues status %d: %s", resp.StatusCode, string(body))
	}
	var payload []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		State     string    `json:"state"`
		HTMLURL   string    `json:"html_url"`
		UpdatedAt time.Time `json:"updated_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		PullRequest *struct{} `json:"pull_request,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	result := make([]Issue, 0, len(payload))
	for _, iss := range payload {
		if iss.PullRequest != nil {
			continue
		}
		result = append(result, Issue{
			Number:    iss.Number,
			Title:     iss.Title,
			State:     iss.State,
			Author:    iss.User.Login,
			UpdatedAt: iss.UpdatedAt,
			URL:       iss.HTMLURL,
		})
	}
	return result, nil
}

// mustEnv fetches a required environment variable or panics.
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("missing env %s", key))
	}
	return val
}
