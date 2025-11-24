package fantasy

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Constants for team constraints
const (
	MaxPlayers     = 11
	Budget         = 100.0
	MinBatsmen     = 3
	MinBowlers     = 3
	MinWicketKeepers = 1
	MinAllRounders = 1
	MaxFromOneTeam = 7
)

// Player Roles
const (
	RoleBatsman      = "BATSMAN"
	RoleBowler       = "BOWLER"
	RoleAllRounder   = "ALL_ROUNDER"
	RoleWicketKeeper = "WICKET_KEEPER"
)

// FantasyPlayer represents a real-world player
type FantasyPlayer struct {
	PlayerID string  `json:"player_id"`
	Name     string  `json:"name"`
	Team     string  `json:"team"` // Real team (e.g., "IND", "AUS")
	Role     string  `json:"role"`
	Cost     float64 `json:"cost"`
}

// FantasyTeam represents a user's team for a contest
type FantasyTeam struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	ContestID   string          `json:"contest_id"`
	MatchID     string          `json:"match_id"`
	Players     []FantasyPlayer `json:"players"`
	CaptainID   string          `json:"captain_id"`
	ViceCaptainID string        `json:"vice_captain_id"`
	TotalPoints float64         `json:"total_points"`
	Rank        int             `json:"rank"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TeamBuilder handles team creation and validation
type TeamBuilder struct{}

// NewTeamBuilder creates a new team builder
func NewTeamBuilder() *TeamBuilder {
	return &TeamBuilder{}
}

// CreateTeam validates and creates a fantasy team
func (tb *TeamBuilder) CreateTeam(userID, contestID, matchID string, players []FantasyPlayer, captainID, viceCaptainID string) (*FantasyTeam, error) {
	// 1. Validate Player Count
	if len(players) != MaxPlayers {
		return nil, errors.New("team must have exactly 11 players")
	}

	// 2. Validate Budget
	totalCost := 0.0
	for _, p := range players {
		totalCost += p.Cost
	}
	if totalCost > Budget {
		return nil, errors.New("team cost exceeds budget")
	}

	// 3. Validate Roles
	batsmen, bowlers, wks, allrounders := 0, 0, 0, 0
	teamCounts := make(map[string]int)

	playerMap := make(map[string]bool)

	for _, p := range players {
		// Check duplicates
		if playerMap[p.PlayerID] {
			return nil, errors.New("duplicate player in team")
		}
		playerMap[p.PlayerID] = true

		// Count roles
		switch p.Role {
		case RoleBatsman:
			batsmen++
		case RoleBowler:
			bowlers++
		case RoleWicketKeeper:
			wks++
		case RoleAllRounder:
			allrounders++
		}

		// Count teams
		teamCounts[p.Team]++
	}

	if batsmen < MinBatsmen {
		return nil, errors.New("must have at least 3 batsmen")
	}
	if bowlers < MinBowlers {
		return nil, errors.New("must have at least 3 bowlers")
	}
	if wks < MinWicketKeepers {
		return nil, errors.New("must have at least 1 wicket keeper")
	}
	if allrounders < MinAllRounders {
		return nil, errors.New("must have at least 1 all-rounder")
	}

	// 4. Validate Max Players from One Team
	for _, count := range teamCounts {
		if count > MaxFromOneTeam {
			return nil, errors.New("cannot have more than 7 players from one team")
		}
	}

	// 5. Validate Captain & Vice Captain
	if !playerMap[captainID] {
		return nil, errors.New("captain must be in the team")
	}
	if !playerMap[viceCaptainID] {
		return nil, errors.New("vice-captain must be in the team")
	}
	if captainID == viceCaptainID {
		return nil, errors.New("captain and vice-captain must be different")
	}

	return &FantasyTeam{
		ID:            uuid.New().String(),
		UserID:        userID,
		ContestID:     contestID,
		MatchID:       matchID,
		Players:       players,
		CaptainID:     captainID,
		ViceCaptainID: viceCaptainID,
		CreatedAt:     time.Now(),
	}, nil
}
