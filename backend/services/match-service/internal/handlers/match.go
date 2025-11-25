package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/match-service/internal/cache"
	"github.com/playkaro/match-service/internal/grpc"
	"github.com/playkaro/match-service/internal/models"
)

type MatchHandler struct {
	DB      *sql.DB
	Cache   *cache.MatchCache
	Clients *grpc.Clients
}

type CreateMatchRequest struct {
	Sport      string    `json:"sport" binding:"required"`
	TeamA      string    `json:"team_a" binding:"required"`
	TeamB      string    `json:"team_b" binding:"required"`
	OddsA      float64   `json:"odds_a" binding:"required"`
	OddsB      float64   `json:"odds_b" binding:"required"`
	OddsDraw   float64   `json:"odds_draw"`
	StartTime  time.Time `json:"start_time" binding:"required"`
	League     string    `json:"league"`
	Venue      string    `json:"venue"`
}

type UpdateOddsRequest struct {
	OddsA    float64 `json:"odds_a" binding:"required"`
	OddsB    float64 `json:"odds_b" binding:"required"`
	OddsDraw float64 `json:"odds_draw"`
}

type SettleMatchRequest struct {
	Result string `json:"result" binding:"required"`
}

func NewMatchHandler(db *sql.DB, cache *cache.MatchCache, clients *grpc.Clients) *MatchHandler {
	return &MatchHandler{
		DB:      db,
		Cache:   cache,
		Clients: clients,
	}
}

// CreateMatch creates a new match (Admin only)
func (h *MatchHandler) CreateMatch(c *gin.Context) {
	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matchID := fmt.Sprintf("match_%d", time.Now().UnixNano())

	var id string
	err := h.DB.QueryRow(`
		INSERT INTO matches
		(match_id, sport, team_a, team_b, odds_a, odds_b, odds_draw, start_time, league, venue, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, matchID, req.Sport, req.TeamA, req.TeamB, req.OddsA, req.OddsB, req.OddsDraw,
		req.StartTime, req.League, req.Venue, models.StatusUpcoming).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
		return
	}

	// Invalidate cache
	h.Cache.InvalidateAll(context.Background())

	// TODO: Publish Kafka event `match.created`

	c.JSON(http.StatusCreated, gin.H{
		"match_id": matchID,
		"status":   models.StatusUpcoming,
	})
}

// GetMatches retrieves all matches with optional filtering
func (h *MatchHandler) GetMatches(c *gin.Context) {
	status := c.Query("status")
	sport := c.Query("sport")

	cacheKey := fmt.Sprintf("matches:%s:%s", status, sport)

	// Try cache first
	cached, err := h.Cache.GetMatches(context.Background(), cacheKey)
	if err == nil && len(cached) > 0 {
		c.JSON(http.StatusOK, gin.H{"matches": cached, "cached": true})
		return
	}

	// Build query
	query := "SELECT id, match_id, sport, team_a, team_b, odds_a, odds_b, odds_draw, status, start_time, league, venue FROM matches WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if sport != "" {
		query += fmt.Sprintf(" AND sport = $%d", argCount)
		args = append(args, sport)
		argCount++
	}

	query += " ORDER BY start_time ASC LIMIT 50"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	matches := []*models.Match{}
	for rows.Next() {
		var match models.Match
		err := rows.Scan(
			&match.ID, &match.MatchID, &match.Sport, &match.TeamA, &match.TeamB,
			&match.OddsA, &match.OddsB, &match.OddsDraw, &match.Status, &match.StartTime,
			&match.League, &match.Venue,
		)
		if err != nil {
			continue
		}
		matches = append(matches, &match)
	}

	// Cache the results
	h.Cache.CacheMatches(context.Background(), cacheKey, matches)

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

// GetMatch retrieves a single match by ID
func (h *MatchHandler) GetMatch(c *gin.Context) {
	matchID := c.Param("match_id")

	// Try cache first
	cached, err := h.Cache.GetMatch(context.Background(), matchID)
	if err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Fallback to database
	var match models.Match
	err = h.DB.QueryRow(`
		SELECT id, match_id, sport, team_a, team_b, odds_a, odds_b, odds_draw,
		       status, start_time, league, venue, result, created_at, updated_at
		FROM matches
		WHERE match_id = $1
	`, matchID).Scan(
		&match.ID, &match.MatchID, &match.Sport, &match.TeamA, &match.TeamB,
		&match.OddsA, &match.OddsB, &match.OddsDraw, &match.Status, &match.StartTime,
		&match.League, &match.Venue, &match.Result, &match.CreatedAt, &match.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Cache the match
	h.Cache.CacheMatch(context.Background(), &match)

	c.JSON(http.StatusOK, match)
}

// UpdateOdds updates match odds (Admin only)
func (h *MatchHandler) UpdateOdds(c *gin.Context) {
	matchID := c.Param("match_id")
	var req UpdateOddsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update database
	_, err := h.DB.Exec(`
		UPDATE matches
		SET odds_a = $1, odds_b = $2, odds_draw = $3, updated_at = $4
		WHERE match_id = $5
	`, req.OddsA, req.OddsB, req.OddsDraw, time.Now(), matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update odds"})
		return
	}

	// Save to odds history
	h.DB.Exec(`
		INSERT INTO odds_history (match_id, odds_a, odds_b, odds_draw)
		SELECT id, $1, $2, $3 FROM matches WHERE match_id = $4
	`, req.OddsA, req.OddsB, req.OddsDraw, matchID)

	// Invalidate cache
	h.Cache.InvalidateMatch(context.Background(), matchID)
	h.Cache.InvalidateAll(context.Background())

	// Publish to Redis Pub/Sub for real-time updates
	h.Cache.PublishOddsUpdate(context.Background(), matchID, req.OddsA, req.OddsB, req.OddsDraw)

	// TODO: Publish Kafka event `match.odds_updated`

	c.JSON(http.StatusOK, gin.H{
		"match_id":   matchID,
		"odds_a":     req.OddsA,
		"odds_b":     req.OddsB,
		"odds_draw":  req.OddsDraw,
		"updated_at": time.Now(),
	})
}

// SettleMatch settles a match with final result (Admin only)
func (h *MatchHandler) SettleMatch(c *gin.Context) {
	matchID := c.Param("match_id")
	var req SettleMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate result
	if req.Result != models.ResultTeamA && req.Result != models.ResultTeamB && req.Result != models.ResultDraw {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result"})
		return
	}

	// Update match
	_, err := h.DB.Exec(`
		UPDATE matches
		SET status = $1, result = $2, settled_at = $3, updated_at = $3
		WHERE match_id = $4
	`, models.StatusCompleted, req.Result, time.Now(), matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to settle match"})
		return
	}

	// Invalidate cache
	h.Cache.InvalidateMatch(context.Background(), matchID)
	h.Cache.InvalidateAll(context.Background())

	// TODO: Publish Kafka event `match.ended`

	c.JSON(http.StatusOK, gin.H{
		"match_id": matchID,
		"status":   models.StatusCompleted,
		"result":   req.Result,
	})
}
