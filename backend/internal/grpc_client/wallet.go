package grpc_client

import (
	"context"
	"log"
	"time"

	pb "github.com/playkaro/backend/proto/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var WalletClient pb.WalletServiceClient

func InitWalletClient() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to wallet service: %v", err)
	}
	// defer conn.Close() // In a real app, handle this better

	WalletClient = pb.NewWalletServiceClient(conn)
	log.Println("Connected to Wallet Microservice via gRPC")
}

func GetBalance(userID string) (*pb.GetBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return WalletClient.GetBalance(ctx, &pb.GetBalanceRequest{UserId: userID})
}

func Deposit(userID string, amount float64, method string) (*pb.DepositResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return WalletClient.Deposit(ctx, &pb.DepositRequest{UserId: userID, Amount: amount, Method: method})
}
