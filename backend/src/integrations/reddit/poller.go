package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 5 * time.Minute

// StartRedditPoller checks reddit_new_post workflows and triggers on new posts.
func StartRedditPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("reddit poller: missing dependencies, skipping")
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
				log.Printf("reddit poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "reddit_new_post")
			if err != nil {
				log.Printf("reddit poller: list workflows: %v", err)
				continue
			}

			for _, wf := range wfs {
				if !wf.Enabled {
					continue
				}

				cfg, err := workflows.RedditNewPostConfigFromJSON(wf.TriggerConfig)
				if err != nil || strings.TrimSpace(cfg.Subreddit) == "" {
					log.Printf("reddit poller wf %d: bad config: %v", wf.ID, err)
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

				posts, err := fetchNewPosts(ctx, cfg.Subreddit, 5)
				if err != nil {
					log.Printf("reddit poller wf %d: fetch posts: %v", wf.ID, err)
					continue
				}
				if len(posts) == 0 {
					continue
				}

				key := wf.ID
				if lastSeen[key] == "" {
					lastSeen[key] = posts[0].ID
					continue
				}

				var toTrigger []redditPost
				for i := 0; i < len(posts); i++ {
					if posts[i].ID == lastSeen[key] {
						break
					}
					toTrigger = append(toTrigger, posts[i])
				}

				for i := len(toTrigger) - 1; i >= 0; i-- {
					p := toTrigger[i]
					payload := map[string]any{
						"subreddit": cfg.Subreddit,
						"id":        p.ID,
						"title":     p.Title,
						"author":    p.Author,
						"url":       p.URL,
						"permalink": p.Permalink,
						"created":   p.CreatedUTC.Format(time.RFC3339),
					}
					for k, v := range cfg.PayloadTemplate {
						payload[k] = v
					}
					if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
						payload["content"] = fmt.Sprintf("New Reddit post in r/%s: %s", cfg.Subreddit, p.Title)
					}
					if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
						log.Printf("reddit poller trigger wf %d: %v", wf.ID, err)
						continue
					}
				}
				lastSeen[key] = posts[0].ID
			}
		}
	}()
}

type redditPost struct {
	ID         string
	Title      string
	Author     string
	URL        string
	Permalink  string
	CreatedUTC time.Time
}

func fetchNewPosts(ctx context.Context, subreddit string, limit int) ([]redditPost, error) {
	u := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=%d", strings.TrimSpace(subreddit), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "area-app/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reddit status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Data struct {
			Children []struct {
				Data struct {
					ID        string  `json:"id"`
					Title     string  `json:"title"`
					Author    string  `json:"author"`
					URL       string  `json:"url"`
					Permalink string  `json:"permalink"`
					Created   float64 `json:"created_utc"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	var posts []redditPost
	for _, child := range payload.Data.Children {
		data := child.Data
		posts = append(posts, redditPost{
			ID:         data.ID,
			Title:      data.Title,
			Author:     data.Author,
			URL:        data.URL,
			Permalink:  "https://www.reddit.com" + data.Permalink,
			CreatedUTC: time.Unix(int64(data.Created), 0),
		})
	}
	return posts, nil
}
