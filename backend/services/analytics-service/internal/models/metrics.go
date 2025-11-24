package models

import (
	"encoding/json"
	"time"
)

// Event Types
const (
	EventTypeGameEnd   = "GAME_END"
	EventTypeDeposit   = "DEPOSIT"
	EventTypeWithdraw  = "WITHDRAW"
	EventTypeLogin     = "LOGIN"
	EventTypeBetPlaced = "BET_PLACED"
	EventTypePayout    = "PAYOUT"
)

// AnalyticsEvent represents a raw event
type AnalyticsEvent struct {
	ID        string          `json:"id" db:"id"`
	UserID    string          `json:"user_id" db:"user_id"`
	EventType string          `json:"event_type" db:"event_type"`
	EventData json.RawMessage `json:"event_data" db:"event_data"`
	Timestamp time.Time       `json:"timestamp" db:"created_at"`
}

// UserMetrics represents daily user stats
type UserMetrics struct {
	UserID           string    `json:"user_id" db:"user_id"`
	Date             time.Time `json:"date" db:"date"`
	TotalDeposits    float64   `json:"total_deposits" db:"total_deposits"`
	TotalWithdrawals float64   `json:"total_withdrawals" db:"total_withdrawals"`
	TotalWagers      float64   `json:"total_wagers" db:"total_wagers"`
	TotalPayouts     float64   `json:"total_payouts" db:"total_payouts"`
	SessionCount     int       `json:"session_count" db:"session_count"`
	LastActive       time.Time `json:"last_active" db:"last_active"`
}

// GameMetrics represents daily game stats
type GameMetrics struct {
	GameType      string    `json:"game_type" db:"game_type"`
	Date          time.Time `json:"date" db:"date"`
	TotalRounds   int       `json:"total_rounds" db:"total_rounds"`
	TotalWagers   float64   `json:"total_wagers" db:"total_wagers"`
	TotalPayouts  float64   `json:"total_payouts" db:"total_payouts"`
	UniquePlayers int       `json:"unique_players" db:"unique_players"`
}

// RevenueStats represents real-time revenue
type RevenueStats struct {
	GGR           float64 `json:"ggr"` // Gross Gaming Revenue (Total Bets)
	NGR           float64 `json:"ngr"` // Net Gaming Revenue (Bets - Payouts)
	ActiveUsers   int64   `json:"active_users"`
	DepositVolume float64 `json:"deposit_volume"`
}
