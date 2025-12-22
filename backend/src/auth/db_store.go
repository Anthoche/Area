package auth

import (
	"database/sql"
	"errors"

	"area/src/database"

	"gorm.io/gorm"
)

// DBStore implements UserStore backed by PostgreSQL.
type DBStore struct{}

// NewDBStore returns a UserStore using the shared database connection.
func NewDBStore() *DBStore {
	return &DBStore{}
}

// Create inserts a new user row and updates the passed user with its ID.
func (DBStore) Create(user *User, passwordHash string) error {
	id, err := database.CreateUser(user.FirstName, user.LastName, user.Email, passwordHash)
	if err != nil {
		if database.UserExists(user.Email) {
			return ErrUserExists
		}
		return err
	}
	user.ID = id
	return nil
}

// GetByEmail returns a user and their hashed password for the given email.
func (DBStore) GetByEmail(email string) (*User, string, error) {
	dbUser, err := database.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}
	return &User{
		ID:        int64(dbUser.ID),
		Email:     dbUser.Email,
		FirstName: dbUser.Firstname,
		LastName:  dbUser.Lastname,
	}, dbUser.PasswordHash, nil
}
