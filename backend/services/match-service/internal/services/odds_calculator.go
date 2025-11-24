package services

import (
	"database/sql"
	"log"
	"math"
)

const (
	MinOdds            = 1.01
	MaxOdds            = 100.0
	MaxSingleAdjustment = 0.20 // Max 20% change per update
	SensitivityFactor  = 0.3   // Controls aggressiveness of adjustments
)

type OddsCalculator struct {
	DB *sql.DB
}

func NewOddsCalculator(db *sql.DB) *OddsCalculator {
	return &OddsCalculator{DB: db}
}

// BettingDistribution represents betting volume per outcome
type BettingDistribution struct {
	TeamAVolume   float64
	TeamBVolume   float64
	DrawVolume    float64
	TeamALiability float64
	TeamBLiability float64
	DrawLiability  float64
	TotalPool     float64
}

// OddsAdjustment represents new calculated odds
type OddsAdjustment struct {
	OddsA float64
	OddsB float64
	OddsDraw float64
	Reason string
}

// CalculateOdds computes new odds based on Kelly Criterion
func (o *OddsCalculator) CalculateOdds(matchID string) (*OddsAdjustment, error) {
	// Get current odds
	var currentOddsA, currentOddsB, currentOddsDraw float64
	err := o.DB.QueryRow(`
		SELECT odds_a, odds_b, odds_draw
		FROM matches
		WHERE match_id = $1
	`, matchID).Scan(&currentOddsA, &currentOddsB, &currentOddsDraw)

	if err != nil {
		return nil, err
	}

	// Get betting distribution
	dist, err := o.GetBettingDistribution(matchID)
	if err != nil {
		return nil, err
	}

	// If no bets, keep current odds
	if dist.TotalPool == 0 {
		return &OddsAdjustment{
			OddsA: currentOddsA,
			OddsB: currentOddsB,
			OddsDraw: currentOddsDraw,
			Reason: "No betting activity",
		}, nil
	}

	// Calculate liability ratios
	teamALiabilityRatio := dist.TeamALiability / dist.TotalPool
	teamBLiabilityRatio := dist.TeamBLiability / dist.TotalPool
	drawLiabilityRatio := dist.DrawLiability / dist.TotalPool

	// Calculate adjustment factors (Kelly Criterion)
	teamAAdjustment := teamALiabilityRatio * SensitivityFactor
	teamBAdjustment := teamBLiabilityRatio * SensitivityFactor
	drawAdjustment := drawLiabilityRatio * SensitivityFactor

	// Apply adjustments (heavy betting = lower odds)
	newOddsA := currentOddsA * (1 - teamAAdjustment)
	newOddsB := currentOddsB * (1 - teamBAdjustment)
	newOddsDraw := currentOddsDraw * (1 - drawAdjustment)

	// Constrain to max single adjustment
	newOddsA = o.constrainAdjustment(currentOddsA, newOddsA)
	newOddsB = o.constrainAdjustment(currentOddsB, newOddsB)
	newOddsDraw = o.constrainAdjustment(currentOddsDraw, newOddsDraw)

	// Enforce min/max odds
	newOddsA = o.clampOdds(newOddsA)
	newOddsB = o.clampOdds(newOddsB)
	newOddsDraw = o.clampOdds(newOddsDraw)

	return &OddsAdjustment{
		OddsA: newOddsA,
		OddsB: newOddsB,
		OddsDraw: newOddsDraw,
		Reason: "Kelly Criterion adjustment",
	}, nil
}

// GetBettingDistribution analyzes current bet volumes
func (o *OddsCalculator) GetBettingDistribution(matchID string) (*BettingDistribution, error) {
	var dist BettingDistribution

	rows, err := o.DB.Query(`
		SELECT
			team,
			SUM(amount) as volume,
			SUM(potential_win) as liability
		FROM bets
		WHERE match_id = $1 AND status = 'ACTIVE'
		GROUP BY team
	`, matchID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var team string
		var volume, liability float64
		rows.Scan(&team, &volume, &liability)

		switch team {
		case "TEAM_A":
			dist.TeamAVolume = volume
			dist.TeamALiability = liability
		case "TEAM_B":
			dist.TeamBVolume = volume
			dist.TeamBLiability = liability
		case "DRAW":
			dist.DrawVolume = volume
			dist.DrawLiability = liability
		}
	}

	dist.TotalPool = dist.TeamAVolume + dist.TeamBVolume + dist.DrawVolume

	return &dist, nil
}

// constrainAdjustment limits odds change to max 20% per update
func (o *OddsCalculator) constrainAdjustment(oldOdds, newOdds float64) float64 {
	change := (newOdds - oldOdds) / oldOdds

	if change > MaxSingleAdjustment {
		return oldOdds * (1 + MaxSingleAdjustment)
	}
	if change < -MaxSingleAdjustment {
		return oldOdds * (1 - MaxSingleAdjustment)
	}

	return newOdds
}

// clampOdds enforces min/max odds boundaries
func (o *OddsCalculator) clampOdds(odds float64) float64 {
	return math.Max(MinOdds, math.Min(MaxOdds, odds))
}

// UpdateMatchOdds applies new odds and logs to history
func (o *OddsCalculator) UpdateMatchOdds(matchID string, adj *OddsAdjustment) error {
	tx, err := o.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update match odds
	_, err = tx.Exec(`
		UPDATE matches
		SET odds_a = $1, odds_b = $2, odds_draw = $3, updated_at = NOW()
		WHERE match_id = $4
	`, adj.OddsA, adj.OddsB, adj.OddsDraw, matchID)

	if err != nil {
		return err
	}

	// Log to odds history
	_, err = tx.Exec(`
		INSERT INTO odds_history (match_id, odds_a, odds_b, odds_draw, triggered_by)
		SELECT $1, $2, $3, $4, $5
	`, matchID, adj.OddsA, adj.OddsB, adj.OddsDraw, adj.Reason)

	if err != nil {
		log.Printf("Failed to log odds history: %v", err)
	}

	return tx.Commit()
}
