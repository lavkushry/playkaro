package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/playkaro/match-service/internal/cache"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type OddsStreamHandler struct {
	Cache *cache.MatchCache
}

func NewOddsStreamHandler(cache *cache.MatchCache) *OddsStreamHandler {
	return &OddsStreamHandler{Cache: cache}
}

// StreamOdds handles WebSocket connections for real-time odds updates
func (h *OddsStreamHandler) StreamOdds(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer ws.Close()

	// Subscribe to Redis Pub/Sub for odds updates
	ctx := context.Background()
	msgChan := h.Cache.SubscribeOddsUpdates(ctx)

	// Send odds updates to connected client
	for  {
		select {
		case msg := <-msgChan:
			if msg == nil {
				return
			}

			// Send message to WebSocket client
			err := ws.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Println("WebSocket write error:", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
