package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func Connect() {
	host := mustEnv("POSTGRES_HOST")
	port := mustEnv("POSTGRES_PORT")
	user := mustEnv("POSTGRES_USER")
	password := mustEnv("POSTGRES_PASSWORD")
	dbname := mustEnv("POSTGRES_DB")
	sslmode := firstNonEmpty(os.Getenv("POSTGRES_SSLMODE"), "disable")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
}

func Disconnect() {
	if db != nil && IsConnected() {
		db.Close()
	}
}

func IsConnected() bool {
	err := db.Ping()
	return err == nil
}

// DB exposes the shared sql.DB handle for packages that need direct queries.
func GetDB() *sql.DB {
	return db
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required env var %s", key)
	}
	return val
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
