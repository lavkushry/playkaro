package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/payment-service/internal/models"
	"github.com/playkaro/payment-service/internal/wallet"
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

// GetBalance returns the current balance with multi-currency breakdown
func (h *PaymentHandler) GetBalance(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		// Allow internal calls to pass UserID via query param if secured
		userID = c.Query("user_id")
	}

	balance, err := h.WalletService.GetBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":          userID,
		"total_balance":    balance.Amount,
		"deposit_balance":  balance.DepositBalance,
		"bonus_balance":    balance.Bonus,
		"winnings_balance": balance.WinningsBalance,
		"currency":         balance.Currency,
	})
}

// ProcessInternalTransaction handles debits/credits from other services
// This uses ACID transactions to ensure integrity
func (h *PaymentHandler) ProcessInternalTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	var result *wallet.TransactionResult

	if req.Type == models.TxTypeBet || req.Type == models.TxTypeWithdrawal {
		result, err = h.WalletService.Debit(req.UserID, req.Amount, req.ReferenceID, req.ReferenceType)
	} else {
		result, err = h.WalletService.Credit(req.UserID, req.Amount, req.ReferenceID, req.ReferenceType)
	}

	if err != nil {
		if err == wallet.ErrInsufficientFunds {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient funds"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"new_balance":    result.BalanceAfter,
		"transaction_id": result.ID,
	})
}
