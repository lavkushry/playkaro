package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
	host := getEnv("MATCH_DB_HOST", "localhost")
	port := getEnv("MATCH_DB_PORT", "5432")
	user := getEnv("MATCH_DB_USER", "postgres")
	password := getEnv("MATCH_DB_PASSWORD", "postgres")
	dbname := getEnv("MATCH_DB_NAME", "matches_db")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}

	log.Println("Successfully connected to Match Service database!")
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
