package auth

import (
	"context"
	"database/sql"
	"errors"

	"area/src/database"
	"github.com/lib/pq"
)

// DBStore implements UserStore backed by PostgreSQL.
type DBStore struct{}

// NewDBStore returns a UserStore using the shared database connection.
func NewDBStore() *DBStore {
	return &DBStore{}
}

// Create inserts a new user row and updates the passed user with its ID.
func (DBStore) Create(ctx context.Context, user *User, passwordHash string) error {
	id, err := database.CreateUser(ctx, user.FirstName, user.LastName, user.Email, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return err
	}
	user.ID = id
	return nil
}

// GetByEmail returns a user and their hashed password for the given email.
func (DBStore) GetByEmail(ctx context.Context, email string) (*User, string, error) {
	dbUser, err := database.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}
	return &User{
		ID:        dbUser.Id,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
	}, dbUser.PasswordHash, nil
}

// isUniqueViolation checks if the error corresponds to a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}
