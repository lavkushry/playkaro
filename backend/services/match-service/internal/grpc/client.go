package grpc

import (
	"context"
	"fmt"
	"time"

	analytics_pb "github.com/playkaro/backend/proto/analytics"
	wallet_pb "github.com/playkaro/backend/proto/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Wallet    wallet_pb.WalletServiceClient
	Analytics analytics_pb.AnalyticsServiceClient
}

func NewClients(walletAddr, analyticsAddr string) (*Clients, error) {
	// Connect to Wallet Service
	walletConn, err := grpc.Dial(walletAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wallet service: %w", err)
	}
	walletClient := wallet_pb.NewWalletServiceClient(walletConn)

	// Connect to Analytics Service
	analyticsConn, err := grpc.Dial(analyticsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to analytics service: %w", err)
	}
	analyticsClient := analytics_pb.NewAnalyticsServiceClient(analyticsConn)

	return &Clients{
		Wallet:    walletClient,
		Analytics: analyticsClient,
	}, nil
}

// Helper methods for Match Service specific operations

func (c *Clients) LogEvent(userID, eventType, dataJSON string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.Analytics.LogEvent(ctx, &analytics_pb.LogEventRequest{
		UserId:        userID,
		EventType:     eventType,
		EventDataJson: dataJSON,
		Timestamp:     time.Now().Format(time.RFC3339),
	})
	return err
}

func (c *Clients) DebitEntryFee(userID string, amount float64, matchID string) (*wallet_pb.DebitResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.Wallet.Debit(ctx, &wallet_pb.DebitRequest{
		UserId:        userID,
		Amount:        amount,
		ReferenceId:   matchID,
		ReferenceType: "ENTRY_FEE",
	})
}

func (c *Clients) CreditPrize(userID string, amount float64, matchID string) (*wallet_pb.CreditResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.Wallet.Credit(ctx, &wallet_pb.CreditRequest{
		UserId:        userID,
		Amount:        amount,
		ReferenceId:   matchID,
		ReferenceType: "PRIZE",
	})
}
