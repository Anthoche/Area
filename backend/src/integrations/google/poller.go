package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"area/src/workflows"
)

// StartGmailPoller launches a background loop that polls Gmail for inbound messages
// and triggers workflows of type "gmail_inbound".
func StartGmailPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, client *Client) {
	lastSeen := make(map[int64]string) // workflowID -> last message id (in-memory)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("gmail poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "gmail_inbound")
			if err != nil {
				log.Printf("gmail poller: list workflows: %v", err)
				continue
			}
			for _, wf := range wfs {
				if !wf.Enabled {
					continue
				}
				var cfg struct {
					TokenID int64 `json:"token_id"`
				}
				_ = json.Unmarshal(wf.TriggerConfig, &cfg)

				lastID := lastSeen[wf.ID]
				if lastID == "" {
					msgs, err := client.ListRecentMessages(ctx, &wf.UserID, cfg.TokenID, 1, "")
					if err != nil {
						log.Printf("gmail poller wf %d (init cursor): %v", wf.ID, err)
						continue
					}
					if len(msgs) > 0 {
						lastSeen[wf.ID] = msgs[0].ID
					}
					continue
				}

				msgs, err := client.ListRecentMessages(ctx, &wf.UserID, cfg.TokenID, 5, lastID)
				if err != nil {
					log.Printf("gmail poller wf %d: %v", wf.ID, err)
					continue
				}
				for i := len(msgs) - 1; i >= 0; i-- {
					msg := msgs[i]
					content := fmt.Sprintf("From: %s\nSubject: %s\nSnippet: %s", msg.From, msg.Subject, msg.Snippet)
					payload := map[string]any{
						"from":    msg.From,
						"subject": msg.Subject,
						"snippet": msg.Snippet,
						"date":    msg.Date,
						"id":      msg.ID,
						"content": content,
					}
					if len(wf.TriggerConfig) > 0 {
						var cfgMap map[string]any
						if err := json.Unmarshal(wf.TriggerConfig, &cfgMap); err == nil {
							if tpl, ok := cfgMap["payload_template"].(map[string]any); ok {
								for k, v := range tpl {
									payload[k] = v
								}
								if userContent, ok := tpl["content"].(string); ok && userContent != "" {
									payload["content"] = userContent + "\n\n" + content
								}
							}
						}
					}
					_, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload)
					if err != nil {
						log.Printf("gmail poller trigger wf %d: %v", wf.ID, err)
						continue
					}
					lastSeen[wf.ID] = msg.ID
				}
			}
		}
	}()
}
