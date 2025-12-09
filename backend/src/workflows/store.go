package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"area/src/database"
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

type Workflow struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	ActionURL     string          `json:"action_url"`
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

type Store struct {
	db *sql.DB
}

// NewStore builds a Store backed by the provided database handle.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// NewDefaultStore uses the shared database connection created in database.Connect().
func NewDefaultStore() *Store {
	return &Store{db: database.GetDB()}
}

// CreateWorkflow persists a new workflow with its trigger configuration.
func (s *Store) CreateWorkflow(ctx context.Context, name, triggerType, actionURL string, triggerConfig json.RawMessage) (*Workflow, error) {
	var id int64
	var nextRun sql.NullTime
	if triggerType == "interval" {
		if cfg, err := intervalConfigFromJSON(triggerConfig); err == nil && cfg.IntervalMinutes > 0 {
			nextRun = sql.NullTime{Time: time.Now().Add(time.Duration(cfg.IntervalMinutes) * time.Minute), Valid: true}
		}
	}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO workflows (name, trigger_type, trigger_config, action_url, next_run_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		name, triggerType, triggerConfig, actionURL, nextRun,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}
	return s.GetWorkflow(ctx, id)
}

// ListWorkflows returns all workflows ordered by creation date.
func (s *Store) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at
		FROM workflows
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()

	var out []Workflow
	for rows.Next() {
		var wf Workflow
		var nextRun sql.NullTime
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.TriggerType, &wf.TriggerConfig, &wf.ActionURL, &nextRun, &wf.CreatedAt); err != nil {
			return nil, fmt.Errorf("list workflows: %w", err)
		}
		if nextRun.Valid {
			wf.NextRunAt = &nextRun.Time
		}
		out = append(out, wf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	return out, nil
}

// GetWorkflow fetches a workflow by ID.
func (s *Store) GetWorkflow(ctx context.Context, id int64) (*Workflow, error) {
	var wf Workflow
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at
		FROM workflows
		WHERE id = $1`,
		id,
	)
	var nextRun sql.NullTime
	if err := row.Scan(&wf.ID, &wf.Name, &wf.TriggerType, &wf.TriggerConfig, &wf.ActionURL, &nextRun, &wf.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get workflow: %w", err)
	}
	if nextRun.Valid {
		wf.NextRunAt = &nextRun.Time
	}
	return &wf, nil
}

// DeleteWorkflow removes a workflow (cascades to runs/jobs via FK).
func (s *Store) DeleteWorkflow(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CreateRun creates a new pending run for a workflow.
func (s *Store) CreateRun(ctx context.Context, workflowID int64) (*Run, error) {
	var id int64
	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO workflow_runs (workflow_id)
		VALUES ($1)
		RETURNING id`,
		workflowID,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}
	now := time.Now()
	return &Run{
		ID:         id,
		WorkflowID: workflowID,
		Status:     RunStatusPending,
		CreatedAt:  now,
	}, nil
}

type RunUpdate struct {
	Status    string
	StartedAt *time.Time
	EndedAt   *time.Time
	Error     *string
}

