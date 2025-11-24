package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/playkaro/streaming-service/internal/models"
)

type WebhookHandler struct {
	DB    *sql.DB
	Redis *redis.Client
}

// OnPublish is called by NGINX when a stream starts
// NGINX sends: POST /hooks/on_publish?name=<stream_key>
func (h *WebhookHandler) OnPublish(c *gin.Context) {
	streamKey := c.Query("name")
	if streamKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing stream key"})
		return
	}

	// Verify stream key exists
	ctx := context.Background()
	streamID, err := h.Redis.Get(ctx, "stream_key:"+streamKey).Result()
	if err != nil {
		// Key not found or expired - reject
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid stream key"})
		return
	}

	// Update stream status to LIVE
	query := `UPDATE streams SET status = $1, updated_at = $2 WHERE id = $3`
	_, err = h.DB.Exec(query, models.StreamStatusLive, time.Now(), streamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stream"})
		return
	}

	// Track live stream in Redis for quick lookups
	h.Redis.Set(ctx, "live_stream:"+streamID, streamKey, 0) // No expiration (manual cleanup)

	c.JSON(http.StatusOK, gin.H{"status": "authorized"})
}

// OnPublishDone is called by NGINX when a stream ends
func (h *WebhookHandler) OnPublishDone(c *gin.Context) {
	streamKey := c.Query("name")
	if streamKey == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ok"}) // Don't fail on cleanup
		return
	}

	ctx := context.Background()
	streamID, err := h.Redis.Get(ctx, "stream_key:"+streamKey).Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// Update stream status to OFFLINE
	query := `UPDATE streams SET status = $1, updated_at = $2 WHERE id = $3`
	h.DB.Exec(query, models.StreamStatusOffline, time.Now(), streamID)

	// Remove from live tracking
	h.Redis.Del(ctx, "live_stream:"+streamID)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// OnPlay is called when a viewer starts watching (optional, for analytics)
func (h *WebhookHandler) OnPlay(c *gin.Context) {
	streamKey := c.Query("name")
	ctx := context.Background()

	streamID, err := h.Redis.Get(ctx, "stream_key:"+streamKey).Result()
	if err == nil {
		// Increment viewer count
		h.Redis.Incr(ctx, "viewers:"+streamID)

		// Update DB viewer count (async in real app)
		var count int64
		count, _ = h.Redis.Get(ctx, "viewers:"+streamID).Int64()
		h.DB.Exec(`UPDATE streams SET viewer_count = $1 WHERE id = $2`, count, streamID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// OnPlayDone is called when a viewer stops watching
func (h *WebhookHandler) OnPlayDone(c *gin.Context) {
	streamKey := c.Query("name")
	ctx := context.Background()

	streamID, err := h.Redis.Get(ctx, "stream_key:"+streamKey).Result()
	if err == nil {
		// Decrement viewer count
		h.Redis.Decr(ctx, "viewers:"+streamID)

		var count int64
		count, _ = h.Redis.Get(ctx, "viewers:"+streamID).Int64()
		if count < 0 {
			count = 0
		}
		h.DB.Exec(`UPDATE streams SET viewer_count = $1 WHERE id = $2`, count, streamID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
