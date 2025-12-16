package database

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupGoogleTokenMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

	// Replace global db for testing
	originalDB := db
	db = gormDB

	cleanup := func() {
		db = originalDB
		mockDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestInsertGoogleToken(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	now := time.Now()
	userID := int64(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "google_tokens" \("created_at","updated_at","deleted_at","user_id","access_token","refresh_token","expiry"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, &userID, "acc", "ref", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	id, err := InsertGoogleToken(&userID, "acc", "ref", now)
	if err != nil {
		t.Fatalf("InsertGoogleToken error: %v", err)
	}
	if id != 5 {
		t.Fatalf("expected id 5, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertGoogleToken_NilUserID(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "google_tokens" \("created_at","updated_at","deleted_at","user_id","access_token","refresh_token","expiry"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, "acc", "ref", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	id, err := InsertGoogleToken(nil, "acc", "ref", now)
	if err != nil {
		t.Fatalf("InsertGoogleToken error: %v", err)
	}
	if id != 10 {
		t.Fatalf("expected id 10, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGoogleToken(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	tokenID := int64(3)
	userID := int64(1)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "access_token", "refresh_token", "expiry"}).
		AddRow(tokenID, now, now, nil, userID, "access123", "refresh456", now.Add(time.Hour))

	mock.ExpectQuery(`^SELECT \* FROM "google_tokens" WHERE id = \$1 AND "google_tokens"\."deleted_at" IS NULL ORDER BY "google_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, sqlmock.AnyArg()).
		WillReturnRows(rows)

	token, err := GetGoogleToken(tokenID)
	if err != nil {
		t.Fatalf("GetGoogleToken error: %v", err)
	}
	if token.AccessToken != "access123" {
		t.Fatalf("expected access token 'access123', got %s", token.AccessToken)
	}
	if token.RefreshToken != "refresh456" {
		t.Fatalf("expected refresh token 'refresh456', got %s", token.RefreshToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGoogleToken_NotFound(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	tokenID := int64(999)

	mock.ExpectQuery(`^SELECT \* FROM "google_tokens" WHERE id = \$1 AND "google_tokens"\."deleted_at" IS NULL ORDER BY "google_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := GetGoogleToken(tokenID)
	if err == nil {
		t.Fatalf("expected error for missing token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGoogleTokenForUser(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	tokenID := int64(5)
	userID := int64(2)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "access_token", "refresh_token", "expiry"}).
		AddRow(tokenID, now, now, nil, userID, "access789", "refresh101", now.Add(time.Hour))

	mock.ExpectQuery(`^SELECT \* FROM "google_tokens" WHERE \(id = \$1 AND user_id = \$2\) AND "google_tokens"\."deleted_at" IS NULL ORDER BY "google_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, userID, sqlmock.AnyArg()).
		WillReturnRows(rows)

	token, err := GetGoogleTokenForUser(tokenID, userID)
	if err != nil {
		t.Fatalf("GetGoogleTokenForUser error: %v", err)
	}
	if token.AccessToken != "access789" {
		t.Fatalf("expected access token 'access789', got %s", token.AccessToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateGoogleToken(t *testing.T) {
	_, mock, cleanup := setupGoogleTokenMockDB(t)
	defer cleanup()

	tokenID := int64(1)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "google_tokens" SET "updated_at"=\$1,"access_token"=\$2,"refresh_token"=\$3,"expiry"=\$4 WHERE id = \$5 AND "google_tokens"\."deleted_at" IS NULL$`).
		WithArgs(sqlmock.AnyArg(), "new_access", "new_refresh", now, tokenID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := UpdateGoogleToken(tokenID, "new_access", "new_refresh", now); err != nil {
		t.Fatalf("UpdateGoogleToken error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
