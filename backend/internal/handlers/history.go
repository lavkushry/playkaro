package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/models"
)

func GetTransactions(c *gin.Context) {
	userID := c.GetString("userID")

	rows, err := db.DB.Query(`
		SELECT t.id, t.wallet_id, t.type, t.amount, t.status, t.reference_id, t.created_at
		FROM transactions t
		JOIN wallets w ON t.wallet_id = w.id
		WHERE w.user_id=$1
		ORDER BY t.created_at DESC
		LIMIT 50
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount, &t.Status, &t.ReferenceID, &t.CreatedAt); err != nil {
			continue
		}
		transactions = append(transactions, t)
	}

	c.JSON(http.StatusOK, transactions)
}

func GetBets(c *gin.Context) {
	userID := c.GetString("userID")

	rows, err := db.DB.Query(`
		SELECT b.id, b.user_id, b.match_id, b.selection, b.amount, b.odds, b.potential_win, b.status, b.created_at,
			   m.team_a, m.team_b
		FROM bets b
		JOIN matches m ON b.match_id = m.id
		WHERE b.user_id=$1
		ORDER BY b.created_at DESC
		LIMIT 50
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type BetWithMatch struct {
		models.Bet
		TeamA string `json:"team_a"`
		TeamB string `json:"team_b"`
	}

	var bets []BetWithMatch
	for rows.Next() {
		var b BetWithMatch
		if err := rows.Scan(&b.ID, &b.UserID, &b.MatchID, &b.Selection, &b.Amount, &b.Odds, &b.PotentialWin, &b.Status, &b.CreatedAt, &b.TeamA, &b.TeamB); err != nil {
			continue
		}
		bets = append(bets, b)
	}

	c.JSON(http.StatusOK, bets)
}
