package registry

import (
	"errors"
	"sync"

	"github.com/playkaro/game-engine/internal/engine"
)

type GameRegistry struct {
	games map[string]engine.IGame
	mu    sync.RWMutex
}

var instance *GameRegistry
var once sync.Once

// GetRegistry returns the singleton instance of GameRegistry
func GetRegistry() *GameRegistry {
	once.Do(func() {
		instance = &GameRegistry{
			games: make(map[string]engine.IGame),
		}
	})
	return instance
}

// RegisterGame adds a new game to the registry
func (r *GameRegistry) RegisterGame(game engine.IGame) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[game.GetGameID()] = game
}

// GetGame retrieves a game by ID
func (r *GameRegistry) GetGame(gameID string) (engine.IGame, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return nil, errors.New("game not found")
	}
	return game, nil
}

// ListGames returns metadata for all registered games
func (r *GameRegistry) ListGames() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	games := []map[string]interface{}{}
	for _, game := range r.games {
		games = append(games, map[string]interface{}{
			"game_id":     game.GetGameID(),
			"name":        game.GetGameName(),
			"type":        game.GetGameType(),
			"min_players": game.GetMinPlayers(),
			"max_players": game.GetMaxPlayers(),
			"entry_fee":   game.GetEntryFee(),
		})
	}
	return games
}
