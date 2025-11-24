package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/realtime"
	"github.com/playkaro/backend/internal/wallet"
)

type CreateMatchRequest struct {
	TeamA string  `json:"team_a" binding:"required"`
	TeamB string  `json:"team_b" binding:"required"`
	OddsA float64 `json:"odds_a" binding:"required,gt=0"`
	OddsB float64 `json:"odds_b" binding:"required,gt=0"`
}

type UpdateOddsRequest struct {
	OddsA float64 `json:"odds_a" binding:"required,gt=0"`
	OddsB float64 `json:"odds_b" binding:"required,gt=0"`
}

type SettleMatchRequest struct {
	Winner string `json:"winner" binding:"required"` // TEAM_A or TEAM_B
}

func CreateMatch(c *gin.Context) {
	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var matchID string
	err := db.DB.QueryRow(
		"INSERT INTO matches (team_a, team_b, odds_a, odds_b, status) VALUES ($1, $2, $3, $4, 'LIVE') RETURNING id",
		req.TeamA, req.TeamB, req.OddsA, req.OddsB,
	).Scan(&matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Match created", "match_id": matchID})
}

func UpdateMatchOdds(c *gin.Context) {
	matchID := c.Param("id")
	var req UpdateOddsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE matches SET odds_a=$1, odds_b=$2 WHERE id=$3", req.OddsA, req.OddsB, matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update odds"})
		return
	}

	// Broadcast odds update via WebSocket
	msg := realtime.WSMessage{
		Type: realtime.TypeOddsUpdate,
		Payload: map[string]interface{}{
			"match_id": matchID,
			"odds_a":   req.OddsA,
			"odds_b":   req.OddsB,
		},
	}
	realtime.MainHub.Broadcast(msg)

	c.JSON(http.StatusOK, gin.H{"message": "Odds updated"})
}

func SettleMatch(c *gin.Context) {
	matchID := c.Param("id")
	var req SettleMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE matches SET status='FINISHED' WHERE id=$1", matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
		return
	}

	rows, err := db.DB.Query("SELECT id, user_id, selection, amount, potential_win FROM bets WHERE match_id=$1 AND status='PENDING'", matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bets"})
		return
	}
	defer rows.Close()

	service := wallet.NewService(db.DB, db.RDB)
	var totalPaid float64
	for rows.Next() {
		var betID, userID, selection string
		var amount, potentialWin float64
		if err := rows.Scan(&betID, &userID, &selection, &amount, &potentialWin); err != nil {
			continue
		}

		if selection == req.Winner {
			_, err = db.DB.Exec("UPDATE bets SET status='WON' WHERE id=$1", betID)
			if err != nil {
				continue
			}

			if _, err := service.SettleBet(c.Request.Context(), userID, amount, potentialWin, true, "WIN-"+betID); err != nil {
				continue
			}

			totalPaid += potentialWin
		} else {
			_, err = db.DB.Exec("UPDATE bets SET status='LOST' WHERE id=$1", betID)
			if err != nil {
				continue
			}
			service.SettleBet(c.Request.Context(), userID, amount, 0, false, "LOSE-"+betID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match settled", "total_paid": totalPaid})
}
