package dice

import (
	"errors"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/fairness"
)

type DiceGame struct {
	gameID   string
	entryFee float64
}

type DiceState struct {
	LastRoll  float64 `json:"last_roll"`
	Target    float64 `json:"target"`
	Condition string  `json:"condition"` // OVER, UNDER
	Win       bool    `json:"win"`
}

func NewDiceGame() *DiceGame {
	return &DiceGame{
		gameID:   "dice_classic",
		entryFee: 1.0,
	}
}

func (g *DiceGame) GetGameID() string { return g.gameID }
func (g *DiceGame) GetGameName() string { return "Dice" }
func (g *DiceGame) GetGameType() engine.GameType { return engine.GameTypeCasino }
func (g *DiceGame) GetMinPlayers() int { return 1 }
func (g *DiceGame) GetMaxPlayers() int { return 1 }
func (g *DiceGame) GetEntryFee() float64 { return g.entryFee }

func (g *DiceGame) Initialize() error {
	return nil
}

func (g *DiceGame) Start(session *engine.GameSession) error {
	session.State = &DiceState{}
	return nil
}

func (g *DiceGame) HandleMove(session *engine.GameSession, move engine.Move) (*engine.MoveResult, error) {
	if move.Type == "ROLL" {
		target := move.Data["target"].(float64)
		condition := move.Data["condition"].(string) // "OVER" or "UNDER"

		// Generate Provably Fair Roll (0.00 - 100.00)
		// In production, use actual client/server seeds
		roll := fairness.GenerateFloat("server_seed", "client_seed", 1) * 100

		win := false
		if condition == "UNDER" && roll < target {
			win = true
		} else if condition == "OVER" && roll > target {
			win = true
		}

		state := &DiceState{
			LastRoll:  roll,
			Target:    target,
			Condition: condition,
			Win:       win,
		}
		session.State = state

		return &engine.MoveResult{
			Success:     true,
			StateUpdate: state,
			GameEnded:   true, // Dice is instant
		}, nil
	}

	return nil, errors.New("invalid move type")
}

func (g *DiceGame) End(session *engine.GameSession) (*engine.GameResult, error) {
	state := session.State.(*DiceState)
	if state.Win {
		return &engine.GameResult{
			WinnerID: session.Players[0].UserID,
			// Calculate prize based on probability
		}, nil
	}
	return &engine.GameResult{}, nil
}

func (g *DiceGame) GetState(session *engine.GameSession) interface{} {
	return session.State
}
