package workflows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTriggererEnqueueRun_Success(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO workflow_runs").
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	mock.ExpectQuery("INSERT INTO jobs").
		WithArgs(int64(5), int64(10), []byte(`{"foo":"bar"}`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(3), now, now))

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
	triggerer := NewTriggerer(&Store{})
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{"ch": make(chan int)}); err == nil {
		t.Fatalf("expected encode error")
	}
}

func TestTriggererEnqueueRun_CreateRunError(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO workflow_runs").
		WithArgs(int64(1)).
		WillReturnError(errors.New("insert fail"))

	triggerer := NewTriggerer(store)
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{}); err == nil {
		t.Fatalf("expected error when CreateRun fails")
	}
}

func TestTriggererEnqueueRun_CreateJobError(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO workflow_runs").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))

	mock.ExpectQuery("INSERT INTO jobs").
		WithArgs(int64(1), int64(2), []byte(`{}`)).
		WillReturnError(errors.New("job insert fail"))

	triggerer := NewTriggerer(store)
	if _, err := triggerer.EnqueueRun(context.Background(), 1, map[string]any{}); err == nil {
		t.Fatalf("expected error when CreateJob fails")
	}
}
