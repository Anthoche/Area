package database

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Firstname    string
	Lastname     string
	Email        string `gorm:"uniqueIndex:uni_users_email"`
	PasswordHash string
}

type Job struct {
	gorm.Model
	WorkflowID uint            `gorm:"not null;index"`
	RunID      uint            `gorm:"not null;index"`
	Payload    json.RawMessage `gorm:"type:jsonb"`
	Status     string          `gorm:"default:'pending';index"`
	Error      string
	StartedAt  *time.Time
	EndedAt    *time.Time
}

type Run struct {
	gorm.Model
	WorkflowID uint   `gorm:"not null;index"`
	Status     string `gorm:"default:'pending'"`
	StartedAt  *time.Time
	EndedAt    *time.Time
	Error      string
}

// TableName aligns with legacy schema initialized from SQL files.
func (Run) TableName() string { return "workflow_runs" }

type Workflow struct {
	gorm.Model
	UserID        uint            `gorm:"not null;index"`
	Name          string          `gorm:"not null"`
	TriggerType   string          `gorm:"not null"`
	TriggerConfig json.RawMessage `gorm:"type:jsonb"`
	ActionURL     string          `gorm:"not null"`
	Enabled       bool            `gorm:"default:false"`
	NextRunAt     *time.Time
}

type GoogleToken struct {
	gorm.Model
	UserID       *int64
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	CreatedAt    time.Time
}

type GithubToken struct {
	gorm.Model
	UserID      *int64
	AccessToken string
	TokenType   string
	Scope       string
	CreatedAt   time.Time
}

type AreaService struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Enabled     bool
	MoreInfo    string
	OAuthScopes json.RawMessage `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (AreaService) TableName() string { return "area_services" }

type AreaCapability struct {
	ID             string `gorm:"primaryKey"`
	ServiceID      string `gorm:"primaryKey;index"`
	Kind           string `gorm:"index"`
	Name           string
	Description    string
	ActionURL      string
	DefaultPayload json.RawMessage `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (AreaCapability) TableName() string { return "area_capabilities" }

type AreaField struct {
	ID           uint   `gorm:"primaryKey"`
	ServiceID    string `gorm:"index"`
	CapabilityID string `gorm:"index"`
	Key          string
	Type         string
	Required     bool
	Description  string
	Example      json.RawMessage `gorm:"type:jsonb"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AreaField) TableName() string { return "area_fields" }
