package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/playkaro/social-service/internal/handlers"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Database Connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Redis Connection
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	// Handlers
	friendHandler := handlers.NewFriendHandler(db)
	chatHandler := handlers.NewChatHandler(rdb)

	// Router
	r := gin.Default()
	r.Use(otelgin.Middleware("social-service"))

	// Friend Routes
	v1 := r.Group("/v1/social")
	{
		friends := v1.Group("/friends")
		{
			friends.POST("/request", friendHandler.SendFriendRequest)
			friends.POST("/accept", friendHandler.AcceptFriendRequest)
			friends.GET("", friendHandler.GetFriends)
			friends.GET("/requests", friendHandler.GetPendingRequests)
		}

		// Chat Routes
		chat := v1.Group("/chat")
		{
			chat.GET("/history", chatHandler.GetChatHistory)
		}

		// WebSocket
		v1.GET("/ws", chatHandler.HandleWebSocket)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	log.Printf("Social Service starting on port %s", port)
	r.Run(":" + port)
}
