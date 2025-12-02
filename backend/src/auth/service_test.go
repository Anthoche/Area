package auth

import (
	"context"
	"testing"
)

type fakeUserStore struct {
	getRespUser *User
	getRespHash string
	getErr      error
	createErr   error

	createdUser *User
	createdHash string
}

func (f *fakeUserStore) Create(_ context.Context, user *User, passwordHash string) error {
	f.createdUser = user
	f.createdHash = passwordHash
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = 42
	return nil
}

func (f *fakeUserStore) GetByEmail(_ context.Context, email string) (*User, string, error) {
	if f.getErr != nil {
		return nil, "", f.getErr
	}
	return f.getRespUser, f.getRespHash, nil
}

func TestServiceAuthenticate_Success(t *testing.T) {
	hash, _ := HashPassword("password")
	store := &fakeUserStore{
		getRespUser: &User{Email: "user@example.com"},
		getRespHash: hash,
	}
	svc := NewService(store)

	user, err := svc.Authenticate(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestServiceAuthenticate_Invalid(t *testing.T) {
	store := &fakeUserStore{getErr: ErrInvalidCredentials}
	svc := NewService(store)

	if _, err := svc.Authenticate(context.Background(), "user@example.com", "password"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestServiceRegister_Success(t *testing.T) {
	store := &fakeUserStore{}
	svc := NewService(store)

	user, err := svc.Register(context.Background(), "user@example.com", "password", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if user.Email != "user@example.com" || store.createdUser == nil {
		t.Fatalf("user not created properly: %+v", user)
	}
	if store.createdHash == "" || store.createdHash == "password" {
		t.Fatalf("hash should be stored and different from password")
	}
}

func TestServiceRegister_UserExists(t *testing.T) {
	store := &fakeUserStore{createErr: ErrUserExists}
	svc := NewService(store)

	if _, err := svc.Register(context.Background(), "user@example.com", "password", "Ada", "Lovelace"); err != ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}
