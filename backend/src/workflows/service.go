package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTriggerUnavailable = errors.New("workflow triggerer not configured")
var ErrWorkflowNotFound = errors.New("workflow not found")
var ErrWorkflowDisabled = errors.New("workflow disabled")

// Service orchestrates workflow CRUD and triggering.
type Service struct {
	Store     *Store
	Triggerer *Triggerer
}

// NewService constructs a workflow service with its store and triggerer.
func NewService(store *Store, triggerer *Triggerer) *Service {
	return &Service{
		Store:     store,
		Triggerer: triggerer,
	}
}

// Trigger enqueues a workflow run with the provided payload.
func (s *Service) Trigger(ctx context.Context, workflowID int64, payload map[string]any) (*Run, error) {
	if s.Triggerer == nil {
		return nil, ErrTriggerUnavailable
	}
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	wf, err := s.Store.GetWorkflowForUser(ctx, workflowID, userID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	if !wf.Enabled && wf.TriggerType != "manual" {
		return nil, ErrWorkflowDisabled
	}
	return s.Triggerer.EnqueueRun(ctx, workflowID, payload)
}

// CreateWorkflow validates input and stores a new workflow.
func (s *Service) CreateWorkflow(ctx context.Context, name, triggerType, actionURL string, triggerConfig json.RawMessage) (*Workflow, error) {
	name = strings.TrimSpace(name)
	triggerType = strings.TrimSpace(triggerType)
	actionURL = strings.TrimSpace(actionURL)
	if name == "" || triggerType == "" || actionURL == "" {
		return nil, errors.New("name, triggerType and actionURL are required")
	}
	switch triggerType {
	case "interval":
		cfg, err := intervalConfigFromJSON(triggerConfig)
		if err != nil || cfg.IntervalMinutes <= 0 {
			return nil, errors.New("interval_minutes must be > 0 for interval trigger")
		}
	case "webhook", "manual", "gmail_inbound":
		if len(triggerConfig) == 0 {
			triggerConfig = []byte(`{}`)
		}
	case "github_commit":
		cfg, err := githubCommitConfigFromJSON(triggerConfig)
		if err != nil || cfg.TokenID <= 0 || cfg.Repo == "" || cfg.Branch == "" {
			return nil, errors.New("github_commit requires token_id, repo and branch")
		}
	case "github_pull_request":
		cfg, err := githubPRConfigFromJSON(triggerConfig)
		if err != nil || cfg.TokenID <= 0 || cfg.Repo == "" {
			return nil, errors.New("github_pull_request requires token_id and repo")
		}
	case "github_issue":
		cfg, err := githubIssueConfigFromJSON(triggerConfig)
		if err != nil || cfg.TokenID <= 0 || cfg.Repo == "" {
			return nil, errors.New("github_issue requires token_id and repo")
		}
	case "weather_temp":
		cfg, err := weatherTempConfigFromJSON(triggerConfig)
		if err != nil || cfg.Direction == "" || cfg.Threshold == 0 || cfg.City == "" {
			return nil, errors.New("weather_temp requires city, threshold and direction")
		}
		switch strings.ToLower(cfg.Direction) {
		case "above", "below":
		default:
			return nil, errors.New("weather_temp direction must be above or below")
		}
	case "weather_report":
		cfg, err := weatherReportConfigFromJSON(triggerConfig)
		if err != nil || cfg.IntervalMin <= 0 || cfg.City == "" {
			return nil, errors.New("weather_report requires city and interval_minutes")
		}
	case "reddit_new_post":
		cfg, err := redditNewPostConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.Subreddit) == "" {
			return nil, errors.New("reddit_new_post requires subreddit")
		}
	case "youtube_new_video":
		cfg, err := youtubeNewVideoConfigFromJSON(triggerConfig)
		if err != nil || (strings.TrimSpace(cfg.ChannelID) == "" && strings.TrimSpace(cfg.Channel) == "") {
			return nil, errors.New("youtube_new_video requires channel")
		}
	case "steam_player_online":
		cfg, err := steamPlayerOnlineConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.SteamID) == "" {
			return nil, errors.New("steam_player_online requires steam_id")
		}
	case "steam_game_sale":
		cfg, err := steamGameSaleConfigFromJSON(triggerConfig)
		if err != nil || cfg.AppID <= 0 {
			return nil, errors.New("steam_game_sale requires app_id")
		}
	case "steam_price_change":
		cfg, err := steamPriceChangeConfigFromJSON(triggerConfig)
		if err != nil || cfg.AppID <= 0 {
			return nil, errors.New("steam_price_change requires app_id")
		}
	case "crypto_price_threshold":
		cfg, err := cryptoPriceThresholdConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.CoinID) == "" || cfg.Threshold == 0 || cfg.Direction == "" {
			return nil, errors.New("crypto_price_threshold requires coin_id, threshold and direction")
		}
		switch strings.ToLower(cfg.Direction) {
		case "above", "below":
		default:
			return nil, errors.New("crypto_price_threshold direction must be above or below")
		}
	case "crypto_percent_change":
		cfg, err := cryptoPercentChangeConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.CoinID) == "" || cfg.Percent == 0 || cfg.Period == "" {
			return nil, errors.New("crypto_percent_change requires coin_id, percent and period")
		}
		switch strings.ToLower(cfg.Period) {
		case "1h", "24h":
		default:
			return nil, errors.New("crypto_percent_change period must be 1h or 24h")
		}
		switch strings.ToLower(strings.TrimSpace(cfg.Direction)) {
		case "", "any", "above", "below":
		default:
			return nil, errors.New("crypto_percent_change direction must be above, below, or any")
		}
	case "nasa_apod":
		if _, err := nasaApodConfigFromJSON(triggerConfig); err != nil {
			return nil, errors.New("nasa_apod config is invalid")
		}
	case "nasa_mars_photo":
		cfg, err := nasaMarsPhotoConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.Rover) == "" {
			return nil, errors.New("nasa_mars_photo requires rover")
		}
	case "nasa_neo_close_approach":
		cfg, err := nasaNeoConfigFromJSON(triggerConfig)
		if err != nil || cfg.ThresholdKM <= 0 {
			return nil, errors.New("nasa_neo_close_approach requires threshold_km")
		}
	case "air_quality_aqi_threshold":
		cfg, err := airQualityAQIConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.City) == "" || cfg.Threshold == 0 || cfg.Direction == "" {
			return nil, errors.New("air_quality_aqi_threshold requires city, threshold and direction")
		}
		switch strings.ToLower(cfg.Direction) {
		case "above", "below":
		default:
			return nil, errors.New("air_quality_aqi_threshold direction must be above or below")
		}
		switch strings.ToLower(cfg.Index) {
		case "", "us_aqi", "european_aqi":
		default:
			return nil, errors.New("air_quality_aqi_threshold index must be us_aqi or european_aqi")
		}
	case "air_quality_pm25_threshold":
		cfg, err := airQualityPM25ConfigFromJSON(triggerConfig)
		if err != nil || strings.TrimSpace(cfg.City) == "" || cfg.Threshold == 0 || cfg.Direction == "" {
			return nil, errors.New("air_quality_pm25_threshold requires city, threshold and direction")
		}
		switch strings.ToLower(cfg.Direction) {
		case "above", "below":
		default:
			return nil, errors.New("air_quality_pm25_threshold direction must be above or below")
		}
	default:
		return nil, fmt.Errorf("unsupported trigger_type %s", triggerType)
	}
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.CreateWorkflow(ctx, userID, name, triggerType, actionURL, triggerConfig)
}

