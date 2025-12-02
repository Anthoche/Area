package main

import (
	"bytes"
	"context"
	"fmt"
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
)

type httpSender struct {
	client *http.Client
}

func newHTTPSender() *httpSender {
	return &httpSender{client: &http.Client{Timeout: 10 * time.Second}}
}

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
		return fmt.Errorf("http sender: status %d", resp.StatusCode)
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database.Connect()
	defer database.Disconnect()
	userStore := auth.NewDBStore() // PostgreSQL-backed user store
	authService := auth.NewService(userStore)

	wfStore := workflows.NewDefaultStore()
	triggerer := workflows.NewTriggerer(wfStore)
	wfService := workflows.NewService(wfStore, triggerer)

	// Start a simple executor loop in background for outgoing webhooks.
	sender := newHTTPSender()
	executor := workflows.NewExecutor(wfStore, sender, 2*time.Second)
	go executor.RunLoop(context.Background())

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
