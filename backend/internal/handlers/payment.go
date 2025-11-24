package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/wallet"
)

type InitiateDepositRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Method  string  `json:"method" binding:"required"` // UPI, CARD, NETBANKING
	Gateway string  `json:"gateway" binding:"required"` // RAZORPAY, MOCK
}

type RazorpayWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID       string  `json:"id"`
				OrderID  string  `json:"order_id"`
				Amount   int     `json:"amount"` // Amount in paise
				Currency string  `json:"currency"`
				Status   string  `json:"status"`
				Method   string  `json:"method"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

// InitiateDeposit creates a payment order
func InitiateDeposit(c *gin.Context) {
	userID := c.GetString("userID")
	var req InitiateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := wallet.NewService(db.DB, db.RDB)

	// Create transaction record
	var txnID string
	err := db.DB.QueryRow(
		`INSERT INTO payment_transactions (user_id, gateway, amount, currency, status, method)
		VALUES ($1, $2, $3, 'INR', 'PENDING', $4)
		RETURNING id`,
		userID, req.Gateway, req.Amount, req.Method,
	).Scan(&txnID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Mock Gateway (for testing)
	if req.Gateway == "MOCK" {
		// Auto-approve for testing
		orderID := "mock_" + txnID
		db.DB.Exec("UPDATE payment_transactions SET order_id=$1, status='SUCCESS' WHERE id=$2", orderID, txnID)

		service.Deposit(c.Request.Context(), userID, req.Amount, orderID)

		c.JSON(http.StatusOK, gin.H{
			"transaction_id": txnID,
			"order_id":       orderID,
			"status":         "SUCCESS",
			"message":        "Mock payment successful",
		})
		return
	}

	// Razorpay Integration (requires Razorpay SDK)
	// For now, return mock response
	c.JSON(http.StatusOK, gin.H{
		"transaction_id": txnID,
		"payment_url":    "https://mock-payment-gateway.com/pay/" + txnID,
		"order_id":       "rzp_order_" + txnID,
	})
}

// RazorpayWebhook handles payment status updates
func RazorpayWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read request body"})
		return
	}

	// Restore the body for subsequent reads (e.g., by c.ShouldBindJSON)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var payload RazorpayWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("X-Razorpay-Signature")
	secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")

	if secret != "" && !verifyWebhookSignature(body, signature, secret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	if payload.Event == "payment.captured" {
		orderID := payload.Payload.Payment.Entity.OrderID
		amount := float64(payload.Payload.Payment.Entity.Amount) / 100 // Convert paise to rupees

		// Update transaction status
		var userID string
		err := db.DB.QueryRow(
			"UPDATE payment_transactions SET status='SUCCESS', reference_id=$1 WHERE order_id=$2 RETURNING user_id",
			payload.Payload.Payment.Entity.ID, orderID,
		).Scan(&userID)

		if err == nil {
			service := wallet.NewService(db.DB, db.RDB)
			service.Deposit(c.Request.Context(), userID, amount, orderID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// InitiateWithdrawal processes payout requests
func InitiateWithdrawal(c *gin.Context) {
	userID := c.GetString("userID")
	var req struct {
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		BankAccountID string  `json:"bank_account_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check KYC status
	var kycLevel int
	db.DB.QueryRow("SELECT COALESCE(kyc_level, 0) FROM users WHERE id=$1", userID).Scan(&kycLevel)

	if kycLevel < 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "KYC verification required for withdrawals"})
		return
	}

	// Create withdrawal request (PENDING approval)
	var txnID string
	db.DB.QueryRow(
		"INSERT INTO payment_transactions (user_id, gateway, amount, currency, status, method) VALUES ($1, 'PAYOUT', $2, 'INR', 'PENDING', 'BANK_TRANSFER') RETURNING id",
		userID, req.Amount,
	).Scan(&txnID)

	// Deduct balance (held until approved)
	service := wallet.NewService(db.DB, db.RDB)
	if _, err := service.Withdraw(c.Request.Context(), userID, req.Amount, "PAYOUT-"+txnID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": txnID,
		"status":         "PENDING",
		"message":        "Withdrawal request submitted for approval",
	})
}

// Helper: Verify webhook signature
func verifyWebhookSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
