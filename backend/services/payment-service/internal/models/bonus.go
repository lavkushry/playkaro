package models

import "time"

// Bonus Status
const (
	BonusStatusActive  = "ACTIVE"
	BonusStatusUsed    = "USED"
	BonusStatusExpired = "EXPIRED"
)

// Bonus represents a promotional bonus granted to a user
type Bonus struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Status    string    `json:"status" db:"status"` // ACTIVE, USED, EXPIRED
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
