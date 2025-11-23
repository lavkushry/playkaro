package models

import (
	"time"
)

type Match struct {
	ID        string    `json:"id"`
	TeamA     string    `json:"team_a"`
	TeamB     string    `json:"team_b"`
	OddsA     float64   `json:"odds_a"`
	OddsB     float64   `json:"odds_b"`
	Status    string    `json:"status"` // LIVE, FINISHED
	StartTime time.Time `json:"start_time"`
}

type Bet struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	MatchID      string    `json:"match_id"`
	Selection    string    `json:"selection"` // TEAM_A, TEAM_B
	Amount       float64   `json:"amount"`
	Odds         float64   `json:"odds"`
	PotentialWin float64   `json:"potential_win"`
	Status       string    `json:"status"` // PENDING, WON, LOST
	CreatedAt    time.Time `json:"created_at"`
}
