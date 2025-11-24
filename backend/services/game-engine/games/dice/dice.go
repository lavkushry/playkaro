package dice

import (
	"errors"
	"fmt"
	"time"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/fairness"
	"github.com/playkaro/game-engine/internal/wallet"
)

type DiceGame struct {
	gameID       string
	entryFee     float64
	walletClient *wallet.WalletClient
}

type DiceState struct {
	LastRoll  float64 `json:"last_roll"`
	Target    float64 `json:"target"`
	Condition string  `json:"condition"` // OVER, UNDER
	Win       bool    `json:"win"`
	Profit    float64 `json:"profit"`
}

func NewDiceGame() *DiceGame {
	return &DiceGame{
		gameID:       "dice_classic",
		entryFee:     1.0,
		walletClient: wallet.NewWalletClient(),
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
		amount := move.Data["amount"].(float64)
		target := move.Data["target"].(float64)
		condition := move.Data["condition"].(string)

		roundID := fmt.Sprintf("dice_%s_%d", session.SessionID, time.Now().UnixNano())

		// 1. Deduct Bet
		err := g.walletClient.Debit(move.PlayerID, amount, roundID, "GAME_DICE")
		if err != nil {
			return nil, fmt.Errorf("bet failed: %v", err)
		}

		// 2. Generate Result
		roll := fairness.GenerateFloat("server_seed", "client_seed", 1) * 100

		win := false
		multiplier := 0.0

		if condition == "UNDER" {
			if roll < target {
				win = true
				multiplier = 99.0 / target // Standard Dice Multiplier Formula
			}
		} else if condition == "OVER" {
			if roll > target {
				win = true
				multiplier = 99.0 / (100.0 - target)
			}
		}

		profit := 0.0
		if win {
			profit = amount * multiplier
			// 3. Credit Winnings
			g.walletClient.Credit(move.PlayerID, profit, roundID, "GAME_DICE")
		}

		state := &DiceState{
			LastRoll:  roll,
			Target:    target,
			Condition: condition,
			Win:       win,
			Profit:    profit,
		}
		session.State = state

		return &engine.MoveResult{
			Success:     true,
			StateUpdate: state,
			GameEnded:   true,
		}, nil
	}

	return nil, errors.New("invalid move type")
}

func (g *DiceGame) End(session *engine.GameSession) (*engine.GameResult, error) {
	return &engine.GameResult{}, nil
}

func (g *DiceGame) GetState(session *engine.GameSession) interface{} {
	return session.State
}
