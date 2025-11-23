package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/playkaro/backend/internal/db"
	pb "github.com/playkaro/backend/proto/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedWalletServiceServer
}

func (s *server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	var balance, bonus float64
	var currency string

	err := db.DB.QueryRow("SELECT balance, COALESCE(bonus, 0), currency FROM wallets WHERE user_id=$1", req.UserId).
		Scan(&balance, &bonus, &currency)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}

	return &pb.GetBalanceResponse{
		Balance:  balance,
		Bonus:    bonus,
		Currency: currency,
	}, nil
}

func (s *server) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Update Wallet
	_, err = tx.Exec("UPDATE wallets SET balance = balance + $1 WHERE user_id=$2", req.Amount, req.UserId)
	if err != nil {
		return nil, err
	}

	// 2. Create Transaction
	var txID string
	err = tx.QueryRow(
		"INSERT INTO transactions (wallet_id, type, amount, status, reference_id) SELECT id, 'DEPOSIT', $1, 'COMPLETED', $2 FROM wallets WHERE user_id=$3 RETURNING id",
		req.Amount, "GRPC-"+time.Now().Format("20060102150405"), req.UserId,
	).Scan(&txID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Get new balance
	var newBalance float64
	db.DB.QueryRow("SELECT balance FROM wallets WHERE user_id=$1", req.UserId).Scan(&newBalance)

	return &pb.DepositResponse{
		Success:       true,
		TransactionId: txID,
		Message:       "Deposit successful",
		NewBalance:    newBalance,
	}, nil
}

func (s *server) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Check Balance
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id=$1 FOR UPDATE", req.UserId).Scan(&currentBalance)
	if err != nil {
		return nil, err
	}

	if currentBalance < req.Amount {
		return &pb.WithdrawResponse{Success: false, Message: "Insufficient funds"}, nil
	}

	// 2. Deduct Balance
	_, err = tx.Exec("UPDATE wallets SET balance = balance - $1 WHERE user_id=$2", req.Amount, req.UserId)
	if err != nil {
		return nil, err
	}

	// 3. Create Transaction
	var txID string
	err = tx.QueryRow(
		"INSERT INTO transactions (wallet_id, type, amount, status, reference_id) SELECT id, 'WITHDRAW', $1, 'COMPLETED', $2 FROM wallets WHERE user_id=$3 RETURNING id",
		req.Amount, "GRPC-"+time.Now().Format("20060102150405"), req.UserId,
	).Scan(&txID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pb.WithdrawResponse{
		Success:       true,
		TransactionId: txID,
		Message:       "Withdrawal successful",
		NewBalance:    currentBalance - req.Amount,
	}, nil
}

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found")
	}

	// Initialize DB connection
	db.Connect()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, &server{})
	reflection.Register(s)

	log.Printf("Wallet Service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
