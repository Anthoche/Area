package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// OutboundSender is implemented by integrations that can send actions (webhooks, email, etc.).
type OutboundSender interface {
	Send(ctx context.Context, url string, payload []byte) error
}

// Executor pulls pending jobs from the store and executes them via an outbound sender.
type Executor struct {
	store    *Store
	sender   OutboundSender
	interval time.Duration
}

// NewExecutor constructs an Executor that polls at the given interval.
func NewExecutor(store *Store, sender OutboundSender, interval time.Duration) *Executor {
	return &Executor{
		store:    store,
		sender:   sender,
		interval: interval,
	}
}

// RunLoop polls for jobs until ctx is canceled.
func (e *Executor) RunLoop(ctx context.Context) {
	t := time.NewTicker(e.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("executor: stop:", ctx.Err())
			return
		default:
			e.processOne(ctx)
		}

		select {
		case <-ctx.Done():
			log.Println("executor: stop:", ctx.Err())
			return
		case <-t.C:
		}
	}
}

// processOne fetches and executes the next pending job if available.
func (e *Executor) processOne(ctx context.Context) {
	job, err := e.store.FetchNextPendingJob(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		log.Println("executor: fetch job:", err)
		return
	}

	wf, err := e.store.GetWorkflow(ctx, job.WorkflowID)
	if err != nil {
		log.Printf("executor: workflow %d missing for job %d: %v", job.WorkflowID, job.ID, err)
		_ = e.store.MarkJobFailed(ctx, job.ID, "workflow missing")
		return
	}

	started := time.Now()
	_ = e.store.UpdateRun(ctx, job.RunID, RunUpdate{
		Status:    RunStatusRunning,
		StartedAt: &started,
	})

	payload := job.Payload
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}

	actionCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := e.sender.Send(actionCtx, wf.ActionURL, payload); err != nil {
		log.Printf("executor: job %d failed: %v", job.ID, err)
		_ = e.store.MarkJobFailed(ctx, job.ID, err.Error())
		failed := time.Now()
		msg := err.Error()
		_ = e.store.UpdateRun(ctx, job.RunID, RunUpdate{
			Status:  RunStatusFailed,
			EndedAt: &failed,
			Error:   &msg,
		})
		return
	}

	if err := e.store.MarkJobSuccess(ctx, job.ID); err != nil {
		log.Printf("executor: mark success job %d: %v", job.ID, err)
	}
	ended := time.Now()
	_ = e.store.UpdateRun(ctx, job.RunID, RunUpdate{
		Status:  RunStatusSucceeded,
		EndedAt: &ended,
	})
	log.Printf("executor: job %d succeeded (workflow %d)", job.ID, job.WorkflowID)
}

// DecodePayload helper for handlers to decode job payload into a typed struct.
func DecodePayload[T any](payload json.RawMessage, target *T) error {
	if len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, target)
}
