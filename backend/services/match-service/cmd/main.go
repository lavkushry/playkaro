package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/playkaro/match-service/internal/cache"
	"github.com/playkaro/match-service/internal/db"
	"github.com/playkaro/match-service/internal/handlers"
	"github.com/playkaro/match-service/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	if err := db.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	// Initialize Redis cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	matchCache, err := cache.NewMatchCache(redisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize handlers
	matchHandler := handlers.NewMatchHandler(db.DB, matchCache)
	oddsStreamHandler := websocket.NewOddsStreamHandler(matchCache)

	// Setup router
	r := gin.Default()

	// CORS middleware (for development)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "match-service"})
	})

	// WebSocket route for real-time odds
	r.GET("/ws/odds", oddsStreamHandler.StreamOdds)

	// Match routes
	v1 := r.Group("/v1/matches")
	{
		// Public routes
		v1.GET("", matchHandler.GetMatches)
		v1.GET("/:match_id", matchHandler.GetMatch)

		// Admin routes (require admin JWT)
		admin := v1.Group("")
		admin.Use(AdminMiddleware())
		{
			admin.POST("", matchHandler.CreateMatch)
			admin.PUT("/:match_id/odds", matchHandler.UpdateOdds)
			admin.POST("/:match_id/settle", matchHandler.SettleMatch)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Match Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// AdminMiddleware validates admin JWT token (simplified version)
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper admin JWT validation
		// For now, just check for presence of admin header
		isAdmin := c.GetHeader("X-Admin-Key")
		if isAdmin == "" {
			c.JSON(401, gin.H{"error": "Unauthorized - Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
