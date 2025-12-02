package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	db = mockDB
	cleanup := func() {
		mockDB.Close()
	}
	return mock, cleanup
}

func TestGetUsers(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "firstname", "lastname", "email", "password_hash"}).
		AddRow(int64(1), "Ada", "Lovelace", "ada@example.com", "hash1").
		AddRow(int64(2), "Alan", "Turing", "alan@example.com", "hash2")
	mock.ExpectQuery("SELECT id, firstname, lastname, email, password_hash FROM users").
		WillReturnRows(rows)

	users, err := GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, firstname, lastname, email, password_hash FROM users WHERE id = \\$1").
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	if _, err := GetUserByID(context.Background(), 99); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Ada", "Lovelace", "ada@example.com", "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	id, err := CreateUser(context.Background(), "Ada", "Lovelace", "ada@example.com", "hash")
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	if id != 10 {
		t.Fatalf("expected id 10, got %d", id)
	}
}

func TestDeleteUser(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := DeleteUser(context.Background(), 5); err != nil {
		t.Fatalf("DeleteUser error: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectExec("UPDATE users SET firstname = \\$1, lastname = \\$2, email = \\$3, password_hash = \\$4 WHERE id = \\$5").
		WithArgs("Grace", "Hopper", "grace@example.com", "hash", int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := UpdateUser(context.Background(), 7, "Grace", "Hopper", "grace@example.com", "hash"); err != nil {
		t.Fatalf("UpdateUser error: %v", err)
	}
}
