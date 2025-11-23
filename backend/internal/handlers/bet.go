package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/models"
)

type PlaceBetRequest struct {
	MatchID   string  `json:"match_id" binding:"required"`
	Selection string  `json:"selection" binding:"required"` // TEAM_A, TEAM_B
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

func GetMatches(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, team_a, team_b, odds_a, odds_b, status, start_time FROM matches WHERE status='LIVE'")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		if err := rows.Scan(&m.ID, &m.TeamA, &m.TeamB, &m.OddsA, &m.OddsB, &m.Status, &m.StartTime); err != nil {
			continue
		}
		matches = append(matches, m)
	}

	c.JSON(http.StatusOK, matches)
}

func PlaceBet(c *gin.Context) {
	userID := c.GetString("userID")
	var req PlaceBetRequest
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

	// 1. Get Wallet & Check Balance
	var walletID string
	var balance float64
	err = tx.QueryRow("SELECT id, balance FROM wallets WHERE user_id=$1 FOR UPDATE", userID).Scan(&walletID, &balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Wallet not found"})
		return
	}
	if balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// 2. Get Match & Odds
	var odds float64
	var oddsA, oddsB float64
	err = tx.QueryRow("SELECT odds_a, odds_b FROM matches WHERE id=$1", req.MatchID).Scan(&oddsA, &oddsB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	if req.Selection == "TEAM_A" {
		odds = oddsA
	} else {
		odds = oddsB
	}

	// 3. Deduct Balance
	newBalance := balance - req.Amount
	_, err = tx.Exec("UPDATE wallets SET balance=$1 WHERE id=$2", newBalance, walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// 4. Create Bet
	potentialWin := req.Amount * odds
	_, err = tx.Exec(
		"INSERT INTO bets (user_id, match_id, selection, amount, odds, potential_win, status) VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')",
		userID, req.MatchID, req.Selection, req.Amount, odds, potentialWin,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place bet"})
		return
	}

	// 5. Record Transaction
	_, err = tx.Exec(
		"INSERT INTO transactions (wallet_id, type, amount, status, reference_id) VALUES ($1, 'BET', $2, 'COMPLETED', $3)",
		walletID, req.Amount, "BET-"+time.Now().Format("20060102150405"),
	)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bet placed successfully", "new_balance": newBalance})
}
