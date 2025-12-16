package auth

import (
	"area/src/auth"
	"testing"
)

type fakeUserStore struct {
	getRespUser *auth.User
	getRespHash string
	getErr      error
	createErr   error

	createdUser *auth.User
	createdHash string
}

func (f *fakeUserStore) Create(user *auth.User, passwordHash string) error {
	f.createdUser = user
	f.createdHash = passwordHash
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = 42
	return nil
}

func (f *fakeUserStore) GetByEmail(email string) (*auth.User, string, error) {
	if f.getErr != nil {
		return nil, "", f.getErr
	}
	return f.getRespUser, f.getRespHash, nil
}

func TestServiceAuthenticate_Success(t *testing.T) {
	hash, _ := auth.HashPassword("password")
	store := &fakeUserStore{
		getRespUser: &auth.User{Email: "user@example.com"},
		getRespHash: hash,
	}
	svc := auth.NewService(store)

	user, err := svc.Authenticate("user@example.com", "password")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if user == nil || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestServiceAuthenticate_Invalid(t *testing.T) {
	store := &fakeUserStore{getErr: auth.ErrInvalidCredentials}
	svc := auth.NewService(store)

	if _, err := svc.Authenticate("user@example.com", "password"); err != auth.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestServiceRegister_Success(t *testing.T) {
	store := &fakeUserStore{}
	svc := auth.NewService(store)

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
}

func TestServiceRegister_UserExists(t *testing.T) {
	store := &fakeUserStore{createErr: auth.ErrUserExists}
	svc := auth.NewService(store)

	if _, err := svc.Register("user@example.com", "password", "Ada", "Lovelace"); err != auth.ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}
