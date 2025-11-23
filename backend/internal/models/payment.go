package models

import (
	"time"
)

type PaymentTransaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Gateway     string    `json:"gateway"` // RAZORPAY, CASHFREE, MOCK
	OrderID     string    `json:"order_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"` // PENDING, SUCCESS, FAILED
	Method      string    `json:"method"` // UPI, CARD, NETBANKING
	ReferenceID string    `json:"reference_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type KYCDocument struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DocumentType string    `json:"document_type"` // PAN, AADHAAR, PASSPORT
	DocumentURL  string    `json:"document_url"`
	Status       string    `json:"status"` // PENDING, APPROVED, REJECTED
	ReviewedBy   string    `json:"reviewed_by"`
	Remarks      string    `json:"remarks"`
	CreatedAt    time.Time `json:"created_at"`
}
