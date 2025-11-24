package models

import "time"

// Transaction Types
const (
	TxTypeDeposit    = "DEPOSIT"
	TxTypeWithdrawal = "WITHDRAWAL"
	TxTypeBet        = "BET"
	TxTypeWin        = "WIN"
	TxTypeRefund     = "REFUND"
)

// Wallet represents a user's balance
type Wallet struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Balance   float64   `json:"balance" db:"balance"`
	Currency  string    `json:"currency" db:"currency"` // Defaults to "PTS"
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// LedgerEntry represents a single financial movement (Double Entry Bookkeeping)
type LedgerEntry struct {
	ID            string    `json:"id" db:"id"`
	TransactionID string    `json:"transaction_id" db:"transaction_id"` // Idempotency Key
	UserID        string    `json:"user_id" db:"user_id"`
	Type          string    `json:"type" db:"type"`
	Amount        float64   `json:"amount" db:"amount"` // Negative for debit, Positive for credit
	ReferenceID   string    `json:"reference_id" db:"reference_id"` // e.g., MatchID or GameSessionID
	ReferenceType string    `json:"reference_type" db:"reference_type"` // e.g., "GAME_CRASH", "MATCH_CRICKET"
	BalanceAfter  float64   `json:"balance_after" db:"balance_after"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
