package database

import (
	"context"
	"database/sql"
	"fmt"
)

type User struct {
	Id           int64
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
}

func GetUsers(ctx context.Context) ([]User, error) {
	var users []User

	rows, err := db.QueryContext(ctx, "SELECT id, firstname, lastname, email, password_hash FROM users")
	if err != nil {
		return nil, fmt.Errorf("getUsers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User

		if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash); err != nil {
			return nil, fmt.Errorf("getUsers: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getUsers: %w", err)
	}
	return users, nil
}

func GetUserByID(ctx context.Context, id int64) (User, error) {
	var user User

	row := db.QueryRowContext(ctx, "SELECT id, firstname, lastname, email, password_hash FROM users WHERE id = $1", id)
	if err := row.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return user, sql.ErrNoRows
		}
		return user, fmt.Errorf("getUserById %d: %w", id, err)
	}
	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User

	row := db.QueryRowContext(ctx, "SELECT id, firstname, lastname, email, password_hash FROM users WHERE email = $1", email)
	if err := row.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return user, sql.ErrNoRows
		}
		return user, fmt.Errorf("getUserByEmail %s: %w", email, err)
	}
	return user, nil
}

func CreateUser(ctx context.Context, firstName string, lastName string, email string, passwordHash string) (int64, error) {
	var id int64

	err := db.QueryRowContext(
		ctx,
		"INSERT INTO users (firstname, lastname, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id",
		firstName, lastName, email, passwordHash,
	).Scan(&id)

	if err != nil {
		return -1, fmt.Errorf("createUser: %w", err)
	}
	return id, nil
}

func DeleteUser(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)

	if err != nil {
		return fmt.Errorf("deleteUser %d: %w", id, err)
	}
	return nil
}

func UpdateUser(ctx context.Context, id int64, firstName string, lastName string, email string, passwordHash string) error {
	_, err := db.ExecContext(
		ctx,
		"UPDATE users SET firstname = $1, lastname = $2, email = $3, password_hash = $4 WHERE id = $5",
		firstName, lastName, email, passwordHash, id,
	)

	if err != nil {
		return fmt.Errorf("updateUser %d: %w", id, err)
	}
	return nil
}
