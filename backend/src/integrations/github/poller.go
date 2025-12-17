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

// StartGithubPoller launches goroutines that watch GitHub workflows (commit/PR/issue) and trigger on changes.
func StartGithubPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client) {
	if wfStore == nil || wfService == nil || client == nil {
		log.Println("github poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()

		lastCommit := make(map[int64]string)
		lastPR := make(map[int64]string)
		lastIssue := make(map[int64]string)

		for {
			select {
			case <-ctx.Done():
				log.Printf("github poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			pollCommits(ctx, wfStore, wfService, client, lastCommit)
			pollPullRequests(ctx, wfStore, wfService, client, lastPR)
			pollIssues(ctx, wfStore, wfService, client, lastIssue)
		}
	}()
}

func pollCommits(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client, lastSeen map[int64]string) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "github_commit")
	if err != nil {
		log.Printf("github poller: list commit workflows: %v", err)
		return
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

		k := wf.ID
		if lastSeen[k] == "" {
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
			for k2, v := range cfg.PayloadTemplate {
				payload[k2] = v
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

func pollPullRequests(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client, lastSeen map[int64]string) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "github_pull_request")
	if err != nil {
		log.Printf("github poller: list PR workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		var cfg workflows.GithubPullRequestConfig
		if err := json.Unmarshal(wf.TriggerConfig, &cfg); err != nil {
			log.Printf("github poller wf %d: bad PR config: %v", wf.ID, err)
			continue
		}
		repoParts := strings.Split(cfg.Repo, "/")
		if len(repoParts) != 2 {
			log.Printf("github poller wf %d: invalid repo %q", wf.ID, cfg.Repo)
			continue
		}
		prs, err := client.ListRecentPullRequests(ctx, &wf.UserID, cfg.TokenID, repoParts[0], repoParts[1], 5)
		if err != nil {
			log.Printf("github poller wf %d: list PRs: %v", wf.ID, err)
			continue
		}
		if len(prs) == 0 {
			continue
		}
		k := wf.ID
		if lastSeen[k] == "" {
			lastSeen[k] = fmt.Sprintf("%d-%s", prs[0].Number, prs[0].UpdatedAt.Format(time.RFC3339Nano))
			continue
		}

		var toTrigger []PullRequest
		last := lastSeen[k]
		found := false
		for i := 0; i < len(prs); i++ {
			curr := fmt.Sprintf("%d-%s", prs[i].Number, prs[i].UpdatedAt.Format(time.RFC3339Nano))
			if curr == last {
				found = true
				break
			}
			toTrigger = append(toTrigger, prs[i])
		}
		if !found && lastSeen[k] != "" && len(toTrigger) > 0 {
			lastSeen[k] = fmt.Sprintf("%d-%s", prs[0].Number, prs[0].UpdatedAt.Format(time.RFC3339Nano))
			continue
		}
		for i := len(toTrigger) - 1; i >= 0; i-- {
			pr := toTrigger[i]
			action := "opened"
			if pr.Merged {
				action = "merged"
			} else if pr.State == "closed" {
				action = "closed"
			}
			if len(cfg.Actions) > 0 && !containsString(cfg.Actions, action) {
				continue
			}
			payload := map[string]any{
				"repo":    cfg.Repo,
				"number":  pr.Number,
				"title":   pr.Title,
				"state":   pr.State,
				"action":  action,
				"author":  pr.Author,
				"base":    pr.Base,
				"head":    pr.Head,
				"url":     pr.URL,
				"updated": pr.UpdatedAt.Format(time.RFC3339),
			}
			for k2, v := range cfg.PayloadTemplate {
				payload[k2] = v
			}
			if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
				payload["content"] = fmt.Sprintf("PR #%d %s on %s: %s", pr.Number, action, cfg.Repo, pr.Title)
			}
			if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
				log.Printf("github poller trigger PR wf %d: %v", wf.ID, err)
				continue
			}
		}
		lastSeen[k] = fmt.Sprintf("%d-%s", prs[0].Number, prs[0].UpdatedAt.Format(time.RFC3339Nano))
	}
}

func pollIssues(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client, lastSeen map[int64]string) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "github_issue")
	if err != nil {
		log.Printf("github poller: list issue workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		var cfg workflows.GithubIssueConfig
		if err := json.Unmarshal(wf.TriggerConfig, &cfg); err != nil {
			log.Printf("github poller wf %d: bad issue config: %v", wf.ID, err)
			continue
		}
		repoParts := strings.Split(cfg.Repo, "/")
		if len(repoParts) != 2 {
			log.Printf("github poller wf %d: invalid repo %q", wf.ID, cfg.Repo)
			continue
		}
		issues, err := client.ListRecentIssues(ctx, &wf.UserID, cfg.TokenID, repoParts[0], repoParts[1], 5)
		if err != nil {
			log.Printf("github poller wf %d: list issues: %v", wf.ID, err)
			continue
		}
		if len(issues) == 0 {
			continue
		}
		k := wf.ID
		if lastSeen[k] == "" {
			lastSeen[k] = fmt.Sprintf("%d-%s", issues[0].Number, issues[0].UpdatedAt.Format(time.RFC3339Nano))
			continue
		}

		var toTrigger []Issue
		last := lastSeen[k]
		found := false
		for i := 0; i < len(issues); i++ {
			curr := fmt.Sprintf("%d-%s", issues[i].Number, issues[i].UpdatedAt.Format(time.RFC3339Nano))
			if curr == last {
				found = true
				break
			}
			toTrigger = append(toTrigger, issues[i])
		}
		if !found && lastSeen[k] != "" && len(toTrigger) > 0 {
			lastSeen[k] = fmt.Sprintf("%d-%s", issues[0].Number, issues[0].UpdatedAt.Format(time.RFC3339Nano))
			continue
		}
		for i := len(toTrigger) - 1; i >= 0; i-- {
			iss := toTrigger[i]
			action := "opened"
			if iss.State == "closed" {
				action = "closed"
			}
			if len(cfg.Actions) > 0 && !containsString(cfg.Actions, action) {
				continue
			}
			payload := map[string]any{
				"repo":    cfg.Repo,
				"number":  iss.Number,
				"title":   iss.Title,
				"state":   iss.State,
				"action":  action,
				"author":  iss.Author,
				"url":     iss.URL,
				"updated": iss.UpdatedAt.Format(time.RFC3339),
			}
			for k2, v := range cfg.PayloadTemplate {
				payload[k2] = v
			}
			if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
				payload["content"] = fmt.Sprintf("Issue #%d %s on %s: %s", iss.Number, action, cfg.Repo, iss.Title)
			}
			if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
				log.Printf("github poller trigger issue wf %d: %v", wf.ID, err)
				continue
			}
		}
		lastSeen[k] = fmt.Sprintf("%d-%s", issues[0].Number, issues[0].UpdatedAt.Format(time.RFC3339Nano))
	}
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}