// ListWorkflows returns all persisted workflows.
func (s *Service) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListWorkflows(ctx, userID)
}

// GetWorkflow fetches a workflow by ID or returns ErrWorkflowNotFound.
func (s *Service) GetWorkflow(ctx context.Context, id int64) (*Workflow, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	wf, err := s.Store.GetWorkflowForUser(ctx, id, userID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	return wf, nil
}

// DeleteWorkflow removes a workflow and its related runs/jobs.
func (s *Service) DeleteWorkflow(ctx context.Context, id int64) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.DeleteWorkflowForUser(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWorkflowNotFound
		}
		return err
	}
	return nil
}

// SetEnabled toggles a workflow (non-manual can be paused); interval workflows are rescheduled on enable.
func (s *Service) SetEnabled(ctx context.Context, id int64, enabled bool, now time.Time) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.SetEnabledForUser(ctx, id, userID, enabled, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWorkflowNotFound
		}
		return err
	}
	return nil
}

// TriggerWebhook finds a webhook workflow by token and enqueues it with payload.
func (s *Service) TriggerWebhook(ctx context.Context, token string, payload map[string]any) (*Run, error) {
	wf, err := s.Store.FindWorkflowByToken(ctx, token)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}
	if !wf.Enabled {
		return nil, ErrWorkflowDisabled
	}
	ctx = WithUserID(ctx, wf.UserID)
	return s.Trigger(ctx, wf.ID, payload)
}

// IntervalConfigFromJSON exposes interval config parsing to callers (e.g., scheduler).
func IntervalConfigFromJSON(raw json.RawMessage) (IntervalConfig, error) {
	return intervalConfigFromJSON(raw)
}

