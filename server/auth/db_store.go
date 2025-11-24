package auth

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"area/server/database"
	"github.com/lib/pq"
)

// DBStore implements UserStore backed by PostgreSQL.
type DBStore struct{}

// NewDBStore returns a UserStore using the shared database connection.
func NewDBStore() *DBStore {
	return &DBStore{}
}

func (DBStore) Create(ctx context.Context, user *User, passwordHash string) error {
	id, err := database.CreateUser(ctx, user.FirstName, user.LastName, user.Email, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return err
	}
	user.ID = formatID(id)
	return nil
}

func (DBStore) GetByEmail(ctx context.Context, email string) (*User, string, error) {
	dbUser, err := database.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}
	return &User{
		ID:        formatID(dbUser.Id),
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
	}, dbUser.PasswordHash, nil
}

func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}

func formatID(id int64) string {
	return "user-" + strconv.FormatInt(id, 10)
}
