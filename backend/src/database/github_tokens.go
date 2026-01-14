package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// InsertGithubToken persists a GitHub access token and returns its ID.
func InsertGithubToken(userID *int64, accessToken, tokenType, scope string) (int64, error) {
	token := &GithubToken{
		UserID:      userID,
		AccessToken: accessToken,
		TokenType:   tokenType,
		Scope:       scope,
	}

	err := gorm.G[GithubToken](Db).Create(GetDBContext(), token)
	if err != nil {
		return -1, fmt.Errorf("insert github token: %w", err)
	}
	return int64(token.ID), nil
}

// GetGithubToken fetches a GitHub token by ID.
func GetGithubToken(id int64) (*GithubToken, error) {
	var token GithubToken

	token, err := gorm.G[GithubToken](Db).Where("id = ?", id).First(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("get github token: %w", err)
	}
	return &token, nil
}

// GetGithubTokenForUser fetches a token if it belongs to the given user.
func GetGithubTokenForUser(id int64, userID int64) (*GithubToken, error) {
	var token GithubToken

	token, err := gorm.G[GithubToken](Db).Where("id = ? AND user_id = ?", id, userID).First(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("get github token for user: %w", err)
	}
	return &token, nil
}

// GetLatestGithubTokenForUser returns the most recently created token for a user.
func GetLatestGithubTokenForUser(ctx context.Context, userID int64) (*GithubToken, error) {
	var token GithubToken

	token, err := gorm.G[GithubToken](Db).Where("user_id = ?", userID).Order("created_at DESC").First(GetDBContext())
	if err != nil {
		return nil, fmt.Errorf("get latest github token for user: %w", err)
	}
	return &token, nil
}
