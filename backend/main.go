package main

import (
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/playkaro/backend/graph"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/grpc_client"
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
	db.ConnectRedis()
	grpc_client.InitWalletClient()

	// Initialize Router
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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

	// Admin Routes (Protected)
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.POST("/matches", handlers.CreateMatch)
		adminGroup.PUT("/matches/:id/odds", handlers.UpdateMatchOdds)
		adminGroup.POST("/matches/:id/settle", handlers.SettleMatch)
	}

	// History Routes (Protected)
	historyGroup := r.Group("/api/v1")
	historyGroup.Use(middleware.AuthMiddleware())
	{
		historyGroup.GET("/transactions", handlers.GetTransactions)
		historyGroup.GET("/bets", handlers.GetBets)
	}

	// Payment Routes
	paymentGroup := r.Group("/api/v1/payment")
	{
		paymentGroup.POST("/deposit", middleware.AuthMiddleware(), handlers.InitiateDeposit)
		paymentGroup.POST("/withdraw", middleware.AuthMiddleware(), handlers.InitiateWithdrawal)
		paymentGroup.POST("/webhook/razorpay", handlers.RazorpayWebhook) // Public webhook
	}

	// KYC Routes
	kycGroup := r.Group("/api/v1/kyc")
	kycGroup.Use(middleware.AuthMiddleware())
	{
		kycGroup.POST("/upload", handlers.UploadKYCDocument)
		kycGroup.GET("/status", handlers.GetKYCStatus)
	}

	// Admin KYC Routes
	adminKYCGroup := r.Group("/api/v1/admin/kyc")
	adminKYCGroup.Use(middleware.AuthMiddleware())
	adminKYCGroup.Use(middleware.AdminMiddleware())
	{
		adminKYCGroup.POST("/approve", handlers.ApproveKYC)
	}

	// Casino/Game Routes
	casinoGroup := r.Group("/api/v1/casino")
	{
		casinoGroup.GET("/games", handlers.GetGames) // Public
		casinoGroup.GET("/launch", middleware.AuthMiddleware(), handlers.LaunchGame)
	}

	// Seamless Wallet API (for game providers)
	walletAPIGroup := r.Group("/api/v1/game-wallet")
	{
		walletAPIGroup.POST("/balance", handlers.GetBalanceForProvider)
		walletAPIGroup.POST("/debit", handlers.DebitWallet)
		walletAPIGroup.POST("/credit", handlers.CreditWallet)
		walletAPIGroup.POST("/rollback", handlers.RollbackWallet)
	}

	// Promotions Routes
	promoGroup := r.Group("/api/v1/promotions")
	promoGroup.Use(middleware.AuthMiddleware())
	{
		promoGroup.GET("/bonuses", handlers.GetBonuses)
		promoGroup.POST("/claim", handlers.ClaimBonus)
		promoGroup.POST("/referral/generate", handlers.GenerateReferralCode)
		promoGroup.POST("/referral/apply", handlers.ApplyReferralCode)
		promoGroup.GET("/leaderboard", handlers.GetLeaderboard)
	}

	// Public Routes
	r.GET("/api/v1/matches", handlers.GetMatches)
	r.GET("/ws", realtime.ServeWS)

	// GraphQL Routes (protected by auth for mutations/queries needing user context)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
	r.POST("/query", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()
		if userID != "" {
			ctx = graph.WithUserID(ctx, userID)
		}
		c.Request = c.Request.WithContext(ctx)
		srv.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})

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
