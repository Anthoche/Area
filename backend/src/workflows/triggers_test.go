package workflows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTriggererEnqueueRun_Success(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	now := time.Now()

	// Mock CreateRun
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "runs" \("created_at","updated_at","deleted_at","workflow_id","status","started_at","ended_at","error"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(5), RunStatusPending, nil, nil, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(10)))
	mock.ExpectCommit()

	// Mock CreateJob
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "jobs" \("created_at","updated_at","deleted_at","workflow_id","run_id","payload","status","error","started_at","ended_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(5), uint(10), []byte(`{"foo":"bar"}`), JobStatusPending, "", nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(uint(3), now, now))
	mock.ExpectCommit()

	triggerer := NewTriggerer(store)
	run, err := triggerer.EnqueueRun(context.Background(), 5, map[string]any{"foo": "bar"})
	if err != nil {
		t.Fatalf("EnqueueRun error: %v", err)
	}
	if run.ID != 10 || run.WorkflowID != 5 {
		t.Fatalf("unexpected run: %+v", run)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTriggererEnqueueRun_EncodeError(t *testing.T) {
	store, _, cleanup := setupMockStore(t)
	defer cleanup()

	triggerer := NewTriggerer(store)
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{"ch": make(chan int)}); err == nil {
		t.Fatalf("expected encode error")
	}
}

func TestTriggererEnqueueRun_CreateRunError(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "runs" \("created_at","updated_at","deleted_at","workflow_id","status","started_at","ended_at","error"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(1), RunStatusPending, nil, nil, "").
		WillReturnError(errors.New("insert fail"))
	mock.ExpectRollback()

	triggerer := NewTriggerer(store)
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{}); err == nil {
		t.Fatalf("expected error when CreateRun fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTriggererEnqueueRun_CreateJobError(t *testing.T) {
	store, mock, cleanup := setupMockStore(t)
	defer cleanup()

	// Mock CreateRun success
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "runs" \("created_at","updated_at","deleted_at","workflow_id","status","started_at","ended_at","error"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(1), RunStatusPending, nil, nil, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(2)))
	mock.ExpectCommit()

	// Mock CreateJob failure
	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "jobs" \("created_at","updated_at","deleted_at","workflow_id","run_id","payload","status","error","started_at","ended_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(1), uint(2), []byte(`{}`), JobStatusPending, "", nil, nil).
		WillReturnError(errors.New("job insert fail"))
	mock.ExpectRollback()

	triggerer := NewTriggerer(store)
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{}); err == nil {
		t.Fatalf("expected error when CreateJob fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTriggererEnqueueRun_NilStore(t *testing.T) {
	triggerer := NewTriggerer(nil)
	_, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{})
	if err == nil {
		t.Fatalf("expected error when store is nil")
	}
}
