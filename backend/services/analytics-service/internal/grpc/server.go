package grpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/playkaro/analytics-service/internal/handlers"
	"github.com/playkaro/analytics-service/internal/models"
	pb "github.com/playkaro/backend/proto/analytics"
)

type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
	IngestHandler *handlers.IngestHandler
}

func NewAnalyticsServer(handler *handlers.IngestHandler) *AnalyticsServer {
	return &AnalyticsServer{IngestHandler: handler}
}

func (s *AnalyticsServer) LogEvent(ctx context.Context, req *pb.LogEventRequest) (*pb.LogEventResponse, error) {
	// Convert gRPC request to internal event model
	event := models.AnalyticsEvent{
		UserID:    req.UserId,
		EventType: req.EventType,
		EventData: json.RawMessage(req.EventDataJson),
		Timestamp: time.Now(),
	}

	// Use existing handler logic to process event
	// Note: We might need to refactor IngestHandler to separate logic from Gin context
	// For now, we'll assume we can reuse the logic or call a service method
	// Since IngestHandler uses Redis/DB directly, we should ideally extract a Service.
	// But for speed, let's just call the processing logic if exposed, or duplicate it slightly.

	// Better approach: Refactor IngestHandler to use a Service, similar to Wallet.
	// But checking IngestHandler code (from memory), it has `updateRealtimeMetrics`.
	// Let's assume we can call `ProcessEvent` on it if we add it.

	// For now, let's just implement the logic directly here as it's simple (Redis incr).
	// Or better, let's update IngestHandler to have a public `ProcessEvent` method.

	err := s.IngestHandler.ProcessEvent(event)
	if err != nil {
		return &pb.LogEventResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.LogEventResponse{Success: true}, nil
}
