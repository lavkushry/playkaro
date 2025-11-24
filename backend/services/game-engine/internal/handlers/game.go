package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/game-engine/internal/engine"
	"github.com/playkaro/game-engine/internal/registry"
	"github.com/playkaro/game-engine/internal/session"
)

type GameHandler struct {
	SessionManager *session.SessionManager
}

func NewGameHandler(sm *session.SessionManager) *GameHandler {
	return &GameHandler{SessionManager: sm}
}

// ListGames returns all available games
func (h *GameHandler) ListGames(c *gin.Context) {
	reg := registry.GetRegistry()
	games := reg.ListGames()
	c.JSON(http.StatusOK, gin.H{"games": games})
}

// CreateSession starts a new game session
func (h *GameHandler) CreateSession(c *gin.Context) {
	var req struct {
		GameID string `json:"game_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	session, err := h.SessionManager.CreateSession(req.GameID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// JoinSession adds a player to a session
func (h *GameHandler) JoinSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.GetString("userID")

	session, err := h.SessionManager.JoinSession(sessionID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// MakeMove handles a player move
func (h *GameHandler) MakeMove(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.GetString("userID")

	var req struct {
		Type string                 `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	move := engine.Move{
		PlayerID: userID,
		Type:     req.Type,
		Data:     req.Data,
	}

	result, err := h.SessionManager.ProcessMove(sessionID, move)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSessionState returns current game state
func (h *GameHandler) GetSessionState(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.SessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}
