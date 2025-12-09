package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

var bcryptCost int

// init configures the bcrypt cost from environment variables.
func init() {
	costStr := os.Getenv("BCRYPT_COST")
	cost, err := strconv.Atoi(costStr)
	if err != nil {
		bcryptCost = bcrypt.DefaultCost
	} else {
		bcryptCost = cost
	}
}

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt hash against a plaintext password.
func CheckPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

// UserStore defines the storage operations needed for auth.
type UserStore interface {
	Create(ctx context.Context, user *User, passwordHash string) error
	GetByEmail(ctx context.Context, email string) (*User, string, error)
}

// Service handles authentication and user creation against a user store.
type Service struct {
	store UserStore
}

// NewService wires the auth service with a backing store (DB, memory, etc.).
func NewService(store UserStore) *Service {
	return &Service{store: store}
}

// Authenticate verifies email/password against the stored hash.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, hashed, err := s.store.GetByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return nil, ErrInvalidCredentials
	case err != nil:
		return nil, err
	}
	if err := CheckPassword(hashed, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// Register creates a new user with a hashed password in the store.
func (s *Service) Register(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	if err := s.store.Create(ctx, user, hashed); err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserDataFromGoogle(code string) ([]byte, error) {
	redirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		return nil, fmt.Errorf("missing GOOGLE_OAUTH_REDIRECT_URI")
	}
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", redirectURI)
	values.Set("client_id", os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("code exchange wrong: %s", string(body))
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

func GetUserDataFromGithub(code string) ([]byte, error) {
	values := url.Values{}
	values.Set("client_id", os.Getenv("GITHUB_OAUTH_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"))
	values.Set("code", code)
	if redirect := os.Getenv("GITHUB_OAUTH_REDIRECT_URI"); redirect != "" {
		values.Set("redirect_uri", redirect)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("code exchange wrong: %s", string(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, err
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("no access_token returned from github")
	}

	userReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	userReq.Header.Set("Accept", "application/json")
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()
	userBody, _ := ioutil.ReadAll(userResp.Body)
	if userResp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed read response: %s", string(userBody))
	}
	return userBody, nil
}
