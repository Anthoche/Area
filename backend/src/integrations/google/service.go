package google

import (
	"bytes"
	"context"
	"encoding/base64"
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

// OAuthToken stores the access credentials we use to call Google APIs.
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// Client wraps token retrieval and Google API calls without third-party OAuth libs.
type Client struct {
	clientID     string
	clientSecret string
	scopes       []string
	httpClient   *http.Client
}

func NewClient() *Client {
	return &Client{
		clientID:     mustEnv("GOOGLE_OAUTH_CLIENT_ID"),
		clientSecret: mustEnv("GOOGLE_OAUTH_CLIENT_SECRET"),
		scopes: []string{
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthURL builds the authorization URL with a given state and redirect URI.
func (c *Client) AuthURL(state, redirectURI string) string {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(c.scopes, " "))
	v.Set("access_type", "offline")
	v.Set("prompt", "consent")
	v.Set("state", state)
	return "https://accounts.google.com/o/oauth2/auth?" + v.Encode()
}

// ExchangeAndStore exchanges an auth code for a token, saves it, and returns the token id, user id, and email.
func (c *Client) ExchangeAndStore(ctx context.Context, code string, redirectURI string, userID *int64) (int64, int64, string, error) {
	tokenResp, err := c.exchangeCode(ctx, code, redirectURI)
	if err != nil {
		return -1, 0, "", err
	}
	email, _ := c.fetchEmail(ctx, tokenResp.AccessToken)
	if userID == nil && email != "" {
		if u, err := database.GetUserByEmail(ctx, email); err == nil {
			userID = &u.Id
		} else {
			uid, createErr := database.CreateUser(ctx, "Google", "User", email, "google-oauth")
			if createErr == nil {
				userID = &uid
			}
		}
	}
	id, err := database.InsertGoogleToken(ctx, userID, tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.Expiry)
	if err != nil {
		return -1, 0, "", err
	}
	var resolvedUserID int64
	if userID != nil {
		resolvedUserID = *userID
	}
	return id, resolvedUserID, email, nil
}

// fetchEmail retrieves the user's email address using the OAuth access token.
func (c *Client) fetchEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("userinfo status %d", resp.StatusCode)
	}
	var data struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Email, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// exchangeCode exchanges an auth code for an OAuth token.
func (c *Client) exchangeCode(ctx context.Context, code, redirectURI string) (*OAuthToken, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", redirectURI)
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("exchange code status %d: %s", resp.StatusCode, string(body))
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}
	expiry := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return &OAuthToken{AccessToken: tr.AccessToken, RefreshToken: tr.RefreshToken, Expiry: expiry}, nil
}

// refresh uses a refresh token to obtain a new access token.
func (c *Client) refresh(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("refresh status %d: %s", resp.StatusCode, string(body))
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}
	expiry := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	if tr.RefreshToken == "" {
		tr.RefreshToken = refreshToken
	}
	return &OAuthToken{AccessToken: tr.AccessToken, RefreshToken: tr.RefreshToken, Expiry: expiry}, nil
}

// SendEmail sends a simple text email via Gmail API.
func (c *Client) SendEmail(ctx context.Context, userID *int64, tokenID int64, to, subject, body string) error {
	oauthToken, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return err
	}
	raw := buildRawEmail(to, subject, body)
	payload := map[string]string{"raw": base64.URLEncoding.EncodeToString([]byte(raw))}
	buf, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.googleapis.com/gmail/v1/users/me/messages/send", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oauthToken.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gmail send status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// CreateCalendarEvent creates a basic event in the primary calendar.
func (c *Client) CreateCalendarEvent(ctx context.Context, userID *int64, tokenID int64, summary string, start time.Time, end time.Time, attendees []string) error {
	oauthToken, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return err
	}
	type event struct {
		Summary     string `json:"summary"`
		Description string `json:"description,omitempty"`
		Start       any    `json:"start"`
		End         any    `json:"end"`
		Attendees   []any  `json:"attendees,omitempty"`
	}
	ev := event{
		Summary: summary,
		Start:   map[string]string{"dateTime": start.Format(time.RFC3339)},
		End:     map[string]string{"dateTime": end.Format(time.RFC3339)},
	}
	for _, a := range attendees {
		ev.Attendees = append(ev.Attendees, map[string]string{"email": a})
	}
	buf, _ := json.Marshal(ev)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.googleapis.com/calendar/v3/calendars/primary/events", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oauthToken.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("calendar create status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// buildRawEmail constructs a raw email string.
func buildRawEmail(to, subject, body string) string {
	return fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
}

// ensureToken retrieves and refreshes the OAuth token as needed.
func (c *Client) ensureToken(ctx context.Context, userID *int64, tokenID int64) (*OAuthToken, error) {
	var t *database.GoogleToken
	var err error
	// If no tokenID provided but user is known, pick the latest token for that user.
	if tokenID == 0 && userID != nil && *userID > 0 {
		t, err = database.GetLatestGoogleTokenForUser(ctx, *userID)
	} else {
		if userID != nil && *userID > 0 {
			t, err = database.GetGoogleTokenForUser(ctx, tokenID, *userID)
		} else {
			t, err = database.GetGoogleToken(ctx, tokenID)
		}
	}
	if err != nil {
		return nil, err
	}
	token := &OAuthToken{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}
	if time.Now().Before(token.Expiry.Add(-30 * time.Second)) {
		return token, nil
	}
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("token expired and no refresh_token stored")
	}
	newToken, err := c.refresh(ctx, token.RefreshToken)
	if err != nil {
		return nil, err
	}
	if err := database.UpdateGoogleToken(ctx, t.Id, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry); err != nil {
		return nil, err
	}
	return newToken, nil
}

type GmailMessage struct {
	ID      string
	From    string
	Subject string
	Snippet string
	Date    string
}

// ListRecentMessages fetches recent messages from Gmail, newest first. If sinceID is provided, stops when that ID is encountered.
func (c *Client) ListRecentMessages(ctx context.Context, userID *int64, tokenID int64, max int, sinceID string) ([]GmailMessage, error) {
	token, err := c.ensureToken(ctx, userID, tokenID)
	if err != nil {
		return nil, err
	}
	listURL := "https://www.googleapis.com/gmail/v1/users/me/messages?labelIds=INBOX&maxResults=%d"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(listURL, max), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list messages status %d: %s", resp.StatusCode, string(body))
	}
	var list struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	var out []GmailMessage
	for _, m := range list.Messages {
		if sinceID != "" && m.ID == sinceID {
			break
		}
		msg, err := c.fetchMessage(ctx, token.AccessToken, m.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, nil
}

func (c *Client) fetchMessage(ctx context.Context, accessToken, msgID string) (GmailMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/gmail/v1/users/me/messages/"+msgID+"?format=metadata&metadataHeaders=Subject&metadataHeaders=From&metadataHeaders=Date", nil)
	if err != nil {
		return GmailMessage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GmailMessage{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return GmailMessage{}, fmt.Errorf("get message status %d: %s", resp.StatusCode, string(body))
	}
	var data struct {
		ID      string `json:"id"`
		Snippet string `json:"snippet"`
		Payload struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return GmailMessage{}, err
	}
	msg := GmailMessage{
		ID:      data.ID,
		Snippet: data.Snippet,
	}
	for _, h := range data.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "from":
			msg.From = h.Value
		case "subject":
			msg.Subject = h.Value
		case "date":
			msg.Date = h.Value
		}
	}
	return msg, nil
}

// mustEnv retrieves an environment variable or panics if not set.
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env " + key)
	}
	return v
}
