package models

import (
	"time"
)

type Wallet struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	DepositBalance   float64   `json:"deposit_balance"`
	BonusBalance     float64   `json:"bonus_balance"`
	WinningsBalance  float64   `json:"winnings_balance"`
	LockedBalance    float64   `json:"locked_balance"`
	Currency         string    `json:"currency"`
	KYCLevel         int       `json:"kyc_level"`
	DailyDepositUsed float64   `json:"daily_deposit_used"`
	LastDepositReset time.Time `json:"last_deposit_reset"`
	Status           string    `json:"status"`
	UpdatedAt        time.Time `json:"updated_at"`
	// Backwards compatibility for existing clients
	Balance float64 `json:"balance,omitempty"`
}

type Transaction struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Type        string    `json:"type"` // DEPOSIT, WITHDRAW, BET, WIN
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // PENDING, COMPLETED, FAILED
	ReferenceID string    `json:"reference_id"`
	Bucket      string    `json:"bucket"`
	CreatedAt   time.Time `json:"created_at"`
}
