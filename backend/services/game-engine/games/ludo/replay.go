package ludo

import (
	"encoding/json"
	"time"
)

// MoveRecord represents a single move in the game
type MoveRecord struct {
	PlayerID  string    `json:"player_id"`
	DiceRoll  int       `json:"dice_roll"`
	PieceID   string    `json:"piece_id"`
	FromPos   int       `json:"from_pos"`
	ToPos     int       `json:"to_pos"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "MOVE", "CAPTURE", "HOME"
}

// BoardState represents the state of the board at a point in time
type BoardState struct {
	Timestamp     time.Time              `json:"timestamp"`
	CurrentPlayer string                 `json:"current_player"`
	Pieces        map[string][]PiecePos  `json:"pieces"` // playerID -> piece positions
	LastDiceRoll  int                    `json:"last_dice_roll"`
}

// PiecePos represents a piece's position
type PiecePos struct {
	PieceID  string `json:"piece_id"`
	Position int    `json:"position"` // 0 = home, 1-56 = board, 57+ = safe zone
}

// GameReplay stores complete game history
type GameReplay struct {
	SessionID   string        `json:"session_id"`
	GameType    string        `json:"game_type"`
	Players     []string      `json:"players"`
	Moves       []MoveRecord  `json:"moves"`
	States      []BoardState  `json:"states"`
	Winner      string        `json:"winner,omitempty"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

// ReplayRecorder records game moves for later playback
type ReplayRecorder struct {
	replay *GameReplay
}

// NewReplayRecorder creates a new replay recorder
func NewReplayRecorder(sessionID string, players []string) *ReplayRecorder {
	return &ReplayRecorder{
		replay: &GameReplay{
			SessionID: sessionID,
			GameType:  "LUDO_SUPREME",
			Players:   players,
			Moves:     []MoveRecord{},
			States:    []BoardState{},
			StartedAt: time.Now(),
		},
	}
}

// RecordMove adds a move to the replay
func (r *ReplayRecorder) RecordMove(move MoveRecord) {
	move.Timestamp = time.Now()
	r.replay.Moves = append(r.replay.Moves, move)
}

// RecordState saves the current board state
func (r *ReplayRecorder) RecordState(state BoardState) {
	state.Timestamp = time.Now()
	r.replay.States = append(r.replay.States, state)
}

// Complete marks the game as finished
func (r *ReplayRecorder) Complete(winner string) {
	now := time.Now()
	r.replay.Winner = winner
	r.replay.CompletedAt = &now
}

// GetReplay returns the complete replay data
func (r *ReplayRecorder) GetReplay() *GameReplay {
	return r.replay
}

// ToJSON converts replay to JSON for storage
func (r *ReplayRecorder) ToJSON() ([]byte, error) {
	return json.Marshal(r.replay)
}

// FromJSON loads a replay from JSON
func FromJSON(data []byte) (*GameReplay, error) {
	var replay GameReplay
	err := json.Unmarshal(data, &replay)
	if err != nil {
		return nil, err
	}
	return &replay, nil
}

// ReplayPlayer plays back a recorded game
type ReplayPlayer struct {
	replay       *GameReplay
	currentMove  int
	currentState int
}

// NewReplayPlayer creates a new replay player
func NewReplayPlayer(replay *GameReplay) *ReplayPlayer {
	return &ReplayPlayer{
		replay:       replay,
		currentMove:  0,
		currentState: 0,
	}
}

// NextMove returns the next move in the replay
func (p *ReplayPlayer) NextMove() (*MoveRecord, bool) {
	if p.currentMove >= len(p.replay.Moves) {
		return nil, false
	}
	move := p.replay.Moves[p.currentMove]
	p.currentMove++
	return &move, true
}

// NextState returns the next board state
func (p *ReplayPlayer) NextState() (*BoardState, bool) {
	if p.currentState >= len(p.replay.States) {
		return nil, false
	}
	state := p.replay.States[p.currentState]
	p.currentState++
	return &state, true
}

// Reset restarts the replay from the beginning
func (p *ReplayPlayer) Reset() {
	p.currentMove = 0
	p.currentState = 0
}

// GetProgress returns current playback progress (0.0 - 1.0)
func (p *ReplayPlayer) GetProgress() float64 {
	if len(p.replay.Moves) == 0 {
		return 0.0
	}
	return float64(p.currentMove) / float64(len(p.replay.Moves))
}
