package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	pb "github.com/playkaro/backend/proto/wallet"
	"github.com/playkaro/payment-service/internal/db"
	"github.com/playkaro/payment-service/internal/gateways/razorpay"
	grpc_impl "github.com/playkaro/payment-service/internal/grpc"
	"github.com/playkaro/payment-service/internal/handlers"
	"github.com/playkaro/payment-service/internal/telemetry"
	"github.com/playkaro/payment-service/internal/wallet"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

	// Initialize Razorpay client
	razorpayKeyID := os.Getenv("RAZORPAY_KEY_ID")
	razorpayKeySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	if razorpayKeyID == "" || razorpayKeySecret == "" {
		log.Fatal("Razorpay credentials not set")
	}
	razorpayClient := razorpay.NewClient(razorpayKeyID, razorpayKeySecret)

	// Initialize handlers
	paymentHandler := handlers.NewPaymentHandler(db.DB, razorpayClient)

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer("payment-service", "otel-collector:4317")
	if err != nil {
		log.Printf("Failed to initialize OpenTelemetry: %v", err)
	} else {
		defer shutdown(context.Background())
	}

	// Setup router
	r := gin.Default()
	r.Use(otelgin.Middleware("payment-service"))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "payment-service"})
	})

	// Payment routes
	v1 := r.Group("/v1/payments")
	{
		// Public routes
		v1.POST("/webhook/razorpay", paymentHandler.HandleRazorpayWebhook)

		// Internal routes (Protected by API Key in production)
		internal := v1.Group("/internal")
		{
			internal.POST("/transaction", paymentHandler.ProcessInternalTransaction)
			internal.GET("/balance", paymentHandler.GetBalance)
		}

		// Protected routes (require JWT)
		authorized := v1.Group("")
		authorized.Use(AuthMiddleware())
		{
			authorized.POST("/deposit", paymentHandler.InitiateDeposit)
			authorized.GET("/order/:order_id", paymentHandler.GetOrderStatus)
			authorized.GET("/balance", paymentHandler.GetBalance)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start HTTP server
	go func() {
		log.Printf("Payment Service HTTP starting on port %s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	// Initialize Wallet Service (Business Logic)
	walletService := wallet.NewService(db.DB)

	// Register gRPC server
	pb.RegisterWalletServiceServer(grpcServer, grpc_impl.NewWalletServer(walletService))

	log.Printf("Payment Service gRPC starting on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// AuthMiddleware validates JWT token (simplified version)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper JWT validation
		// For now, just extract user ID from header
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}
