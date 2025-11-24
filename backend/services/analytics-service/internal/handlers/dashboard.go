package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/playkaro/analytics-service/internal/models"
)

type DashboardHandler struct {
	Redis *redis.Client
}

func NewDashboardHandler(rdb *redis.Client) *DashboardHandler {
	return &DashboardHandler{Redis: rdb}
}

// GetRevenueStats returns GGR, NGR, and Active Users
func (h *DashboardHandler) GetRevenueStats(c *gin.Context) {
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	// Fetch metrics from Redis
	pipeline := h.Redis.Pipeline()
	ggrCmd := pipeline.Get(ctx, "analytics:ggr:"+today)
	payoutsCmd := pipeline.Get(ctx, "analytics:payouts:"+today)
	depositsCmd := pipeline.Get(ctx, "analytics:deposits:"+today)
	usersCmd := pipeline.PFCount(ctx, "analytics:active_users:"+today)

	_, _ = pipeline.Exec(ctx)

	ggr, _ := ggrCmd.Float64()
	payouts, _ := payoutsCmd.Float64()
	deposits, _ := depositsCmd.Float64()
	activeUsers := usersCmd.Val()

	stats := models.RevenueStats{
		GGR:           ggr,
		NGR:           ggr - payouts,
		ActiveUsers:   activeUsers,
		DepositVolume: deposits,
	}

	c.JSON(http.StatusOK, stats)
}

// GetGameMetrics returns stats per game
func (h *DashboardHandler) GetGameMetrics(c *gin.Context) {
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	// For now, hardcoded list of games. In production, fetch from DB or discovery.
	games := []string{"LUDO", "TEEN_PATTI", "FANTASY", "CRASH"}

	metrics := []models.GameMetrics{}

	for _, game := range games {
		pipeline := h.Redis.Pipeline()
		wagersCmd := pipeline.Get(ctx, "analytics:game:"+game+":wagers:"+today)
		payoutsCmd := pipeline.Get(ctx, "analytics:game:"+game+":payouts:"+today)
		roundsCmd := pipeline.Get(ctx, "analytics:game:"+game+":rounds:"+today)

		_, _ = pipeline.Exec(ctx)

		wagers, _ := wagersCmd.Float64()
		payouts, _ := payoutsCmd.Float64()
		rounds, _ := roundsCmd.Int()

		metrics = append(metrics, models.GameMetrics{
			GameType:     game,
			Date:         time.Now(),
			TotalWagers:  wagers,
			TotalPayouts: payouts,
			TotalRounds:  rounds,
		})
	}

	c.JSON(http.StatusOK, gin.H{"games": metrics})
}

// GetChurnRiskUsers returns users at risk of churning
// Simplified implementation: Returns mock data or queries DB (omitted here)
func (h *DashboardHandler) GetChurnRiskUsers(c *gin.Context) {
	// In real implementation:
	// Query DB for users with LastActive > 7 days ago AND Balance > 0

	c.JSON(http.StatusOK, gin.H{
		"risk_users": []string{}, // Placeholder
		"count":      0,
	})
}
