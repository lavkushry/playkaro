package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/playkaro/streaming-service/internal/models"
)

type StreamHandler struct {
	DB    *sql.DB
	Redis *redis.Client
}

// CreateStream generates a new stream key for a match
func (h *StreamHandler) CreateStream(c *gin.Context) {
	var req models.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate secure stream key
	streamKey := uuid.New().String()
	streamID := uuid.New().String()
	playbackURL := os.Getenv("CDN_URL") + "/hls/" + streamKey + ".m3u8"

	stream := models.Stream{
		ID:          streamID,
		MatchID:     req.MatchID,
		StreamKey:   streamKey,
		Status:      models.StreamStatusOffline,
		PlaybackURL: playbackURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store in DB
	query := `
		INSERT INTO streams (id, match_id, stream_key, status, playback_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := h.DB.Exec(query, stream.ID, stream.MatchID, stream.StreamKey, stream.Status, stream.PlaybackURL, stream.CreatedAt, stream.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
		return
	}

	// Cache key mapping for fast auth
	h.Redis.Set(context.Background(), "stream_key:"+streamKey, streamID, 24*time.Hour)

	resp := models.StreamResponse{
		ID:          stream.ID,
		MatchID:     stream.MatchID,
		StreamKey:   stream.StreamKey,
		Status:      stream.Status,
		PlaybackURL: stream.PlaybackURL,
	}

	c.JSON(http.StatusCreated, resp)
}

// GetStream returns public stream info
func (h *StreamHandler) GetStream(c *gin.Context) {
	matchID := c.Param("match_id")

	var stream models.Stream
	query := `SELECT id, match_id, status, playback_url, viewer_count FROM streams WHERE match_id = $1`
	err := h.DB.QueryRow(query, matchID).Scan(&stream.ID, &stream.MatchID, &stream.Status, &stream.PlaybackURL, &stream.ViewerCount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, stream)
}
