package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Temporaire, à remplacer par des secrets (.env)
const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "area"
)

func Connect() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
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
