package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/playkaro/social-service/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type ChatHandler struct {
	Redis   *redis.Client
	Clients map[string]*websocket.Conn // userID -> conn
	Mutex   sync.RWMutex
}

func NewChatHandler(rdb *redis.Client) *ChatHandler {
	return &ChatHandler{
		Redis:   rdb,
		Clients: make(map[string]*websocket.Conn),
	}
}

// HandleWebSocket upgrades HTTP to WebSocket
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	h.Mutex.Lock()
	h.Clients[userID] = conn
	h.Mutex.Unlock()

	defer func() {
		h.Mutex.Lock()
		delete(h.Clients, userID)
		h.Mutex.Unlock()
		conn.Close()
	}()

	// Subscribe to Redis channels
	go h.subscribeToChannels(userID, conn)

	// Read loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Handle incoming message (e.g., send to global chat)
		var chatMsg models.ChatMessage
		if err := json.Unmarshal(msg, &chatMsg); err != nil {
			continue
		}

		chatMsg.SenderID = userID
		chatMsg.Timestamp = time.Now()
		chatMsg.ID = uuid.New().String()

		h.publishMessage(chatMsg)
	}
}

// subscribeToChannels listens for Redis messages
func (h *ChatHandler) subscribeToChannels(userID string, conn *websocket.Conn) {
	ctx := context.Background()
	pubsub := h.Redis.Subscribe(ctx, "chat:global", fmt.Sprintf("chat:user:%s", userID))
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			break
		}
	}
}

// publishMessage publishes a message to Redis
func (h *ChatHandler) publishMessage(msg models.ChatMessage) {
	ctx := context.Background()
	payload, _ := json.Marshal(msg)

	var channel string
	if msg.Type == models.ChatTypeGlobal {
		channel = "chat:global"
	} else if msg.Type == models.ChatTypePrivate && msg.RecipientID != nil {
		channel = fmt.Sprintf("chat:user:%s", *msg.RecipientID)
		// Also send to sender so they see their own message
		// Or client handles optimistic UI
	} else {
		return
	}

	h.Redis.Publish(ctx, channel, payload)

	// Persist to history (Redis List for now)
	if msg.Type == models.ChatTypeGlobal {
		h.Redis.LPush(ctx, "chat:history:global", payload)
		h.Redis.LTrim(ctx, "chat:history:global", 0, 49) // Keep last 50
	}
}

// GetChatHistory returns recent global messages
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	ctx := context.Background()
	messages, err := h.Redis.LRange(ctx, "chat:history:global", 0, 49).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
		return
	}

	var history []models.ChatMessage
	for _, msg := range messages {
		var m models.ChatMessage
		json.Unmarshal([]byte(msg), &m)
		history = append(history, m)
	}

	c.JSON(http.StatusOK, gin.H{"messages": history})
}
