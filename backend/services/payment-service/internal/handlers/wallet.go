package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/payment-service/internal/models"
)

// TransactionRequest for internal microservice calls
type TransactionRequest struct {
	UserID        string  `json:"user_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"` // Positive value
	Type          string  `json:"type" binding:"required"`   // BET, WIN, REFUND
	TransactionID string  `json:"transaction_id" binding:"required"`
	ReferenceID   string  `json:"reference_id" binding:"required"`
	ReferenceType string  `json:"reference_type" binding:"required"`
}

// GetBalance returns the current balance
func (h *PaymentHandler) GetBalance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		// Allow internal calls to pass UserID via query param if secured
		userID = c.Query("user_id")
	}

	var balance float64
	err := h.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err == sql.ErrNoRows {
		// Create wallet if not exists
		h.DB.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0)", userID)
		balance = 0
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "balance": balance})
}

// ProcessInternalTransaction handles debits/credits from other services
// This uses ACID transactions to ensure integrity
func (h *PaymentHandler) ProcessInternalTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start Database Transaction
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// 1. Check Idempotency (Has this transaction_id been processed?)
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM ledger WHERE transaction_id = $1)", req.TransactionID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusOK, gin.H{"status": "already_processed"})
		return
	}

	// 2. Determine Amount (+ for Credit, - for Debit)
	var finalAmount float64
	if req.Type == models.TxTypeBet || req.Type == models.TxTypeWithdrawal {
		finalAmount = -req.Amount
	} else {
		finalAmount = req.Amount
	}

	// 3. Lock Wallet Row & Update Balance
	// "FOR UPDATE" locks the row to prevent race conditions
	var currentBalance float64
	err = tx.QueryRow(`
		SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE
	`, req.UserID).Scan(&currentBalance)

	if err == sql.ErrNoRows {
		// Create wallet if missing
		_, err = tx.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0)", req.UserID)
		currentBalance = 0
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock wallet"})
		return
	}

	// 4. Check Sufficient Funds (for Debits)
	if finalAmount < 0 && currentBalance+finalAmount < 0 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient funds"})
		return
	}

	newBalance := currentBalance + finalAmount

	// 5. Update Wallet
	_, err = tx.Exec(`
		UPDATE wallets SET balance = $1, updated_at = $2 WHERE user_id = $3
	`, newBalance, time.Now(), req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// 6. Insert Ledger Entry
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, reference_id, reference_type, balance_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, req.TransactionID, req.UserID, req.Type, finalAmount, req.ReferenceID, req.ReferenceType, newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ledger entry"})
		return
	}

	// Commit Transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"new_balance": newBalance,
		"transaction_id": req.TransactionID,
	})
}
