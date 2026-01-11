package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sqlDB *sql.DB
var Db *gorm.DB
var dbContext context.Context

// Connect initializes the shared PostgreSQL connection using environment variables.
func Connect() {
	host := mustEnv("POSTGRES_HOST")
	port := mustEnv("POSTGRES_PORT")
	user := mustEnv("POSTGRES_USER")
	password := mustEnv("POSTGRES_PASSWORD")
	dbname := mustEnv("POSTGRES_DB")
	sslmode := FirstNonEmpty(os.Getenv("POSTGRES_SSLMODE"), "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	var err error

	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err = Db.DB()
	if err != nil {
		log.Fatal(err)
	}

	dbContext = context.Background()
	Db.AutoMigrate(&User{})
	Db.AutoMigrate(&GoogleToken{})
	Db.AutoMigrate(&GithubToken{})
	Db.AutoMigrate(&AreaService{})
	Db.AutoMigrate(&AreaCapability{})
	Db.AutoMigrate(&AreaField{})
	Db.AutoMigrate(&Job{})
	Db.AutoMigrate(&Run{})
	Db.AutoMigrate(&Workflow{})
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
	return Db
}

// SetDBForTesting allows tests to inject a mock database instance.
func SetDBForTesting(testDB *gorm.DB) {
	Db = testDB
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

// FirstNonEmpty returns the first non-empty string in the provided list.
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
