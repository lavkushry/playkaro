package models

import "time"

// Transaction Types
const (
	TxTypeDeposit    = "DEPOSIT"
	TxTypeWithdrawal = "WITHDRAWAL"
	TxTypeBet        = "BET"
	TxTypeWin        = "WIN"
	TxTypeRefund     = "REFUND"
	TxTypeBonus      = "BONUS"
)

// Transaction States
const (
	TxStatePending    = "PENDING"
	TxStateProcessing = "PROCESSING"
	TxStateSettled    = "SETTLED"
	TxStateFailed     = "FAILED"
	TxStateReversed   = "REVERSED"
)

// Balance Types
const (
	BalanceTypeDeposit  = "DEPOSIT"
	BalanceTypeBonus    = "BONUS"
	BalanceTypeWinnings = "WINNINGS"
)

// Wallet represents a user's balance with multi-currency support
type Wallet struct {
	UserID          string    `json:"user_id" db:"user_id"`
	Balance         float64   `json:"balance" db:"balance"`                   // Total balance (sum of all)
	DepositBalance  float64   `json:"deposit_balance" db:"deposit_balance"`   // Real money deposits
	BonusBalance    float64   `json:"bonus_balance" db:"bonus_balance"`       // Promotional bonuses
	WinningsBalance float64   `json:"winnings_balance" db:"winnings_balance"` // Game winnings
	Currency        string    `json:"currency" db:"currency"`                 // Defaults to "PTS"
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// LedgerEntry represents a single financial movement (Double Entry Bookkeeping)
type LedgerEntry struct {
	ID            string    `json:"id" db:"id"`
	TransactionID string    `json:"transaction_id" db:"transaction_id"` // Idempotency Key
	UserID        string    `json:"user_id" db:"user_id"`
	Type          string    `json:"type" db:"type"`
	Amount        float64   `json:"amount" db:"amount"` // Negative for debit, Positive for credit
	BalanceType   string    `json:"balance_type" db:"balance_type"` // Which balance was affected
	ReferenceID   string    `json:"reference_id" db:"reference_id"` // e.g., MatchID or GameSessionID
	ReferenceType string    `json:"reference_type" db:"reference_type"` // e.g., "GAME_CRASH", "MATCH_CRICKET"
	BalanceAfter  float64   `json:"balance_after" db:"balance_after"`
	State         string    `json:"state" db:"state"` // Transaction state
	Metadata      string    `json:"metadata,omitempty" db:"metadata"` // JSON metadata for audit trail
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
