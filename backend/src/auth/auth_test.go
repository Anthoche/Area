package auth

import (
	"area/src/database"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDBForAuth(t *testing.T) (sqlmock.Sqlmock, func()) {
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

func TestHashAndCheckPassword(t *testing.T) {
	const password = "supersecret"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hashed == "" || hashed == password {
		t.Fatalf("hashed password should be non-empty and different from input")
	}
	if err := CheckPassword(hashed, password); err != nil {
		t.Fatalf("CheckPassword should succeed: %v", err)
	}
	if err := CheckPassword(hashed, "wrong"); err == nil {
		t.Fatalf("CheckPassword should fail for wrong password")
	}
}

type stubStore struct {
	user         *User
	hash         string
	getErr       error
	createErr    error
	lastCreate   *User
	lastPassword string
}

func (s *stubStore) Create(user *User, passwordHash string) error {
	s.lastCreate = user
	s.lastPassword = passwordHash
	return s.createErr
}
func (s *stubStore) GetByEmail(email string) (*User, string, error) {
	return s.user, s.hash, s.getErr
}

func TestAuthenticate_Success(t *testing.T) {
	mock, cleanup := setupMockDBForAuth(t)
	defer cleanup()

	// Mock UserExists call
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	pw := "mypassword"
	hash, _ := HashPassword(pw)
	store := &stubStore{
		user: &User{ID: 1, Email: "a@b.com"},
		hash: hash,
	}
	svc := NewService(store)

	user, err := svc.Authenticate("a@b.com", pw)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "a@b.com" {
		t.Fatalf("unexpected user returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	mock, cleanup := setupMockDBForAuth(t)
	defer cleanup()

	// Mock UserExists call returning 0 (user doesn't exist)
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	store := &stubStore{getErr: ErrInvalidCredentials}
	svc := NewService(store)
	if _, err := svc.Authenticate("a@b.com", "pw"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthenticate_StoreError(t *testing.T) {
	mock, cleanup := setupMockDBForAuth(t)
	defer cleanup()

	// Mock UserExists call
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	store := &stubStore{getErr: errors.New("db down")}
	svc := NewService(store)
	if _, err := svc.Authenticate("a@b.com", "pw"); err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected store error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRegister_Success(t *testing.T) {
	mock, cleanup := setupMockDBForAuth(t)
	defer cleanup()

	// Mock UserExists call returning 0 (user doesn't exist)
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	store := &stubStore{}
	svc := NewService(store)
	user, err := svc.Register("a@b.com", "pw", "f", "l")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if user == nil || store.lastCreate == nil {
		t.Fatalf("store.Create not called or user nil")
	}
	if store.lastPassword == "" || store.lastPassword == "pw" {
		t.Fatalf("password should be hashed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRegister_CreateError(t *testing.T) {
	mock, cleanup := setupMockDBForAuth(t)
	defer cleanup()

	// Mock UserExists call returning 0 (user doesn't exist)
	mock.ExpectQuery(`^SELECT count\(\*\) FROM "users" WHERE email = \$1 AND "users"\."deleted_at" IS NULL$`).
		WithArgs("a@b.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	store := &stubStore{createErr: errors.New("insert fail")}
	svc := NewService(store)
	if _, err := svc.Register("a@b.com", "pw", "f", "l"); err == nil {
		t.Fatalf("expected create error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
