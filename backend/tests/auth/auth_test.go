package auth

import (
	"area/src/auth"
	"errors"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	const password = "supersecret"

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hashed == "" || hashed == password {
		t.Fatalf("hashed password should be non-empty and different from input")
	}
	if err := auth.CheckPassword(hashed, password); err != nil {
		t.Fatalf("CheckPassword should succeed: %v", err)
	}
	if err := auth.CheckPassword(hashed, "wrong"); err == nil {
		t.Fatalf("CheckPassword should fail for wrong password")
	}
}

type stubStore struct {
	user         *auth.User
	hash         string
	getErr       error
	createErr    error
	lastCreate   *auth.User
	lastPassword string
}

func (s *stubStore) Create(user *auth.User, passwordHash string) error {
	s.lastCreate = user
	s.lastPassword = passwordHash
	return s.createErr
}
func (s *stubStore) GetByEmail(email string) (*auth.User, string, error) {
	return s.user, s.hash, s.getErr
}

func TestAuthenticate_Success(t *testing.T) {
	pw := "mypassword"
	hash, _ := auth.HashPassword(pw)
	store := &stubStore{
		user: &auth.User{ID: 1, Email: "a@b.com"},
		hash: hash,
	}
	svc := auth.NewService(store)

	user, err := svc.Authenticate("a@b.com", pw)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "a@b.com" {
		t.Fatalf("unexpected user returned")
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	store := &stubStore{getErr: auth.ErrInvalidCredentials}
	svc := auth.NewService(store)
	if _, err := svc.Authenticate("a@b.com", "pw"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_StoreError(t *testing.T) {
	store := &stubStore{getErr: errors.New("db down")}
	svc := auth.NewService(store)
	if _, err := svc.Authenticate("a@b.com", "pw"); err == nil || errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected store error, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	store := &stubStore{}
	svc := auth.NewService(store)
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
}

func TestRegister_CreateError(t *testing.T) {
	store := &stubStore{createErr: errors.New("insert fail")}
	svc := auth.NewService(store)
	if _, err := svc.Register("a@b.com", "pw", "f", "l"); err == nil {
		t.Fatalf("expected create error")
	}
}
