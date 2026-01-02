package notion

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

const apiBase = "https://api.notion.com/v1"
const notionVersion = "2022-06-28"

type Client struct {
	token  string
	client *http.Client
}

// NewClient builds a Notion API client.
func NewClient() *Client {
	return &Client{
		token:  strings.TrimSpace(os.Getenv("NOTION_TOKEN")),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// CreatePage creates a Notion page with optional blocks.
func (c *Client) CreatePage(ctx context.Context, parentPageID, title, content string, blocks json.RawMessage) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	payload := map[string]any{
		"parent": map[string]string{
			"page_id": parentPageID,
		},
		"properties": map[string]any{
			"title": []map[string]any{
				{
					"type": "text",
					"text": map[string]any{"content": title},
				},
			},
		},
	}
	children, err := buildChildrenBlocks(content, blocks)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		payload["children"] = children
	}
	return c.doJSON(ctx, http.MethodPost, "/pages", payload)
}

// AppendBlocks appends blocks to a Notion block/page.
func (c *Client) AppendBlocks(ctx context.Context, blockID string, blocks json.RawMessage) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	children, err := decodeBlocks(blocks)
	if err != nil {
		return err
	}
	if len(children) == 0 {
		return fmt.Errorf("blocks payload is required")
	}
	payload := map[string]any{"children": children}
	return c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/blocks/%s/children", blockID), payload)
}

// CreateDatabaseRow creates a Notion database row.
func (c *Client) CreateDatabaseRow(ctx context.Context, databaseID string, properties json.RawMessage, children json.RawMessage) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	props, err := decodeObject(properties, "properties")
	if err != nil {
		return err
	}
	payload := map[string]any{
		"parent": map[string]string{
			"database_id": databaseID,
		},
		"properties": props,
	}
	if len(children) > 0 {
		blocks, err := decodeBlocks(children)
		if err != nil {
			return err
		}
		if len(blocks) > 0 {
			payload["children"] = blocks
		}
	}
	return c.doJSON(ctx, http.MethodPost, "/pages", payload)
}

// UpdatePage updates a Notion page properties.
func (c *Client) UpdatePage(ctx context.Context, pageID string, properties json.RawMessage) error {
	if err := c.ensureToken(); err != nil {
		return err
	}
	props, err := decodeObject(properties, "properties")
	if err != nil {
		return err
	}
	payload := map[string]any{"properties": props}
	return c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/pages/%s", pageID), payload)
}

// ensureToken checks that the Notion token is set.
func (c *Client) ensureToken() error {
	if c.token == "" {
		return fmt.Errorf("missing NOTION_TOKEN")
	}
	return nil
}

// doJSON performs an HTTP request with JSON payload.
func (c *Client) doJSON(ctx context.Context, method, path string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionVersion)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notion status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// buildChildrenBlocks builds Notion blocks from content and raw JSON.
func buildChildrenBlocks(content string, raw json.RawMessage) ([]map[string]any, error) {
	var children []map[string]any
	if strings.TrimSpace(content) != "" {
		children = append(children, map[string]any{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]any{
				"rich_text": []map[string]any{
					{
						"type": "text",
						"text": map[string]any{"content": content},
					},
				},
			},
		})
	}
	if len(raw) > 0 {
		extra, err := decodeBlocks(raw)
		if err != nil {
			return nil, err
		}
		children = append(children, extra...)
	}
	return children, nil
}

// decodeBlocks decodes Notion blocks from raw JSON.
func decodeBlocks(raw json.RawMessage) ([]map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var blocks []map[string]any
	if err := json.Unmarshal(raw, &blocks); err == nil {
		return blocks, nil
	}
	str, err := decodeJSONString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid blocks JSON: %w", err)
	}
	if strings.TrimSpace(str) == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(str), &blocks); err != nil {
		return nil, fmt.Errorf("invalid blocks JSON: %w", err)
	}
	return blocks, nil
}

// decodeObject decodes a JSON object from raw JSON.
func decodeObject(raw json.RawMessage, label string) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%s is required", label)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj, nil
	}
	str, err := decodeJSONString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", label, err)
	}
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", label, err)
	}
	return obj, nil
}

// decodeJSONString decodes a JSON string from raw JSON.
func decodeJSONString(raw json.RawMessage) (string, error) {
	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return "", err
	}
	return str, nil
}
