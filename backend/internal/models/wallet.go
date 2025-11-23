package models

import (
	"time"
)

type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Type        string    `json:"type"` // DEPOSIT, WITHDRAW, BET, WIN
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // PENDING, COMPLETED, FAILED
	ReferenceID string    `json:"reference_id"`
	CreatedAt   time.Time `json:"created_at"`
}
