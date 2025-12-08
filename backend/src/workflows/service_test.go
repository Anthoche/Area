package workflows

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock, func()) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return &Store{db: mockDB}, mock, func() { mockDB.Close() }
}

func TestServiceTrigger_Success(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	// GetWorkflow query
	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows WHERE id = \\$1").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "next_run_at", "created_at"}).
			AddRow(int64(2), "wf", "manual", []byte(`{}`), "https://example.com", sql.NullTime{}, now))

	// CreateRun insert
	mock.ExpectQuery("INSERT INTO workflow_runs").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))

	// CreateJob insert
	mock.ExpectQuery("INSERT INTO jobs").
		WithArgs(int64(2), int64(7), []byte(`{"k":"v"}`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(9), now, now))

	triggerer := NewTriggerer(store)
	svc := NewService(store, triggerer)

	if _, err := svc.Trigger(context.Background(), 2, map[string]any{"k": "v"}); err != nil {
		t.Fatalf("Trigger error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceTrigger_NotFound(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows WHERE id = \\$1").
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	svc := NewService(store, NewTriggerer(store))
	if _, err := svc.Trigger(context.Background(), 99, nil); err != ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound, got %v", err)
	}
}

func TestServiceCreateWorkflow_ManualDefaultsConfig(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO workflows").
		WithArgs("name", "manual", []byte("{}"), "https://example.com", sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows WHERE id = \\$1").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "next_run_at", "created_at"}).
			AddRow(int64(1), "name", "manual", []byte(`{}`), "https://example.com", sql.NullTime{}, now))

	svc := NewService(store, NewTriggerer(store))
	wf, err := svc.CreateWorkflow(context.Background(), "name", "manual", "https://example.com", nil)
	if err != nil {
		t.Fatalf("CreateWorkflow error: %v", err)
	}
	if string(wf.TriggerConfig) != "{}" {
		t.Fatalf("expected default trigger_config '{}', got %q", string(wf.TriggerConfig))
	}
}

func TestServiceCreateWorkflow_InvalidInterval(t *testing.T) {
	svc := NewService(&Store{}, nil)
	if _, err := svc.CreateWorkflow(context.Background(), "name", "interval", "https://example.com", []byte(`{"interval_minutes":0}`)); err == nil {
		t.Fatalf("expected error for invalid interval config")
	}
}

func TestServiceCreateWorkflow_Unsupported(t *testing.T) {
	svc := NewService(&Store{}, nil)
	if _, err := svc.CreateWorkflow(context.Background(), "name", "unknown", "url", []byte(`{}`)); err == nil {
		t.Fatalf("expected error for unsupported trigger type")
	}
}

func TestServiceListWorkflows(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "next_run_at", "created_at"}).
			AddRow(int64(1), "wf1", "manual", []byte(`{}`), "url1", sql.NullTime{}, now).
			AddRow(int64(2), "wf2", "manual", []byte(`{}`), "url2", sql.NullTime{}, now))

	svc := NewService(store, NewTriggerer(store))
	items, err := svc.ListWorkflows(context.Background())
	if err != nil {
		t.Fatalf("ListWorkflows error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(items))
	}
}

func TestServiceGetWorkflow_ErrorMapping(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows WHERE id = \\$1").
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	svc := NewService(store, NewTriggerer(store))
	if _, err := svc.GetWorkflow(context.Background(), 1); err != ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound mapping, got %v", err)
	}
}

func TestServiceTriggerWebhook_NotFound(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, name, trigger_type, trigger_config, action_url, next_run_at, created_at FROM workflows WHERE trigger_type = 'webhook' AND trigger_config->>'token' = \\$1").
		WithArgs("abc").
		WillReturnError(sql.ErrNoRows)

	svc := NewService(store, NewTriggerer(store))
	if _, err := svc.TriggerWebhook(context.Background(), "abc", nil); err != ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound for missing token, got %v", err)
	}
}
