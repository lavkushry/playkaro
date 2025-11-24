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
	host := getEnv("PAYMENT_DB_HOST", "localhost")
	port := getEnv("PAYMENT_DB_PORT", "5432")
	user := getEnv("PAYMENT_DB_USER", "postgres")
	password := getEnv("PAYMENT_DB_PASSWORD", "postgres")
	dbname := getEnv("PAYMENT_DB_NAME", "payments_db")

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

	log.Println("Successfully connected to Payment Service database!")
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
