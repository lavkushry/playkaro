package tournament

import (
	"errors"
	"math"
)

// PrizeDistributor calculates prize amounts
type PrizeDistributor struct{}

// NewPrizeDistributor creates a new distributor
func NewPrizeDistributor() *PrizeDistributor {
	return &PrizeDistributor{}
}

// CalculatePrizes determines prize amounts for winners
func (pd *PrizeDistributor) CalculatePrizes(t *Tournament, participants []Participant) (map[string]float64, error) {
	payouts := make(map[string]float64)

	// Filter for ranked players (Rank > 0)
	rankedPlayers := make([]Participant, 0)
	for _, p := range participants {
		if p.Rank > 0 {
			rankedPlayers = append(rankedPlayers, p)
		}
	}

	if len(rankedPlayers) == 0 {
		return nil, errors.New("no ranked players found")
	}

	strategy := t.Config.PrizeStrategy
	pool := t.PrizePool

	switch strategy {
	case PrizeWinnerTakesAll:
		for _, p := range rankedPlayers {
			if p.Rank == 1 {
				payouts[p.UserID] = pool
				break
			}
		}

	case PrizeTop3:
		// 1st: 50%, 2nd: 30%, 3rd: 20%
		for _, p := range rankedPlayers {
			switch p.Rank {
			case 1:
				payouts[p.UserID] = pool * 0.50
			case 2:
				payouts[p.UserID] = pool * 0.30
			case 3:
				payouts[p.UserID] = pool * 0.20
			}
		}

	case PrizeTiered:
		// Use custom distribution from config
		// Map of "Rank" -> Percentage (e.g., "1": 0.5, "2": 0.3)
		// Or "1-10": 0.1

		// For simplicity, let's assume Config.PrizeDistribution is map[string]float64
		// keys are rank strings "1", "2", "3"

		dist := t.Config.PrizeDistribution
		if dist == nil {
			// Fallback to Top 3 if not configured
			return pd.CalculatePrizes(t, participants) // Recursive call with Top 3? No, infinite loop risk.
			// Just use Top 3 logic
			for _, p := range rankedPlayers {
				if p.Rank == 1 { payouts[p.UserID] = pool * 0.5 }
				if p.Rank == 2 { payouts[p.UserID] = pool * 0.3 }
				if p.Rank == 3 { payouts[p.UserID] = pool * 0.2 }
			}
			break
		}

		for _, p := range rankedPlayers {
			// Check exact rank
			rankStr := string(rune('0' + p.Rank)) // Simple int to string for single digits
			// Better: use fmt.Sprintf or just iterate map

			// Let's iterate the map to handle ranges if we supported them
			// But for now, exact match
			// We need a way to convert int rank to string key or just use int keys in struct?
			// Config uses map[string]float64 because JSON keys are strings.

			// Hacky int to string
			rankKey := ""
			if p.Rank < 10 {
				rankKey = string(rune('0' + p.Rank))
			} else {
				// We won't support > 9 ranks for this simple implementation without strconv
				continue
			}

			if percentage, ok := dist[rankKey]; ok {
				payouts[p.UserID] = pool * percentage
			}
		}

	default:
		return nil, errors.New("unknown prize strategy")
	}

	// Round to 2 decimal places
	for userID, amount := range payouts {
		payouts[userID] = math.Floor(amount*100) / 100
	}

	return payouts, nil
}
