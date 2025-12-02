package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrTriggerUnavailable = errors.New("workflow triggerer not configured")
var ErrWorkflowNotFound = errors.New("workflow not found")

// Service orchestrates workflow CRUD and triggering.
type Service struct {
	Store     *Store
	Triggerer *Triggerer
}

func NewService(store *Store, triggerer *Triggerer) *Service {
	return &Service{
		Store:     store,
		Triggerer: triggerer,
	}
}

func (s *Service) Trigger(ctx context.Context, workflowID int64, payload map[string]any) (*Run, error) {
	if s.Triggerer == nil {
		return nil, ErrTriggerUnavailable
	}
	if _, err := s.Store.GetWorkflow(ctx, workflowID); err != nil {
		return nil, ErrWorkflowNotFound
	}
	return s.Triggerer.EnqueueRun(ctx, workflowID, payload)
}

func (s *Service) CreateWorkflow(ctx context.Context, name, triggerType, actionURL string, triggerConfig json.RawMessage) (*Workflow, error) {
	name = strings.TrimSpace(name)
	triggerType = strings.TrimSpace(triggerType)
	actionURL = strings.TrimSpace(actionURL)
	if name == "" || triggerType == "" || actionURL == "" {
		return nil, errors.New("name, triggerType and actionURL are required")
	}
	switch triggerType {
	case "interval":
		if minutes, err := intervalFromConfig(triggerConfig); err != nil || minutes <= 0 {
			return nil, errors.New("interval_minutes must be > 0 for interval trigger")
		}
	case "webhook", "manual":
		// no-op, but ensure valid JSON
		if len(triggerConfig) == 0 {
			triggerConfig = []byte(`{}`)
		}
	default:
		return nil, fmt.Errorf("unsupported trigger_type %s", triggerType)
	}
	return s.Store.CreateWorkflow(ctx, name, triggerType, actionURL, triggerConfig)
}

func (s *Service) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	return s.Store.ListWorkflows(ctx)
}

func (s *Service) GetWorkflow(ctx context.Context, id int64) (*Workflow, error) {
	wf, err := s.Store.GetWorkflow(ctx, id)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	return wf, nil
}

// TriggerWebhook finds a webhook workflow by token and enqueues it with payload.
func (s *Service) TriggerWebhook(ctx context.Context, token string, payload map[string]any) (*Run, error) {
	wf, err := s.Store.FindWorkflowByToken(ctx, token)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	return s.Trigger(ctx, wf.ID, payload)
}
