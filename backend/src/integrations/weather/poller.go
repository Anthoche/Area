package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"area/src/workflows"
)

// StartWeatherPoller checks weather_temp workflows and triggers on threshold crossings.
func StartWeatherPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("weather poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		lastState := make(map[int64]bool)
		lastCheck := make(map[int64]time.Time)
		lastReport := make(map[int64]time.Time)
		cityCache := make(map[string][2]float64)

		for {
			select {
			case <-ctx.Done():
				log.Printf("weather poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "weather_temp")
			if err != nil {
				log.Printf("weather poller: list workflows: %v", err)
				continue
			}
			for _, wf := range wfs {
				if !wf.Enabled {
					continue
				}
				var cfg workflows.WeatherTempConfig
				if err := json.Unmarshal(wf.TriggerConfig, &cfg); err != nil {
					log.Printf("weather poller wf %d: bad config: %v", wf.ID, err)
					continue
				}
				if cfg.City == "" {
					log.Printf("weather poller wf %d: missing city", wf.ID)
					continue
				}
				coords, ok := cityCache[cfg.City]
				if !ok {
					lat, lon, err := geocodeCity(ctx, cfg.City)
					if err != nil {
						log.Printf("weather poller wf %d: geocode %q: %v", wf.ID, cfg.City, err)
						continue
					}
					coords = [2]float64{lat, lon}
					cityCache[cfg.City] = coords
				}
				cfg.Lat = coords[0]
				cfg.Lon = coords[1]
				if cfg.IntervalMin > 0 {
					if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < time.Duration(cfg.IntervalMin)*time.Minute {
						continue
					}
				}
				temp, err := fetchCurrentTemp(ctx, cfg.Lat, cfg.Lon)
				if err != nil {
					log.Printf("weather poller wf %d: fetch: %v", wf.ID, err)
					continue
				}
				lastCheck[wf.ID] = time.Now()

				above := temp >= cfg.Threshold
				prev, hasPrev := lastState[wf.ID]
				if !hasPrev {
					lastState[wf.ID] = above
					payload := map[string]any{
						"lat":       cfg.Lat,
						"lon":       cfg.Lon,
						"threshold": cfg.Threshold,
						"direction": cfg.Direction,
						"temp":      temp,
						"event":     "current",
						"timestamp": time.Now().Format(time.RFC3339),
					}
					for k, v := range cfg.PayloadTemplate {
						payload[k] = v
					}
					if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
						payload["content"] = fmt.Sprintf("%s | temp: %.1f°C", content, temp)
					} else {
						payload["content"] = fmt.Sprintf("Temp: %.1f°C (%s %g)", temp, cfg.Direction, cfg.Threshold)
					}
					if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
						log.Printf("weather poller trigger wf %d: %v", wf.ID, err)
					}
					continue
				}
				lastState[wf.ID] = above

				dir := cfg.Direction
				trigger := (dir == "above" && above && !prev) || (dir == "below" && !above && prev)
				if !trigger {
					continue
				}

				payload := map[string]any{
					"lat":       cfg.Lat,
					"lon":       cfg.Lon,
					"threshold": cfg.Threshold,
					"direction": cfg.Direction,
					"temp":      temp,
					"event":     "threshold",
					"timestamp": time.Now().Format(time.RFC3339),
				}
				for k, v := range cfg.PayloadTemplate {
					payload[k] = v
				}
				if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
					payload["content"] = fmt.Sprintf("%s | temp: %.1f°C", content, temp)
				} else {
					payload["content"] = fmt.Sprintf("Temp: %.1f°C (%s %g)", temp, cfg.Direction, cfg.Threshold)
				}
				if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
					log.Printf("weather poller trigger wf %d: %v", wf.ID, err)
				}
			}

			reports, err := wfStore.ListWorkflowsByTrigger(ctx, "weather_report")
			if err != nil {
				log.Printf("weather poller: list report workflows: %v", err)
				continue
			}
			for _, wf := range reports {
				if !wf.Enabled {
					continue
				}
				var cfg workflows.WeatherReportConfig
				if err := json.Unmarshal(wf.TriggerConfig, &cfg); err != nil {
					log.Printf("weather poller wf %d: bad report config: %v", wf.ID, err)
					continue
				}
				if cfg.City == "" {
					log.Printf("weather poller wf %d: missing city", wf.ID)
					continue
				}
				if cfg.IntervalMin <= 0 {
					cfg.IntervalMin = 10
				}
				if last, ok := lastReport[wf.ID]; ok && time.Since(last) < time.Duration(cfg.IntervalMin)*time.Minute {
					continue
				}
				coords, ok := cityCache[cfg.City]
				if !ok {
					lat, lon, err := geocodeCity(ctx, cfg.City)
					if err != nil {
						log.Printf("weather poller wf %d: geocode %q: %v", wf.ID, cfg.City, err)
						continue
					}
					coords = [2]float64{lat, lon}
					cityCache[cfg.City] = coords
				}
				temp, err := fetchCurrentTemp(ctx, coords[0], coords[1])
				if err != nil {
					log.Printf("weather poller wf %d: fetch report temp: %v", wf.ID, err)
					continue
				}
				lastReport[wf.ID] = time.Now()
				payload := map[string]any{
					"city":      cfg.City,
					"lat":       coords[0],
					"lon":       coords[1],
					"temp":      temp,
					"event":     "report",
					"timestamp": time.Now().Format(time.RFC3339),
				}
				for k, v := range cfg.PayloadTemplate {
					payload[k] = v
				}
				if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
					payload["content"] = fmt.Sprintf("%s | temp: %.1f°C", content, temp)
				} else {
					payload["content"] = fmt.Sprintf("Temp: %.1f°C (%s)", temp, cfg.City)
				}
				if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
					log.Printf("weather poller trigger wf %d: %v", wf.ID, err)
				}
			}
		}
	}()
}

// fetchCurrentTemp retrieves the current temperature for given latitude and longitude.
func fetchCurrentTemp(ctx context.Context, lat, lon float64) (float64, error) {
	u := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m", lat, lon)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("weather status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
		} `json:"current"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	return payload.Current.Temperature, nil
}

// geocodeCity retrieves the latitude and longitude for a given city name.
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
