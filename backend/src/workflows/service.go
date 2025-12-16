package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTriggerUnavailable = errors.New("workflow triggerer not configured")
var ErrWorkflowNotFound = errors.New("workflow not found")
var ErrWorkflowDisabled = errors.New("workflow disabled")

// Service orchestrates workflow CRUD and triggering.
type Service struct {
	Store     *Store
	Triggerer *Triggerer
}

// NewService constructs a workflow service with its store and triggerer.
func NewService(store *Store, triggerer *Triggerer) *Service {
	return &Service{
		Store:     store,
		Triggerer: triggerer,
	}
}

// Trigger enqueues a workflow run with the provided payload.
func (s *Service) Trigger(ctx context.Context, workflowID int64, payload map[string]any) (*Run, error) {
	if s.Triggerer == nil {
		return nil, ErrTriggerUnavailable
	}
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	wf, err := s.Store.GetWorkflowForUser(ctx, workflowID, userID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	if !wf.Enabled && wf.TriggerType != "manual" {
		return nil, ErrWorkflowDisabled
	}
	return s.Triggerer.EnqueueRun(ctx, workflowID, payload)
}

// CreateWorkflow validates input and stores a new workflow.
func (s *Service) CreateWorkflow(ctx context.Context, name, triggerType, actionURL string, triggerConfig json.RawMessage) (*Workflow, error) {
	name = strings.TrimSpace(name)
	triggerType = strings.TrimSpace(triggerType)
	actionURL = strings.TrimSpace(actionURL)
	if name == "" || triggerType == "" || actionURL == "" {
		return nil, errors.New("name, triggerType and actionURL are required")
	}
	switch triggerType {
	case "interval":
		cfg, err := intervalConfigFromJSON(triggerConfig)
		if err != nil || cfg.IntervalMinutes <= 0 {
			return nil, errors.New("interval_minutes must be > 0 for interval trigger")
		}
	case "webhook", "manual", "gmail_inbound":
		// no-op, but ensure valid JSON
		if len(triggerConfig) == 0 {
			triggerConfig = []byte(`{}`)
		}
	default:
		return nil, fmt.Errorf("unsupported trigger_type %s", triggerType)
	}
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.CreateWorkflow(ctx, userID, name, triggerType, actionURL, triggerConfig)
}

// ListWorkflows returns all persisted workflows.
func (s *Service) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListWorkflows(ctx, userID)
}

// GetWorkflow fetches a workflow by ID or returns ErrWorkflowNotFound.
func (s *Service) GetWorkflow(ctx context.Context, id int64) (*Workflow, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	wf, err := s.Store.GetWorkflowForUser(ctx, id, userID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	return wf, nil
}

// DeleteWorkflow removes a workflow and its related runs/jobs.
func (s *Service) DeleteWorkflow(ctx context.Context, id int64) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.DeleteWorkflowForUser(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWorkflowNotFound
		}
		return err
	}
	return nil
}

// SetEnabled toggles a workflow (non-manual can be paused); interval workflows are rescheduled on enable.
func (s *Service) SetEnabled(ctx context.Context, id int64, enabled bool, now time.Time) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.SetEnabledForUser(ctx, id, userID, enabled, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWorkflowNotFound
		}
		return err
	}
	return nil
}

// TriggerWebhook finds a webhook workflow by token and enqueues it with payload.
func (s *Service) TriggerWebhook(ctx context.Context, token string, payload map[string]any) (*Run, error) {
	wf, err := s.Store.FindWorkflowByToken(ctx, token)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	if !wf.Enabled {
		return nil, ErrWorkflowDisabled
	}
	ctx = WithUserID(ctx, wf.UserID)
	return s.Trigger(ctx, wf.ID, payload)
}

// IntervalConfigFromJSON exposes interval config parsing to callers (e.g., scheduler).
func IntervalConfigFromJSON(raw json.RawMessage) (IntervalConfig, error) {
	return intervalConfigFromJSON(raw)
}

type ctxUserIDKey struct{}

// WithUserID returns a context carrying the user id for authz in workflow operations.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

func userIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(ctxUserIDKey{})
	if val == nil {
		return 0, fmt.Errorf("missing user id in context")
	}
	uid, ok := val.(int64)
	if !ok || uid <= 0 {
		return 0, fmt.Errorf("invalid user id in context")
	}
	return uid, nil
}
