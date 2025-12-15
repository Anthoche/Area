package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sqlDB *sql.DB
var db *gorm.DB
var dbContext context.Context

// Connect initializes the shared PostgreSQL connection using environment variables.
func Connect() {
	host := mustEnv("POSTGRES_HOST")
	port := mustEnv("POSTGRES_PORT")
	user := mustEnv("POSTGRES_USER")
	password := mustEnv("POSTGRES_PASSWORD")
	dbname := mustEnv("POSTGRES_DB")
	sslmode := firstNonEmpty(os.Getenv("POSTGRES_SSLMODE"), "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err = db.DB()
	if err != nil {
		log.Fatal(err)
	}

	dbContext = context.Background()
	db.AutoMigrate(&User{})
	db.AutoMigrate(&GoogleToken{})
	db.AutoMigrate(&Job{})
	db.AutoMigrate(&Run{})
	db.AutoMigrate(&Workflow{})
}

// Disconnect closes the database connection if it is open.
func Disconnect() {
	if sqlDB != nil && IsConnected() {
		sqlDB.Close()
	}
}

// IsConnected reports whether the database connection is alive.
func IsConnected() bool {
	err := sqlDB.Ping()
	return err == nil
}

// GetDB exposes the shared sql.DB handle for packages that need direct queries.
func GetDB() *gorm.DB {
	return db
}

// SetDBForTesting allows tests to inject a mock database instance.
// This should only be used in test code.
func SetDBForTesting(testDB *gorm.DB) {
	db = testDB
}

// GetDBContext TODO: doc
func GetDBContext() context.Context {
	return dbContext
}

// mustEnv reads an environment variable or exits if it is missing.
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required env var %s", key)
	}
	return val
}

// firstNonEmpty returns the first non-empty string in the provided list.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
