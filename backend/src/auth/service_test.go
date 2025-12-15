package auth

import (
	"area/src/database"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDBForService(t *testing.T) (sqlmock.Sqlmock, func()) {
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

	originalDB := database.GetDB()
	database.SetDBForTesting(gormDB)

	cleanup := func() {
		database.SetDBForTesting(originalDB)
		mockDB.Close()
	}
	return mock, cleanup
}

type fakeUserStore struct {
	getRespUser *User
	getRespHash string
	getErr      error
	createErr   error

	createdUser *User
	createdHash string
}

func (f *fakeUserStore) Create(user *User, passwordHash string) error {
	f.createdUser = user
	f.createdHash = passwordHash
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = 42
	return nil
}

func (f *fakeUserStore) GetByEmail(email string) (*User, string, error) {
	if f.getErr != nil {
		return nil, "", f.getErr
	}
	return f.getRespUser, f.getRespHash, nil
}

func TestServiceAuthenticate_Success(t *testing.T) {
	mock, cleanup := setupMockDBForService(t)
	defer cleanup()

	// Mock UserExists call - user exists
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	hash, _ := HashPassword("password")
	store := &fakeUserStore{
		getRespUser: &User{Email: "user@example.com"},
		getRespHash: hash,
	}
	svc := NewService(store)

	user, err := svc.Authenticate("user@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestServiceAuthenticate_Invalid(t *testing.T) {
	mock, cleanup := setupMockDBForService(t)
	defer cleanup()

	// Mock UserExists call - user doesn't exist
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	store := &fakeUserStore{getErr: ErrInvalidCredentials}
	svc := NewService(store)

	if _, err := svc.Authenticate("user@example.com", "password"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestServiceRegister_Success(t *testing.T) {
	mock, cleanup := setupMockDBForService(t)
	defer cleanup()

	// Mock UserExists call - user doesn't exist
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	store := &fakeUserStore{}
	svc := NewService(store)

	user, err := svc.Register("user@example.com", "password", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if user.Email != "user@example.com" || store.createdUser == nil {
		t.Fatalf("user not created properly: %+v", user)
	}
	if store.createdHash == "" || store.createdHash == "password" {
		t.Fatalf("hash should be stored and different from password")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestServiceRegister_UserExists(t *testing.T) {
	mock, cleanup := setupMockDBForService(t)
	defer cleanup()

	// Mock UserExists call - user already exists
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	store := &fakeUserStore{createErr: ErrUserExists}
	svc := NewService(store)

	if _, err := svc.Register("user@example.com", "password", "Ada", "Lovelace"); err != ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
