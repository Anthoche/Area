package google

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"area/src/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Client wraps token retrieval and Google API calls without extra dependencies.
type Client struct {
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

func NewClient() *Client {
	return &Client{
		oauthConfig: &oauth2.Config{
			ClientID:     mustEnv("GOOGLE_OAUTH_CLIENT_ID"),
			ClientSecret: mustEnv("GOOGLE_OAUTH_CLIENT_SECRET"),
			Scopes: []string{
				"https://www.googleapis.com/auth/gmail.send",
				"https://www.googleapis.com/auth/calendar.events",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthURL builds the authorization URL with a given state and redirect URI.
func (c *Client) AuthURL(state, redirectURI string) string {
	cfg := *c.oauthConfig
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// Exchange saves the token and returns its id.
func (c *Client) ExchangeAndStore(ctx context.Context, code string, redirectURI string, userID *int64) (int64, error) {
	cfg := *c.oauthConfig
	cfg.RedirectURL = redirectURI
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return -1, fmt.Errorf("exchange code: %w", err)
	}
	return database.InsertGoogleToken(ctx, userID, token.AccessToken, token.RefreshToken, token.Expiry)
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

func (c *Client) ensureToken(ctx context.Context, userID *int64, tokenID int64) (*oauth2.Token, error) {
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
	token := &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}
	if token.Valid() {
		return token, nil
	}
	src := c.oauthConfig.TokenSource(ctx, token)
	newToken, err := src.Token()
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
