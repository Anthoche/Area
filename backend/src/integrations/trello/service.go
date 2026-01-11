package trello

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const apiBase = "https://api.trello.com/1"

// Client wraps Trello API operations.
type Client struct {
	key    string
	token  string
	client *http.Client
}

// NewClient builds a Trello client from environment credentials.
func NewClient() *Client {
	return NewClientWithCredentials(os.Getenv("TRELLO_API_KEY"), os.Getenv("TRELLO_TOKEN"))
}

// NewClientWithCredentials builds a Trello client with explicit credentials.
func NewClientWithCredentials(key, token string) *Client {
	return &Client{
		key:    strings.TrimSpace(key),
		token:  strings.TrimSpace(token),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateCard creates a new Trello card in a list.
func (c *Client) CreateCard(ctx context.Context, listID, name, desc, pos string) error {
	if err := c.ensureCredentials(); err != nil {
		return err
	}
	params := url.Values{}
	params.Set("idList", strings.TrimSpace(listID))
	params.Set("name", strings.TrimSpace(name))
	if strings.TrimSpace(desc) != "" {
		params.Set("desc", desc)
	}
	if strings.TrimSpace(pos) != "" {
		params.Set("pos", strings.TrimSpace(pos))
	}
	return c.doForm(ctx, http.MethodPost, apiBase+"/cards", params)
}

// MoveCard moves a Trello card to another list.
func (c *Client) MoveCard(ctx context.Context, cardID, listID, pos string) error {
	if err := c.ensureCredentials(); err != nil {
		return err
	}
	params := url.Values{}
	params.Set("idList", strings.TrimSpace(listID))
	if strings.TrimSpace(pos) != "" {
		params.Set("pos", strings.TrimSpace(pos))
	}
	endpoint := fmt.Sprintf("%s/cards/%s", apiBase, url.PathEscape(strings.TrimSpace(cardID)))
	return c.doForm(ctx, http.MethodPut, endpoint, params)
}

// CreateList creates a list on a board.
func (c *Client) CreateList(ctx context.Context, boardID, name, pos string) error {
	if err := c.ensureCredentials(); err != nil {
		return err
	}
	params := url.Values{}
	params.Set("idBoard", strings.TrimSpace(boardID))
	params.Set("name", strings.TrimSpace(name))
	if strings.TrimSpace(pos) != "" {
		params.Set("pos", strings.TrimSpace(pos))
	}
	return c.doForm(ctx, http.MethodPost, apiBase+"/lists", params)
}

// ensureCredentials checks that API credentials are set.
func (c *Client) ensureCredentials() error {
	if c.key == "" || c.token == "" {
		return fmt.Errorf("missing TRELLO_API_KEY or TRELLO_TOKEN")
	}
	return nil
}

// doForm performs an HTTP request with form-encoded parameters.
func (c *Client) doForm(ctx context.Context, method, endpoint string, params url.Values) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("key", c.key)
	params.Set("token", c.token)

	var body io.Reader
	if method == http.MethodGet {
		endpoint = endpoint + "?" + params.Encode()
	} else {
		body = strings.NewReader(params.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("trello status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
