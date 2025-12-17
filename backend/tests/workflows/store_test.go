package workflows

import (
	"area/src/workflows"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockStore(t *testing.T) (*workflows.Store, sqlmock.Sqlmock, func()) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	store := workflows.NewStore(gormDB)

	cleanup := func() {
		mockDB.Close()
	}

	return store, mock, cleanup
}

func TestIntervalConfigFromJSON(t *testing.T) {
	raw := []byte(`{"interval_minutes":5,"payload":{"foo":"bar"}}`)
	cfg, err := workflows.IntervalConfigFromJSON(raw)
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
	_, err := workflows.IntervalConfigFromJSON([]byte(`not json`))
	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestSetEnabled_EnableIntervalSetsNextRun(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	now := time.Now()
	triggerCfg := []byte(`{"interval_minutes":2}`)

	// Mock GetWorkflow call
	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(id = \$1 AND user_id = \$2\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs(int64(1), int64(99), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at", "user_id"}).
			AddRow(uint(1), "wf", "interval", triggerCfg, "url", false, nil, now, 99))

	// Mock Updates call
	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "workflows" SET "enabled"=\$1,"next_run_at"=\$2,"updated_at"=\$3 WHERE \(id = \$4 AND user_id = \$5\) AND "workflows"\."deleted_at" IS NULL$`).
		WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), uint(1), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.SetEnabledForUser(context.Background(), 1, 99, true, now); err != nil {
		t.Fatalf("SetEnabledForUser() enable interval: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetEnabled_Disable(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "workflows" SET "enabled"=\$1,"next_run_at"=\$2,"updated_at"=\$3 WHERE \(id = \$4 AND user_id = \$5\) AND "workflows"\."deleted_at" IS NULL$`).
		WithArgs(false, nil, sqlmock.AnyArg(), uint(2), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.SetEnabledForUser(context.Background(), 2, 99, false, time.Now()); err != nil {
		t.Fatalf("SetEnabledForUser() disable: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetEnabled_NotFound(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "workflows" SET "enabled"=\$1,"next_run_at"=\$2,"updated_at"=\$3 WHERE \(id = \$4 AND user_id = \$5\) AND "workflows"\."deleted_at" IS NULL$`).
		WithArgs(false, nil, sqlmock.AnyArg(), uint(42), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := store.SetEnabledForUser(context.Background(), 42, 99, false, time.Now())
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestGetWorkflow_DisablesNonManualWithoutNextRun(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE "workflows"\."id" = \$1 AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs(uint(3), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at", "user_id"}).
			AddRow(uint(3), "wf", "interval", []byte(`{"interval_minutes":5}`), "url", true, nil, now, 99))

	wf, err := store.GetWorkflow(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetWorkflow: %v", err)
	}
	if wf.Enabled {
		t.Fatalf("expected enabled=false when no next_run_at, got true")
	}
}

func TestCreateWorkflow(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	triggerCfg := []byte(`{"interval_minutes":10}`)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "workflows" \("created_at","updated_at","deleted_at","user_id","name","trigger_type","trigger_config","action_url","enabled","next_run_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(99), "test-workflow", "interval", triggerCfg, "http://example.com/action", false, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(1)))
	mock.ExpectCommit()

	wf, err := store.CreateWorkflow(context.Background(), 99, "test-workflow", "interval", "http://example.com/action", triggerCfg)
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if wf.ID != 1 {
		t.Fatalf("expected id 1, got %d", wf.ID)
	}
	if wf.Name != "test-workflow" {
		t.Fatalf("expected name test-workflow, got %s", wf.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListWorkflows(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE user_id = \$1 AND "workflows"\."deleted_at" IS NULL ORDER BY created_at DESC$`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at", "user_id"}).
			AddRow(uint(1), "wf1", "manual", []byte(`{}`), "url1", true, nil, now, 99).
			AddRow(uint(2), "wf2", "interval", []byte(`{"interval_minutes":5}`), "url2", true, nil, now, 99))

	workflows, err := store.ListWorkflows(context.Background(), 99)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(workflows))
	}

	// First workflow (manual) should stay enabled
	if !workflows[0].Enabled {
		t.Fatalf("expected manual workflow to stay enabled")
	}

	// Second workflow (interval without next_run_at) should be disabled
	if workflows[1].Enabled {
		t.Fatalf("expected interval workflow without next_run_at to be disabled")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteWorkflow(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "workflows" SET "deleted_at"=\$1 WHERE "workflows"\."id" = \$2 AND "workflows"\."deleted_at" IS NULL$`).
		WithArgs(sqlmock.AnyArg(), uint(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.DeleteWorkflow(context.Background(), 5); err != nil {
		t.Fatalf("DeleteWorkflow: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateRun(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "workflow_runs" \("created_at","updated_at","deleted_at","workflow_id","status","started_at","ended_at","error"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(1), workflows.RunStatusPending, nil, nil, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(100)))
	mock.ExpectCommit()

	run, err := store.CreateRun(context.Background(), 1)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if run.ID != 100 {
		t.Fatalf("expected id 100, got %d", run.ID)
	}
	if run.Status != workflows.RunStatusPending {
		t.Fatalf("expected status pending, got %s", run.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateJob(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	payload := []byte(`{"key":"value"}`)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "jobs" \("created_at","updated_at","deleted_at","workflow_id","run_id","payload","status","error","started_at","ended_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(1), uint(10), payload, workflows.JobStatusPending, "", nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(200)))
	mock.ExpectCommit()

	job, err := store.CreateJob(context.Background(), 1, 10, payload)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID != 200 {
		t.Fatalf("expected id 200, got %d", job.ID)
	}
	if job.Status != workflows.JobStatusPending {
		t.Fatalf("expected status pending, got %s", job.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMarkJobSuccess(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "jobs" SET "ended_at"=\$1,"error"=\$2,"status"=\$3,"updated_at"=\$4 WHERE id = \$5 AND "jobs"\."deleted_at" IS NULL$`).
		WithArgs(sqlmock.AnyArg(), "", workflows.JobStatusSucceeded, sqlmock.AnyArg(), uint(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.MarkJobSuccess(context.Background(), 50); err != nil {
		t.Fatalf("MarkJobSuccess: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMarkJobFailed(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "jobs" SET "ended_at"=\$1,"error"=\$2,"status"=\$3,"updated_at"=\$4 WHERE id = \$5 AND "jobs"\."deleted_at" IS NULL$`).
		WithArgs(sqlmock.AnyArg(), "test error", workflows.JobStatusFailed, sqlmock.AnyArg(), uint(51)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.MarkJobFailed(context.Background(), 51, "test error"); err != nil {
		t.Fatalf("MarkJobFailed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindWorkflowByToken(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	now := time.Now()
	triggerCfg := []byte(`{"token":"secret123"}`)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(trigger_type = \$1 AND enabled = \$2 AND trigger_config->>'token' = \$3\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs("webhook", true, "secret123", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "trigger_type", "trigger_config", "action_url", "enabled", "next_run_at", "created_at", "user_id"}).
			AddRow(uint(7), "webhook-wf", "webhook", triggerCfg, "url", true, nil, now, 99))

	wf, err := store.FindWorkflowByToken(context.Background(), "secret123")
	if err != nil {
		t.Fatalf("FindWorkflowByToken: %v", err)
	}
	if wf.ID != 7 {
		t.Fatalf("expected id 7, got %d", wf.ID)
	}
	if wf.Name != "webhook-wf" {
		t.Fatalf("expected name webhook-wf, got %s", wf.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
