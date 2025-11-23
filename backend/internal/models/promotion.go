package models

import (
	"time"
)

type Bonus struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Type                 string    `json:"type"` // WELCOME, REFERRAL, DAILY
	Amount               float64   `json:"amount"`
	WageringRequirement  float64   `json:"wagering_requirement"`
	Wagered              float64   `json:"wagered"`
	Status               string    `json:"status"` // ACTIVE, COMPLETED, EXPIRED
	ExpiresAt            time.Time `json:"expires_at"`
	CreatedAt            time.Time `json:"created_at"`
}

type Referral struct {
	ID          string    `json:"id"`
	ReferrerID  string    `json:"referrer_id"`
	ReferredID  string    `json:"referred_id"`
	Code        string    `json:"code"`
	BonusAwarded float64  `json:"bonus_awarded"`
	CreatedAt   time.Time `json:"created_at"`
}
