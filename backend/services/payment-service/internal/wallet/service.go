package wallet

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWalletNotFound    = errors.New("wallet not found")
)

type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

type Balance struct {
	Amount          float64
	Bonus           float64
	Currency        string
	DepositBalance  float64
	WinningsBalance float64
}

type TransactionResult struct {
	ID           string
	BalanceAfter float64
}

func (s *Service) GetBalance(userID string) (*Balance, error) {
	var b Balance
	err := s.DB.QueryRow(`
		SELECT balance, deposit_balance, bonus_balance, winnings_balance, currency
		FROM wallets WHERE user_id = $1
	`, userID).Scan(&b.Amount, &b.DepositBalance, &b.Bonus, &b.WinningsBalance, &b.Currency)

	if err == sql.ErrNoRows {
		// Create wallet if not exists
		_, err = s.DB.Exec(`
			INSERT INTO wallets (user_id, balance, deposit_balance, bonus_balance, winnings_balance, currency)
			VALUES ($1, 0, 0, 0, 0, 'PTS')
		`, userID)
		if err != nil {
			return nil, err
		}
		return &Balance{Currency: "PTS"}, nil
	} else if err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *Service) Debit(userID string, amount float64, refID, refType string) (*TransactionResult, error) {
	return s.processTransaction(userID, -amount, "DEBIT", refID, refType)
}

func (s *Service) Credit(userID string, amount float64, refID, refType string) (*TransactionResult, error) {
	return s.processTransaction(userID, amount, "CREDIT", refID, refType)
}

// Deposit adds funds (usually from payment gateway)
func (s *Service) Deposit(userID string, amount float64, method string) (*TransactionResult, error) {
	return s.processTransaction(userID, amount, "DEPOSIT", method, "PAYMENT_GATEWAY")
}

// Withdraw deducts funds (usually to bank account)
func (s *Service) Withdraw(userID string, amount float64, accountID string) (*TransactionResult, error) {
	return s.processTransaction(userID, -amount, "WITHDRAWAL", accountID, "BANK_ACCOUNT")
}

func (s *Service) processTransaction(userID string, amount float64, txType, refID, refType string) (*TransactionResult, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock Wallet Row & Update Balance
	var currentBalance float64
	err = tx.QueryRow(`
		SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&currentBalance)

	if err == sql.ErrNoRows {
		// Create wallet if missing
		_, err = tx.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0)", userID)
		if err != nil {
			return nil, err
		}
		currentBalance = 0
	} else if err != nil {
		return nil, err
	}

	// Check Sufficient Funds (for Debits)
	if amount < 0 && currentBalance+amount < 0 {
		return nil, ErrInsufficientFunds
	}

	newBalance := currentBalance + amount
	txID := uuid.New().String()

	// Update Wallet
	_, err = tx.Exec(`
		UPDATE wallets SET balance = $1, updated_at = $2 WHERE user_id = $3
	`, newBalance, time.Now(), userID)
	if err != nil {
		return nil, err
	}

	// Insert Ledger Entry
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, reference_id, reference_type, balance_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, txID, userID, txType, amount, refID, refType, newBalance)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &TransactionResult{
		ID:           txID,
		BalanceAfter: newBalance,
	}, nil
}
