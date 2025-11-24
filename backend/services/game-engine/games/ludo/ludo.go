package ludo

import (
	"errors"
	"math/rand"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/wallet"
)

type LudoGame struct {
	gameID       string
	entryFee     float64
	walletClient *wallet.WalletClient
}

type LudoState struct {
	Board       map[string]int // Token positions
	CurrentTurn string
	DiceValue   int
	Winner      string
}

func NewLudoGame() *LudoGame {
	return &LudoGame{
		gameID:       "ludo_classic",
		entryFee:     50.0,
		walletClient: wallet.NewWalletClient(),
	}
}

func (g *LudoGame) GetGameID() string { return g.gameID }
func (g *LudoGame) GetGameName() string { return "Ludo Classic" }
func (g *LudoGame) GetGameType() engine.GameType { return engine.GameTypeSkill }
func (g *LudoGame) GetMinPlayers() int { return 2 }
func (g *LudoGame) GetMaxPlayers() int { return 4 }
func (g *LudoGame) GetEntryFee() float64 { return g.entryFee }

func (g *LudoGame) Initialize() error {
	return nil
}

func (g *LudoGame) Start(session *engine.GameSession) error {
	// Deduct entry fee from all players
	for _, p := range session.Players {
		err := g.walletClient.Debit(p.UserID, g.entryFee, session.SessionID, "GAME_LUDO")
		if err != nil {
			// In production, we would refund other players and cancel session
			return errors.New("failed to deduct entry fee for " + p.UserID)
		}
	}

	session.State = &LudoState{
		Board:       make(map[string]int),
		CurrentTurn: session.Players[0].UserID,
	}
	return nil
}

func (g *LudoGame) HandleMove(session *engine.GameSession, move engine.Move) (*engine.MoveResult, error) {
	state := session.State.(*LudoState)

	if move.PlayerID != state.CurrentTurn {
		return nil, errors.New("not your turn")
	}

	if move.Type == "ROLL_DICE" {
		dice := rand.Intn(6) + 1
		state.DiceValue = dice

		// Simple logic: pass turn to next player
		nextPlayerIndex := 0
		for i, p := range session.Players {
			if p.UserID == state.CurrentTurn {
				nextPlayerIndex = (i + 1) % len(session.Players)
				break
			}
		}
		state.CurrentTurn = session.Players[nextPlayerIndex].UserID

		// Check for win condition (simplified: roll 6 to win for demo)
		gameEnded := false
		if dice == 6 {
			state.Winner = move.PlayerID
			gameEnded = true
		}

		return &engine.MoveResult{
			Success:     true,
			NextTurn:    state.CurrentTurn,
			StateUpdate: map[string]interface{}{"dice": dice},
			GameEnded:   gameEnded,
		}, nil
	}

	return nil, errors.New("invalid move type")
}

func (g *LudoGame) End(session *engine.GameSession) (*engine.GameResult, error) {
	state := session.State.(*LudoState)

	// Calculate Prize Pool (Platform fee 10%)
	totalPool := g.entryFee * float64(len(session.Players))
	platformFee := totalPool * 0.10
	prize := totalPool - platformFee

	// Credit Winner
	if state.Winner != "" {
		g.walletClient.Credit(state.Winner, prize, session.SessionID, "GAME_LUDO")
	}

	return &engine.GameResult{
		WinnerID: state.Winner,
		Prizes:   map[string]float64{state.Winner: prize},
	}, nil
}

func (g *LudoGame) GetState(session *engine.GameSession) interface{} {
	return session.State
}
