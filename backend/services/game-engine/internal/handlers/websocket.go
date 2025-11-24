package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/playkaro/game-engine/internal/session"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

type WebSocketHandler struct {
	SessionManager *session.SessionManager
}

func NewWebSocketHandler(sm *session.SessionManager) *WebSocketHandler {
	return &WebSocketHandler{SessionManager: sm}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	sessionID := c.Param("session_id")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer ws.Close()

	// Simple polling loop for demo purposes
	// In production, use Redis Pub/Sub or internal event bus
	for {
		session, err := h.SessionManager.GetSession(sessionID)
		if err != nil {
			return
		}

		if err := ws.WriteJSON(session); err != nil {
			return
		}

		time.Sleep(1 * time.Second) // Poll every second
	}
}
