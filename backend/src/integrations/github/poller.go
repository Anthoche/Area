package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"area/src/workflows"
)

// StartGithubPoller launches a goroutine that watches github_commit workflows and triggers on new commits.
func StartGithubPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client) {
	if wfStore == nil || wfService == nil || client == nil {
		log.Println("github poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()

		type key struct {
			WorkflowID int64
		}
		lastSeen := make(map[key]string)

		for {
			select {
			case <-ctx.Done():
				log.Printf("github poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "github_commit")
			if err != nil {
				log.Printf("github poller: list workflows: %v", err)
				continue
			}

			for _, wf := range wfs {
				if !wf.Enabled {
					continue
				}

				var cfg workflows.GithubCommitConfig
				if err := json.Unmarshal(wf.TriggerConfig, &cfg); err != nil {
					log.Printf("github poller wf %d: bad config: %v", wf.ID, err)
					continue
				}
				repoParts := strings.Split(cfg.Repo, "/")
				if len(repoParts) != 2 {
					log.Printf("github poller wf %d: invalid repo %q", wf.ID, cfg.Repo)
					continue
				}

				commits, err := client.ListRecentCommits(ctx, &wf.UserID, cfg.TokenID, repoParts[0], repoParts[1], cfg.Branch, 5)
				if err != nil {
					log.Printf("github poller wf %d: list commits: %v", wf.ID, err)
					continue
				}
				if len(commits) == 0 {
					continue
				}

				k := key{WorkflowID: wf.ID}
				if lastSeen[k] == "" {
					// initialize cursor without triggering
					lastSeen[k] = commits[0].SHA
					continue
				}

				var toTrigger []Commit
				for i := 0; i < len(commits); i++ {
					if commits[i].SHA == lastSeen[k] {
						break
					}
					toTrigger = append(toTrigger, commits[i])
				}

				// trigger oldest-first among new commits
				for i := len(toTrigger) - 1; i >= 0; i-- {
					cmt := toTrigger[i]
					payload := map[string]any{
						"repo":    cfg.Repo,
						"branch":  cfg.Branch,
						"sha":     cmt.SHA,
						"author":  cmt.Author,
						"message": cmt.Message,
						"date":    cmt.Date.Format(time.RFC3339),
					}
					for k, v := range cfg.PayloadTemplate {
						payload[k] = v
					}
					if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
						payload["content"] = fmt.Sprintf("New commit on %s (%s): %s", cfg.Repo, cfg.Branch, cmt.Message)
					}
					if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
						log.Printf("github poller trigger wf %d: %v", wf.ID, err)
						continue
					}
				}
				lastSeen[k] = commits[0].SHA
			}
		}
	}()
}
