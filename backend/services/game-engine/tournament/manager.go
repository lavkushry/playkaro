package tournament

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TournamentManager handles tournament lifecycle
type TournamentManager struct {
	DB               *sql.DB
	BracketGen       *BracketGenerator
	PrizeDistributor *PrizeDistributor
}

// NewTournamentManager creates a new manager
func NewTournamentManager(db *sql.DB) *TournamentManager {
	return &TournamentManager{
		DB:               db,
		BracketGen:       NewBracketGenerator(),
		PrizeDistributor: NewPrizeDistributor(),
	}
}

// CreateTournament creates a new tournament
func (tm *TournamentManager) CreateTournament(name, gameType string, entryFee, prizePool float64, maxPlayers int, config TournamentConfig) (*Tournament, error) {
	t := &Tournament{
		ID:         uuid.New().String(),
		Name:       name,
		GameType:   gameType,
		EntryFee:   entryFee,
		PrizePool:  prizePool,
		MaxPlayers: maxPlayers,
		Status:     StatusRegistration,
		Config:     config,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to DB (omitted for brevity, assume standard SQL insert)
	// In real implementation:
	// _, err := tm.DB.Exec(...)

	return t, nil
}

// RegisterParticipant adds a user to the tournament
func (tm *TournamentManager) RegisterParticipant(tournamentID, userID string) error {
	// Check if tournament is in REGISTRATION state
	// Check if max players reached
	// Deduct entry fee (via wallet service)

	// Add to DB
	return nil
}

// StartTournament generates brackets and starts the tournament
func (tm *TournamentManager) StartTournament(tournamentID string) error {
	// Get participants
	// participants := tm.getParticipants(tournamentID)
	// Mock participants for now
	participants := []string{"p1", "p2", "p3", "p4"} // Example

	// Generate Bracket
	matches, err := tm.BracketGen.GenerateBracket(tournamentID, participants)
	if err != nil {
		return err
	}
	_ = matches // Mock save

	// Save matches to DB
	// Update tournament status to ACTIVE

	return nil
}

// AdvanceMatch updates a match result and progresses the winner
func (tm *TournamentManager) AdvanceMatch(matchID, winnerID string) error {
	// Get match
	// match := tm.getMatch(matchID)

	// Update match status to COMPLETED, set WinnerID

	// If NextMatchID is set:
	//   Get next match
	//   If Player1ID is nil, set Player1ID = winnerID
	//   Else if Player2ID is nil, set Player2ID = winnerID
	//   Update next match in DB

	// If NextMatchID is nil, this was the final.
	//   CompleteTournament(tournamentID, winnerID)

	return nil
}

// CompleteTournament finishes the tournament and distributes prizes
func (tm *TournamentManager) CompleteTournament(tournamentID, winnerID string) error {
	// Update tournament status to COMPLETED

	// Calculate ranks
	// Winner = Rank 1
	// Runner-up = Rank 2 (Loser of final)
	// Semi-final losers = Rank 3/4

	// Distribute prizes
	// prizes, err := tm.PrizeDistributor.CalculatePrizes(tournament, rankedParticipants)
	// Credit wallets

	return nil
}
