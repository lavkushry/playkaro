package teenpatti

import (
	"math"
)

// SidePot represents a separate pot for all-in scenarios
type SidePot struct {
	Amount          float64  `json:"amount"`
	EligiblePlayers []string `json:"eligible_players"`
}

// PotManager handles main pot and side pots
type PotManager struct {
	Pots []SidePot
}

// NewPotManager creates a new pot manager
func NewPotManager() *PotManager {
	return &PotManager{
		Pots: []SidePot{},
	}
}

// CalculatePots distributes bets into main and side pots
// bets: map of playerID -> total amount bet in this hand
func (pm *PotManager) CalculatePots(bets map[string]float64) {
	pm.Pots = []SidePot{}

	// Continue until all bets are processed
	for len(bets) > 0 {
		// Find minimum non-zero bet
		minBet := math.MaxFloat64
		for _, amount := range bets {
			if amount > 0 && amount < minBet {
				minBet = amount
			}
		}

		if minBet == math.MaxFloat64 {
			break // No more bets
		}

		// Create pot for this level
		potAmount := 0.0
		eligiblePlayers := []string{}

		activeBetters := []string{}
		for pID, amount := range bets {
			if amount > 0 {
				contribution := minBet
				if amount < minBet {
					contribution = amount
				}

				potAmount += contribution
				bets[pID] -= contribution

				// If player still has money or is all-in at this level, they are eligible
				// Simplified: anyone who contributed to this pot is eligible
				eligiblePlayers = append(eligiblePlayers, pID)

				if bets[pID] == 0 {
					delete(bets, pID)
				} else {
					activeBetters = append(activeBetters, pID)
				}
			}
		}

		if potAmount > 0 {
			pm.Pots = append(pm.Pots, SidePot{
				Amount:          potAmount,
				EligiblePlayers: eligiblePlayers,
			})
		}
	}
}

// DistributeWinnings calculates payouts for each pot
// players: map of playerID -> Player object (to get hand strength)
func (pm *PotManager) DistributeWinnings(players map[string]*Player) map[string]float64 {
	payouts := make(map[string]float64)

	for _, pot := range pm.Pots {
		if pot.Amount == 0 {
			continue
		}

		// Find winner among eligible players
		var winners []string
		var bestRank HandRank

		for _, pID := range pot.EligiblePlayers {
			player := players[pID]
			if player.Status == PlayerStatusFolded {
				continue
			}

			rank := GetHandRank(player.Cards)

			if len(winners) == 0 {
				winners = []string{pID}
				bestRank = rank
			} else {
				// Compare with current best
				// Note: This duplicates EvaluateWinner logic, ideally refactor
				// For brevity, assuming EvaluateWinner returns the better player
				// If pID beats current best, replace winners
				// If tie, append to winners

				// Simplified comparison
				if rank.Type > bestRank.Type {
					winners = []string{pID}
					bestRank = rank
				} else if rank.Type == bestRank.Type {
					// Compare values
					isBetter := false
					isTie := true
					for i := 0; i < 3; i++ {
						if rank.Values[i] > bestRank.Values[i] {
							isBetter = true
							isTie = false
							break
						}
						if rank.Values[i] < bestRank.Values[i] {
							isTie = false
							break
						}
					}

					if isBetter {
						winners = []string{pID}
						bestRank = rank
					} else if isTie {
						winners = append(winners, pID)
					}
				}
			}
		}

		// Split pot among winners
		if len(winners) > 0 {
			share := pot.Amount / float64(len(winners))
			for _, winnerID := range winners {
				payouts[winnerID] += share
			}
		} else {
			// Should not happen unless everyone folded (handled by game logic)
			// Return to last eligible player?
		}
	}

	return payouts
}
