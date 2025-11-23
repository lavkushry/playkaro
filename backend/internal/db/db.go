package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to open database connection: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	log.Println("Successfully connected to PostgreSQL!")

	InitSchema()
}

func InitSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		mobile VARCHAR(20),
		is_admin BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS wallets (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		balance DECIMAL(15, 2) DEFAULT 0.00,
		currency VARCHAR(3) DEFAULT 'INR',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_user_wallet UNIQUE (user_id)
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
		type VARCHAR(20) NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		status VARCHAR(20) DEFAULT 'PENDING',
		reference_id VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS matches (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		team_a VARCHAR(100) NOT NULL,
		team_b VARCHAR(100) NOT NULL,
		odds_a DECIMAL(5, 2) NOT NULL,
		odds_b DECIMAL(5, 2) NOT NULL,
		status VARCHAR(20) DEFAULT 'LIVE',
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bets (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		match_id UUID REFERENCES matches(id) ON DELETE CASCADE,
		selection VARCHAR(10) NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		odds DECIMAL(5, 2) NOT NULL,
		potential_win DECIMAL(15, 2) NOT NULL,
		status VARCHAR(20) DEFAULT 'PENDING',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Seed some data if empty
	INSERT INTO matches (team_a, team_b, odds_a, odds_b)
	SELECT 'India', 'Australia', 1.80, 2.10
	WHERE NOT EXISTS (SELECT 1 FROM matches);

	CREATE TABLE IF NOT EXISTS payment_transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		gateway VARCHAR(50) NOT NULL,
		order_id VARCHAR(100),
		amount DECIMAL(15, 2) NOT NULL,
		currency VARCHAR(3) DEFAULT 'INR',
		status VARCHAR(20) DEFAULT 'PENDING',
		method VARCHAR(50),
		reference_id VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS kyc_documents (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		document_type VARCHAR(20) NOT NULL,
		document_url TEXT NOT NULL,
		status VARCHAR(20) DEFAULT 'PENDING',
		reviewed_by UUID REFERENCES users(id),
		remarks TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS kyc_level INT DEFAULT 0;
	`

	_, err := DB.Exec(schema)

	if err != nil {
		log.Fatal("Failed to create schema: ", err)
	}
	log.Println("Database schema initialized!")
}
