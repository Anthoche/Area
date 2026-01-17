package areas

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
	Hidden     bool         `json:"hidden,omitempty"`
}
