package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const apiBase = "https://slack.com/api"

type Client struct {
	token  string
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		token:  strings.TrimSpace(os.Getenv("SLACK_BOT_TOKEN")),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SendMessage(ctx context.Context, channelID, text string) error {
	payload := map[string]any{
		"channel": channelID,
		"text":    text,
	}
	return c.post(ctx, "chat.postMessage", payload)
}

func (c *Client) SendBlocks(ctx context.Context, channelID, text string, blocks json.RawMessage) error {
	if len(blocks) == 0 {
		return fmt.Errorf("blocks payload is required")
	}
	payload := map[string]any{
		"channel": channelID,
		"blocks":  json.RawMessage(blocks),
	}
	if strings.TrimSpace(text) != "" {
		payload["text"] = text
	}
	return c.post(ctx, "chat.postMessage", payload)
}

func (c *Client) UpdateMessage(ctx context.Context, channelID, messageTS, text string) error {
	payload := map[string]any{
		"channel": channelID,
		"ts":      messageTS,
		"text":    text,
	}
	return c.post(ctx, "chat.update", payload)
}

func (c *Client) DeleteMessage(ctx context.Context, channelID, messageTS string) error {
	payload := map[string]any{
		"channel": channelID,
		"ts":      messageTS,
	}
	return c.post(ctx, "chat.delete", payload)
}

func (c *Client) AddReaction(ctx context.Context, channelID, messageTS, emoji string) error {
	payload := map[string]any{
		"channel":   channelID,
		"timestamp": messageTS,
		"name":      emoji,
	}
	return c.post(ctx, "reactions.add", payload)
}

func (c *Client) post(ctx context.Context, path string, payload any) error {
	if c.token == "" {
		return fmt.Errorf("missing SLACK_BOT_TOKEN")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/"+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack status %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err == nil {
		if !result.OK {
			if result.Error == "" {
				result.Error = "unknown_slack_error"
			}
			return fmt.Errorf("slack error: %s", result.Error)
		}
	}
	return nil
}
