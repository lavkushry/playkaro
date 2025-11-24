package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/playkaro/analytics-service/internal/handlers"
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
	ingestHandler := handlers.NewIngestHandler(db, rdb)
	dashboardHandler := handlers.NewDashboardHandler(rdb)

	// Router
	r := gin.Default()
	r.Use(otelgin.Middleware("analytics-service"))

	// Ingestion Routes (Internal)
	v1 := r.Group("/v1/analytics")
	{
		v1.POST("/events", ingestHandler.IngestEvent)

		// Dashboard Routes
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/revenue", dashboardHandler.GetRevenueStats)
			dashboard.GET("/games", dashboardHandler.GetGameMetrics)
		}

		// User Analytics
		users := v1.Group("/users")
		{
			users.GET("/churn-risk", dashboardHandler.GetChurnRiskUsers)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	log.Printf("Analytics Service starting on port %s", port)
	r.Run(":" + port)
}
