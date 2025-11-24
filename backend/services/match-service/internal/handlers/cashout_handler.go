package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/match-service/internal/models"
)

// CashOutBet allows users to exit a bet early
func (h *BetHandler) CashOutBet(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	betID := c.Param("bet_id")

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Lock bet
	var bet models.Bet
	err = tx.QueryRow(`
		SELECT id, user_id, match_id, team, amount, odds, potential_win, status, cashed_out, version
		FROM bets
		WHERE id = $1
		FOR UPDATE
	`, betID).Scan(
		&bet.ID, &bet.UserID, &bet.MatchID, &bet.Team, &bet.Amount,
		&bet.Odds, &bet.PotentialWin, &bet.Status, &bet.CashedOut, &bet.Version,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bet not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify ownership
	if bet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Verify bet is ACTIVE
	if bet.Status != models.BetStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bet is not active"})
		return
	}

	// Verify not already cashed out
	if bet.CashedOut {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bet already cashed out"})
		return
	}

	// Get current match status and odds
	var matchStatus string
	var currentOdds float64
	err = tx.QueryRow(`
		SELECT status,
		CASE
			WHEN $1 = 'TEAM_A' THEN odds_a
			WHEN $1 = 'TEAM_B' THEN odds_b
			ELSE odds_draw
		END as current_odds
		FROM matches
		WHERE match_id = $2
	`, bet.Team, bet.MatchID).Scan(&matchStatus, &currentOdds)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch match"})
		return
	}

	// Verify match is still LIVE (can't cash out if match ended)
	if matchStatus != models.StatusLive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cash-out only available during live matches"})
		return
	}

	// Calculate cash-out amount
	// Formula: StakeAmount × (CurrentOdds / OriginalOdds) × 0.9
	// 0.9 = 10% cash-out fee
	cashOutMultiplier := (currentOdds / bet.Odds) * 0.9
	cashOutAmount := bet.Amount * cashOutMultiplier

	// Ensure minimum cash-out (at least 10% of stake)
	if cashOutAmount < bet.Amount*0.1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cash-out value too low (less than 10% of stake)",
			"cash_out_amount": cashOutAmount,
		})
		return
	}

	// Update bet
	now := time.Now()
	_, err = tx.Exec(`
		UPDATE bets
		SET status = $1,
		    cashed_out = TRUE,
		    cash_out_at = $2,
		    cash_out_odds = $3,
		    cash_out_amount = $4,
		    version = version + 1,
		    updated_at = $2
		WHERE id = $5 AND version = $6
	`, models.BetStatusCashedOut, now, currentOdds, cashOutAmount, betID, bet.Version)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bet"})
		return
	}

	// Credit wallet
	transactionID := fmt.Sprintf("cashout_%s", betID)
	creditErr := h.creditWallet(bet.UserID, cashOutAmount, transactionID, bet.MatchID)
	if creditErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit wallet"})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit cash-out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bet_id":                betID,
		"status":                models.BetStatusCashedOut,
		"original_stake":        bet.Amount,
		"original_odds":         bet.Odds,
		"cash_out_odds":         currentOdds,
		"cash_out_amount":       cashOutAmount,
		"potential_win_forgone": bet.PotentialWin - cashOutAmount,
	})
}

// GetBetHistory returns user's betting history
func (h *BetHandler) GetBetHistory(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	rows, err := h.DB.Query(`
		SELECT b.id, b.match_id, m.team_a, m.team_b, b.team, b.amount, b.odds,
		       b.potential_win, b.status, b.result, b.cashed_out, b.cash_out_amount,
		       b.created_at, b.settled_at
		FROM bets b
		JOIN matches m ON b.match_id = m.match_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT 50
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var bets []map[string]interface{}
	for rows.Next() {
		var b struct {
			ID             string
			MatchID        string
			TeamA          string
			TeamB          string
			Team           string
			Amount         float64
			Odds           float64
			PotentialWin   float64
			Status         string
			Result         *string
			CashedOut      bool
			CashOutAmount  *float64
			CreatedAt      time.Time
			SettledAt      *time.Time
		}

		rows.Scan(&b.ID, &b.MatchID, &b.TeamA, &b.TeamB, &b.Team, &b.Amount, &b.Odds,
			&b.PotentialWin, &b.Status, &b.Result, &b.CashedOut, &b.CashOutAmount,
			&b.CreatedAt, &b.SettledAt)

		bets = append(bets, map[string]interface{}{
			"bet_id":         b.ID,
			"match_id":       b.MatchID,
			"match_name":     fmt.Sprintf("%s vs %s", b.TeamA, b.TeamB),
			"team":           b.Team,
			"amount":         b.Amount,
			"odds":           b.Odds,
			"potential_win":  b.PotentialWin,
			"status":         b.Status,
			"result":         b.Result,
			"cashed_out":     b.CashedOut,
			"cash_out_amount": b.CashOutAmount,
			"created_at":     b.CreatedAt,
			"settled_at":     b.SettledAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"bets": bets})
}
