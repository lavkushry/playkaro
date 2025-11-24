package teenpatti

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Game Constants
const (
	MinPlayers = 2
	MaxPlayers = 6
	TurnTime   = 30 * time.Second
)

// Player Status
const (
	PlayerStatusActive  = "ACTIVE"
	PlayerStatusFolded  = "FOLDED"
	PlayerStatusAllIn   = "ALL_IN"
	PlayerStatusLeft    = "LEFT"
)

// Game State
const (
	StateWaiting  = "WAITING"
	StateDealing  = "DEALING"
	StateBetting  = "BETTING"
	StateShowdown = "SHOWDOWN"
	StateFinished = "FINISHED"
)

// Card represents a playing card
type Card struct {
	Suit  string `json:"suit"`  // "H", "D", "C", "S"
	Value int    `json:"value"` // 2-14 (14 = Ace)
}

// Player represents a participant in the game
type Player struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Avatar    string  `json:"avatar"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
	Cards     []Card  `json:"cards"`
	IsBlind   bool    `json:"is_blind"`
	SeenCards bool    `json:"seen_cards"`
	TotalBet  float64 `json:"total_bet"`
	RoundBet  float64 `json:"round_bet"`
	SeatIndex int     `json:"seat_index"`
}

// TeenPattiGame represents a game session
type TeenPattiGame struct {
	ID            string             `json:"id"`
	TableID       string             `json:"table_id"`
	State         string             `json:"state"`
	Players       map[string]*Player `json:"players"`
	ActivePlayers []string           `json:"active_players"` // IDs of players in turn order
	Deck          []Card             `json:"-"`
	Pot           float64            `json:"pot"`
	BootAmount    float64            `json:"boot_amount"`
	CurrentTurn   string             `json:"current_turn"` // Player ID
	CurrentStake  float64            `json:"current_stake"` // Current bet amount to match
	DealerIndex   int                `json:"dealer_index"`
	TurnIndex     int                `json:"turn_index"`
	SidePots      []SidePot          `json:"side_pots"`
	Winner        string             `json:"winner,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// NewGame creates a new Teen Patti game
