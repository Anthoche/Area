package database

import (
	"context"
	"fmt"
	"time"
)

type GoogleToken struct {
	Id           int64
	UserID       *int64
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	CreatedAt    time.Time
}

// InsertGoogleToken persists a new token and returns its ID.
func InsertGoogleToken(ctx context.Context, userID *int64, access, refresh string, expiry time.Time) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO google_tokens (user_id, access_token, refresh_token, expiry)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		userID, access, refresh, expiry).
		Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("insert google token: %w", err)
	}
	return id, nil
}

// GetGoogleToken fetches a token by its ID.
func GetGoogleToken(ctx context.Context, id int64) (*GoogleToken, error) {
	var t GoogleToken
	var userID *int64
	err := db.QueryRowContext(ctx, `
		SELECT id, user_id, access_token, refresh_token, expiry, created_at
		FROM google_tokens
		WHERE id = $1`, id).
		Scan(&t.Id, &userID, &t.AccessToken, &t.RefreshToken, &t.Expiry, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get google token: %w", err)
	}
	if userID != nil {
		t.UserID = userID
	}
	return &t, nil
}

// GetGoogleTokenForUser fetches a token only if it matches the provided user id.
func GetGoogleTokenForUser(ctx context.Context, id int64, userID int64) (*GoogleToken, error) {
	var t GoogleToken
	var uid *int64
	err := db.QueryRowContext(ctx, `
		SELECT id, user_id, access_token, refresh_token, expiry, created_at
		FROM google_tokens
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&t.Id, &uid, &t.AccessToken, &t.RefreshToken, &t.Expiry, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get google token for user: %w", err)
	}
	if uid != nil {
		t.UserID = uid
	}
	return &t, nil
}

// UpdateGoogleToken stores new access/refresh tokens and expiry for a token row.
func UpdateGoogleToken(ctx context.Context, id int64, access, refresh string, expiry time.Time) error {
	_, err := db.ExecContext(ctx, `
		UPDATE google_tokens
		SET access_token = $1, refresh_token = $2, expiry = $3
		WHERE id = $4`,
		access, refresh, expiry, id)
	if err != nil {
		return fmt.Errorf("update google token: %w", err)
	}
	return nil
}
