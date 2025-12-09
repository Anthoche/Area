package auth

import (
	"context"
	"errors"
	"testing"
)

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

func (s *stubStore) Create(ctx context.Context, user *User, passwordHash string) error {
	s.lastCreate = user
	s.lastPassword = passwordHash
	return s.createErr
}
func (s *stubStore) GetByEmail(ctx context.Context, email string) (*User, string, error) {
	return s.user, s.hash, s.getErr
}

func TestAuthenticate_Success(t *testing.T) {
	pw := "mypassword"
	hash, _ := HashPassword(pw)
	store := &stubStore{
		user: &User{ID: 1, Email: "a@b.com"},
		hash: hash,
	}
	svc := NewService(store)

	user, err := svc.Authenticate(context.Background(), "a@b.com", pw)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "a@b.com" {
		t.Fatalf("unexpected user returned")
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	store := &stubStore{getErr: ErrInvalidCredentials}
	svc := NewService(store)
	if _, err := svc.Authenticate(context.Background(), "a@b.com", "pw"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_StoreError(t *testing.T) {
	store := &stubStore{getErr: errors.New("db down")}
	svc := NewService(store)
	if _, err := svc.Authenticate(context.Background(), "a@b.com", "pw"); err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected store error, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	store := &stubStore{}
	svc := NewService(store)
	user, err := svc.Register(context.Background(), "a@b.com", "pw", "f", "l")
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
	svc := NewService(store)
	if _, err := svc.Register(context.Background(), "a@b.com", "pw", "f", "l"); err == nil {
		t.Fatalf("expected create error")
	}
}
