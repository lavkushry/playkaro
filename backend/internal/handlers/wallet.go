package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/models"
	"github.com/playkaro/backend/internal/wallet"
)

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	// Optional: allow clients to pass idempotency keys
	IdempotencyKey string `json:"idempotency_key"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func GetBalance(c *gin.Context) {
	userID := c.GetString("userID")

	service := wallet.NewService(db.DB, db.RDB)
	w, err := service.Get(c.Request.Context(), userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load wallet"})
		return
	}
	// Provide aggregate balance for backward compatibility
	w.Balance = w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance
	c.JSON(http.StatusOK, w)
}

func Deposit(c *gin.Context) {
	userID := c.GetString("userID")
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	service := wallet.NewService(db.DB, db.RDB)
	reference := req.IdempotencyKey
	if reference == "" {
		reference = "REF-" + time.Now().Format("20060102150405")
	}
	w, err := service.Deposit(c.Request.Context(), userID, req.Amount, reference)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w.Balance = w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance
	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "new_balance": w.Balance, "wallet": w})
}

func Withdraw(c *gin.Context) {
	userID := c.GetString("userID")
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	service := wallet.NewService(db.DB, db.RDB)
	ref := "REF-" + time.Now().Format("20060102150405")
	w, err := service.Withdraw(c.Request.Context(), userID, req.Amount, ref)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w.Balance = w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance
	c.JSON(http.StatusOK, gin.H{"message": "Withdraw successful", "new_balance": w.Balance, "wallet": w})
}
