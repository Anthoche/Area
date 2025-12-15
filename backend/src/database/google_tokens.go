package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InsertGoogleToken persists a new token and returns its ID.
func InsertGoogleToken(userID *int64, access, refresh string, expiry time.Time) (int64, error) {
	token := &GoogleToken{
		UserID:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
		Expiry:       expiry,
	}

	err := gorm.G[GoogleToken](db).Create(GetDBContext(), token)
	if err != nil {
		return -1, fmt.Errorf("insert google token: %w", err)
	}
	return int64(token.ID), nil
}

// GetGoogleToken fetches a token by its ID.
func GetGoogleToken(id int64) (*GoogleToken, error) {
	var t GoogleToken

	t, err := gorm.G[GoogleToken](db).Where("id = ?", id).First(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("get google token: %w", err)
	}
	return &t, nil
}

// GetGoogleTokenForUser fetches a token only if it matches the provided user id.
func GetGoogleTokenForUser(id int64, userID int64) (*GoogleToken, error) {
	var t GoogleToken

	t, err := gorm.G[GoogleToken](db).Where("id = ? AND user_id", id, userID).First(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("get google token for user: %w", err)
	}
	return &t, nil
}

// UpdateGoogleToken stores new access/refresh tokens and expiry for a token row.
func UpdateGoogleToken(id int64, access, refresh string, expiry time.Time) error {
	token := &GoogleToken{
		AccessToken:  access,
		RefreshToken: refresh,
		Expiry:       expiry,
	}

	_, err := gorm.G[GoogleToken](db).Where("id = ?", id).Updates(GetDBContext(), *token)
	if err != nil {
		return fmt.Errorf("update google token: %w", err)
	}
	return nil
}
