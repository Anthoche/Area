package workflows

import (
	"area/src/workflows"
	"context"
	"time"

	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

	cleanup := func() {
		mockDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestServiceTrigger_Success(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	ctx := workflows.WithUserID(context.Background(), 99)

	store := workflows.NewStore(gormDB)

	// Store.GetWorkflow -> gorm First()
	rowsWF := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "user_id",
		"name", "trigger_type", "trigger_config", "action_url",
		"enabled", "next_run_at",
	}).AddRow(
		2, time.Now(), time.Now(), nil, 99,
		"wf", "manual", []byte(`{}`), "https://example.com",
		true, nil,
	)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(id = \$1 AND user_id = \$2\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs(int64(2), int64(99), sqlmock.AnyArg()).
		WillReturnRows(rowsWF)

	// Triggerer.EnqueueRun -> Store.CreateRun (gorm Create => begin/insert/commit)
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "workflow_runs" \("created_at","updated_at","deleted_at","workflow_id","status","started_at","ended_at","error"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(2), workflows.RunStatusPending, nil, nil, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	// Triggerer.EnqueueRun -> Store.CreateJob (gorm Create => begin/insert/commit)
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "jobs" \("created_at","updated_at","deleted_at","workflow_id","run_id","payload","status","error","started_at","ended_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(2), uint(7), []byte(`{"k":"v"}`), workflows.JobStatusPending, "", nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	triggerer := workflows.NewTriggerer(store)
	svc := workflows.NewService(store, triggerer)

	if _, err := svc.Trigger(ctx, 2, map[string]any{"k": "v"}); err != nil {
		t.Fatalf("Trigger error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceTrigger_NotFound(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	ctx := workflows.WithUserID(context.Background(), 99)

	store := workflows.NewStore(gormDB)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(id = \$1 AND user_id = \$2\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs(int64(99), int64(99), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	svc := workflows.NewService(store, workflows.NewTriggerer(store))
	if _, err := svc.Trigger(ctx, 99, nil); err != workflows.ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceCreateWorkflow_ManualDefaultsConfig(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	ctx := workflows.WithUserID(context.Background(), 99)

	store := workflows.NewStore(gormDB)

	svc := workflows.NewService(store, workflows.NewTriggerer(store))
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "workflows" \("created_at","updated_at","deleted_at","user_id","name","trigger_type","trigger_config","action_url","enabled","next_run_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(99), "name", "manual", []byte(`{}`), "https://example.com", true, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	wf, err := svc.CreateWorkflow(ctx, "name", "manual", "https://example.com", nil)
	if err != nil {
		t.Fatalf("CreateWorkflow error: %v", err)
	}
	if string(wf.TriggerConfig) != "{}" {
		t.Fatalf("expected default trigger_config '{}', got %q", string(wf.TriggerConfig))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceCreateWorkflow_InvalidInterval(t *testing.T) {
	svc := workflows.NewService(&workflows.Store{}, nil)
	if _, err := svc.CreateWorkflow(context.Background(), "name", "interval", "https://example.com", []byte(`{"interval_minutes":0}`)); err == nil {
		t.Fatalf("expected error for invalid interval config")
	}
}

func TestServiceCreateWorkflow_Unsupported(t *testing.T) {
	svc := workflows.NewService(&workflows.Store{}, nil)
	if _, err := svc.CreateWorkflow(context.Background(), "name", "unknown", "url", []byte(`{}`)); err == nil {
		t.Fatalf("expected error for unsupported trigger type")
	}
}

func TestServiceListWorkflows(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	ctx := workflows.WithUserID(context.Background(), 99)

	store := workflows.NewStore(gormDB)

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "user_id",
		"name", "trigger_type", "trigger_config", "action_url",
		"enabled", "next_run_at",
	}).
		AddRow(1, time.Now(), time.Now(), nil, 99, "wf1", "manual", []byte(`{}`), "url1", true, nil).
		AddRow(2, time.Now(), time.Now(), nil, 99, "wf2", "manual", []byte(`{}`), "url2", true, nil)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE user_id = \$1 AND "workflows"\."deleted_at" IS NULL ORDER BY created_at DESC$`).
		WithArgs(int64(99)).
		WillReturnRows(rows)

	svc := workflows.NewService(store, workflows.NewTriggerer(store))
	items, err := svc.ListWorkflows(ctx)
	if err != nil {
		t.Fatalf("ListWorkflows error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceGetWorkflow_ErrorMapping(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	ctx := workflows.WithUserID(context.Background(), 99)

	store := workflows.NewStore(gormDB)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(id = \$1 AND user_id = \$2\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs(int64(1), int64(99), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	svc := workflows.NewService(store, workflows.NewTriggerer(store))
	if _, err := svc.GetWorkflow(ctx, 1); err != workflows.ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound mapping, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceTriggerWebhook_NotFound(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	store := workflows.NewStore(gormDB)

	mock.ExpectQuery(`^SELECT \* FROM "workflows" WHERE \(trigger_type = \$1 AND enabled = \$2 AND trigger_config->>'token' = \$3\) AND "workflows"\."deleted_at" IS NULL ORDER BY "workflows"\."id" LIMIT \$[0-9]+$`).
		WithArgs("webhook", true, "abc", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	svc := workflows.NewService(store, workflows.NewTriggerer(store))
	if _, err := svc.TriggerWebhook(context.Background(), "abc", nil); err != workflows.ErrWorkflowNotFound {
		t.Fatalf("expected ErrWorkflowNotFound for missing token, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
