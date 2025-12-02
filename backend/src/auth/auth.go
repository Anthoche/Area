package auth

import (
	"context"
	"errors"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

var bcryptCost int

func init() {
	costStr := os.Getenv("BCRYPT_COST")
	cost, err := strconv.Atoi(costStr)
	if err != nil {
		bcryptCost = bcrypt.DefaultCost
	} else {
		bcryptCost = cost
	}
}

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt hash against a plaintext password.
func CheckPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

// UserStore defines the storage operations needed for auth.
type UserStore interface {
	Create(ctx context.Context, user *User, passwordHash string) error
	GetByEmail(ctx context.Context, email string) (*User, string, error)
}

// Service handles authentication and user creation against a user store.
type Service struct {
	store UserStore
}

// NewService wires the auth service with a backing store (DB, memory, etc.).
func NewService(store UserStore) *Service {
	return &Service{store: store}
}

// Authenticate verifies email/password against the stored hash.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, hashed, err := s.store.GetByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return nil, ErrInvalidCredentials
	case err != nil:
		return nil, err
	}
	if err := CheckPassword(hashed, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// Register creates a new user with a hashed password in the store.
func (s *Service) Register(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	if err := s.store.Create(ctx, user, hashed); err != nil {
		return nil, err
	}
	return user, nil
}
