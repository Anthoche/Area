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
		{
			ID:          "weather_temp",
			Name:        "When temperature crosses a threshold",
			Description: "Triggers when current temperature crosses above/below a threshold.",
			Fields: []Field{
				{Key: "city", Type: "string", Required: true, Description: "City name (e.g. Paris)"},
				{Key: "threshold", Type: "number", Required: true, Description: "Temperature threshold (°C)", Example: 20},
				{Key: "direction", Type: "string", Required: true, Description: "above or below", Example: "above"},
				{Key: "interval_minutes", Type: "number", Required: false, Description: "Minimum polling interval in minutes", Example: 5},
			},
		},
		{
			ID:          "weather_report",
			Name:        "Weather report (interval)",
			Description: "Sends current weather for a city every X minutes.",
			Fields: []Field{
				{Key: "city", Type: "string", Required: true, Description: "City name (e.g. Paris)"},
				{Key: "interval_minutes", Type: "number", Required: true, Description: "Polling interval in minutes", Example: 10},
			},
		},
		{
			ID:          "reddit_new_post",
			Name:        "Reddit new post",
			Description: "Triggers when a new post appears in a subreddit.",
			Fields: []Field{
				{Key: "subreddit", Type: "string", Required: true, Description: "Subreddit name (without r/)", Example: "golang"},
				{Key: "interval_minutes", Type: "number", Required: false, Description: "Polling interval in minutes", Example: 5},
			},
		},
		{
			ID:          "youtube_new_video",
			Name:        "YouTube new video",
			Description: "Triggers when a channel publishes a new video.",
			Fields: []Field{
				{Key: "channel", Type: "string", Required: true, Description: "YouTube channel name, handle, or ID", Example: "@GoogleDevelopers"},
				{Key: "interval_minutes", Type: "number", Required: false, Description: "Polling interval in minutes", Example: 5},
			},
		},
	}

	discord := Service{
		ID:      "discord",
		Name:    "Discord",
		Enabled: true,
		Reactions: []Capability{
			{
				ID:             "discord_message",
				Name:           "Send message",
				Description:    "Send a message to a channel using the bot.",
				ActionURL:      "/actions/discord/message",
				DefaultPayload: map[string]any{"content": "Hello from Area"},
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "123456789012345678"},
					{Key: "content", Type: "string", Required: true, Description: "Message content", Example: "Hello from Area"},
				},
			},
			{
				ID:             "discord_embed",
				Name:           "Send embed",
				Description:    "Send an embed to a channel.",
				ActionURL:      "/actions/discord/embed",
				DefaultPayload: map[string]any{"title": "Area update", "description": "Something happened"},
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "123456789012345678"},
					{Key: "title", Type: "string", Required: true, Description: "Embed title"},
					{Key: "description", Type: "string", Required: true, Description: "Embed description"},
					{Key: "url", Type: "string", Required: false, Description: "Embed URL"},
					{Key: "color", Type: "string", Required: false, Description: "Hex color (e.g. #5865F2)"},
					{Key: "content", Type: "string", Required: false, Description: "Optional message content"},
				},
			},
			{
				ID:          "discord_edit_message",
				Name:        "Edit message",
				Description: "Edit a previously sent message.",
				ActionURL:   "/actions/discord/message/edit",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID"},
					{Key: "message_id", Type: "string", Required: true, Description: "Message ID to edit"},
					{Key: "content", Type: "string", Required: true, Description: "New message content"},
				},
			},
			{
				ID:          "discord_delete_message",
				Name:        "Delete message",
				Description: "Delete a message by ID.",
				ActionURL:   "/actions/discord/message/delete",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID"},
					{Key: "message_id", Type: "string", Required: true, Description: "Message ID to delete"},
				},
			},
			{
				ID:          "discord_add_reaction",
				Name:        "Add reaction",
				Description: "Add a reaction emoji to a message.",
				ActionURL:   "/actions/discord/message/react",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID"},
					{Key: "message_id", Type: "string", Required: true, Description: "Message ID to react to"},
					{Key: "emoji", Type: "string", Required: true, Description: "Emoji (e.g. 😀 or name:id)"},
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

	slackEnabled := os.Getenv("SLACK_BOT_TOKEN") != ""
	slack := Service{
		ID:      "slack",
		Name:    "Slack",
		Enabled: slackEnabled,
		Reactions: []Capability{
			{
				ID:             "slack_message",
				Name:           "Send message",
				Description:    "Send a message to a Slack channel.",
				ActionURL:      "/actions/slack/message",
				DefaultPayload: map[string]any{"text": "Hello from Area"},
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "C1234567890"},
					{Key: "text", Type: "string", Required: true, Description: "Message text", Example: "Hello from Area"},
				},
			},
			{
				ID:          "slack_blocks",
				Name:        "Send blocks message",
				Description: "Send a message with Block Kit payload.",
				ActionURL:   "/actions/slack/blocks",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "C1234567890"},
					{Key: "text", Type: "string", Required: false, Description: "Fallback text"},
					{Key: "blocks", Type: "array<object>", Required: true, Description: "Block Kit JSON array"},
				},
			},
			{
				ID:          "slack_update",
				Name:        "Update message",
				Description: "Update an existing message.",
				ActionURL:   "/actions/slack/message/update",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "C1234567890"},
					{Key: "message_ts", Type: "string", Required: true, Description: "Message timestamp"},
					{Key: "text", Type: "string", Required: true, Description: "New message text"},
				},
			},
			{
				ID:          "slack_delete",
				Name:        "Delete message",
				Description: "Delete a message by timestamp.",
				ActionURL:   "/actions/slack/message/delete",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID", Example: "C1234567890"},
					{Key: "message_ts", Type: "string", Required: true, Description: "Message timestamp"},
				},
			},
			{
				ID:          "slack_reaction",
				Name:        "Add reaction",
				Description: "Add an emoji reaction to a message.",
				ActionURL:   "/actions/slack/message/react",
				Fields: []Field{
					{Key: "channel_id", Type: "string", Required: true, Description: "Target channel ID"},
					{Key: "message_ts", Type: "string", Required: true, Description: "Message timestamp"},
					{Key: "emoji", Type: "string", Required: true, Description: "Emoji name without colons", Example: "thumbsup"},
				},
			},
		},
	}

	notionEnabled := os.Getenv("NOTION_TOKEN") != ""
	notion := Service{
		ID:      "notion",
		Name:    "Notion",
		Enabled: notionEnabled,
		Reactions: []Capability{
			{
				ID:          "notion_create_page",
				Name:        "Create page",
				Description: "Create a page under a parent page.",
				ActionURL:   "/actions/notion/page",
				Fields: []Field{
					{Key: "parent_page_id", Type: "string", Required: true, Description: "Parent page ID"},
					{Key: "title", Type: "string", Required: true, Description: "Page title"},
					{Key: "content", Type: "string", Required: false, Description: "Optional paragraph content"},
					{Key: "blocks", Type: "array<object>", Required: false, Description: "Optional Notion blocks (JSON array)"},
				},
			},
			{
				ID:          "notion_append_blocks",
				Name:        "Append blocks",
				Description: "Append blocks to a page or block.",
				ActionURL:   "/actions/notion/blocks",
				Fields: []Field{
					{Key: "block_id", Type: "string", Required: true, Description: "Page or block ID"},
					{Key: "blocks", Type: "array<object>", Required: true, Description: "Notion blocks (JSON array)"},
				},
			},
			{
				ID:          "notion_create_database_row",
				Name:        "Create database row",
				Description: "Create a new page in a database.",
				ActionURL:   "/actions/notion/database",
				Fields: []Field{
					{Key: "database_id", Type: "string", Required: true, Description: "Database ID"},
					{Key: "properties", Type: "object", Required: true, Description: "Notion properties JSON object"},
					{Key: "children", Type: "array<object>", Required: false, Description: "Optional blocks (JSON array)"},
				},
			},
			{
				ID:          "notion_update_page",
				Name:        "Update page",
				Description: "Update page properties.",
				ActionURL:   "/actions/notion/page/update",
				Fields: []Field{
					{Key: "page_id", Type: "string", Required: true, Description: "Page ID"},
					{Key: "properties", Type: "object", Required: true, Description: "Notion properties JSON object"},
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
		slack,
		notion,
		webhook,
	}
}
