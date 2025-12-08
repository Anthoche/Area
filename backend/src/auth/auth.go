package auth

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	gh "github.com/google/go-github/github"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
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

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

var githubOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"user:email", "repo"},
	Endpoint:     github.Endpoint,
}

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
	token, err := googleOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
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

func GetUserDataFromGithub(code string) (*gh.User, error) {
	token, err := githubOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	oauthClient := githubOauthConfig.Client(context.Background(), token)
	client := gh.NewClient(oauthClient)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return user, nil
}
