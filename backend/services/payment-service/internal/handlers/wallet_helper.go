package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/playkaro/payment-service/internal/models"
)

// WalletHelper provides helper methods for wallet operations
type WalletHelper struct {
	DB *sql.DB
}

func NewWalletHelper(db *sql.DB) *WalletHelper {
	return &WalletHelper{DB: db}
}

// CreditDeposit credits amount to deposit_balance
func (w *WalletHelper) CreditDeposit(tx *sql.Tx, userID string, amount float64, transactionID, referenceID, referenceType string) error {
	// Lock wallet
	var deposit, bonus, winnings float64
	err := tx.QueryRow(`
		SELECT deposit_balance, bonus_balance, winnings_balance
		FROM wallets WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&deposit, &bonus, &winnings)

	if err == sql.ErrNoRows {
		// Create wallet
		_, err = tx.Exec(`
			INSERT INTO wallets (user_id, balance, deposit_balance, bonus_balance, winnings_balance, currency)
			VALUES ($1, $2, $2, 0, 0, 'PTS')
		`, userID, amount)
		deposit = 0
	} else if err != nil {
		return err
	}

	newDeposit := deposit + amount
	newTotal := newDeposit + bonus + winnings

	// Update wallet
	_, err = tx.Exec(`
		UPDATE wallets
		SET deposit_balance = $1, balance = $2, updated_at = $3
		WHERE user_id = $4
	`, newDeposit, newTotal, time.Now(), userID)
	if err != nil {
		return err
	}

	// Create ledger entry
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, balance_type, reference_id, reference_type, balance_after, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, transactionID, userID, models.TxTypeDeposit, amount, models.BalanceTypeDeposit,
		referenceID, referenceType, newTotal, models.TxStateSettled)

	return err
}

// CreditWinnings credits amount to winnings_balance
func (w *WalletHelper) CreditWinnings(tx *sql.Tx, userID string, amount float64, transactionID, referenceID, referenceType string) error {
	var deposit, bonus, winnings float64
	err := tx.QueryRow(`
		SELECT deposit_balance, bonus_balance, winnings_balance
		FROM wallets WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&deposit, &bonus, &winnings)

	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO wallets (user_id, balance, deposit_balance, bonus_balance, winnings_balance, currency)
			VALUES ($1, $2, 0, 0, $2, 'PTS')
		`, userID, amount)
		winnings = 0
	} else if err != nil {
		return err
	}

	newWinnings := winnings + amount
	newTotal := deposit + bonus + newWinnings

	_, err = tx.Exec(`
		UPDATE wallets
		SET winnings_balance = $1, balance = $2, updated_at = $3
		WHERE user_id = $4
	`, newWinnings, newTotal, time.Now(), userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, balance_type, reference_id, reference_type, balance_after, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, transactionID, userID, models.TxTypeWin, amount, models.BalanceTypeWinnings,
		referenceID, referenceType, newTotal, models.TxStateSettled)

	return err
}

// DebitWallet debits from wallet with priority: Deposit → Winnings → Bonus
func (w *WalletHelper) DebitWallet(tx *sql.Tx, userID string, amount float64, transactionID, referenceID, referenceType string) error {
	var deposit, bonus, winnings float64
	err := tx.QueryRow(`
		SELECT deposit_balance, bonus_balance, winnings_balance
		FROM wallets WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&deposit, &bonus, &winnings)

	if err != nil {
		return err
	}

	totalAvailable := deposit + bonus + winnings
	if totalAvailable < amount {
		return fmt.Errorf("insufficient funds: need %.2f, have %.2f", amount, totalAvailable)
	}

	remaining := amount
	deductions := make(map[string]float64)

	// Priority 1: Deduct from Deposit
	if deposit > 0 {
		deductFromDeposit := min(remaining, deposit)
		deposit -= deductFromDeposit
		remaining -= deductFromDeposit
		deductions[models.BalanceTypeDeposit] = deductFromDeposit
	}

	// Priority 2: Deduct from Winnings
	if remaining > 0 && winnings > 0 {
		deductFromWinnings := min(remaining, winnings)
		winnings -= deductFromWinnings
		remaining -= deductFromWinnings
		deductions[models.BalanceTypeWinnings] = deductFromWinnings
	}

	// Priority 3: Deduct from Bonus
	if remaining > 0 && bonus > 0 {
		deductFromBonus := min(remaining, bonus)
		bonus -= deductFromBonus
		remaining -= deductFromBonus
		deductions[models.BalanceTypeBonus] = deductFromBonus
	}

	newTotal := deposit + bonus + winnings

	// Update wallet
	_, err = tx.Exec(`
		UPDATE wallets
		SET deposit_balance = $1, bonus_balance = $2, winnings_balance = $3, balance = $4, updated_at = $5
		WHERE user_id = $6
	`, deposit, bonus, winnings, newTotal, time.Now(), userID)
	if err != nil {
		return err
	}

	// Create ledger entries for each balance type affected
	metadata, _ := json.Marshal(deductions)
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, balance_type, reference_id, reference_type, balance_after, state, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, transactionID, userID, models.TxTypeBet, -amount, "MIXED",
		referenceID, referenceType, newTotal, models.TxStateSettled, string(metadata))

	return err
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
