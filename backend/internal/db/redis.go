package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func ConnectRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // DragonflyDB default port
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis/DragonflyDB:", err)
	}
	log.Println("Successfully connected to Redis/DragonflyDB!")
}
