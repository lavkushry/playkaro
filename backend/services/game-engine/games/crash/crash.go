package crash

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/fairness"
	"github.com/playkaro/game-engine/internal/wallet"
)

type CrashGame struct {
	gameID       string
	entryFee     float64
	walletClient *wallet.WalletClient
}

type CrashState struct {
	Status          string             `json:"status"` // WAITING, FLYING, CRASHED
	Multiplier      float64            `json:"multiplier"`
	StartTime       time.Time          `json:"start_time"`
	NextRoundIn     int                `json:"next_round_in"` // Seconds
	Bets            map[string]*Bet    `json:"bets"`
	History         []float64          `json:"history"`
	ServerSeedHash  string             `json:"server_seed_hash"`
	CurrentRoundID  string             `json:"current_round_id"`
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
		gameID:       "crash_aviator",
		entryFee:     10.0, // Min bet
		walletClient: wallet.NewWalletClient(),
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
	session.State = &CrashState{
		Status:         "WAITING",
		Multiplier:     1.00,
		Bets:           make(map[string]*Bet),
		History:        []float64{},
		ServerSeedHash: fairness.HashServerSeed("secret_seed"),
		CurrentRoundID: fmt.Sprintf("round_%d", time.Now().Unix()),
	}

	go g.RunGameLoop(session)

	return nil
}

func (g *CrashGame) RunGameLoop(session *engine.GameSession) {
	state := session.State.(*CrashState)

	for {
		// 1. WAITING PHASE
		state.Status = "WAITING"
		state.Multiplier = 1.00
		state.Bets = make(map[string]*Bet)
		state.CurrentRoundID = fmt.Sprintf("round_%d", time.Now().Unix())

		for i := 5; i > 0; i-- {
			state.NextRoundIn = i
			time.Sleep(1 * time.Second)
		}

		// 2. CALCULATE CRASH POINT
		crashPoint := fairness.CalculateCrashPoint("secret_seed", "public_seed", int(time.Now().Unix()))

		// 3. FLYING PHASE
		state.Status = "FLYING"
		startTime := time.Now()

		for {
			elapsed := time.Since(startTime).Seconds()
			currentMult := math.Exp(0.06 * elapsed)

			if currentMult >= crashPoint {
				state.Multiplier = crashPoint
				break
			}

			state.Multiplier = currentMult
			g.processAutoCashouts(state)
			time.Sleep(100 * time.Millisecond)
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
			g.cashoutUser(bet, bet.AutoCashout, state.CurrentRoundID)
		}
	}
}

func (g *CrashGame) cashoutUser(bet *Bet, multiplier float64, roundID string) {
	bet.CashoutAt = multiplier
	bet.Profit = bet.Amount * multiplier

	// Credit winnings to wallet
	// Note: We credit the FULL amount (Stake + Profit) because we debited the stake earlier
	g.walletClient.Credit(bet.UserID, bet.Profit, roundID, "GAME_CRASH")
}

func (g *CrashGame) HandleMove(session *engine.GameSession, move engine.Move) (*engine.MoveResult, error) {
	state := session.State.(*CrashState)

	if move.Type == "BET" {
		if state.Status != "WAITING" {
			return nil, errors.New("can only bet during waiting phase")
		}

		amount := move.Data["amount"].(float64)
		if amount < g.entryFee {
			return nil, errors.New("bet amount too low")
		}

		// Deduct bet from wallet
		err := g.walletClient.Debit(move.PlayerID, amount, state.CurrentRoundID, "GAME_CRASH")
		if err != nil {
			return nil, fmt.Errorf("bet failed: %v", err)
		}

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

		g.cashoutUser(bet, state.Multiplier, state.CurrentRoundID)

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
	return nil, nil
}

func (g *CrashGame) GetState(session *engine.GameSession) interface{} {
	return session.State
}
