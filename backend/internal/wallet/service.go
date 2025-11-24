package wallet

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/playkaro/backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type BalanceBucket string

const (
	BucketDeposit  BalanceBucket = "deposit"
	BucketBonus    BalanceBucket = "bonus"
	BucketWinnings BalanceBucket = "winnings"
	BucketLocked   BalanceBucket = "locked"
)

// Simple KYC-tiered limits; tune as needed.
var kycDailyLimits = map[int]float64{
	0: 10000,   // INR
	1: 100000,  // INR
	2: 1000000, // INR
}

type Service struct {
	db    *sql.DB
	redis *redis.Client
}

func NewService(database *sql.DB, rdb *redis.Client) *Service {
	return &Service{db: database, redis: rdb}
}

func (s *Service) Get(ctx context.Context, userID string) (*models.Wallet, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	wallet, err := s.ensureWallet(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return wallet, nil
}

// Deposit credits the deposit bucket and enforces daily limits based on KYC level.
func (s *Service) Deposit(ctx context.Context, userID string, amount float64, reference string) (*models.Wallet, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := s.ensureWallet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != "ACTIVE" {
		return nil, errors.New("wallet is not active")
	}

	if err := s.resetDailyCounters(ctx, tx, wallet); err != nil {
		return nil, err
	}
	limit, ok := kycDailyLimits[wallet.KYCLevel]
	if !ok {
		limit = kycDailyLimits[0]
	}
	if wallet.DailyDepositUsed+amount > limit {
		return nil, errors.New("daily deposit limit exceeded for current KYC level")
	}

	wallet.DepositBalance += amount
	wallet.DailyDepositUsed += amount
	wallet.UpdatedAt = time.Now()

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET deposit_balance=$1, daily_deposit_used=$2, last_deposit_reset=$3, updated_at=$4
		WHERE id=$5`,
		wallet.DepositBalance, wallet.DailyDepositUsed, wallet.LastDepositReset, wallet.UpdatedAt, wallet.ID); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, status, reference_id, bucket)
		VALUES ($1, 'DEPOSIT', $2, 'COMPLETED', $3, 'deposit')`,
		wallet.ID, amount, reference); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.cacheAvailable(ctx, userID, wallet)
	return wallet, nil
}

// Withdraw debits winnings first, then deposit if required.
func (s *Service) Withdraw(ctx context.Context, userID string, amount float64, reference string) (*models.Wallet, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := s.ensureWallet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != "ACTIVE" {
		return nil, errors.New("wallet is not active")
	}

	available := wallet.WinningsBalance + wallet.DepositBalance + wallet.BonusBalance - wallet.LockedBalance
	if amount > available {
		return nil, errors.New("insufficient balance")
	}

	remaining := amount

	// Winnings first
	if wallet.WinningsBalance > 0 {
		use := minFloat(wallet.WinningsBalance, remaining)
		wallet.WinningsBalance -= use
		remaining -= use
	}
	// Then deposit
	if remaining > 0 && wallet.DepositBalance > 0 {
		use := minFloat(wallet.DepositBalance, remaining)
		wallet.DepositBalance -= use
		remaining -= use
	}
	// Bonus should not be withdrawable, so we leave it intact

	if remaining > 0 {
		return nil, errors.New("insufficient withdrawable balance")
	}

	wallet.UpdatedAt = time.Now()

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET winnings_balance=$1, deposit_balance=$2, updated_at=$3
		WHERE id=$4`,
		wallet.WinningsBalance, wallet.DepositBalance, wallet.UpdatedAt, wallet.ID); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, status, reference_id, bucket)
		VALUES ($1, 'WITHDRAW', $2, 'COMPLETED', $3, 'winnings')`,
		wallet.ID, amount, reference); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.cacheAvailable(ctx, userID, wallet)
	return wallet, nil
}

