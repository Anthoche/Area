package discord

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
)

const apiBase = "https://discord.com/api/v10"

type Client struct {
	token  string
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		token:  strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN")),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type Embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Color       int    `json:"color,omitempty"`
}

func (c *Client) SendMessage(ctx context.Context, channelID, content string, embeds []Embed) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	body := map[string]any{
		"content": content,
	}
	if len(embeds) > 0 {
		body["embeds"] = embeds
	}
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("%s/channels/%s/messages", apiBase, channelID), body)
}

func (c *Client) EditMessage(ctx context.Context, channelID, messageID, content string) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	body := map[string]any{
		"content": content,
	}
	return c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/channels/%s/messages/%s", apiBase, channelID, messageID), body)
}

func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/channels/%s/messages/%s", apiBase, channelID, messageID), nil)
}

func (c *Client) AddReaction(ctx context.Context, channelID, messageID, emoji string) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	emoji = strings.TrimSpace(emoji)
	emoji = strings.Trim(emoji, ":")
	if emoji == "" {
		return fmt.Errorf("emoji is required")
	}
	escaped := url.PathEscape(emoji)
	endpoint := fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", apiBase, channelID, messageID, escaped)
	return c.doJSON(ctx, http.MethodPut, endpoint, nil)
}

func (c *Client) ensureToken() error {
	if c.token == "" {
		return fmt.Errorf("missing DISCORD_BOT_TOKEN")
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, payload any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
