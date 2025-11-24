package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBetHistory returns the betting history for a user
func (h *MatchHandler) GetBetHistory(c *gin.Context) {
	userID := c.GetString("userID")

	// In a real microservice, bets might be stored in a separate 'betting-service' or 'ledger'
	// For this architecture, we'll query the Payment Service's ledger via internal API
	// OR (Simpler for now) we assume we have a local 'bets' table if we want to track match specifics

	// Let's assume we have a local bets table for match-specific details
	rows, err := h.DB.Query(`
		SELECT id, match_id, team, amount, odds, potential_win, status, created_at
		FROM bets
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 50
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bets"})
		return
	}
	defer rows.Close()

	var bets []map[string]interface{}
	for rows.Next() {
		var b struct {
			ID           string
			MatchID      string
			Team         string
			Amount       float64
			Odds         float64
			PotentialWin float64
			Status       string
			CreatedAt    string
		}
		rows.Scan(&b.ID, &b.MatchID, &b.Team, &b.Amount, &b.Odds, &b.PotentialWin, &b.Status, &b.CreatedAt)

		bets = append(bets, map[string]interface{}{
			"id": b.ID,
			"match_id": b.MatchID,
			"team": b.Team,
			"amount": b.Amount,
			"odds": b.Odds,
			"potential_win": b.PotentialWin,
			"status": b.Status,
			"created_at": b.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"bets": bets})
}
