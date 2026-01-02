package database

import (
	"area/src/database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupGithubTokenMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

	originalDB := database.Db
	database.Db = gormDB

	cleanup := func() {
		database.Db = originalDB
		mockDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestInsertGithubToken(t *testing.T) {
	_, mock, cleanup := setupGithubTokenMockDB(t)
	defer cleanup()

	userID := int64(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "github_tokens" \("created_at","updated_at","deleted_at","user_id","access_token","token_type","scope"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, &userID, "acc", "bearer", "repo").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	id, err := database.InsertGithubToken(&userID, "acc", "bearer", "repo")
	if err != nil {
		t.Fatalf("InsertGithubToken error: %v", err)
	}
	if id != 7 {
		t.Fatalf("expected id 7, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertGithubToken_NilUserID(t *testing.T) {
	_, mock, cleanup := setupGithubTokenMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "github_tokens" \("created_at","updated_at","deleted_at","user_id","access_token","token_type","scope"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"$`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, "acc", "bearer", "repo").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	id, err := database.InsertGithubToken(nil, "acc", "bearer", "repo")
	if err != nil {
		t.Fatalf("InsertGithubToken error: %v", err)
	}
	if id != 9 {
		t.Fatalf("expected id 9, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGithubToken(t *testing.T) {
	_, mock, cleanup := setupGithubTokenMockDB(t)
	defer cleanup()

	tokenID := int64(3)
	userID := int64(1)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "access_token", "token_type", "scope"}).
		AddRow(tokenID, now, now, nil, userID, "access123", "bearer", "repo")

	mock.ExpectQuery(`^SELECT \* FROM "github_tokens" WHERE id = \$1 AND "github_tokens"\."deleted_at" IS NULL ORDER BY "github_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, sqlmock.AnyArg()).
		WillReturnRows(rows)

	token, err := database.GetGithubToken(tokenID)
	if err != nil {
		t.Fatalf("GetGithubToken error: %v", err)
	}
	if token.AccessToken != "access123" {
		t.Fatalf("expected access token 'access123', got %s", token.AccessToken)
	}
	if token.TokenType != "bearer" {
		t.Fatalf("expected token type 'bearer', got %s", token.TokenType)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGithubToken_NotFound(t *testing.T) {
	_, mock, cleanup := setupGithubTokenMockDB(t)
	defer cleanup()

	tokenID := int64(999)

	mock.ExpectQuery(`^SELECT \* FROM "github_tokens" WHERE id = \$1 AND "github_tokens"\."deleted_at" IS NULL ORDER BY "github_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := database.GetGithubToken(tokenID)
	if err == nil {
		t.Fatalf("expected error for missing token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGithubTokenForUser(t *testing.T) {
	_, mock, cleanup := setupGithubTokenMockDB(t)
	defer cleanup()

	tokenID := int64(5)
	userID := int64(2)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "access_token", "token_type", "scope"}).
		AddRow(tokenID, now, now, nil, userID, "access789", "bearer", "repo")

	mock.ExpectQuery(`^SELECT \* FROM "github_tokens" WHERE \(id = \$1 AND user_id = \$2\) AND "github_tokens"\."deleted_at" IS NULL ORDER BY "github_tokens"\."id" LIMIT \$[0-9]+$`).
		WithArgs(tokenID, userID, sqlmock.AnyArg()).
		WillReturnRows(rows)

	token, err := database.GetGithubTokenForUser(tokenID, userID)
	if err != nil {
		t.Fatalf("GetGithubTokenForUser error: %v", err)
	}
	if token.AccessToken != "access789" {
		t.Fatalf("expected access token 'access789', got %s", token.AccessToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
