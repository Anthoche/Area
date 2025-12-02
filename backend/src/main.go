package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"area/src/auth"
	"area/src/database"
	"area/src/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database.Connect()
	defer database.Disconnect()
	store := auth.NewDBStore() // PostgreSQL-backed user store
	service := auth.NewService(store)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewMux(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
