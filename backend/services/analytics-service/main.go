package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"net"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	grpc_impl "github.com/playkaro/analytics-service/internal/grpc"
	"github.com/playkaro/analytics-service/internal/handlers"
	pb "github.com/playkaro/backend/proto/analytics"
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

	// Start HTTP server
	go func() {
		log.Printf("Analytics Service HTTP starting on port %s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	// Register gRPC server
	pb.RegisterAnalyticsServiceServer(grpcServer, grpc_impl.NewAnalyticsServer(ingestHandler))

	log.Printf("Analytics Service gRPC starting on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
