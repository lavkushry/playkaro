package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/playkaro/match-service/internal/cache"
	"github.com/playkaro/match-service/internal/models"
)

type BetHandler struct {
	DB              *sql.DB
	Cache           *cache.MatchCache
	PaymentSvcURL   string
}

func NewBetHandler(db *sql.DB, cache *cache.MatchCache, paymentSvcURL string) *BetHandler {
	return &BetHandler{
		DB:            db,
		Cache:         cache,
		PaymentSvcURL: paymentSvcURL,
	}
}

type PlaceBetRequest struct {
	MatchID string  `json:"match_id" binding:"required"`
	Team    string  `json:"team" binding:"required"`    // "TEAM_A", "TEAM_B", "DRAW"
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type PlaceBetResponse struct {
	BetID        string  `json:"bet_id"`
	MatchID      string  `json:"match_id"`
	Team         string  `json:"team"`
	Amount       float64 `json:"amount"`
	Odds         float64 `json:"odds"`
	PotentialWin float64 `json:"potential_win"`
	Status       string  `json:"status"`
}

// PlaceBet handles bet placement with optimistic locking
func (h *BetHandler) PlaceBet(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	var req PlaceBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate team selection
	if req.Team != "TEAM_A" && req.Team != "TEAM_B" && req.Team != "DRAW" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team selection"})
		return
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Lock match and get current odds + version (Optimistic Locking)
	var match models.Match
	err = tx.QueryRow(`
		SELECT id, match_id, team_a, team_b, odds_a, odds_b, odds_draw, status, version
		FROM matches
		WHERE match_id = $1
		FOR UPDATE
	`, req.MatchID).Scan(
		&match.ID, &match.MatchID, &match.TeamA, &match.TeamB,
		&match.OddsA, &match.OddsB, &match.OddsDraw, &match.Status, &match.Version,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if match is still accepting bets
	if match.Status != models.StatusUpcoming && match.Status != models.StatusLive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Match is not accepting bets"})
		return
	}

	// Get odds for selected team
	var odds float64
	switch req.Team {
	case "TEAM_A":
		odds = match.OddsA
	case "TEAM_B":
		odds = match.OddsB
	case "DRAW":
		odds = match.OddsDraw
	}

	if odds < 1.01 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid odds"})
		return
	}

	// Calculate potential win
	potentialWin := req.Amount * odds

	// Generate bet ID
	betID := uuid.New().String()
	transactionID := fmt.Sprintf("bet_%s", betID)

	// Debit wallet via Payment Service
	debitErr := h.debitWallet(userID, req.Amount, transactionID, req.MatchID)
	if debitErr != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": debitErr.Error()})
		return
	}

	// Create bet record
	_, err = tx.Exec(`
		INSERT INTO bets (id, user_id, match_id, team, amount, odds, potential_win, status, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1)
	`, betID, userID, req.MatchID, req.Team, req.Amount, odds, potentialWin, models.BetStatusActive)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bet"})
		return
	}

	// Increment match version (optimistic lock check)
	result, err := tx.Exec(`
		UPDATE matches SET version = version + 1, updated_at = $1
		WHERE match_id = $2 AND version = $3
	`, time.Now(), req.MatchID, match.Version)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Concurrent modification detected
		c.JSON(http.StatusConflict, gin.H{"error": "Odds changed, please retry"})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit bet"})
		return
	}

	c.JSON(http.StatusCreated, PlaceBetResponse{
		BetID:        betID,
		MatchID:      req.MatchID,
		Team:         req.Team,
		Amount:       req.Amount,
		Odds:         odds,
		PotentialWin: potentialWin,
		Status:       models.BetStatusActive,
	})
}

// debitWallet calls Payment Service to debit user wallet
func (h *BetHandler) debitWallet(userID string, amount float64, transactionID, referenceID string) error {
	reqBody := map[string]interface{}{
		"user_id":        userID,
		"amount":         amount,
		"type":           "BET",
		"transaction_id": transactionID,
		"reference_id":   referenceID,
		"reference_type": "MATCH_CRICKET",
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		fmt.Sprintf("%s/v1/payments/internal/transaction", h.PaymentSvcURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return fmt.Errorf("payment service unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf(errResp["error"])
	}

	return nil
}

// creditWallet calls Payment Service to credit user wallet
func (h *BetHandler) creditWallet(userID string, amount float64, transactionID, referenceID string) error {
	reqBody := map[string]interface{}{
		"user_id":        userID,
		"amount":         amount,
		"type":           "WIN",
		"transaction_id": transactionID,
		"reference_id":   referenceID,
		"reference_type": "MATCH_CRICKET",
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		fmt.Sprintf("%s/v1/payments/internal/transaction", h.PaymentSvcURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return fmt.Errorf("payment service unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to credit wallet")
	}

	return nil
}