// GithubCommitConfigFromJSON exposes parsing for github_commit trigger config.
func GithubCommitConfigFromJSON(raw json.RawMessage) (GithubCommitConfig, error) {
	return githubCommitConfigFromJSON(raw)
}

// GithubPRConfigFromJSON exposes parsing for github_pull_request trigger config.
func GithubPRConfigFromJSON(raw json.RawMessage) (GithubPullRequestConfig, error) {
	return githubPRConfigFromJSON(raw)
}

// GithubIssueConfigFromJSON exposes parsing for github_issue trigger config.
func GithubIssueConfigFromJSON(raw json.RawMessage) (GithubIssueConfig, error) {
	return githubIssueConfigFromJSON(raw)
}

// WeatherTempConfigFromJSON exposes parsing for weather_temp trigger config.
func WeatherTempConfigFromJSON(raw json.RawMessage) (WeatherTempConfig, error) {
	return weatherTempConfigFromJSON(raw)
}

// RedditNewPostConfigFromJSON exposes parsing for reddit_new_post trigger config.
func RedditNewPostConfigFromJSON(raw json.RawMessage) (RedditNewPostConfig, error) {
	return redditNewPostConfigFromJSON(raw)
}

// YouTubeNewVideoConfigFromJSON exposes parsing for youtube_new_video trigger config.
func YouTubeNewVideoConfigFromJSON(raw json.RawMessage) (YouTubeNewVideoConfig, error) {
	return youtubeNewVideoConfigFromJSON(raw)
}

// SteamPlayerOnlineConfigFromJSON exposes parsing for steam_player_online trigger config.
func SteamPlayerOnlineConfigFromJSON(raw json.RawMessage) (SteamPlayerOnlineConfig, error) {
	return steamPlayerOnlineConfigFromJSON(raw)
}

// SteamGameSaleConfigFromJSON exposes parsing for steam_game_sale trigger config.
func SteamGameSaleConfigFromJSON(raw json.RawMessage) (SteamGameSaleConfig, error) {
	return steamGameSaleConfigFromJSON(raw)
}

// SteamPriceChangeConfigFromJSON exposes parsing for steam_price_change trigger config.
func SteamPriceChangeConfigFromJSON(raw json.RawMessage) (SteamPriceChangeConfig, error) {
	return steamPriceChangeConfigFromJSON(raw)
}

// CryptoPriceThresholdConfigFromJSON exposes parsing for crypto_price_threshold trigger config.
func CryptoPriceThresholdConfigFromJSON(raw json.RawMessage) (CryptoPriceThresholdConfig, error) {
	return cryptoPriceThresholdConfigFromJSON(raw)
}

// CryptoPercentChangeConfigFromJSON exposes parsing for crypto_percent_change trigger config.
func CryptoPercentChangeConfigFromJSON(raw json.RawMessage) (CryptoPercentChangeConfig, error) {
	return cryptoPercentChangeConfigFromJSON(raw)
}

// NasaApodConfigFromJSON exposes parsing for nasa_apod trigger config.
func NasaApodConfigFromJSON(raw json.RawMessage) (NasaApodConfig, error) {
	return nasaApodConfigFromJSON(raw)
}

// NasaMarsPhotoConfigFromJSON exposes parsing for nasa_mars_photo trigger config.
func NasaMarsPhotoConfigFromJSON(raw json.RawMessage) (NasaMarsPhotoConfig, error) {
	return nasaMarsPhotoConfigFromJSON(raw)
}

// NasaNeoConfigFromJSON exposes parsing for nasa_neo_close_approach trigger config.
func NasaNeoConfigFromJSON(raw json.RawMessage) (NasaNeoConfig, error) {
	return nasaNeoConfigFromJSON(raw)
}

// AirQualityAQIConfigFromJSON exposes parsing for air_quality_aqi_threshold trigger config.
func AirQualityAQIConfigFromJSON(raw json.RawMessage) (AirQualityAQIConfig, error) {
	return airQualityAQIConfigFromJSON(raw)
}

// AirQualityPM25ConfigFromJSON exposes parsing for air_quality_pm25_threshold trigger config.
func AirQualityPM25ConfigFromJSON(raw json.RawMessage) (AirQualityPM25Config, error) {
	return airQualityPM25ConfigFromJSON(raw)
}

type ctxUserIDKey struct{}

// WithUserID returns a context carrying the user id for authz in workflow operations.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

func userIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(ctxUserIDKey{})
	if val == nil {
		return 0, fmt.Errorf("missing user id in context")
	}
	uid, ok := val.(int64)
	if !ok || uid <= 0 {
		return 0, fmt.Errorf("invalid user id in context")
	}
	return uid, nil
}
