package nasa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 30 * time.Minute

// StartNasaPoller checks NASA workflows and triggers on new space data.
func StartNasaPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("nasa poller: missing dependencies, skipping")
		return
	}
	apiKey := strings.TrimSpace(os.Getenv("NASA_API_KEY"))
	if apiKey == "" {
		apiKey = "DEMO_KEY"
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		lastAPOD := make(map[int64]string)
		lastMars := make(map[int64]int)
		lastNEO := make(map[int64]string)
		lastCheck := make(map[int64]time.Time)

		for {
			select {
			case <-ctx.Done():
				log.Printf("nasa poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			pollAPOD(ctx, wfStore, wfService, apiKey, lastAPOD, lastCheck)
			pollMarsPhotos(ctx, wfStore, wfService, apiKey, lastMars, lastCheck)
			pollNEO(ctx, wfStore, wfService, apiKey, lastNEO, lastCheck)
		}
	}()
}

// pollAPOD checks NASA APOD and triggers workflows on new images.
func pollAPOD(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, apiKey string, lastAPOD map[int64]string, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "nasa_apod")
	if err != nil {
		log.Printf("nasa poller: list apod workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.NasaApodConfigFromJSON(wf.TriggerConfig)
		if err != nil {
			log.Printf("nasa poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		apod, err := fetchAPOD(ctx, apiKey)
		if err != nil {
			log.Printf("nasa poller wf %d: apod: %v", wf.ID, err)
			continue
		}
		if last, ok := lastAPOD[wf.ID]; ok && last == apod.Date {
			continue
		}
		lastAPOD[wf.ID] = apod.Date

		payload := map[string]any{
			"date":        apod.Date,
			"title":       apod.Title,
			"explanation": apod.Explanation,
			"url":         apod.URL,
			"hd_url":      apod.HDURL,
			"media_type":  apod.MediaType,
			"timestamp":   time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		augmentContent(payload, fmt.Sprintf("APOD: %s", apod.Title), apod.URL)
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("nasa poller trigger apod wf %d: %v", wf.ID, err)
		}
	}
}

// pollMarsPhotos checks NASA Mars photos and triggers workflows on new images.
func pollMarsPhotos(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, apiKey string, lastMars map[int64]int, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "nasa_mars_photo")
	if err != nil {
		log.Printf("nasa poller: list mars workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.NasaMarsPhotoConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.Rover) == "" {
			log.Printf("nasa poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		photo, err := fetchLatestMarsPhoto(ctx, apiKey, cfg.Rover, cfg.Camera)
		if err != nil {
			log.Printf("nasa poller wf %d: mars photo: %v", wf.ID, err)
			continue
		}
		if photo == nil {
			continue
		}
		if last, ok := lastMars[wf.ID]; ok && last == photo.ID {
			continue
		}
		lastMars[wf.ID] = photo.ID

		payload := map[string]any{
			"rover":      photo.Rover.Name,
			"camera":     photo.Camera.Name,
			"camera_id":  photo.Camera.ID,
			"earth_date": photo.EarthDate,
			"img_src":    photo.ImgSrc,
			"photo_id":   photo.ID,
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		augmentContent(payload, fmt.Sprintf("Mars photo (%s)", photo.Rover.Name), photo.ImgSrc)
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("nasa poller trigger mars wf %d: %v", wf.ID, err)
		}
	}
}

// pollNEO checks NASA NEOs and triggers workflows on close approaches.
func pollNEO(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, apiKey string, lastNEO map[int64]string, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "nasa_neo_close_approach")
	if err != nil {
		log.Printf("nasa poller: list neo workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.NasaNeoConfigFromJSON(wf.TriggerConfig)
		if err != nil || cfg.ThresholdKM <= 0 {
			log.Printf("nasa poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		days := cfg.DaysAhead
		if days <= 0 {
			days = 1
		}
		neo, err := fetchNearestNEO(ctx, apiKey, cfg.ThresholdKM, days)
		if err != nil {
			log.Printf("nasa poller wf %d: neo: %v", wf.ID, err)
			continue
		}
		if neo == nil {
			continue
		}
		key := fmt.Sprintf("%s:%s", neo.ID, neo.CloseApproachDate)
		if last, ok := lastNEO[wf.ID]; ok && last == key {
			continue
		}
		lastNEO[wf.ID] = key

		payload := map[string]any{
			"id":                  neo.ID,
			"name":                neo.Name,
			"hazardous":           neo.IsHazardous,
			"close_approach_date": neo.CloseApproachDate,
			"miss_distance_km":    neo.MissDistanceKM,
			"velocity_kps":        neo.VelocityKPS,
			"timestamp":           time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		augmentContent(payload, fmt.Sprintf("NEO %s at %.0f km", neo.Name, neo.MissDistanceKM), "")
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("nasa poller trigger neo wf %d: %v", wf.ID, err)
		}
	}
}

type apodResponse struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	URL         string `json:"url"`
	HDURL       string `json:"hdurl"`
	MediaType   string `json:"media_type"`
}

// fetchAPOD retrieves the Astronomy Picture of the Day.
func fetchAPOD(ctx context.Context, apiKey string) (*apodResponse, error) {
	u := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s", url.QueryEscape(apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nasa apod status %d: %s", resp.StatusCode, string(body))
	}
	var out apodResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type marsLatestResponse struct {
	Latest []marsPhoto `json:"latest_photos"`
}

type marsPhoto struct {
	ID        int    `json:"id"`
	ImgSrc    string `json:"img_src"`
	EarthDate string `json:"earth_date"`
	Camera    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"camera"`
	Rover struct {
		Name string `json:"name"`
	} `json:"rover"`
}

// fetchLatestMarsPhoto retrieves the latest Mars photo from a rover and optional camera.
func fetchLatestMarsPhoto(ctx context.Context, apiKey, rover, camera string) (*marsPhoto, error) {
	rover = strings.ToLower(strings.TrimSpace(rover))
	u := fmt.Sprintf("https://api.nasa.gov/mars-photos/api/v1/rovers/%s/latest_photos?api_key=%s",
		url.PathEscape(rover),
		url.QueryEscape(apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nasa mars status %d: %s", resp.StatusCode, string(body))
	}
	var out marsLatestResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if len(out.Latest) == 0 {
		return nil, nil
	}
	if strings.TrimSpace(camera) == "" {
		return &out.Latest[0], nil
	}
	target := strings.ToUpper(strings.TrimSpace(camera))
	for _, photo := range out.Latest {
		if strings.EqualFold(photo.Camera.Name, target) {
			return &photo, nil
		}
	}
	return nil, nil
}

type neoFeedResponse struct {
	Objects map[string][]neoObject `json:"near_earth_objects"`
}

type neoObject struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Hazardous         bool          `json:"is_potentially_hazardous_asteroid"`
	EstimatedDiameter neoDiameter   `json:"estimated_diameter"`
	CloseApproachData []neoApproach `json:"close_approach_data"`
}

type neoDiameter struct {
	Kilometers struct {
		EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
	} `json:"kilometers"`
}

type neoApproach struct {
	CloseApproachDate string `json:"close_approach_date"`
	MissDistance      struct {
		Kilometers string `json:"kilometers"`
	} `json:"miss_distance"`
	RelativeVelocity struct {
		KilometersPerSecond string `json:"kilometers_per_second"`
	} `json:"relative_velocity"`
}

type neoEvent struct {
	ID                string
	Name              string
	IsHazardous       bool
	CloseApproachDate string
	MissDistanceKM    float64
	VelocityKPS       float64
}

// augmentContent adds info and link to the content field of the payload.
func augmentContent(payload map[string]any, info, link string) {
	content, _ := payload["content"].(string)
	trimmed := strings.TrimSpace(content)
	var suffix string
	if link != "" {
		suffix = fmt.Sprintf("%s\n%s", info, link)
	} else {
		suffix = info
	}
	if trimmed == "" {
		payload["content"] = suffix
		return
	}
	if !strings.Contains(trimmed, info) {
		payload["content"] = trimmed + "\n" + suffix
	}
}

// fetchNearestNEO retrieves the nearest NEO within the threshold in the next daysAhead days.
func fetchNearestNEO(ctx context.Context, apiKey string, thresholdKM float64, daysAhead int) (*neoEvent, error) {
	start := time.Now().UTC()
	end := start.AddDate(0, 0, daysAhead)
	u := fmt.Sprintf("https://api.nasa.gov/neo/rest/v1/feed?start_date=%s&end_date=%s&api_key=%s",
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
		url.QueryEscape(apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nasa neo status %d: %s", resp.StatusCode, string(body))
	}
	var out neoFeedResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	var candidates []neoEvent
	for _, list := range out.Objects {
		for _, obj := range list {
			for _, approach := range obj.CloseApproachData {
				distance, err := strconv.ParseFloat(approach.MissDistance.Kilometers, 64)
				if err != nil {
					continue
				}
				if distance > thresholdKM {
					continue
				}
				velocity, _ := strconv.ParseFloat(approach.RelativeVelocity.KilometersPerSecond, 64)
				candidates = append(candidates, neoEvent{
					ID:                obj.ID,
					Name:              obj.Name,
					IsHazardous:       obj.Hazardous,
					CloseApproachDate: approach.CloseApproachDate,
					MissDistanceKM:    distance,
					VelocityKPS:       velocity,
				})
			}
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].MissDistanceKM < candidates[j].MissDistanceKM
	})
	return &candidates[0], nil
}
