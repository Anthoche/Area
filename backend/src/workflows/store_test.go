package workflows

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"context"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIntervalConfigFromJSON(t *testing.T) {
	raw := []byte(`{"interval_minutes":5,"payload":{"foo":"bar"}}`)
	cfg, err := intervalConfigFromJSON(raw)
	if err != nil {
		t.Fatalf("intervalConfigFromJSON returned error: %v", err)
	}
	if cfg.IntervalMinutes != 5 {
		t.Fatalf("interval minutes = %d, want 5", cfg.IntervalMinutes)
	}
	if cfg.Payload["foo"] != "bar" {
		t.Fatalf("payload foo = %v, want bar", cfg.Payload["foo"])
	}
}

func TestIntervalConfigFromJSON_Invalid(t *testing.T) {
	_, err := intervalConfigFromJSON([]byte(`not json`))
	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestSetEnabled_EnableIntervalSetsNextRun(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	store := &Store{db: mockDB}

	now := time.Now()
	triggerCfg := []byte(`{"interval_minutes":2}`)
	mock.ExpectQuery("SELECT id, user_id, name, trigger_type, trigger_config, action_url, enabled, next_run_at, created_at FROM workflows WHERE id = \\$1 AND user_id = \\$2").
		WithArgs(int64(1), int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at"}).
			AddRow(int64(1), int64(99), "wf", "interval", triggerCfg, "url", false, sql.NullTime{}, now))

	mock.ExpectExec("UPDATE workflows\\s+SET enabled = TRUE, next_run_at = \\$1\\s+WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(sqlmock.AnyArg(), int64(1), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.SetEnabledForUser(context.Background(), 1, 99, true, now); err != nil {
		t.Fatalf("SetEnabled enable interval: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetEnabled_Disable(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	store := &Store{db: mockDB}

	mock.ExpectExec("UPDATE workflows\\s+SET enabled = FALSE, next_run_at = NULL\\s+WHERE id = \\$1 AND user_id = \\$2").
		WithArgs(int64(2), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.SetEnabledForUser(context.Background(), 2, 99, false, time.Now()); err != nil {
		t.Fatalf("SetEnabled disable: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetEnabled_NotFound(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	store := &Store{db: mockDB}

	mock.ExpectExec("UPDATE workflows\\s+SET enabled = FALSE, next_run_at = NULL\\s+WHERE id = \\$1 AND user_id = \\$2").
		WithArgs(int64(42), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := store.SetEnabledForUser(context.Background(), 42, 99, false, time.Now()); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetWorkflow_DisablesNonManualWithoutNextRun(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	store := &Store{db: mockDB}

	now := time.Now()
	mock.ExpectQuery("SELECT id, user_id, name, trigger_type, trigger_config, action_url, enabled, next_run_at, created_at FROM workflows WHERE id = \\$1").
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at"}).
			AddRow(int64(3), int64(99), "wf", "interval", []byte(`{"interval_minutes":5}`), "url", true, sql.NullTime{}, now))

	wf, err := store.GetWorkflow(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetWorkflow: %v", err)
	}
	if wf.Enabled {
		t.Fatalf("expected enabled=false when no next_run_at, got true")
	}
}
