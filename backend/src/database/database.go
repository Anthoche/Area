package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sqlDB *sql.DB
var db *gorm.DB
var dbContext context.Context

// Temporaire, à remplacer par des secrets (.env)
const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "area"
)

func Connect() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
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
}

func Disconnect() {
	if sqlDB != nil && IsConnected() {
		sqlDB.Close()
	}
}

func IsConnected() bool {
	err := sqlDB.Ping()
	return err == nil
}

func GetDB() *gorm.DB {
	return db
}

func getDBContext() context.Context {
	return dbContext
}
