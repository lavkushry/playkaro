package tournament

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// BracketGenerator handles tournament bracket creation
type BracketGenerator struct{}

// NewBracketGenerator creates a new generator
func NewBracketGenerator() *BracketGenerator {
	return &BracketGenerator{}
}

// GenerateBracket creates a single-elimination bracket
func (bg *BracketGenerator) GenerateBracket(tournamentID string, participants []string) ([]TournamentMatch, error) {
	if len(participants) < 2 {
		return nil, errors.New("need at least 2 participants")
	}

	// Shuffle participants
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	// Calculate bracket size (next power of 2)
	numPlayers := len(participants)
	bracketSize := 1
	for bracketSize < numPlayers {
		bracketSize *= 2
	}

	// Calculate rounds
	numRounds := int(math.Log2(float64(bracketSize)))
	matches := []TournamentMatch{}

	// Create matches level by level, starting from the final (Round 1) up to the first round
	// Note: Round 1 = Final, Round 2 = Semis, etc.
	// Or: Round 1 = First Round, Round N = Final
	// Let's use: Round 1 = First Round, Round N = Final

	// We need to generate matches and link them.
	// Easier approach: Generate from Round 1 (leaves) to Round N (root)

	// But we need to know next_match_id.
	// So we generate from Final (Round N) down to Round 1?
	// Or generate all and then link.

	// Let's generate from Round 1 (First Round)
	// Round 1 has bracketSize/2 matches

	// Wait, handling Byes is tricky.
	// Standard approach:
	// 1. Determine number of byes: bracketSize - numPlayers
	// 2. Place byes in the first round.

	// Let's use a recursive approach or a layer-based approach.

	// Create all match placeholders first
	matchMap := make(map[string]*TournamentMatch) // Key: "Round-Index"

	currentRoundMatches := bracketSize / 2
	totalRounds := numRounds

	// Generate IDs for all matches
	for r := 1; r <= totalRounds; r++ {
		for i := 0; i < currentRoundMatches; i++ {
			matchID := uuid.New().String()
			match := TournamentMatch{
				ID:           matchID,
				TournamentID: tournamentID,
				Round:        r,
				MatchIndex:   i,
				Status:       MatchScheduled,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			key := fmt.Sprintf("%d-%d", r, i)
			matchMap[key] = &match
			matches = append(matches, match)
		}
		currentRoundMatches /= 2
	}

	// Link matches (NextMatchID)
	// Match at Round R, Index I feeds into Match at Round R+1, Index I/2
	for i := range matches {
		m := &matches[i]
		if m.Round < totalRounds {
			nextRound := m.Round + 1
			nextIndex := m.MatchIndex / 2
			key := fmt.Sprintf("%d-%d", nextRound, nextIndex)
			if nextMatch, ok := matchMap[key]; ok {
				nextID := nextMatch.ID
				m.NextMatchID = &nextID
			}
		}
	}

	// Assign players to Round 1
	// We have 'bracketSize' slots in Round 1.
	// Slots 0 to bracketSize-1.
	// Match 0 has slots 0, 1. Match 1 has slots 2, 3.

	// Distribute Byes.
	// Byes should ideally be distributed to top seeds, but here random.
	// Number of byes = bracketSize - numPlayers.
	// Byes mean the opponent is NULL, and the player auto-advances.

	// We place players in the first 'numPlayers' slots? No.
	// Standard seeding for byes:
	// If 4 slots, 3 players. Bye = 1.
	// Match 0: P1 vs P2
	// Match 1: P3 vs Bye

	// Let's fill slots 0 to numPlayers-1 with players.
	// Slots numPlayers to bracketSize-1 are Byes (Empty).

	// Wait, if Match 1 is P3 vs Bye, P3 auto-wins.
	// If Match 1 is Bye vs Bye, that's invalid (shouldn't happen if bracketSize < 2*numPlayers).

	// Assign players to matches in Round 1
	playerIdx := 0
	round1Matches := bracketSize / 2

	for i := 0; i < round1Matches; i++ {
		// Find the match in the slice
		var match *TournamentMatch
		for j := range matches {
			if matches[j].Round == 1 && matches[j].MatchIndex == i {
				match = &matches[j]
				break
			}
		}

		// Slot 1 (2*i)
		if playerIdx < len(participants) {
			pID := participants[playerIdx]
			match.Player1ID = &pID
			playerIdx++
		}

		// Slot 2 (2*i + 1)
		if playerIdx < len(participants) {
			pID := participants[playerIdx]
			match.Player2ID = &pID
			playerIdx++
		}

		// Handle Byes / Auto-win
		if match.Player1ID != nil && match.Player2ID == nil {
			// Player 1 gets a bye
			match.Status = MatchCompleted
			match.WinnerID = match.Player1ID
			// We need to propagate this to next round immediately?
			// Or let the manager handle it. Let's let the manager handle it.
			// But for initial generation, we can mark it.
		} else if match.Player1ID == nil && match.Player2ID != nil {
			// Should not happen with sequential filling
			match.Status = MatchCompleted
			match.WinnerID = match.Player2ID
		} else if match.Player1ID == nil && match.Player2ID == nil {
			// Double bye? Should not happen if bracketSize is correct
			match.Status = MatchCompleted
		}
	}

	return matches, nil
}
