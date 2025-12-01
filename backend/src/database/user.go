package database

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Firstname    string
	Lastname     string
	Email        string
	PasswordHash string
}

func GetUsers() ([]User, error) {
	var users []User

	users, err := gorm.G[User](db).Find(getDBContext())
	if err != nil {
		return nil, fmt.Errorf("getUsers: %w", err)
	}
	return users, nil
}

func GetUserByID(id int64) (User, error) {
	var user User

	user, err := gorm.G[User](db).Where("id = ?", 1).First(getDBContext())
	if err != nil {
		return user, fmt.Errorf("getUserById %d: %w", id, err)
	}
	return user, nil
}

func GetUserByEmail(email string) (User, error) {
	var user User

	user, err := gorm.G[User](db).Where("email = ?", email).First(getDBContext())
	if err != nil {
		return user, fmt.Errorf("getUserByEmail %s: %w", email, err)
	}
	return user, nil
}

func CreateUser(firstName string, lastName string, email string, passwordHash string) (int64, error) {
	user := &User{
		Firstname:    firstName,
		Lastname:     lastName,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := gorm.G[User](db).Create(getDBContext(), user); err != nil {
		return -1, fmt.Errorf("createUser: %w", err)
	}
	return int64(user.ID), nil
}

func DeleteUser(id int64) error {
	_, err := gorm.G[User](db).Where("id = ?", id).Delete(getDBContext())

	if err != nil {
		return fmt.Errorf("deleteUser %d: %w", id, err)
	}
	return nil
}

func UpdateUser(id int64, firstName string, lastName string, email string, passwordHash string) error {
	_, err := gorm.G[User](db).Where("id = ?", id).Updates(getDBContext(), User{Firstname: firstName, Lastname: lastName, Email: email, PasswordHash: passwordHash})

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
