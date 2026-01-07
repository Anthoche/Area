package youtube

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 5 * time.Minute

// StartYouTubePoller checks youtube_new_video workflows and triggers on new videos.
func StartYouTubePoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("youtube poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		lastSeen := make(map[int64]string)
		lastCheck := make(map[int64]time.Time)

		for {
			select {
			case <-ctx.Done():
				log.Printf("youtube poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "youtube_new_video")
			if err != nil {
				log.Printf("youtube poller: list workflows: %v", err)
				continue
			}

			for _, wf := range wfs {
				if !wf.Enabled {
					continue
				}

				cfg, err := workflows.YouTubeNewVideoConfigFromJSON(wf.TriggerConfig)
				if err != nil || (strings.TrimSpace(cfg.ChannelID) == "" && strings.TrimSpace(cfg.Channel) == "") {
					log.Printf("youtube poller wf %d: bad config: %v", wf.ID, err)
					continue
				}

				channelID, err := resolveChannelID(ctx, cfg)
				if err != nil {
					log.Printf("youtube poller wf %d: resolve channel: %v", wf.ID, err)
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

				videos, err := fetchNewVideos(ctx, channelID, 5)
				if err != nil {
					log.Printf("youtube poller wf %d: fetch videos: %v", wf.ID, err)
					continue
				}
				if len(videos) == 0 {
					continue
				}

				key := wf.ID
				if lastSeen[key] == "" {
					lastSeen[key] = videos[0].ID
					continue
				}

				var toTrigger []youtubeVideo
				for i := 0; i < len(videos); i++ {
					if videos[i].ID == lastSeen[key] {
						break
					}
					toTrigger = append(toTrigger, videos[i])
				}

				for i := len(toTrigger) - 1; i >= 0; i-- {
					v := toTrigger[i]
					payload := map[string]any{
						"channel_id": channelID,
						"channel":    cfg.Channel,
						"id":         v.ID,
						"title":      v.Title,
						"author":     v.Author,
						"url":        v.URL,
						"published":  v.Published.Format(time.RFC3339),
					}
					for k, val := range cfg.PayloadTemplate {
						payload[k] = val
					}
					if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
						payload["content"] = fmt.Sprintf("New YouTube video: %s", v.Title)
					}
					if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
						log.Printf("youtube poller trigger wf %d: %v", wf.ID, err)
						continue
					}
				}
				lastSeen[key] = videos[0].ID
			}
		}
	}()
}

type youtubeFeed struct {
	Entries []youtubeEntry `xml:"entry"`
}

type youtubeEntry struct {
	ID        string        `xml:"id"`
	Title     string        `xml:"title"`
	Published string        `xml:"published"`
	Updated   string        `xml:"updated"`
	Link      youtubeLink   `xml:"link"`
	Author    youtubeAuthor `xml:"author"`
}

type youtubeLink struct {
	Href string `xml:"href,attr"`
}

type youtubeAuthor struct {
	Name string `xml:"name"`
}

type youtubeVideo struct {
	ID        string
	Title     string
	Author    string
	URL       string
	Published time.Time
}

// fetchNewVideos retrieves the latest videos from a YouTube channel's RSS feed.
func fetchNewVideos(ctx context.Context, channelID string, limit int) ([]youtubeVideo, error) {
	u := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", strings.TrimSpace(channelID))
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
		return nil, fmt.Errorf("youtube status %d: %s", resp.StatusCode, string(body))
	}
	body, _ := io.ReadAll(resp.Body)
	var feed youtubeFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	var videos []youtubeVideo
	seen := make(map[string]struct{})
	for _, entry := range feed.Entries {
		if limit > 0 && len(videos) >= limit {
			break
		}
		published := parseTime(entry.Published, entry.Updated)
		videoID := strings.TrimPrefix(entry.ID, "yt:video:")
		if videoID == "" {
			continue
		}
		if _, ok := seen[videoID]; ok {
			continue
		}
		seen[videoID] = struct{}{}
		videos = append(videos, youtubeVideo{
			ID:        videoID,
			Title:     entry.Title,
			Author:    entry.Author.Name,
			URL:       entry.Link.Href,
			Published: published,
		})
	}
	return videos, nil
}

// parseTime tries to parse primary time string, falling back to fallback if needed.
func parseTime(primary, fallback string) time.Time {
	if primary != "" {
		if t, err := time.Parse(time.RFC3339, primary); err == nil {
			return t
		}
	}
	if fallback != "" {
		if t, err := time.Parse(time.RFC3339, fallback); err == nil {
			return t
		}
	}
	return time.Now()
}

// resolveChannelID resolves the YouTube channel ID from the workflow config.
func resolveChannelID(ctx context.Context, cfg workflows.YouTubeNewVideoConfig) (string, error) {
	id := strings.TrimSpace(cfg.ChannelID)
	if id == "" {
		id = strings.TrimSpace(cfg.Channel)
	}
	if id == "" {
		return "", fmt.Errorf("missing channel identifier")
	}
	if strings.HasPrefix(id, "UC") {
		return id, nil
	}
	return lookupChannelID(ctx, id)
}

// lookupChannelID resolves a YouTube channel ID from various inputs.
func lookupChannelID(ctx context.Context, input string) (string, error) {
	candidate := strings.TrimSpace(input)
	if candidate == "" {
		return "", fmt.Errorf("empty channel value")
	}
	if strings.HasPrefix(candidate, "@") {
		urls := []string{
			"https://www.youtube.com/" + candidate,
			"https://www.youtube.com/" + candidate + "/about",
			"https://www.youtube.com/" + candidate + "/videos",
		}
		for _, u := range urls {
			id, err := fetchChannelIDFromPage(ctx, u)
			if err == nil && id != "" {
				return id, nil
			}
		}
		return "", fmt.Errorf("unable to resolve channel id from %q", input)
	}
	if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
		return fetchChannelIDFromPage(ctx, candidate)
	}
	urls := []string{
		"https://www.youtube.com/@" + candidate,
		"https://www.youtube.com/c/" + candidate,
		"https://www.youtube.com/user/" + candidate,
		"https://www.youtube.com/" + candidate,
	}
	for _, u := range urls {
		id, err := fetchChannelIDFromPage(ctx, u)
		if err == nil && id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("unable to resolve channel id from %q", input)
}

// fetchChannelIDFromPage retrieves the channel ID by scraping the channel page.
func fetchChannelIDFromPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "area-app/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("youtube channel page status %d", resp.StatusCode)
	}
	id := findChannelID(string(body))
	if id == "" {
		return "", fmt.Errorf("channel id not found")
	}
	return id, nil
}

// findChannelID extracts the channel ID from the page body.
func findChannelID(body string) string {
	patterns := []string{
		`"channelId":"(UC[a-zA-Z0-9_-]+)"`,
		`"externalId":"(UC[a-zA-Z0-9_-]+)"`,
		`"browseId":"(UC[a-zA-Z0-9_-]+)"`,
		`channel_id=(UC[a-zA-Z0-9_-]+)`,
		`/channel/(UC[a-zA-Z0-9_-]+)`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(body)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
