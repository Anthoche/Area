package auth

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory user store for dev/testing.
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]record
}

type record struct {
	user         *User
	passwordHash string
}

// NewMemoryStore initializes an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users: make(map[string]record),
	}
}

func (m *MemoryStore) Create(_ context.Context, user *User, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Email]; exists {
		return ErrUserExists
	}

	if user.ID == "" {
		user.ID = "user-" + user.Email
	}
	m.users[user.Email] = record{user: user, passwordHash: passwordHash}
	return nil
}

func (m *MemoryStore) GetByEmail(_ context.Context, email string) (*User, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, ok := m.users[email]
	if !ok {
		return nil, "", ErrInvalidCredentials
	}
	return rec.user, rec.passwordHash, nil
}
