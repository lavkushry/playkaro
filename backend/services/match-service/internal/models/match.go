package models

import "time"

type Match struct {
	ID         string    `json:"id" db:"id"`
	MatchID    string    `json:"match_id" db:"match_id"`
	Sport      string    `json:"sport" db:"sport"`
	TeamA      string    `json:"team_a" db:"team_a"`
	TeamB      string    `json:"team_b" db:"team_b"`
	OddsA      float64   `json:"odds_a" db:"odds_a"`
	OddsB      float64   `json:"odds_b" db:"odds_b"`
	OddsDraw   float64   `json:"odds_draw" db:"odds_draw"`
	Status     string    `json:"status" db:"status"`
	StartTime  time.Time `json:"start_time" db:"start_time"`
	League     string    `json:"league" db:"league"`
	Venue      string    `json:"venue" db:"venue"`
	Result     string    `json:"result" db:"result"`
	Metadata   string    `json:"metadata" db:"metadata"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	SettledAt  *time.Time `json:"settled_at" db:"settled_at"`
}

type OddsHistory struct {
	ID        int       `json:"id" db:"id"`
	MatchID   string    `json:"match_id" db:"match_id"`
	OddsA     float64   `json:"odds_a" db:"odds_a"`
	OddsB     float64   `json:"odds_b" db:"odds_b"`
	OddsDraw  float64   `json:"odds_draw" db:"odds_draw"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type Market struct {
	ID         string    `json:"id" db:"id"`
	MatchID    string    `json:"match_id" db:"match_id"`
	MarketType string    `json:"market_type" db:"market_type"`
	MarketName string    `json:"market_name" db:"market_name"`
	Status     string    `json:"status" db:"status"`
	Metadata   string    `json:"metadata" db:"metadata"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Match status constants
const (
	StatusUpcoming  = "UPCOMING"
	StatusLive      = "LIVE"
	StatusCompleted = "COMPLETED"
	StatusCancelled = "CANCELLED"
	StatusSuspended = "SUSPENDED"
)

// Sports constants
const (
	SportCricket   = "CRICKET"
	SportFootball  = "FOOTBALL"
	SportTennis    = "TENNIS"
	SportBasketball = "BASKETBALL"
)

// Result constants
const (
	ResultTeamA = "TEAM_A"
	ResultTeamB = "TEAM_B"
	ResultDraw  = "DRAW"
)
