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

	CREATE TABLE IF NOT EXISTS game_sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		game_id VARCHAR(100) NOT NULL,
		provider_id VARCHAR(50) NOT NULL,
		start_balance DECIMAL(15, 2) NOT NULL,
		end_balance DECIMAL(15, 2),
		status VARCHAR(20) DEFAULT 'ACTIVE',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ended_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_rounds (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		session_id VARCHAR(100),
		round_id VARCHAR(100) UNIQUE NOT NULL,
		bet DECIMAL(15, 2) NOT NULL,
		win DECIMAL(15, 2) DEFAULT 0,
		status VARCHAR(20) DEFAULT 'PENDING',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS games (
		id VARCHAR(100) PRIMARY KEY,
		provider_id VARCHAR(50) NOT NULL,
		name VARCHAR(200) NOT NULL,
		type VARCHAR(50) NOT NULL,
		thumbnail_url TEXT,
		min_bet DECIMAL(15, 2) DEFAULT 10,
		max_bet DECIMAL(15, 2) DEFAULT 10000,
		rtp DECIMAL(5, 2) DEFAULT 96.00,
		is_active BOOLEAN DEFAULT TRUE
	);

	-- Seed some demo games
	INSERT INTO games (id, provider_id, name, type, thumbnail_url, min_bet, max_bet, rtp) VALUES
	('evolution-roulette', 'EVOLUTION', 'Live Roulette', 'LIVE_CASINO', 'https://picsum.photos/seed/roulette/300/200', 10, 5000, 97.30),
	('evolution-blackjack', 'EVOLUTION', 'Live Blackjack', 'LIVE_CASINO', 'https://picsum.photos/seed/blackjack/300/200', 10, 2500, 99.50),
	('pragmatic-wolf-gold', 'PRAGMATIC', 'Wolf Gold', 'SLOT', 'https://picsum.photos/seed/wolfgold/300/200', 5, 1000, 96.01),
	('pragmatic-sweet-bonanza', 'PRAGMATIC', 'Sweet Bonanza', 'SLOT', 'https://picsum.photos/seed/bonanza/300/200', 5, 2000, 96.48),
	('ezugi-baccarat', 'EZUGI', 'Live Baccarat', 'LIVE_CASINO', 'https://picsum.photos/seed/baccarat/300/200', 25, 10000, 98.94)
	ON CONFLICT (id) DO NOTHING;
	`

	_, err := DB.Exec(schema)

	if err != nil {
		log.Fatal("Failed to create schema: ", err)
	}
	log.Println("Database schema initialized!")
}
