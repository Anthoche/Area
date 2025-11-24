package database

import (
	"database/sql"
	"fmt"
)

type User struct {
	Id int64
	FirstName string
	LastName string
	Email string
	PasswordHash string
}

func getUserByID(id int64) (User, error) {
	var user User

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	if err := row.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("getUserById %d: no such user", id)
		}
		return user, fmt.Errorf("getUserById %d: %v", id, err)
	}
	return user, nil
}

func getUserByEmail(email string) (User, error) {
	var user User

	row := db.QueryRow("SELECT * FROM users WHERE email = $1", email)
	if err := row.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("getUserByEmail %s: no such user", email)
		}
		return user, fmt.Errorf("getUserByEmail %s: %v", email, err)
	}
	return user, nil
}

func createUser(firstName string, lastName string, email string, passwordHash string) (int64, error) {
	var id int64

	err := db.QueryRow(
		"INSERT INTO users (firstname, lastname, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id",
		firstName, lastName, email, passwordHash,
	).Scan(&id)

	if err != nil {
		return -1, fmt.Errorf("createUser: %v", err)
	}
	return id, nil
}

func deleteUser(id int64) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)

	if err != nil {
		return fmt.Errorf("deleteUser %d: %v", id, err)
	}
	return nil
}

func updateUser(id int64, firstName string, lastName string, email string, passwordHash string) error {
	_, err := db.Exec(
		"UPDATE users SET firstname = $1, lastname = $2, email = $3, password_hash = $4 WHERE id = $5",
		firstName, lastName, email, passwordHash, id,
	)

	if err != nil {
		return fmt.Errorf("updateUser %d: %v", id, err)
	}
	return nil
}
