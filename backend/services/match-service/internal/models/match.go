package models

import "time"

const (
	StatusUpcoming  = "UPCOMING"
	StatusLive      = "LIVE"
	StatusCompleted = "COMPLETED"

	ResultTeamA = "TEAM_A"
	ResultTeamB = "TEAM_B"
	ResultDraw  = "DRAW"
)

type Match struct {
	ID        string    `json:"id" db:"id"`
	MatchID   string    `json:"match_id" db:"match_id"`
	Sport     string    `json:"sport" db:"sport"`
	TeamA     string    `json:"team_a" db:"team_a"`
	TeamB     string    `json:"team_b" db:"team_b"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	Status    string    `json:"status" db:"status"` // SCHEDULED, LIVE, COMPLETED
	OddsA     float64   `json:"odds_a" db:"odds_a"`
	OddsB     float64   `json:"odds_b" db:"odds_b"`
	OddsDraw  float64   `json:"odds_draw" db:"odds_draw"`
	League    string    `json:"league" db:"league"`
	Venue     string    `json:"venue" db:"venue"`
	Result    *string   `json:"result,omitempty" db:"result"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
