package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"area/src/auth"
	"area/src/database"
	"area/src/httpapi"
	"area/src/workflows"

	"github.com/joho/godotenv"
)

type httpSender struct {
	client *http.Client
}

// newHTTPSender builds an httpSender with a default timeout.
func newHTTPSender() *httpSender {
	return &httpSender{client: &http.Client{Timeout: 10 * time.Second}}
}

// Send posts the given payload as JSON to the target URL.
func (s *httpSender) Send(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http sender: status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// main boots the API server, background workers, and graceful shutdown handling.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file, ignoring it.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database.Connect()
	defer database.Disconnect()
	userStore := auth.NewDBStore()
	authService := auth.NewService(userStore)

	wfStore := workflows.NewDefaultStore()
	triggerer := workflows.NewTriggerer(wfStore)
	wfService := workflows.NewService(wfStore, triggerer)

	// Start a simple executor loop in background for outgoing webhooks.
	sender := newHTTPSender()
	executor := workflows.NewExecutor(wfStore, sender, 2*time.Second)
	go executor.RunLoop(context.Background())

	// Interval scheduler: triggers workflows of type "interval" when due.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			due, err := wfStore.ClaimDueIntervalWorkflows(context.Background(), now)
			if err != nil {
				log.Printf("scheduler: %v", err)
				continue
			}
			for _, wf := range due {
				ctx := workflows.WithUserID(context.Background(), wf.UserID)
				cfg, err := workflows.IntervalConfigFromJSON(wf.TriggerConfig)
				payload := map[string]any{}
				if err == nil && len(cfg.Payload) > 0 {
					for k, v := range cfg.Payload {
						payload[k] = v
					}
				}
				_, err = wfService.Trigger(ctx, wf.ID, payload)
				if err != nil {
					log.Printf("scheduler trigger wf %d: %v", wf.ID, err)
				}
			}
		}
	}()

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewMux(authService, wfService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on http://0.0.0.0:%s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server")
	if err := server.Close(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
