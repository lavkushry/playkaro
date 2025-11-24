package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/playkaro/game-engine/games/crash"
	"github.com/playkaro/game-engine/games/dice"
	"github.com/playkaro/game-engine/games/ludo"
	"github.com/playkaro/game-engine/internal/handlers"
	"github.com/playkaro/game-engine/internal/registry"
	"github.com/playkaro/game-engine/internal/session"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize Registry
	reg := registry.GetRegistry()

	// Register built-in games
	reg.RegisterGame(ludo.NewLudoGame())
	reg.RegisterGame(crash.NewCrashGame())
	reg.RegisterGame(dice.NewDiceGame())
	log.Println("Registered games:", reg.ListGames())

	// Initialize Session Manager
	sessionManager := session.NewSessionManager()

	// Initialize Handlers
	gameHandler := handlers.NewGameHandler(sessionManager)
	wsHandler := handlers.NewWebSocketHandler(sessionManager)

	// Setup Router
	r := gin.Default()

	// CORS
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
		c.JSON(200, gin.H{"status": "ok", "service": "game-engine"})
	})

	// API Routes
	v1 := r.Group("/v1")
	{
		// Game Catalog
		v1.GET("/games", gameHandler.ListGames)

		// Session Management
		// In production, use middleware to extract userID from JWT
		// For demo, we'll simulate auth via header
		authorized := v1.Group("")
		authorized.Use(MockAuthMiddleware())
		{
			authorized.POST("/games/sessions", gameHandler.CreateSession)
			authorized.POST("/sessions/:session_id/join", gameHandler.JoinSession)
			authorized.POST("/sessions/:session_id/move", gameHandler.MakeMove)
			authorized.GET("/sessions/:session_id", gameHandler.GetSessionState)
		}
	}

	// WebSocket
	r.GET("/ws/sessions/:session_id", wsHandler.HandleWebSocket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Game Engine Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "demo_user" // Fallback for easy testing
		}
		c.Set("userID", userID)
		c.Next()
	}
}
