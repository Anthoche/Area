package areas

import "os"

// Field describes an input required to use a trigger or reaction.
type Field struct {
	Key         string      `json:"key"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Capability represents either a trigger or a reaction.
type Capability struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	ActionURL      string         `json:"action_url,omitempty"`
	DefaultPayload map[string]any `json:"default_payload,omitempty"`
	Fields         []Field        `json:"fields,omitempty"`
}

// Service groups all triggers and reactions available for a provider.
type Service struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Enabled    bool         `json:"enabled"`
	Triggers   []Capability `json:"triggers,omitempty"`
	Reactions  []Capability `json:"reactions,omitempty"`
	MoreInfo   string       `json:"more_info,omitempty"`
	OAuthScope []string     `json:"oauth_scopes,omitempty"`
}

// List returns the catalog of available AREAs (triggers + reactions).
func List() []Service {
	coreTriggers := []Capability{
		{
			ID:          "manual",
			Name:        "Manual trigger",
			Description: "Trigger launched manually from the UI.",
		},
		{
			ID:          "interval",
			Name:        "Timer (interval)",
			Description: "Runs every N minutes.",
			Fields: []Field{
				{Key: "interval_minutes", Type: "number", Required: true, Description: "Delay between runs in minutes", Example: 5},
			},
		},
		{
			ID:          "gmail_inbound",
			Name:        "When a Gmail is received",
			Description: "Triggers on new unread messages in Gmail inbox.",
			Fields:      []Field{},
		},
		{
			ID:          "github_commit",
			Name:        "When a GitHub commit is pushed",
			Description: "Triggers on new commits on a branch.",
			Fields: []Field{
				{Key: "token_id", Type: "number", Required: true, Description: "Stored GitHub token id"},
				{Key: "repo", Type: "string", Required: true, Description: "Repository in owner/name format", Example: "owner/repo"},
				{Key: "branch", Type: "string", Required: true, Description: "Branch to watch", Example: "main"},
			},
		},
		{
			ID:          "github_pull_request",
			Name:        "When a GitHub pull request changes",
			Description: "Triggers on PR updates (opened/closed/merged).",
			Fields: []Field{
				{Key: "token_id", Type: "number", Required: true, Description: "Stored GitHub token id"},
				{Key: "repo", Type: "string", Required: true, Description: "Repository in owner/name format", Example: "owner/repo"},
				{Key: "actions", Type: "array<string>", Required: false, Description: "Actions to watch (opened,closed,merged)", Example: []string{"opened", "closed", "merged"}},
			},
		},
		{
			ID:          "github_issue",
			Name:        "When a GitHub issue changes",
			Description: "Triggers on issue updates (opened/closed/reopened).",
			Fields: []Field{
				{Key: "token_id", Type: "number", Required: true, Description: "Stored GitHub token id"},
				{Key: "repo", Type: "string", Required: true, Description: "Repository in owner/name format", Example: "owner/repo"},
				{Key: "actions", Type: "array<string>", Required: false, Description: "Actions to watch (opened,closed,reopened)", Example: []string{"opened", "closed"}},
			},
		},
	}

	discord := Service{
		ID:      "discord",
		Name:    "Discord",
		Enabled: true,
		Reactions: []Capability{
			{
				ID:             "discord_webhook",
				Name:           "Send message via webhook",
				Description:    "POST a message to a Discord webhook URL.",
				ActionURL:      "",
				DefaultPayload: map[string]any{"content": "Hello from Area"},
				Fields: []Field{
					{Key: "webhook_url", Type: "string", Required: true, Description: "Discord webhook URL", Example: "https://discord.com/api/webhooks/..."},
					{Key: "content", Type: "string", Required: true, Description: "Message content", Example: "Hello from Area"},
				},
			},
		},
	}

	googleEnabled := os.Getenv("GOOGLE_OAUTH_CLIENT_ID") != "" && os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET") != ""
	google := Service{
		ID:      "google",
		Name:    "Google",
		Enabled: googleEnabled,
		OAuthScope: []string{
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Reactions: []Capability{
			{
				ID:          "google_gmail_send",
				Name:        "Send Gmail",
				Description: "Send an email from the authenticated Google account.",
				ActionURL:   "/actions/google/email",
				Fields: []Field{
					{Key: "token_id", Type: "number", Required: true, Description: "Stored Google token id"},
					{Key: "to", Type: "string", Required: true, Description: "Recipient email", Example: "dest@example.com"},
					{Key: "subject", Type: "string", Required: true, Description: "Email subject", Example: "Hello"},
					{Key: "body", Type: "string", Required: true, Description: "Email body", Example: "Hello from Area"},
				},
			},
			{
				ID:          "google_calendar_event",
				Name:        "Create Calendar event",
				Description: "Create an event in the primary calendar.",
				ActionURL:   "/actions/google/calendar",
				Fields: []Field{
					{Key: "token_id", Type: "number", Required: true, Description: "Stored Google token id"},
					{Key: "summary", Type: "string", Required: true, Description: "Event title", Example: "Area event"},
					{Key: "start", Type: "string", Required: true, Description: "Start datetime (RFC3339)", Example: "2025-12-09T14:00:00Z"},
					{Key: "end", Type: "string", Required: true, Description: "End datetime (RFC3339)", Example: "2025-12-09T15:00:00Z"},
					{Key: "attendees", Type: "array<string>", Required: false, Description: "Attendee emails", Example: []string{"a@example.com"}},
				},
			},
		},
	}

	githubEnabled := os.Getenv("GITHUB_OAUTH_CLIENT_ID") != "" && os.Getenv("GITHUB_OAUTH_CLIENT_SECRET") != ""
	github := Service{
		ID:      "github",
		Name:    "GitHub",
		Enabled: githubEnabled,
		Reactions: []Capability{
			{
				ID:          "github_issue",
				Name:        "Create issue",
				Description: "Create a new issue in a repository.",
				ActionURL:   "/actions/github/issue",
				Fields: []Field{
					{Key: "token_id", Type: "number", Required: true, Description: "Stored GitHub token id"},
					{Key: "repo", Type: "string", Required: true, Description: "Repository in owner/name format", Example: "owner/repo"},
					{Key: "title", Type: "string", Required: true, Description: "Issue title"},
					{Key: "body", Type: "string", Required: false, Description: "Issue body"},
					{Key: "labels", Type: "array<string>", Required: false, Description: "Labels to add"},
				},
			},
			{
				ID:          "github_pull_request",
				Name:        "Create pull request",
				Description: "Create a pull request from a branch.",
				ActionURL:   "/actions/github/pr",
				Fields: []Field{
					{Key: "token_id", Type: "number", Required: true, Description: "Stored GitHub token id"},
					{Key: "repo", Type: "string", Required: true, Description: "Repository in owner/name format", Example: "owner/repo"},
					{Key: "title", Type: "string", Required: true, Description: "Pull request title"},
					{Key: "head", Type: "string", Required: true, Description: "Source branch (or owner:branch)", Example: "feature-branch"},
					{Key: "base", Type: "string", Required: true, Description: "Base branch", Example: "main"},
					{Key: "body", Type: "string", Required: false, Description: "Pull request body"},
				},
			},
		},
	}

	webhook := Service{
		ID:      "webhook",
		Name:    "Webhook",
		Enabled: true,
		Reactions: []Capability{
			{
				ID:          "http_webhook",
				Name:        "HTTP POST",
				Description: "Send raw JSON to a custom HTTP endpoint.",
				ActionURL:   "",
				Fields: []Field{
					{Key: "url", Type: "string", Required: true, Description: "Target URL", Example: "https://example.com/hook"},
					{Key: "payload", Type: "object", Required: false, Description: "JSON payload to send"},
				},
			},
		},
	}

	return []Service{
		{
			ID:       "core",
			Name:     "Core",
			Enabled:  true,
			Triggers: coreTriggers,
		},
		discord,
		google,
		github,
		webhook,
	}
}
