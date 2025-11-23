package realtime

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Message types
const (
	TypeOddsUpdate  = "odds_update"
	TypeChatMessage = "chat_message"
)

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ChatMessage struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a single websocket connection
type Client struct {
	hub *Hub
	conn *websocket.Conn
	send chan WSMessage
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		var msg WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Handle incoming messages (e.g., chat messages)
		if msg.Type == TypeChatMessage {
			// Broadcast chat message to all clients
			// In a real app, you'd validate the user here
			payloadMap, ok := msg.Payload.(map[string]interface{})
			if ok {
				username, userOk := payloadMap["username"].(string)
				messageContent, msgOk := payloadMap["message"].(string)
				if userOk && msgOk {
					chatMsg := ChatMessage{
						Username:  username,
						Message:   messageContent,
						Timestamp: time.Now(),
					}
					c.hub.broadcast <- WSMessage{
						Type:    TypeChatMessage,
						Payload: chatMsg,
					}
				} else {
					log.Printf("Invalid chat message payload format: missing username or message in %v", msg.Payload)
				}
			} else {
				log.Printf("Invalid chat message payload type: expected map[string]interface{}, got %T", msg.Payload)
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		err := c.conn.WriteJSON(message)
		if err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan WSMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

var MainHub = newHub()

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func ServeWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: MainHub, conn: ws, send: make(chan WSMessage, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message WSMessage) {
	h.broadcast <- message
}

// Simulate Odds Updates
func StartOddsSimulation() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			// In a real app, this would come from a data provider
			msg := WSMessage{
				Type: TypeOddsUpdate,
				Payload: map[string]interface{}{
					"match_id": "1",
					"odds_a":   1.95,
					"odds_b":   1.95,
				},
			}
			MainHub.broadcast <- msg
		}
	}()
}
