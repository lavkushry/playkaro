package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/handlers"
	"github.com/playkaro/backend/internal/middleware"
	"github.com/playkaro/backend/internal/realtime"
)




func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Database
	db.Connect()

	// Initialize Router
	r := gin.Default()

	// Auth Routes
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
	}

	// Wallet Routes (Protected)
	walletGroup := r.Group("/api/v1/wallet")
	walletGroup.Use(middleware.AuthMiddleware())
	{
		walletGroup.GET("/", handlers.GetBalance)
		walletGroup.POST("/deposit", handlers.Deposit)
		walletGroup.POST("/withdraw", handlers.Withdraw)
	}

	// Betting Routes (Protected)
	betGroup := r.Group("/api/v1/bet")
	betGroup.Use(middleware.AuthMiddleware())
	{
		betGroup.POST("/", handlers.PlaceBet)
	}

	// Public Routes
	r.GET("/api/v1/matches", handlers.GetMatches)
	r.GET("/ws", realtime.ServeWS)

	// Start Real-time Simulation
	realtime.StartOddsSimulation()
	go realtime.MainHub.Run()

	// Health Check


	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "playkaro-backend",
		})
	})

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
