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
	Email        string
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

type Workflow struct {
	gorm.Model
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
