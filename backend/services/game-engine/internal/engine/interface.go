package engine

import (
	"time"
)

// GameType defines the category of game
type GameType string

const (
	GameTypeSkill  GameType = "SKILL"
	GameTypeCasino GameType = "CASINO"
	GameTypePuzzle GameType = "PUZZLE"
)

// IGame is the interface that all games must implement
type IGame interface {
	// Metadata
	GetGameID() string
	GetGameName() string
	GetGameType() GameType
	GetMinPlayers() int
	GetMaxPlayers() int
	GetEntryFee() float64

	// Lifecycle
	Initialize() error
	Start(session *GameSession) error
	HandleMove(session *GameSession, move Move) (*MoveResult, error)
	End(session *GameSession) (*GameResult, error)

	// State
	GetState(session *GameSession) interface{}
}

// GameSession represents an active game instance
type GameSession struct {
	SessionID string
	GameID    string
	Players   []*Player
	State     interface{} // Game-specific state
	Status    string
	EntryFee  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Player represents a participant in a game session
type Player struct {
	UserID   string
	Username string
	Score    int
	IsTurn   bool
}

// Move represents an action taken by a player
type Move struct {
	PlayerID string
	Type     string
	Data     map[string]interface{}
}

// MoveResult represents the outcome of a move
type MoveResult struct {
	Success     bool
	NextTurn    string
	StateUpdate interface{}
	GameEnded   bool
}

// GameResult represents the final outcome of a game
type GameResult struct {
	WinnerID string
	Scores   map[string]int
	Prizes   map[string]float64
}
