package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInsertGoogleToken(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	db = mockDB

	now := time.Now()
	mock.ExpectQuery("INSERT INTO google_tokens").
		WithArgs(sqlmock.AnyArg(), "acc", "ref", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))

	id, err := InsertGoogleToken(context.Background(), nil, "acc", "ref", now)
	if err != nil {
		t.Fatalf("InsertGoogleToken error: %v", err)
	}
	if id != 5 {
		t.Fatalf("expected id 5, got %d", id)
	}
}

func TestGetGoogleToken_NotFound(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	db = mockDB

	mock.ExpectQuery("SELECT id, user_id, access_token, refresh_token, expiry, created_at FROM google_tokens WHERE id = \\$1").
		WithArgs(int64(3)).
		WillReturnError(sql.ErrNoRows)

	if _, err := GetGoogleToken(context.Background(), 3); err == nil {
		t.Fatalf("expected error for missing token")
	}
}

func TestUpdateGoogleToken(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	db = mockDB

	now := time.Now()
	mock.ExpectExec("UPDATE google_tokens SET access_token = \\$1, refresh_token = \\$2, expiry = \\$3 WHERE id = \\$4").
		WithArgs("acc", "ref", now, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := UpdateGoogleToken(context.Background(), 1, "acc", "ref", now); err != nil {
		t.Fatalf("UpdateGoogleToken error: %v", err)
	}
}
