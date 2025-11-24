package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/playkaro/streaming-service/internal/handlers"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis connection
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Initialize Gin
	r := gin.Default()

	// Handlers
	streamHandler := &handlers.StreamHandler{DB: db, Redis: rdb}
	webhookHandler := &handlers.WebhookHandler{DB: db, Redis: rdb}

	// API Routes
	v1 := r.Group("/v1/streaming")
	{
		v1.POST("/keys", streamHandler.CreateStream)
		v1.GET("/live/:match_id", streamHandler.GetStream)
	}

	// Webhook Routes (Internal - NGINX only)
	hooks := r.Group("/v1/streaming/hooks")
	{
		hooks.POST("/on_publish", webhookHandler.OnPublish)
		hooks.POST("/on_publish_done", webhookHandler.OnPublishDone)
		hooks.POST("/on_play", webhookHandler.OnPlay)
		hooks.POST("/on_play_done", webhookHandler.OnPlayDone)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("Streaming Service listening on port %s", port)
	r.Run(":" + port)
}