func NewGame(tableID string, bootAmount float64) *TeenPattiGame {
	return &TeenPattiGame{
		ID:         uuid.New().String(),
		TableID:    tableID,
		State:      StateWaiting,
		Players:    make(map[string]*Player),
		BootAmount: bootAmount,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// AddPlayer adds a player to the game
func (g *TeenPattiGame) AddPlayer(playerID, name string, balance float64) error {
	if g.State != StateWaiting {
		return errors.New("game already started")
	}
	if len(g.Players) >= MaxPlayers {
		return errors.New("table full")
	}

	g.Players[playerID] = &Player{
		ID:        playerID,
		Name:      name,
		Balance:   balance,
		Status:    PlayerStatusActive,
		IsBlind:   true,
		SeatIndex: len(g.Players),
	}
	g.ActivePlayers = append(g.ActivePlayers, playerID)
	return nil
}

// StartGame begins the game
func (g *TeenPattiGame) StartGame() error {
	if len(g.Players) < MinPlayers {
		return errors.New("not enough players")
	}

	g.State = StateDealing
	g.Deck = NewDeck()
	Shuffle(g.Deck)

	// Collect Boot Amount
	for _, pID := range g.ActivePlayers {
		player := g.Players[pID]
		if player.Balance < g.BootAmount {
			return errors.New("player " + player.Name + " has insufficient balance")
		}
		player.Balance -= g.BootAmount
		player.TotalBet += g.BootAmount
		g.Pot += g.BootAmount
	}

	// Deal Cards
	for i := 0; i < 3; i++ {
		for _, pID := range g.ActivePlayers {
			card := g.Deck[0]
			g.Deck = g.Deck[1:]
			g.Players[pID].Cards = append(g.Players[pID].Cards, card)
		}
	}

	g.State = StateBetting
	g.CurrentStake = g.BootAmount
	g.TurnIndex = (g.DealerIndex + 1) % len(g.ActivePlayers)
	g.CurrentTurn = g.ActivePlayers[g.TurnIndex]

	return nil
}

// SeeCards marks a player as having seen their cards
func (g *TeenPattiGame) SeeCards(playerID string) error {
	player, ok := g.Players[playerID]
	if !ok {
		return errors.New("player not found")
	}

	player.SeenCards = true
	player.IsBlind = false
	return nil
}

// Pack folds the player's hand
func (g *TeenPattiGame) Pack(playerID string) error {
	if g.CurrentTurn != playerID {
		return errors.New("not your turn")
	}

	player := g.Players[playerID]
	player.Status = PlayerStatusFolded

	return g.nextTurn()
}

// PlaceBet handles betting logic
func (g *TeenPattiGame) PlaceBet(playerID string, amount float64) error {
	if g.CurrentTurn != playerID {
		return errors.New("not your turn")
	}

	player := g.Players[playerID]

	// Validate bet amount based on Blind/Seen status
	// minBet := g.CurrentStake // Unused
	if player.IsBlind {
		// Blind player pays 50% of seen stake (or current stake if it was set by blind player)
		// Simplified: Blind bet = X, Seen bet = 2X
		// If current stake is 100 (Seen), Blind needs to pay 50
		// If current stake is 50 (Blind), Blind needs to pay 50
	}

	// For simplicity in this implementation:
	// CurrentStake is always the "Seen" value.
	// Blind players pay 50% of CurrentStake.
	// Seen players pay 100% of CurrentStake.

	requiredAmount := g.CurrentStake
	if player.IsBlind {
		requiredAmount = g.CurrentStake / 2
	}

	if amount < requiredAmount {
		return errors.New("bet amount too low")
	}

	// If player raises
	// Limit raise to 2x current stake
	if amount > requiredAmount * 2 {
		return errors.New("bet limit exceeded")
	}

	if player.Balance < amount {
		// All-in logic would go here
		return errors.New("insufficient balance")
	}

	player.Balance -= amount
	player.TotalBet += amount
	player.RoundBet += amount
	g.Pot += amount

	// Update CurrentStake if raised
	// If Blind player bets X, Seen value is 2X
	// If Seen player bets Y, Seen value is Y
	newSeenStake := amount
	if player.IsBlind {
		newSeenStake = amount * 2
	}

	if newSeenStake > g.CurrentStake {
		g.CurrentStake = newSeenStake
	}

	return g.nextTurn()
}

// Showdown compares hands of remaining players
func (g *TeenPattiGame) Showdown(initiatorID, targetID string) error {
	// Only allowed when 2 players remain
	activeCount := 0
	for _, p := range g.Players {
		if p.Status == PlayerStatusActive {
			activeCount++
		}
	}

	if activeCount != 2 {
		return errors.New("showdown only allowed with 2 players")
	}

	// Compare hands
	p1 := g.Players[initiatorID]
	p2 := g.Players[targetID]

	// Cost of sideshow
	cost := g.CurrentStake
	if p1.IsBlind {
		cost = cost / 2
	}

	if p1.Balance < cost {
		return errors.New("insufficient balance for showdown")
	}

	p1.Balance -= cost
	g.Pot += cost

	winner := EvaluateWinner(p1, p2)
	g.Winner = winner.ID
	g.State = StateFinished

	return nil
}

func (g *TeenPattiGame) nextTurn() error {
	// Find next active player
	start := g.TurnIndex
	for {
		g.TurnIndex = (g.TurnIndex + 1) % len(g.ActivePlayers)
		nextPlayerID := g.ActivePlayers[g.TurnIndex]
		if g.Players[nextPlayerID].Status == PlayerStatusActive {
			g.CurrentTurn = nextPlayerID
			break
		}
		if g.TurnIndex == start {
			// Should not happen if game logic is correct
			return errors.New("no active players left")
		}
	}

	// Check if only one player remains
	activeCount := 0
	var lastPlayerID string
	for _, p := range g.Players {
		if p.Status == PlayerStatusActive {
			activeCount++
			lastPlayerID = p.ID
		}
	}

	if activeCount == 1 {
		g.Winner = lastPlayerID
		g.State = StateFinished
	}

	return nil
}

// Helper functions for Deck
func NewDeck() []Card {
	suits := []string{"H", "D", "C", "S"}
	deck := []Card{}
	for _, suit := range suits {
		for v := 2; v <= 14; v++ {
			deck = append(deck, Card{Suit: suit, Value: v})
		}
	}
	return deck
}

func Shuffle(deck []Card) {
	// In production use crypto/rand
	// For now simple shuffle logic
}
