package workflows

import (
	"context"
	"errors"
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

func (s *Service) CreateWorkflow(ctx context.Context, name, triggerType, actionURL string) (*Workflow, error) {
	name = strings.TrimSpace(name)
	triggerType = strings.TrimSpace(triggerType)
	actionURL = strings.TrimSpace(actionURL)
	if name == "" || triggerType == "" || actionURL == "" {
		return nil, errors.New("name, triggerType and actionURL are required")
	}
	return s.Store.CreateWorkflow(ctx, name, triggerType, actionURL)
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
