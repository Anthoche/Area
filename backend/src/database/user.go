package database

import (
	"fmt"

	"gorm.io/gorm"
)

// GetUsers returns all users from the database.
func GetUsers() ([]User, error) {
	var users []User

	users, err := gorm.G[User](db).Find(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("getUsers: %w", err)
	}
	return users, nil
}

// GetUserByID fetches a single user by numeric ID.
func GetUserByID(id int64) (User, error) {
	var user User

	user, err := gorm.G[User](db).Where("id = ?", 1).First(GetDBContext())
	if err != nil {
		return user, fmt.Errorf("getUserById %d: %w", id, err)
	}
	return user, nil
}

// GetUserByEmail fetches a user record matching the provided email.
func GetUserByEmail(email string) (User, error) {
	var user User

	user, err := gorm.G[User](db).Where("email = ?", email).First(GetDBContext())
	if err != nil {
		return user, fmt.Errorf("getUserByEmail %s: %w", email, err)
	}
	return user, nil
}

// CreateUser inserts a new user and returns its generated ID.
func CreateUser(firstName string, lastName string, email string, passwordHash string) (int64, error) {
	user := &User{
		Firstname:    firstName,
		Lastname:     lastName,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := gorm.G[User](db).Create(GetDBContext(), user); err != nil {
		return -1, fmt.Errorf("createUser: %w", err)
	}
	return int64(user.ID), nil
}

// DeleteUser removes the user with the given ID.
func DeleteUser(id int64) error {
	_, err := gorm.G[User](db).Where("id = ?", id).Delete(GetDBContext())

	if err != nil {
		return fmt.Errorf("deleteUser %d: %w", id, err)
	}
	return nil
}

// UpdateUser updates user attributes and password hash.
func UpdateUser(id int64, firstName string, lastName string, email string, passwordHash string) error {
	_, err := gorm.G[User](db).Where("id = ?", id).Updates(GetDBContext(), User{Firstname: firstName, Lastname: lastName, Email: email, PasswordHash: passwordHash})

	if err != nil {
		return fmt.Errorf("updateUser %d: %w", id, err)
	}
	return nil
}

func UserExists(email string) bool {
	var count int64

	db.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}