// LockForBet deducts stake by priority Bonus -> Deposit -> Winnings and moves it to locked_balance.
func (s *Service) LockForBet(ctx context.Context, userID string, stake float64, reference string) (*models.Wallet, error) {
	if stake <= 0 {
		return nil, errors.New("stake must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := s.ensureWallet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != "ACTIVE" {
		return nil, errors.New("wallet is not active")
	}

	available := wallet.DepositBalance + wallet.BonusBalance + wallet.WinningsBalance - wallet.LockedBalance
	if stake > available {
		return nil, errors.New("insufficient balance")
	}

	remaining := stake
	if wallet.BonusBalance > 0 {
		use := minFloat(wallet.BonusBalance, remaining)
		wallet.BonusBalance -= use
		remaining -= use
	}
	if remaining > 0 && wallet.DepositBalance > 0 {
		use := minFloat(wallet.DepositBalance, remaining)
		wallet.DepositBalance -= use
		remaining -= use
	}
	if remaining > 0 && wallet.WinningsBalance > 0 {
		use := minFloat(wallet.WinningsBalance, remaining)
		wallet.WinningsBalance -= use
		remaining -= use
	}
	if remaining > 0 {
		return nil, errors.New("insufficient balance after deductions")
	}

	wallet.LockedBalance += stake
	wallet.UpdatedAt = time.Now()

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET bonus_balance=$1, deposit_balance=$2, winnings_balance=$3, locked_balance=$4, updated_at=$5
		WHERE id=$6`,
		wallet.BonusBalance, wallet.DepositBalance, wallet.WinningsBalance, wallet.LockedBalance, wallet.UpdatedAt, wallet.ID); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, status, reference_id, bucket)
		VALUES ($1, 'BET', $2, 'COMPLETED', $3, 'locked')`,
		wallet.ID, stake, reference); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.cacheAvailable(ctx, userID, wallet)
	return wallet, nil
}

// SettleBet releases locked stake and credits winnings if applicable.
func (s *Service) SettleBet(ctx context.Context, userID string, stake float64, payout float64, win bool, reference string) (*models.Wallet, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := s.ensureWallet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.LockedBalance < stake {
		return nil, errors.New("locked balance insufficient for settlement")
	}

	wallet.LockedBalance -= stake
	if win && payout > 0 {
		wallet.WinningsBalance += payout
	}
	wallet.UpdatedAt = time.Now()

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET locked_balance=$1, winnings_balance=$2, updated_at=$3
		WHERE id=$4`,
		wallet.LockedBalance, wallet.WinningsBalance, wallet.UpdatedAt, wallet.ID); err != nil {
		return nil, err
	}

	if win && payout > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transactions (wallet_id, type, amount, status, reference_id, bucket)
			VALUES ($1, 'WIN', $2, 'COMPLETED', $3, 'winnings')`,
			wallet.ID, payout, reference); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transactions (wallet_id, type, amount, status, reference_id, bucket)
			VALUES ($1, 'BET_SETTLE', $2, 'COMPLETED', $3, 'locked')`,
			wallet.ID, stake, reference); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.cacheAvailable(ctx, userID, wallet)
	return wallet, nil
}

func (s *Service) ensureWallet(ctx context.Context, tx *sql.Tx, userID string) (*models.Wallet, error) {
	var w models.Wallet
	err := tx.QueryRowContext(ctx, `
		SELECT id, user_id, deposit_balance, bonus_balance, winnings_balance, locked_balance,
		       currency, kyc_level, daily_deposit_used, last_deposit_reset, status, updated_at
		FROM wallets WHERE user_id=$1 FOR UPDATE`,
		userID).Scan(
		&w.ID, &w.UserID, &w.DepositBalance, &w.BonusBalance, &w.WinningsBalance, &w.LockedBalance,
		&w.Currency, &w.KYCLevel, &w.DailyDepositUsed, &w.LastDepositReset, &w.Status, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Create wallet
		err = tx.QueryRowContext(ctx, `
			INSERT INTO wallets (user_id, currency)
			VALUES ($1, 'INR')
			RETURNING id, user_id, deposit_balance, bonus_balance, winnings_balance, locked_balance,
			          currency, kyc_level, daily_deposit_used, last_deposit_reset, status, updated_at`,
			userID).Scan(
			&w.ID, &w.UserID, &w.DepositBalance, &w.BonusBalance, &w.WinningsBalance, &w.LockedBalance,
			&w.Currency, &w.KYCLevel, &w.DailyDepositUsed, &w.LastDepositReset, &w.Status, &w.UpdatedAt,
		)
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Service) resetDailyCounters(ctx context.Context, tx *sql.Tx, w *models.Wallet) error {
	if time.Since(w.LastDepositReset) > 24*time.Hour {
		w.DailyDepositUsed = 0
		w.LastDepositReset = time.Now()
		_, err := tx.ExecContext(ctx, `
			UPDATE wallets SET daily_deposit_used=$1, last_deposit_reset=$2 WHERE id=$3`,
			w.DailyDepositUsed, w.LastDepositReset, w.ID)
		return err
	}
	return nil
}

func (s *Service) cacheAvailable(ctx context.Context, userID string, w *models.Wallet) {
	if s.redis == nil {
		return
	}
	available := w.DepositBalance + w.BonusBalance + w.WinningsBalance - w.LockedBalance
	_ = s.redis.Set(ctx, s.availableKey(userID), available, 30*time.Second).Err()
}

func (s *Service) availableKey(userID string) string {
	return "wallet:available:" + userID
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
