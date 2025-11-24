package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/playkaro/payment-service/internal/models"
)

type BonusService struct {
	DB *sql.DB
}

func NewBonusService(db *sql.DB) *BonusService {
	return &BonusService{DB: db}
}

// GrantBonus grants a bonus to a user
func (s *BonusService) GrantBonus(userID string, amount float64, expiryDays int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	expiresAt := time.Now().AddDate(0, 0, expiryDays)

	// 1. Create bonus record
	_, err = tx.Exec(`
		INSERT INTO bonuses (user_id, amount, status, expires_at)
		VALUES ($1, $2, $3, $4)
	`, userID, amount, models.BonusStatusActive, expiresAt)
	if err != nil {
		return err
	}

	// 2. Credit bonus_balance in wallet
	var currentBonus float64
	err = tx.QueryRow("SELECT bonus_balance FROM wallets WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBonus)
	if err == sql.ErrNoRows {
		// Create wallet if doesn't exist
		_, err = tx.Exec(`
			INSERT INTO wallets (user_id, balance, deposit_balance, bonus_balance, winnings_balance, currency)
			VALUES ($1, $2, 0, $2, 0, 'PTS')
		`, userID, amount)
		currentBonus = 0
	} else if err != nil {
		return err
	} else {
		// Update existing wallet
		newBonusBalance := currentBonus + amount
		_, err = tx.Exec(`
			UPDATE wallets
			SET bonus_balance = $1, balance = balance + $2, updated_at = $3
			WHERE user_id = $4
		`, newBonusBalance, amount, time.Now(), userID)
		if err != nil {
			return err
		}
	}

	// 3. Create ledger entry
	transactionID := fmt.Sprintf("bonus_%d", time.Now().UnixNano())
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, balance_type, reference_type, balance_after, state)
		SELECT $1, $2, $3, $4, $5, $6, balance, $7
		FROM wallets WHERE user_id = $2
	`, transactionID, userID, models.TxTypeBonus, amount, models.BalanceTypeBonus, "PROMOTION", models.TxStateSettled)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ExpireBonuses marks all expired bonuses as EXPIRED and deducts from wallet
func (s *BonusService) ExpireBonuses() error {
	// Find expired bonuses
	rows, err := s.DB.Query(`
		SELECT id, user_id, amount
		FROM bonuses
		WHERE status = $1 AND expires_at < $2
	`, models.BonusStatusActive, time.Now())
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var bonusID, userID string
		var amount float64
		if err := rows.Scan(&bonusID, &userID, &amount); err != nil {
			continue
		}

		// Mark bonus as expired
		tx, err := s.DB.Begin()
		if err != nil {
			continue
		}

		_, err = tx.Exec("UPDATE bonuses SET status = $1, updated_at = $2 WHERE id = $3",
			models.BonusStatusExpired, time.Now(), bonusID)
		if err != nil {
			tx.Rollback()
			continue
		}

		// Deduct from bonus_balance
		_, err = tx.Exec(`
			UPDATE wallets
			SET bonus_balance = GREATEST(bonus_balance - $1, 0),
			    balance = GREATEST(balance - $1, 0),
			    updated_at = $2
			WHERE user_id = $3
		`, amount, time.Now(), userID)
		if err != nil {
			tx.Rollback()
			continue
		}

		tx.Commit()
		count++
	}

	fmt.Printf("Expired %d bonuses\n", count)
	return nil
}
