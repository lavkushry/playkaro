package services

import (
	"database/sql"
	"time"
)

const (
	MaxLiabilityThreshold = 500000.0 // ₹5 lakh max exposure per match
	MaxOddsVolatility     = 0.30     // 30% change triggers suspension
	MaxUserBetRatio       = 0.20     // Single user can't bet >20% of pool
)

type MarketController struct {
	DB *sql.DB
}

func NewMarketController(db *sql.DB) *MarketController {
	return &MarketController{DB: db}
}

// SuspensionCheck result
type SuspensionCheck struct {
	ShouldSuspend bool
	Reason        string
}

// CheckSuspensionTriggers evaluates if market should be suspended
func (m *MarketController) CheckSuspensionTriggers(matchID string) (*SuspensionCheck, error) {
	// Rule 1: High Liability
	totalLiability, err := m.getTotalLiability(matchID)
	if err != nil {
		return nil, err
	}

	if totalLiability > MaxLiabilityThreshold {
		return &SuspensionCheck{
			ShouldSuspend: true,
			Reason:        "Maximum liability threshold exceeded",
		}, nil
	}

	// Rule 2: High Odds Volatility
	volatility, err := m.getOddsVolatility(matchID, 5*time.Minute)
	if err == nil && volatility > MaxOddsVolatility {
		return &SuspensionCheck{
			ShouldSuspend: true,
			Reason:        "High odds volatility detected",
		}, nil
	}

	// Rule 3: Suspicious Betting Pattern
	suspicious, err := m.detectSuspiciousActivity(matchID)
	if err == nil && suspicious {
		return &SuspensionCheck{
			ShouldSuspend: true,
			Reason:        "Suspicious betting pattern detected",
		}, nil
	}

	return &SuspensionCheck{
		ShouldSuspend: false,
	}, nil
}

// getTotalLiability calculates total potential payout
func (m *MarketController) getTotalLiability(matchID string) (float64, error) {
	var totalLiability float64
	err := m.DB.QueryRow(`
		SELECT COALESCE(SUM(potential_win), 0)
		FROM bets
		WHERE match_id = $1 AND status = 'ACTIVE'
	`, matchID).Scan(&totalLiability)

	return totalLiability, err
}

// getOddsVolatility calculates odds change % over time window
func (m *MarketController) getOddsVolatility(matchID string, window time.Duration) (float64, error) {
	cutoff := time.Now().Add(-window)

	rows, err := m.DB.Query(`
		SELECT odds_a, odds_b, odds_draw, created_at
		FROM odds_history
		WHERE match_id = $1 AND created_at > $2
		ORDER BY created_at ASC
		LIMIT 2
	`, matchID, cutoff)

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var oddsRecords []struct {
		OddsA, OddsB, OddsDraw float64
		CreatedAt              time.Time
	}

	for rows.Next() {
		var rec struct {
			OddsA, OddsB, OddsDraw float64
			CreatedAt              time.Time
		}
		rows.Scan(&rec.OddsA, &rec.OddsB, &rec.OddsDraw, &rec.CreatedAt)
		oddsRecords = append(oddsRecords, rec)
	}

	if len(oddsRecords) < 2 {
		return 0, nil // Not enough data
	}

	// Calculate max volatility across all odds
	first := oddsRecords[0]
	last := oddsRecords[len(oddsRecords)-1]

	volatilityA := abs((last.OddsA - first.OddsA) / first.OddsA)
	volatilityB := abs((last.OddsB - first.OddsB) / first.OddsB)
	volatilityDraw := abs((last.OddsDraw - first.OddsDraw) / first.OddsDraw)

	return max(volatilityA, max(volatilityB, volatilityDraw)), nil
}

// detectSuspiciousActivity checks for unusual betting patterns
func (m *MarketController) detectSuspiciousActivity(matchID string) (bool, error) {
	// Get total pool
	var totalPool float64
	err := m.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM bets
		WHERE match_id = $1 AND status = 'ACTIVE'
	`, matchID).Scan(&totalPool)

	if err != nil || totalPool == 0 {
		return false, err
	}

	// Check if any single user has bet >20% of total
	var maxUserBet float64
	err = m.DB.QueryRow(`
		SELECT COALESCE(MAX(user_total), 0)
		FROM (
			SELECT SUM(amount) as user_total
			FROM bets
			WHERE match_id = $1 AND status = 'ACTIVE'
			GROUP BY user_id
		) user_bets
	`, matchID).Scan(&maxUserBet)

	if err != nil {
		return false, err
	}

	ratio := maxUserBet / totalPool
	return ratio > MaxUserBetRatio, nil
}

// SuspendMarket suspends betting on a match
func (m *MarketController) SuspendMarket(matchID, reason string) error {
	_, err := m.DB.Exec(`
		UPDATE matches
		SET suspended = TRUE, suspension_reason = $1, updated_at = NOW()
		WHERE match_id = $2
	`, reason, matchID)

	return err
}

// ResumeMarket resumes betting on a match
func (m *MarketController) ResumeMarket(matchID string) error {
	_, err := m.DB.Exec(`
		UPDATE matches
		SET suspended = FALSE, suspension_reason = NULL, updated_at = NOW()
		WHERE match_id = $1
	`, matchID)

	return err
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
