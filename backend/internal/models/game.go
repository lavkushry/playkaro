package models

import (
	"time"
)

type GameProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // SLOTS, LIVE_CASINO, TABLE_GAMES
	LogoURL     string `json:"logo_url"`
	IsActive    bool   `json:"is_active"`
}

type Game struct {
	ID           string  `json:"id"`
	ProviderID   string  `json:"provider_id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"` // SLOT, ROULETTE, BLACKJACK, BACCARAT
	ThumbnailURL string  `json:"thumbnail_url"`
	MinBet       float64 `json:"min_bet"`
	MaxBet       float64 `json:"max_bet"`
	RTP          float64 `json:"rtp"` // Return to Player %
	IsActive     bool    `json:"is_active"`
}

type GameSession struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	GameID       string    `json:"game_id"`
	ProviderID   string    `json:"provider_id"`
	StartBalance float64   `json:"start_balance"`
	EndBalance   float64   `json:"end_balance"`
	Status       string    `json:"status"` // ACTIVE, ENDED
	CreatedAt    time.Time `json:"created_at"`
	EndedAt      time.Time `json:"ended_at"`
}

type GameRound struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	RoundID     string    `json:"round_id"` // Provider's round ID
	Bet         float64   `json:"bet"`
	Win         float64   `json:"win"`
	Status      string    `json:"status"` // PENDING, COMPLETED, CANCELLED
	CreatedAt   time.Time `json:"created_at"`
}
