package fraud

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Detector struct {
	DB *sql.DB
}

type CheckResult struct {
	Passed    bool
	RiskScore int
	Reason    string
}

func NewDetector(db *sql.DB) *Detector {
	return &Detector{DB: db}
}

// CheckDepositVelocity checks if user has exceeded deposit limits
func (d *Detector) CheckDepositVelocity(ctx context.Context, userID string) (*CheckResult, error) {
	// Check: Max 5 deposits per hour
	var count int
	hourAgo := time.Now().Add(-1 * time.Hour)

	err := d.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM payment_orders
		WHERE user_id = $1
		AND type = 'DEPOSIT'
		AND created_at >= $2
	`, userID, hourAgo).Scan(&count)

	if err != nil {
		return nil, err
	}

	if count >= 5 {
		return &CheckResult{
			Passed:    false,
			RiskScore: 80,
			Reason:    fmt.Sprintf("Exceeded hourly deposit limit: %d deposits in last hour", count),
		}, nil
	}

	return &CheckResult{Passed: true, RiskScore: 20}, nil
}

// CheckDailyDepositLimit checks if user has exceeded daily deposit amount limit
func (d *Detector) CheckDailyDepositLimit(ctx context.Context, userID string, amount float64) (*CheckResult, error) {
	var totalAmount float64
	dayAgo := time.Now().Add(-24 * time.Hour)

	err := d.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_orders
		WHERE user_id = $1
		AND type = 'DEPOSIT'
		AND status = 'SUCCESS'
		AND created_at >= $2
	`, userID, dayAgo).Scan(&totalAmount)

	if err != nil {
		return nil, err
	}

	// Max ₹50,000 per day
	maxDaily := 50000.0
	if totalAmount+amount > maxDaily {
		return &CheckResult{
			Passed:    false,
			RiskScore: 90,
			Reason:    fmt.Sprintf("Exceeded daily deposit limit: ₹%.2f/₹%.2f", totalAmount, maxDaily),
		}, nil
	}

	// Medium risk if close to limit
	if totalAmount+amount > maxDaily*0.8 {
		return &CheckResult{
			Passed:    true,
			RiskScore: 60,
			Reason:    "Approaching daily deposit limit",
		}, nil
	}

	return &CheckResult{Passed: true, RiskScore: 10}, nil
}

// CheckSuspiciousAmount checks if amount is unusually high
func (d *Detector) CheckSuspiciousAmount(ctx context.Context, userID string, amount float64) (*CheckResult, error) {
	// Check average deposit amount for this user
	var avgAmount sql.NullFloat64

	err := d.DB.QueryRowContext(ctx, `
		SELECT AVG(amount)
		FROM payment_orders
		WHERE user_id = $1
		AND type = 'DEPOSIT'
		AND status = 'SUCCESS'
	`, userID).Scan(&avgAmount)

	if err != nil {
		return nil, err
	}

	// If first deposit or no history
	if !avgAmount.Valid {
		// Flag deposits over ₹10,000 for new users
		if amount > 10000 {
			return &CheckResult{
				Passed:    true,
				RiskScore: 50,
				Reason:    "First deposit with high amount",
			}, nil
		}
		return &CheckResult{Passed: true, RiskScore: 10}, nil
	}

	// If current amount is 5x the average, flag it
	if amount > avgAmount.Float64*5 {
		return &CheckResult{
			Passed:    true,
			RiskScore: 70,
			Reason:    fmt.Sprintf("Unusual amount: ₹%.2f vs avg ₹%.2f", amount, avgAmount.Float64),
		}, nil
	}

	return &CheckResult{Passed: true, RiskScore: 10}, nil
}

// RunAllChecks runs all fraud detection checks
func (d *Detector) RunAllChecks(ctx context.Context, userID string, amount float64) (*CheckResult, error) {
	checks := []func(context.Context, string) (*CheckResult, error){
		func(ctx context.Context, uid string) (*CheckResult, error) {
			return d.CheckDepositVelocity(ctx, uid)
		},
	}

	checksWithAmount := []func(context.Context, string, float64) (*CheckResult, error){
		d.CheckDailyDepositLimit,
		d.CheckSuspiciousAmount,
	}

	maxRiskScore := 0
	var failReason string

	// Run checks without amount
	for _, check := range checks {
		result, err := check(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !result.Passed {
			return result, nil
		}
		if result.RiskScore > maxRiskScore {
			maxRiskScore = result.RiskScore
			failReason = result.Reason
		}
	}

	// Run checks with amount
	for _, check := range checksWithAmount {
		result, err := check(ctx, userID, amount)
		if err != nil {
			return nil, err
		}
		if !result.Passed {
			return result, nil
		}
		if result.RiskScore > maxRiskScore {
			maxRiskScore = result.RiskScore
			failReason = result.Reason
		}
	}

	return &CheckResult{
		Passed:    true,
		RiskScore: maxRiskScore,
		Reason:    failReason,
	}, nil
}
