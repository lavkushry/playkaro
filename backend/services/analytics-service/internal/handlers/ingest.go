package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/playkaro/analytics-service/internal/models"
)

type IngestHandler struct {
	DB    *sql.DB
	Redis *redis.Client
}

func NewIngestHandler(db *sql.DB, rdb *redis.Client) *IngestHandler {
	return &IngestHandler{DB: db, Redis: rdb}
}

// IngestEvent receives raw events from other services
func (h *IngestHandler) IngestEvent(c *gin.Context) {
	var event models.AnalyticsEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ID = uuid.New().String()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 1. Store raw event (Async in production, sync here for simplicity)
	go h.storeEvent(event)

	// 2. Update Real-time Metrics
	go h.updateRealtimeMetrics(event)

	c.JSON(http.StatusAccepted, gin.H{"id": event.ID})
}

func (h *IngestHandler) storeEvent(event models.AnalyticsEvent) {
	data, _ := json.Marshal(event.EventData)
	_, err := h.DB.Exec(`
		INSERT INTO analytics_events (id, user_id, event_type, event_data, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, event.ID, event.UserID, event.EventType, data, event.Timestamp)

	if err != nil {
		log.Printf("Failed to store event: %v", err)
	}
}

func (h *IngestHandler) updateRealtimeMetrics(event models.AnalyticsEvent) {
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	// Parse event data
	var data map[string]interface{}
	json.Unmarshal(event.EventData, &data)

	pipeline := h.Redis.Pipeline()

	// Track Active Users (HyperLogLog)
	pipeline.PFAdd(ctx, "analytics:active_users:"+today, event.UserID)

	switch event.EventType {
	case models.EventTypeBetPlaced:
		amount, _ := data["amount"].(float64)
		pipeline.IncrByFloat(ctx, "analytics:ggr:"+today, amount)
		pipeline.IncrByFloat(ctx, "analytics:wagers:"+today, amount)

		// Game specific
		if gameType, ok := data["game_type"].(string); ok {
			pipeline.IncrByFloat(ctx, "analytics:game:"+gameType+":wagers:"+today, amount)
			pipeline.Incr(ctx, "analytics:game:"+gameType+":rounds:"+today)
		}

	case models.EventTypePayout:
		amount, _ := data["amount"].(float64)
		pipeline.IncrByFloat(ctx, "analytics:payouts:"+today, amount)

		// Game specific
		if gameType, ok := data["game_type"].(string); ok {
			pipeline.IncrByFloat(ctx, "analytics:game:"+gameType+":payouts:"+today, amount)
		}

	case models.EventTypeDeposit:
		amount, _ := data["amount"].(float64)
		pipeline.IncrByFloat(ctx, "analytics:deposits:"+today, amount)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		log.Printf("Failed to update metrics: %v", err)
	}
}
