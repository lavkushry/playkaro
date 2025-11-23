package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/models"
)

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func GetBalance(c *gin.Context) {
	userID := c.GetString("userID")

	var wallet models.Wallet
	err := db.DB.QueryRow("SELECT id, user_id, balance, currency FROM wallets WHERE user_id=$1", userID).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency,
	)

	if err == sql.ErrNoRows {
		// Create wallet if not exists (lazy initialization)
		err = db.DB.QueryRow(
			"INSERT INTO wallets (user_id, balance, currency) VALUES ($1, 0, 'INR') RETURNING id, user_id, balance, currency",
			userID,
		).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func Deposit(c *gin.Context) {
	userID := c.GetString("userID")
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	// 1. Get or Create Wallet
	var walletID string
	var currentBalance float64
	err = tx.QueryRow("SELECT id, balance FROM wallets WHERE user_id=$1 FOR UPDATE", userID).Scan(&walletID, &currentBalance)
	if err == sql.ErrNoRows {
		err = tx.QueryRow("INSERT INTO wallets (user_id, balance, currency) VALUES ($1, 0, 'INR') RETURNING id, balance", userID).Scan(&walletID, &currentBalance)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 2. Update Balance
	newBalance := currentBalance + req.Amount
	_, err = tx.Exec("UPDATE wallets SET balance=$1, updated_at=$2 WHERE id=$3", newBalance, time.Now(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// 3. Create Transaction Record
	_, err = tx.Exec(
		"INSERT INTO transactions (wallet_id, type, amount, status, reference_id) VALUES ($1, 'DEPOSIT', $2, 'COMPLETED', $3)",
		walletID, req.Amount, "REF-"+time.Now().Format("20060102150405"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "new_balance": newBalance})
}

func Withdraw(c *gin.Context) {
	userID := c.GetString("userID")
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	// 1. Get Wallet
	var walletID string
	var currentBalance float64
	err = tx.QueryRow("SELECT id, balance FROM wallets WHERE user_id=$1 FOR UPDATE", userID).Scan(&walletID, &currentBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Wallet not found"})
		return
	}

	// 2. Check Balance
	if currentBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// 3. Deduct Balance
	newBalance := currentBalance - req.Amount
	_, err = tx.Exec("UPDATE wallets SET balance=$1, updated_at=$2 WHERE id=$3", newBalance, time.Now(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// 4. Create Transaction Record
	_, err = tx.Exec(
		"INSERT INTO transactions (wallet_id, type, amount, status, reference_id) VALUES ($1, 'WITHDRAW', $2, 'COMPLETED', $3)",
		walletID, req.Amount, "REF-"+time.Now().Format("20060102150405"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdraw successful", "new_balance": newBalance})
}
