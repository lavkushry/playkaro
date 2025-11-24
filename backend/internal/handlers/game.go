package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/wallet"
)

// Seamless Wallet API - Standard callbacks for game providers

type SeamlessWalletRequest struct {
	UserID     string  `json:"user_id" binding:"required"`
	GameID     string  `json:"game_id" binding:"required"`
	RoundID    string  `json:"round_id" binding:"required"`
	Amount     float64 `json:"amount"`
	ProviderID string  `json:"provider_id" binding:"required"`
	Signature  string  `json:"signature"` // For security
}

// GetBalanceForProvider - Provider requests user balance
func GetBalanceForProvider(c *gin.Context) {
	var req SeamlessWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := wallet.NewService(db.DB, db.RDB)
	w, err := service.Get(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	w.Balance = w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance

	c.JSON(http.StatusOK, gin.H{
		"user_id":  req.UserID,
		"balance":  w.Balance,
		"currency": "INR",
	})
}

// DebitWallet - Provider deducts bet amount
func DebitWallet(c *gin.Context) {
	var req SeamlessWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := wallet.NewService(db.DB, db.RDB)
	w, err := service.LockForBet(c.Request.Context(), req.UserID, req.Amount, "GAME-"+req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Record game round
	db.DB.Exec(
		"INSERT INTO game_rounds (session_id, round_id, bet, status) VALUES ($1, $2, $3, 'PENDING')",
		req.GameID+"-"+req.UserID, req.RoundID, req.Amount,
	)

	c.JSON(http.StatusOK, gin.H{
		"user_id":        req.UserID,
		"new_balance":    w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance,
		"transaction_id": req.RoundID,
	})
}

// CreditWallet - Provider adds win amount
func CreditWallet(c *gin.Context) {
	var req SeamlessWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var betAmount float64
	db.DB.QueryRow("SELECT bet FROM game_rounds WHERE round_id=$1", req.RoundID).Scan(&betAmount)
	if betAmount == 0 {
		betAmount = req.Amount
	}

	service := wallet.NewService(db.DB, db.RDB)
	payout := betAmount + req.Amount
	w, err := service.SettleBet(c.Request.Context(), req.UserID, betAmount, payout, true, "GAME-WIN-"+req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.DB.Exec("UPDATE game_rounds SET win=$1, status='COMPLETED' WHERE round_id=$2", req.Amount, req.RoundID)

	c.JSON(http.StatusOK, gin.H{
		"user_id":        req.UserID,
		"new_balance":    w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance,
		"transaction_id": req.RoundID,
	})
}

// RollbackWallet - Provider cancels a round
func RollbackWallet(c *gin.Context) {
	var req SeamlessWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var betAmount float64
	if err := db.DB.QueryRow("SELECT bet FROM game_rounds WHERE round_id=$1", req.RoundID).Scan(&betAmount); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	service := wallet.NewService(db.DB, db.RDB)
	w, err := service.SettleBet(c.Request.Context(), req.UserID, betAmount, betAmount, true, "GAME-ROLLBACK-"+req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.DB.Exec("UPDATE game_rounds SET status='CANCELLED' WHERE round_id=$1", req.RoundID)

	c.JSON(http.StatusOK, gin.H{
		"user_id":     req.UserID,
		"new_balance": w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance,
		"refunded":    betAmount,
	})
}

// LaunchGame - Creates game session and returns launch URL
func LaunchGame(c *gin.Context) {
	userID := c.GetString("userID")
	gameID := c.Query("game_id")

	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id required"})
		return
	}

	// Get user balance
	var balance float64
	service := wallet.NewService(db.DB, db.RDB)
	w, _ := service.Get(c.Request.Context(), userID)
	if w != nil {
		balance = w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance
	}

	// Create game session
	var sessionID string
	err := db.DB.QueryRow(
		"INSERT INTO game_sessions (user_id, game_id, provider_id, start_balance, status) VALUES ($1, $2, 'EVOLUTION', $3, 'ACTIVE') RETURNING id",
		userID, gameID, balance,
	).Scan(&sessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Mock game URL (in production, this would call provider API)
	gameURL := "https://demo-casino.com/game/" + gameID + "?session=" + sessionID + "&user=" + userID

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"game_url":   gameURL,
		"balance":    balance,
	})
}

// GetGames - Returns list of available games
func GetGames(c *gin.Context) {
	providerFilter := c.Query("provider")
	typeFilter := c.Query("type")

	query := "SELECT id, provider_id, name, type, thumbnail_url, min_bet, max_bet, rtp FROM games WHERE is_active=true"
	args := []interface{}{}
	argCount := 1

	if providerFilter != "" {
		query += " AND provider_id=$" + string(rune(argCount+'0'))
		args = append(args, providerFilter)
		argCount++
	}

	if typeFilter != "" {
		query += " AND type=$" + string(rune(argCount+'0'))
		args = append(args, typeFilter)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type GameInfo struct {
		ID           string  `json:"id"`
		ProviderID   string  `json:"provider_id"`
		Name         string  `json:"name"`
		Type         string  `json:"type"`
		ThumbnailURL string  `json:"thumbnail_url"`
		MinBet       float64 `json:"min_bet"`
		MaxBet       float64 `json:"max_bet"`
		RTP          float64 `json:"rtp"`
	}

	var games []GameInfo
	for rows.Next() {
		var g GameInfo
		rows.Scan(&g.ID, &g.ProviderID, &g.Name, &g.Type, &g.ThumbnailURL, &g.MinBet, &g.MaxBet, &g.RTP)
		games = append(games, g)
	}

	c.JSON(http.StatusOK, games)
}
