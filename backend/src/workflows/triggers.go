package workflows

import (
	"context"
	"encoding/json"
	"fmt"
)

// Triggerer creates runs + jobs when an event (manual/webhook/GitHub) happens.
type Triggerer struct {
	store *Store
}

// NewTriggerer creates a Triggerer bound to a Store.
func NewTriggerer(store *Store) *Triggerer {
	return &Triggerer{store: store}
}

// EnqueueRun creates a run and a pending job with the provided payload.
func (t *Triggerer) EnqueueRun(ctx context.Context, workflowID int64, payload map[string]any) (*Run, error) {
	if t.store == nil {
		return nil, fmt.Errorf("triggerer: store is nil")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}

	run, err := t.store.CreateRun(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if _, err := t.store.CreateJob(ctx, workflowID, run.ID, encoded); err != nil {
		return nil, err
	}
	return run, nil
}
