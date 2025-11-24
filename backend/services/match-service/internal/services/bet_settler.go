package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/playkaro/match-service/internal/models"
)

type BetSettler struct {
	DB            *sql.DB
	PaymentSvcURL string
}

func NewBetSettler(db *sql.DB, paymentSvcURL string) *BetSettler {
	return &BetSettler{
		DB:            db,
		PaymentSvcURL: paymentSvcURL,
	}
}

// SettleMatchBets settles all bets for a completed match
func (s *BetSettler) SettleMatchBets(matchID, winningTeam string) error {
	log.Printf("Settling bets for match %s, winner: %s", matchID, winningTeam)

	// Get all ACTIVE bets for this match
	rows, err := s.DB.Query(`
		SELECT id, user_id, team, amount, potential_win
		FROM bets
		WHERE match_id = $1 AND status = $2
	`, matchID, models.BetStatusActive)

	if err != nil {
		return err
	}
	defer rows.Close()

	settledCount := 0
	for rows.Next() {
		var betID, userID, team string
		var amount, potentialWin float64

		if err := rows.Scan(&betID, &userID, &team, &amount, &potentialWin); err != nil {
			continue
		}

		// Determine if bet won or lost
		var result string
		var payout float64
		if team == winningTeam {
			result = models.BetResultWon
			payout = potentialWin
		} else {
			result = models.BetResultLost
			payout = 0
		}

		// Update bet status
		_, err = s.DB.Exec(`
			UPDATE bets
			SET status = $1, result = $2, settled_at = $3, updated_at = $3
			WHERE id = $4
		`, models.BetStatusSettled, result, time.Now(), betID)

		if err != nil {
			log.Printf("Failed to update bet %s: %v", betID, err)
			continue
		}

		// Credit winners
		if result == models.BetResultWon && payout > 0 {
			transactionID := fmt.Sprintf("bet_win_%s", betID)
			creditErr := s.creditWallet(userID, payout, transactionID, matchID)
			if creditErr != nil {
				log.Printf("Failed to credit user %s for bet %s: %v", userID, betID, creditErr)
			}
		}

		settledCount++
	}

	log.Printf("Settled %d bets for match %s", settledCount, matchID)
	return nil
}

// creditWallet calls Payment Service to credit user wallet
func (s *BetSettler) creditWallet(userID string, amount float64, transactionID, referenceID string) error {
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
		fmt.Sprintf("%s/v1/payments/internal/transaction", s.PaymentSvcURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return fmt.Errorf("payment service unavailable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("payment error: %s", errResp["error"])
	}

	return nil
}

// VoidBets voids all bets for a match (e.g., if match is cancelled)
func (s *BetSettler) VoidBets(matchID string) error {
	log.Printf("Voiding all bets for match %s", matchID)

	// Get all ACTIVE bets
	rows, err := s.DB.Query(`
		SELECT id, user_id, amount
		FROM bets
		WHERE match_id = $1 AND status = $2
	`, matchID, models.BetStatusActive)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var betID, userID string
		var amount float64

		if err := rows.Scan(&betID, &userID, &amount); err != nil {
			continue
		}

		// Mark as voided
		_, err = s.DB.Exec(`
			UPDATE bets
			SET status = $1, result = $2, settled_at = $3, updated_at = $3
			WHERE id = $4
		`, models.BetStatusSettled, models.BetResultVoid, time.Now(), betID)

		// Refund stake
		transactionID := fmt.Sprintf("bet_void_%s", betID)
		s.creditWallet(userID, amount, transactionID, matchID)
	}

	return nil
}
