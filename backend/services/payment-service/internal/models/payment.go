package models

import "time"

type PaymentOrder struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	OrderID        string    `json:"order_id" db:"order_id"`
	Gateway        string    `json:"gateway" db:"gateway"`
	GatewayOrderID string    `json:"gateway_order_id" db:"gateway_order_id"`
	Amount         float64   `json:"amount" db:"amount"`
	Currency       string    `json:"currency" db:"currency"`
	Type           string    `json:"type" db:"type"`
	Status         string    `json:"status" db:"status"`
	PaymentMethod  string    `json:"payment_method" db:"payment_method"`
	Metadata       string    `json:"metadata" db:"metadata"` // JSONB stored as string
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
}

type PaymentTransaction struct {
	ID            string    `json:"id" db:"id"`
	OrderID       string    `json:"order_id" db:"order_id"`
	GatewayTxnID  string    `json:"gateway_txn_id" db:"gateway_txn_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Fee           float64   `json:"fee" db:"fee"`
	Tax           float64   `json:"tax" db:"tax"`
	NetAmount     float64   `json:"net_amount" db:"net_amount"`
	Reconciled    bool      `json:"reconciled" db:"reconciled"`
	ReconciledAt  *time.Time `json:"reconciled_at" db:"reconciled_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type FraudCheck struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CheckType string    `json:"check_type" db:"check_type"`
	RiskScore int       `json:"risk_score" db:"risk_score"`
	Flagged   bool      `json:"flagged" db:"flagged"`
	Details   string    `json:"details" db:"details"` // JSONB stored as string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Payment status constants
const (
	StatusInitiated = "INITIATED"
	StatusPending   = "PENDING"
	StatusSuccess   = "SUCCESS"
	StatusFailed    = "FAILED"
	StatusRefunded  = "REFUNDED"
)

// Payment types
const (
	TypeDeposit    = "DEPOSIT"
	TypeWithdrawal = "WITHDRAWAL"
)

// Gateways
const (
	GatewayRazorpay = "razorpay"
	GatewayCashfree = "cashfree"
)
