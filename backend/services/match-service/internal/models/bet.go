package models

import "time"

// Bet Status Constants
const (
	BetStatusPending   = "PENDING"
	BetStatusActive    = "ACTIVE"
	BetStatusSettled   = "SETTLED"
	BetStatusRejected  = "REJECTED"
	BetStatusCashedOut = "CASHED_OUT"
)

// Bet Result Constants
const (
	BetResultWon  = "WON"
	BetResultLost = "LOST"
	BetResultVoid = "VOID"
)

// Bet represents a user's bet on a match
type Bet struct {
	ID            string     `json:"id" db:"id"`
	UserID        string     `json:"user_id" db:"user_id"`
	MatchID       string     `json:"match_id" db:"match_id"`
	Team          string     `json:"team" db:"team"` // "TEAM_A", "TEAM_B", "DRAW"
	Amount        float64    `json:"amount" db:"amount"`
	Odds          float64    `json:"odds" db:"odds"`
	PotentialWin  float64    `json:"potential_win" db:"potential_win"`
	Status        string     `json:"status" db:"status"`
	Result        *string    `json:"result,omitempty" db:"result"` // WON, LOST, VOID
	CashedOut     bool       `json:"cashed_out" db:"cashed_out"`
	CashOutAt     *time.Time `json:"cash_out_at,omitempty" db:"cash_out_at"`
	CashOutOdds   *float64   `json:"cash_out_odds,omitempty" db:"cash_out_odds"`
	CashOutAmount *float64   `json:"cash_out_amount,omitempty" db:"cash_out_amount"`
	SettledAt     *time.Time `json:"settled_at,omitempty" db:"settled_at"`
	Version       int        `json:"version" db:"version"` // For optimistic locking
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