// UpdateRun updates run metadata such as status or timestamps.
func (s *Store) UpdateRun(ctx context.Context, runID int64, upd RunUpdate) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE workflow_runs
		SET status = COALESCE(NULLIF($1, ''), status),
		    started_at = COALESCE($2, started_at),
		    ended_at = COALESCE($3, ended_at),
		    error = COALESCE($4, error)
		WHERE id = $5`,
		upd.Status, nullableTime(upd.StartedAt), nullableTime(upd.EndedAt), nullableString(upd.Error), runID,
	)
	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}
	return nil
}

// CreateJob inserts a pending job belonging to a workflow run.
func (s *Store) CreateJob(ctx context.Context, workflowID, runID int64, payload json.RawMessage) (*Job, error) {
	var id int64
	var createdAt, updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO jobs (workflow_id, run_id, payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		workflowID, runID, payload,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	return &Job{
		ID:         id,
		WorkflowID: workflowID,
		RunID:      runID,
		Payload:    payload,
		Status:     JobStatusPending,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// FetchNextPendingJob locks and returns the oldest pending job.
func (s *Store) FetchNextPendingJob(ctx context.Context) (*Job, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		SELECT id, workflow_id, run_id, payload, status, error, created_at, updated_at, started_at, ended_at
		FROM jobs
		WHERE status = $1
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1`,
		JobStatusPending,
	)

	var job Job
	var errMsg sql.NullString
	var started, ended sql.NullTime
	if err := row.Scan(&job.ID, &job.WorkflowID, &job.RunID, &job.Payload, &job.Status, &errMsg, &job.CreatedAt, &job.UpdatedAt, &started, &ended); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan job: %w", err)
	}
	if errMsg.Valid {
		job.Error = errMsg.String
	}
	if started.Valid {
		job.StartedAt = &started.Time
	}
	if ended.Valid {
		job.EndedAt = &ended.Time
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
		UPDATE jobs
		SET status = $1, started_at = $2, updated_at = $2
		WHERE id = $3`,
		JobStatusProcessing, now, job.ID,
	); err != nil {
		return nil, fmt.Errorf("mark processing: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit job lock: %w", err)
	}
	job.Status = JobStatusProcessing
	job.StartedAt = &now
	job.UpdatedAt = now
	return &job, nil
}

// MarkJobSuccess marks a job as succeeded and closes its timestamps.
func (s *Store) MarkJobSuccess(ctx context.Context, jobID int64) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = $1, updated_at = $2, ended_at = $2, error = NULL
		WHERE id = $3`,
		JobStatusSucceeded, now, jobID,
	)
	if err != nil {
		return fmt.Errorf("mark job success: %w", err)
	}
	return nil
}

// MarkJobFailed marks a job as failed with the provided reason.
func (s *Store) MarkJobFailed(ctx context.Context, jobID int64, reason string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = $1, updated_at = $2, ended_at = $2, error = $3
		WHERE id = $4`,
		JobStatusFailed, now, reason, jobID,
	)
	if err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	return nil
}

// ClaimDueIntervalWorkflows locks and returns interval workflows whose next_run_at <= now, and advances next_run_at.
func (s *Store) ClaimDueIntervalWorkflows(ctx context.Context, now time.Time) ([]Workflow, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at
		FROM workflows
		WHERE trigger_type = 'interval'
		  AND next_run_at IS NOT NULL
		  AND next_run_at <= $1
		FOR UPDATE SKIP LOCKED`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("claim due interval: %w", err)
	}
	defer rows.Close()

	var due []Workflow
	for rows.Next() {
		var wf Workflow
		var nextRun sql.NullTime
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.TriggerType, &wf.TriggerConfig, &wf.ActionURL, &nextRun, &wf.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan interval wf: %w", err)
		}
		if nextRun.Valid {
			wf.NextRunAt = &nextRun.Time
		}
		due = append(due, wf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows interval: %w", err)
	}

	for _, wf := range due {
		cfg, err := intervalConfigFromJSON(wf.TriggerConfig)
		if err != nil || cfg.IntervalMinutes <= 0 {
			continue
		}
		next := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
		if _, err := tx.ExecContext(ctx, `
			UPDATE workflows SET next_run_at = $1 WHERE id = $2`, next, wf.ID); err != nil {
			return nil, fmt.Errorf("update next_run_at: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit interval claim: %w", err)
	}
	return due, nil
}

// FindWorkflowByToken returns a webhook workflow matching the token stored in trigger_config.
func (s *Store) FindWorkflowByToken(ctx context.Context, token string) (*Workflow, error) {
	var wf Workflow
	var nextRun sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at
		FROM workflows
		WHERE trigger_type = 'webhook' AND trigger_config->>'token' = $1
		LIMIT 1`,
		token,
	).Scan(&wf.ID, &wf.Name, &wf.TriggerType, &wf.TriggerConfig, &wf.ActionURL, &nextRun, &wf.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("find workflow token: %w", err)
	}
	if nextRun.Valid {
		wf.NextRunAt = &nextRun.Time
	}
	return &wf, nil
}

// nullableTime converts a pointer time to a value usable in SQL queries.
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// nullableString converts a string pointer to a SQL-compatible nullable value.
func nullableString(s *string) any {
	if s == nil {
		return nil
	}
	return sql.NullString{String: *s, Valid: true}
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
