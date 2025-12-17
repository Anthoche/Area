package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"area/src/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusSucceeded  = "succeeded"
	JobStatusFailed     = "failed"

	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSucceeded = "succeeded"
	RunStatusFailed    = "failed"
)

// Workflow API Response Types (keep existing for API compatibility)
type Workflow struct {
	ID            int64           `json:"id"`
	UserID        int64           `json:"user_id"`
	Name          string          `json:"name"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	ActionURL     string          `json:"action_url"`
	Enabled       bool            `json:"enabled"`
	NextRunAt     *time.Time      `json:"next_run_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Run struct {
	ID         int64      `json:"id"`
	WorkflowID int64      `json:"workflow_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type Job struct {
	ID         int64           `json:"id"`
	WorkflowID int64           `json:"workflow_id"`
	RunID      int64           `json:"run_id"`
	Payload    json.RawMessage `json:"payload"`
	Status     string          `json:"status"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	StartedAt  *time.Time      `json:"started_at,omitempty"`
	EndedAt    *time.Time      `json:"ended_at,omitempty"`
}

type IntervalConfig struct {
	IntervalMinutes int                    `json:"interval_minutes"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
}

type GithubCommitConfig struct {
	TokenID         int64                  `json:"token_id"`
	Repo            string                 `json:"repo"`
	Branch          string                 `json:"branch"`
	PayloadTemplate map[string]interface{} `json:"payload_template,omitempty"`
}

type GithubPullRequestConfig struct {
	TokenID         int64                  `json:"token_id"`
	Repo            string                 `json:"repo"`
	Actions         []string               `json:"actions,omitempty"`
	PayloadTemplate map[string]interface{} `json:"payload_template,omitempty"`
}

type GithubIssueConfig struct {
	TokenID         int64                  `json:"token_id"`
	Repo            string                 `json:"repo"`
	Actions         []string               `json:"actions,omitempty"`
	PayloadTemplate map[string]interface{} `json:"payload_template,omitempty"`
}

type Store struct {
	db *gorm.DB
}

// NewStore builds a Store backed by the provided database handle.
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// NewDefaultStore uses the shared database connection created in database.Connect().
func NewDefaultStore() *Store {
	return &Store{db: database.GetDB()}
}

// Helper functions to convert between models and API types
func workflowModelToAPI(model database.Workflow) Workflow {
	return Workflow{
		ID:            int64(model.ID),
		UserID:        int64(model.UserID),
		Name:          model.Name,
		TriggerType:   model.TriggerType,
		TriggerConfig: model.TriggerConfig,
		ActionURL:     model.ActionURL,
		Enabled:       model.Enabled,
		NextRunAt:     model.NextRunAt,
		CreatedAt:     model.CreatedAt,
	}
}

// runModelToAPI converts a database.Run model to the API Run type.
func runModelToAPI(model database.Run) Run {
	return Run{
		ID:         int64(model.ID),
		WorkflowID: int64(model.WorkflowID),
		Status:     model.Status,
		CreatedAt:  model.CreatedAt,
		StartedAt:  model.StartedAt,
		EndedAt:    model.EndedAt,
		Error:      model.Error,
	}
}

// jobModelToAPI converts a database.Job model to the API Job type.
func jobModelToAPI(model database.Job) Job {
	return Job{
		ID:         int64(model.ID),
		WorkflowID: int64(model.WorkflowID),
		RunID:      int64(model.RunID),
		Payload:    model.Payload,
		Status:     model.Status,
		Error:      model.Error,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		StartedAt:  model.StartedAt,
		EndedAt:    model.EndedAt,
	}
}

// CreateWorkflow persists a new workflow with its trigger configuration.
func (s *Store) CreateWorkflow(ctx context.Context, userID int64, name, triggerType, actionURL string, triggerConfig json.RawMessage) (*Workflow, error) {
	initialEnabled := triggerType == "manual"

	model := database.Workflow{
		UserID:        uint(userID),
		Name:          name,
		TriggerType:   triggerType,
		TriggerConfig: triggerConfig,
		ActionURL:     actionURL,
		Enabled:       initialEnabled,
	}

	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}

	workflow := workflowModelToAPI(model)
	return &workflow, nil
}

// ListWorkflows returns all workflows for a user ordered by creation date.
func (s *Store) ListWorkflows(ctx context.Context, userID int64) ([]Workflow, error) {
	var models []database.Workflow
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}

	workflows := make([]Workflow, len(models))
	for i, model := range models {
		workflows[i] = workflowModelToAPI(model)
		// Apply business logic: disable interval workflows without a next_run_at
		if workflows[i].TriggerType == "interval" && workflows[i].Enabled && workflows[i].NextRunAt == nil {
			workflows[i].Enabled = false
		}
	}

	return workflows, nil
}

// ListWorkflowsByTrigger returns workflows filtered by trigger type (all users).
func (s *Store) ListWorkflowsByTrigger(ctx context.Context, triggerType string) ([]Workflow, error) {
	var models []database.Workflow

	if err := s.db.WithContext(ctx).
		Where("trigger_type = ?", triggerType).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list workflows by trigger: %w", err)
	}

	workflows := make([]Workflow, len(models))
	for i, model := range models {
		workflows[i] = workflowModelToAPI(model)

		// Apply business logic: disable invalid interval workflows
		if workflows[i].TriggerType == "interval" &&
			workflows[i].Enabled &&
			workflows[i].NextRunAt == nil {
			workflows[i].Enabled = false
		}
	}
	return workflows, nil
}

// GetWorkflow fetches a workflow by ID (no user check).
func (s *Store) GetWorkflow(ctx context.Context, id int64) (*Workflow, error) {
	var model database.Workflow
	if err := s.db.WithContext(ctx).First(&model, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get workflow: %w", err)
	}

	workflow := workflowModelToAPI(model)

	if workflow.TriggerType == "interval" && workflow.Enabled && workflow.NextRunAt == nil {
		workflow.Enabled = false
	}
	return &workflow, nil
}

// DeleteWorkflow removes a workflow (cascades to runs/jobs via FK).
func (s *Store) DeleteWorkflow(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Delete(&database.Workflow{}, uint(id))
	if result.Error != nil {
		return fmt.Errorf("delete workflow: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetWorkflowForUser fetches a workflow by ID constrained to the owner.
func (s *Store) GetWorkflowForUser(ctx context.Context, id int64, userID int64) (*Workflow, error) {
	var model database.Workflow

	if err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&model).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get workflow: %w", err)
	}

	workflow := workflowModelToAPI(model)

	// Apply business logic
	if workflow.TriggerType == "interval" &&
		workflow.Enabled &&
		workflow.NextRunAt == nil {
		workflow.Enabled = false
	}
	return &workflow, nil
}

// DeleteWorkflowForUser deletes a workflow if it belongs to the user.
func (s *Store) DeleteWorkflowForUser(ctx context.Context, id int64, userID int64) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&database.Workflow{})

	if res.Error != nil {
		return fmt.Errorf("delete workflow: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEnabledForUser toggles the enabled flag for a user's workflow; interval gets next_run_at.
func (s *Store) SetEnabledForUser(ctx context.Context, id int64, userID int64, enabled bool, now time.Time) error {
	if enabled {
		wf, err := s.GetWorkflowForUser(ctx, id, userID)
		if err != nil {
			return err
		}

		updates := map[string]interface{}{
			"enabled": true,
		}

		if wf.TriggerType == "interval" {
			cfg, err := intervalConfigFromJSON(wf.TriggerConfig)
			if err != nil || cfg.IntervalMinutes <= 0 {
				return fmt.Errorf("invalid interval config")
			}
			nextRun := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
			updates["next_run_at"] = nextRun
		}

		result := s.db.WithContext(ctx).Model(&database.Workflow{}).Where("id = ? AND user_id = ?", uint(id), userID).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("enable workflow: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}

	// Disable workflow
	result := s.db.WithContext(ctx).Model(&database.Workflow{}).Where("id = ? AND user_id = ?", uint(id), userID).Updates(map[string]interface{}{
		"enabled":     false,
		"next_run_at": nil,
	})
	if result.Error != nil {
		return fmt.Errorf("disable workflow: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateRun creates a new pending run for a workflow.
func (s *Store) CreateRun(ctx context.Context, workflowID int64) (*Run, error) {
	model := database.Run{
		WorkflowID: uint(workflowID),
		Status:     RunStatusPending,
	}

	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	run := runModelToAPI(model)
	return &run, nil
}

type RunUpdate struct {
	Status    string
	StartedAt *time.Time
	EndedAt   *time.Time
	Error     *string
}

// UpdateRun updates run metadata such as status or timestamps.
func (s *Store) UpdateRun(ctx context.Context, runID int64, upd RunUpdate) error {
	updates := make(map[string]interface{})

	if upd.Status != "" {
		updates["status"] = upd.Status
	}
	if upd.StartedAt != nil {
		updates["started_at"] = *upd.StartedAt
	}
	if upd.EndedAt != nil {
		updates["ended_at"] = *upd.EndedAt
	}
	if upd.Error != nil {
		updates["error"] = *upd.Error
	}

	if err := s.db.WithContext(ctx).Model(&database.Run{}).Where("id = ?", uint(runID)).Updates(updates).Error; err != nil {
		return fmt.Errorf("update run: %w", err)
	}
	return nil
}

// CreateJob inserts a pending job belonging to a workflow run.
func (s *Store) CreateJob(ctx context.Context, workflowID, runID int64, payload json.RawMessage) (*Job, error) {
	model := database.Job{
		WorkflowID: uint(workflowID),
		RunID:      uint(runID),
		Payload:    payload,
		Status:     JobStatusPending,
	}

	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	job := jobModelToAPI(model)
	return &job, nil
}

// FetchNextPendingJob locks and returns the oldest pending job.
func (s *Store) FetchNextPendingJob(ctx context.Context) (*Job, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer tx.Rollback()

	var model database.Job
	result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ?", JobStatusPending).
		Order("created_at, id").
		Limit(1).
		Find(&model)

	if result.Error != nil {
		return nil, fmt.Errorf("scan job: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	now := time.Now()
	if err := tx.Model(&model).Updates(map[string]interface{}{
		"status":     JobStatusProcessing,
		"started_at": now,
		"updated_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("mark processing: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit job lock: %w", err)
	}

	// Update the model with the new values
	model.Status = JobStatusProcessing
	model.StartedAt = &now
	model.UpdatedAt = now

	job := jobModelToAPI(model)
	return &job, nil
}

// MarkJobSuccess marks a job as succeeded and closes its timestamps.
func (s *Store) MarkJobSuccess(ctx context.Context, jobID int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     JobStatusSucceeded,
		"updated_at": now,
		"ended_at":   now,
		"error":      "",
	}

	if err := s.db.WithContext(ctx).Model(&database.Job{}).Where("id = ?", uint(jobID)).Updates(updates).Error; err != nil {
		return fmt.Errorf("mark job success: %w", err)
	}
	return nil
}

// MarkJobFailed marks a job as failed with the provided reason.
func (s *Store) MarkJobFailed(ctx context.Context, jobID int64, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     JobStatusFailed,
		"updated_at": now,
		"ended_at":   now,
		"error":      reason,
	}

	if err := s.db.WithContext(ctx).Model(&database.Job{}).Where("id = ?", uint(jobID)).Updates(updates).Error; err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	return nil
}

// ClaimDueIntervalWorkflows locks and returns interval workflows whose next_run_at <= now, and advances next_run_at.
func (s *Store) ClaimDueIntervalWorkflows(ctx context.Context, now time.Time) ([]Workflow, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer tx.Rollback()

	var models []database.Workflow
	err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("trigger_type = ? AND enabled = ? AND next_run_at IS NOT NULL AND next_run_at <= ?",
			"interval", true, now).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("claim due interval: %w", err)
	}

	var workflows []Workflow
	for _, model := range models {
		cfg, err := intervalConfigFromJSON(model.TriggerConfig)
		if err != nil || cfg.IntervalMinutes <= 0 {
			continue
		}

		nextRun := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
		if err := tx.Model(&model).Update("next_run_at", nextRun).Error; err != nil {
			return nil, fmt.Errorf("update next_run_at: %w", err)
		}

		workflows = append(workflows, workflowModelToAPI(model))
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit interval claim: %w", err)
	}

	return workflows, nil
}

// FindWorkflowByToken returns a webhook workflow matching the token stored in trigger_config.
func (s *Store) FindWorkflowByToken(ctx context.Context, token string) (*Workflow, error) {
	var model database.Workflow
	err := s.db.WithContext(ctx).
		Where("trigger_type = ? AND enabled = ? AND trigger_config->>'token' = ?", "webhook", true, token).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find workflow token: %w", err)
	}

	workflow := workflowModelToAPI(model)
	return &workflow, nil
}

// intervalConfigFromJSON parses an IntervalConfig from raw JSON.
func intervalConfigFromJSON(raw json.RawMessage) (IntervalConfig, error) {
	if len(raw) == 0 {
		return IntervalConfig{}, errors.New("empty config")
	}
	var cfg IntervalConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return IntervalConfig{}, err
	}
	return cfg, nil
}

// githubCommitConfigFromJSON parses GitHub commit trigger config.
func githubCommitConfigFromJSON(raw json.RawMessage) (GithubCommitConfig, error) {
	if len(raw) == 0 {
		return GithubCommitConfig{}, errors.New("empty config")
	}
	var cfg GithubCommitConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return GithubCommitConfig{}, err
	}
	return cfg, nil
}

// githubPRConfigFromJSON parses GitHub pull request trigger config.
func githubPRConfigFromJSON(raw json.RawMessage) (GithubPullRequestConfig, error) {
	if len(raw) == 0 {
		return GithubPullRequestConfig{}, errors.New("empty config")
	}
	var cfg GithubPullRequestConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return GithubPullRequestConfig{}, err
	}
	return cfg, nil
}

// githubIssueConfigFromJSON parses GitHub issue trigger config.
func githubIssueConfigFromJSON(raw json.RawMessage) (GithubIssueConfig, error) {
	if len(raw) == 0 {
		return GithubIssueConfig{}, errors.New("empty config")
	}
	var cfg GithubIssueConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return GithubIssueConfig{}, err
	}
	return cfg, nil
}
