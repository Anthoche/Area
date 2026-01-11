package airquality

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 10 * time.Minute

// StartAirQualityPoller checks air quality workflows and triggers on threshold crossings.
func StartAirQualityPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("air quality poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		lastAQI := make(map[int64]bool)
		lastPM25 := make(map[int64]bool)
		lastCheck := make(map[int64]time.Time)
		cityCache := make(map[string][2]float64)

		for {
			select {
			case <-ctx.Done():
				log.Printf("air quality poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			pollAQI(ctx, wfStore, wfService, lastAQI, lastCheck, cityCache)
			pollPM25(ctx, wfStore, wfService, lastPM25, lastCheck, cityCache)
		}
	}()
}

func pollAQI(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastState map[int64]bool, lastCheck map[int64]time.Time, cityCache map[string][2]float64) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "air_quality_aqi_threshold")
	if err != nil {
		log.Printf("air quality poller: list aqi workflows: %v", err)
		return
	}
	last := make(map[int64]bool)
	for k, v := range lastState {
		last[k] = v
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.AirQualityAQIConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.City) == "" {
			log.Printf("air quality poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if lastTime, ok := lastCheck[wf.ID]; ok && time.Since(lastTime) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		coords, ok := cityCache[cfg.City]
		if !ok {
			lat, lon, err := geocodeCity(ctx, cfg.City)
			if err != nil {
				log.Printf("air quality poller wf %d: geocode %q: %v", wf.ID, cfg.City, err)
				continue
			}
			coords = [2]float64{lat, lon}
			cityCache[cfg.City] = coords
		}
		snap, err := fetchAirQuality(ctx, coords[0], coords[1])
		if err != nil {
			log.Printf("air quality poller wf %d: fetch: %v", wf.ID, err)
			continue
		}
		index := strings.ToLower(strings.TrimSpace(cfg.Index))
		if index == "" {
			index = "us_aqi"
		}
		aqi := snap.USAQI
		if index == "european_aqi" {
			aqi = snap.EUAQI
		}
		above := aqi >= cfg.Threshold
		prev, hasPrev := last[wf.ID]
		lastState[wf.ID] = above
		payload := map[string]any{
			"city":      cfg.City,
			"lat":       coords[0],
			"lon":       coords[1],
			"index":     index,
			"aqi":       aqi,
			"threshold": cfg.Threshold,
			"direction": cfg.Direction,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
			payload["content"] = fmt.Sprintf("%s | AQI: %.0f", content, aqi)
		} else {
			payload["content"] = fmt.Sprintf("%s AQI: %.0f", strings.ToUpper(index), aqi)
		}
		if !hasPrev {
			payload["event"] = "current"
			if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
				log.Printf("air quality poller trigger wf %d: %v", wf.ID, err)
			}
			continue
		}
		dir := strings.ToLower(cfg.Direction)
		trigger := (dir == "above" && above && !prev) || (dir == "below" && !above && prev)
		if !trigger {
			continue
		}
		payload["event"] = "threshold"
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("air quality poller trigger wf %d: %v", wf.ID, err)
		}
	}
}

func pollPM25(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastState map[int64]bool, lastCheck map[int64]time.Time, cityCache map[string][2]float64) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "air_quality_pm25_threshold")
	if err != nil {
		log.Printf("air quality poller: list pm2_5 workflows: %v", err)
		return
	}
	last := make(map[int64]bool)
	for k, v := range lastState {
		last[k] = v
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.AirQualityPM25ConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.City) == "" {
			log.Printf("air quality poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if lastTime, ok := lastCheck[wf.ID]; ok && time.Since(lastTime) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		coords, ok := cityCache[cfg.City]
		if !ok {
			lat, lon, err := geocodeCity(ctx, cfg.City)
			if err != nil {
				log.Printf("air quality poller wf %d: geocode %q: %v", wf.ID, cfg.City, err)
				continue
			}
			coords = [2]float64{lat, lon}
			cityCache[cfg.City] = coords
		}
		snap, err := fetchAirQuality(ctx, coords[0], coords[1])
		if err != nil {
			log.Printf("air quality poller wf %d: fetch: %v", wf.ID, err)
			continue
		}
		pm25 := snap.PM25
		above := pm25 >= cfg.Threshold
		prev, hasPrev := last[wf.ID]
		lastState[wf.ID] = above

		payload := map[string]any{
			"city":      cfg.City,
			"lat":       coords[0],
			"lon":       coords[1],
			"pm2_5":     pm25,
			"threshold": cfg.Threshold,
			"direction": cfg.Direction,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
			payload["content"] = fmt.Sprintf("%s | PM2.5: %.1f", content, pm25)
		} else {
			payload["content"] = fmt.Sprintf("PM2.5: %.1f µg/m³", pm25)
		}
		if !hasPrev {
			payload["event"] = "current"
			if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
				log.Printf("air quality poller trigger wf %d: %v", wf.ID, err)
			}
			continue
		}
		dir := strings.ToLower(cfg.Direction)
		trigger := (dir == "above" && above && !prev) || (dir == "below" && !above && prev)
		if !trigger {
			continue
		}
		payload["event"] = "threshold"
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("air quality poller trigger wf %d: %v", wf.ID, err)
		}
	}
}

type airQualitySnapshot struct {
	USAQI float64
	EUAQI float64
	PM25  float64
	PM10  float64
}

func fetchAirQuality(ctx context.Context, lat, lon float64) (*airQualitySnapshot, error) {
	u := fmt.Sprintf("https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%g&longitude=%g&hourly=us_aqi,european_aqi,pm2_5,pm10&timezone=UTC",
		lat,
		lon,
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
		return nil, fmt.Errorf("air quality status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Hourly struct {
			Time  []string  `json:"time"`
			USAQI []float64 `json:"us_aqi"`
			EUAQI []float64 `json:"european_aqi"`
			PM25  []float64 `json:"pm2_5"`
			PM10  []float64 `json:"pm10"`
		} `json:"hourly"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	idx := latestIndex(payload.Hourly.Time)
	if idx < 0 {
		return nil, fmt.Errorf("air quality no data")
	}
	return &airQualitySnapshot{
		USAQI: valueAt(payload.Hourly.USAQI, idx),
		EUAQI: valueAt(payload.Hourly.EUAQI, idx),
		PM25:  valueAt(payload.Hourly.PM25, idx),
		PM10:  valueAt(payload.Hourly.PM10, idx),
	}, nil
}

func latestIndex(times []string) int {
	now := time.Now().UTC()
	for i := len(times) - 1; i >= 0; i-- {
		ts, err := time.Parse("2006-01-02T15:04", times[i])
		if err != nil {
			continue
		}
		if !ts.After(now) {
			return i
		}
	}
	return len(times) - 1
}

func valueAt(values []float64, idx int) float64 {
	if idx < 0 || idx >= len(values) {
		return 0
	}
	return values[idx]
}

func geocodeCity(ctx context.Context, city string) (float64, float64, error) {
	u := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=fr&format=json", url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("geocode status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, 0, err
	}
	if len(payload.Results) == 0 {
		return 0, 0, fmt.Errorf("no results for city")
	}
	return payload.Results[0].Latitude, payload.Results[0].Longitude, nil
}
