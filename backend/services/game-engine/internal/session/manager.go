package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/registry"
)

type SessionManager struct {
	sessions map[string]*engine.GameSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*engine.GameSession),
	}
}

// CreateSession initializes a new game session
func (sm *SessionManager) CreateSession(gameID string, userID string) (*engine.GameSession, error) {
	reg := registry.GetRegistry()
	game, err := reg.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	session := &engine.GameSession{
		SessionID: sessionID,
		GameID:    gameID,
		Players:   []*engine.Player{{UserID: userID, IsTurn: true}}, // Creator starts first (simplified)
		Status:    "WAITING",
		EntryFee:  game.GetEntryFee(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initialize game state
	if err := game.Start(session); err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return session, nil
}

// JoinSession adds a player to an existing session
func (sm *SessionManager) JoinSession(sessionID, userID string) (*engine.GameSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	if session.Status != "WAITING" {
		return nil, errors.New("session not open for joining")
	}

	reg := registry.GetRegistry()
	game, _ := reg.GetGame(session.GameID)

	if len(session.Players) >= game.GetMaxPlayers() {
		return nil, errors.New("session full")
	}

	// Check if user already joined
	for _, p := range session.Players {
		if p.UserID == userID {
			return nil, errors.New("user already in session")
		}
	}

	session.Players = append(session.Players, &engine.Player{UserID: userID})
	session.UpdatedAt = time.Now()

	// Start game if min players reached
	if len(session.Players) >= game.GetMinPlayers() {
		session.Status = "IN_PROGRESS"
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*engine.GameSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

// ProcessMove handles a player's move
func (sm *SessionManager) ProcessMove(sessionID string, move engine.Move) (*engine.MoveResult, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	if session.Status != "IN_PROGRESS" {
		return nil, errors.New("game not in progress")
	}

	reg := registry.GetRegistry()
	game, _ := reg.GetGame(session.GameID)

	result, err := game.HandleMove(session, move)
	if err != nil {
		return nil, err
	}

	session.UpdatedAt = time.Now()

	if result.GameEnded {
		session.Status = "COMPLETED"
		// TODO: Handle game end (prizes, etc.)
	}

	return result, nil
}
