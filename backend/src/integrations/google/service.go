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

// Exchange saves the token and returns its id and the user email if available.
func (c *Client) ExchangeAndStore(ctx context.Context, code string, redirectURI string, userID *int64) (int64, string, error) {
	tokenResp, err := c.exchangeCode(ctx, code, redirectURI)
	if err != nil {
		return -1, "", err
	}
	email, _ := c.fetchEmail(ctx, tokenResp.AccessToken)
	id, err := database.InsertGoogleToken(ctx, userID, tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.Expiry)
	if err != nil {
		return -1, "", err
	}
	return id, email, nil
}

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

func buildRawEmail(to, subject, body string) string {
	return fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
}

func (c *Client) ensureToken(ctx context.Context, userID *int64, tokenID int64) (*OAuthToken, error) {
	var t *database.GoogleToken
	var err error
	if userID != nil && *userID > 0 {
		t, err = database.GetGoogleTokenForUser(ctx, tokenID, *userID)
	} else {
		t, err = database.GetGoogleToken(ctx, tokenID)
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
	if err := database.UpdateGoogleToken(ctx, tokenID, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry); err != nil {
		return nil, err
	}
	return newToken, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env " + key)
	}
	return v
}
