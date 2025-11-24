package crash

import (
	"errors"
	"math"
	"time"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/fairness"
)

type CrashGame struct {
	gameID   string
	entryFee float64
}

type CrashState struct {
	Status          string             `json:"status"` // WAITING, FLYING, CRASHED
	Multiplier      float64            `json:"multiplier"`
	StartTime       time.Time          `json:"start_time"`
	NextRoundIn     int                `json:"next_round_in"` // Seconds
	Bets            map[string]*Bet    `json:"bets"`
	History         []float64          `json:"history"`
	ServerSeedHash  string             `json:"server_seed_hash"`
}

type Bet struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"`
	AutoCashout float64 `json:"auto_cashout"`
	CashoutAt   float64 `json:"cashout_at"` // 0 if not cashed out
	Profit      float64 `json:"profit"`
}

func NewCrashGame() *CrashGame {
	return &CrashGame{
		gameID:   "crash_aviator",
		entryFee: 10.0, // Min bet
	}
}

func (g *CrashGame) GetGameID() string { return g.gameID }
func (g *CrashGame) GetGameName() string { return "Crash (Aviator)" }
func (g *CrashGame) GetGameType() engine.GameType { return engine.GameTypeCasino }
func (g *CrashGame) GetMinPlayers() int { return 1 }
func (g *CrashGame) GetMaxPlayers() int { return 1000 } // Unlimited
func (g *CrashGame) GetEntryFee() float64 { return g.entryFee }

func (g *CrashGame) Initialize() error {
	return nil
}

func (g *CrashGame) Start(session *engine.GameSession) error {
	// Crash is a continuous game loop, handled by a separate runner
	// For this architecture, we initialize the state
	session.State = &CrashState{
		Status:     "WAITING",
		Multiplier: 1.00,
		Bets:       make(map[string]*Bet),
		History:    []float64{},
		ServerSeedHash: fairness.HashServerSeed("secret_seed"), // Demo seed
	}

	// Start the game loop in a goroutine
	go g.RunGameLoop(session)

	return nil
}

func (g *CrashGame) RunGameLoop(session *engine.GameSession) {
	state := session.State.(*CrashState)

	for {
		// 1. WAITING PHASE (5 seconds)
		state.Status = "WAITING"
		state.Multiplier = 1.00
		state.Bets = make(map[string]*Bet) // Clear bets

		for i := 5; i > 0; i-- {
			state.NextRoundIn = i
			time.Sleep(1 * time.Second)
		}

		// 2. CALCULATE CRASH POINT
		// In production, use rotated seeds
		crashPoint := fairness.CalculateCrashPoint("secret_seed", "public_seed", int(time.Now().Unix()))

		// 3. FLYING PHASE
		state.Status = "FLYING"
		startTime := time.Now()

		for {
			elapsed := time.Since(startTime).Seconds()

			// Growth function: 1.00 * e^(0.06 * t)
			// This makes it grow slowly then fast
			currentMult := math.Exp(0.06 * elapsed)

			if currentMult >= crashPoint {
				state.Multiplier = crashPoint
				break
			}

			state.Multiplier = currentMult

			// Check auto-cashouts
			g.processAutoCashouts(state)

			time.Sleep(100 * time.Millisecond) // 10 updates/sec
		}

		// 4. CRASH PHASE
		state.Status = "CRASHED"
		state.History = append(state.History, state.Multiplier)
		if len(state.History) > 10 {
			state.History = state.History[1:]
		}

		time.Sleep(3 * time.Second)
	}
}

func (g *CrashGame) processAutoCashouts(state *CrashState) {
	for _, bet := range state.Bets {
		if bet.CashoutAt == 0 && bet.AutoCashout > 0 && state.Multiplier >= bet.AutoCashout {
			bet.CashoutAt = bet.AutoCashout
			bet.Profit = bet.Amount * bet.AutoCashout
		}
	}
}

func (g *CrashGame) HandleMove(session *engine.GameSession, move engine.Move) (*engine.MoveResult, error) {
	state := session.State.(*CrashState)

	if move.Type == "BET" {
		if state.Status != "WAITING" {
			return nil, errors.New("can only bet during waiting phase")
		}

		amount := move.Data["amount"].(float64)
		autoCashout := 0.0
		if val, ok := move.Data["auto_cashout"]; ok {
			autoCashout = val.(float64)
		}

		state.Bets[move.PlayerID] = &Bet{
			UserID:      move.PlayerID,
			Amount:      amount,
			AutoCashout: autoCashout,
		}

		return &engine.MoveResult{Success: true}, nil
	}

	if move.Type == "CASHOUT" {
		if state.Status != "FLYING" {
			return nil, errors.New("can only cashout while flying")
		}

		bet, exists := state.Bets[move.PlayerID]
		if !exists {
			return nil, errors.New("no active bet")
		}

		if bet.CashoutAt > 0 {
			return nil, errors.New("already cashed out")
		}

		bet.CashoutAt = state.Multiplier
		bet.Profit = bet.Amount * state.Multiplier

		return &engine.MoveResult{
			Success: true,
			StateUpdate: map[string]interface{}{
				"profit": bet.Profit,
			},
		}, nil
	}

	return nil, errors.New("invalid move type")
}

func (g *CrashGame) End(session *engine.GameSession) (*engine.GameResult, error) {
	// Crash never "ends" in the traditional sense, it loops
	return nil, nil
}

func (g *CrashGame) GetState(session *engine.GameSession) interface{} {
	return session.State
}
