package grpc

import (
	"context"
	"errors"

	pb "github.com/playkaro/backend/proto/wallet"
	"github.com/playkaro/payment-service/internal/wallet"
)

type WalletServer struct {
	pb.UnimplementedWalletServiceServer
	WalletService *wallet.Service
}

func NewWalletServer(service *wallet.Service) *WalletServer {
	return &WalletServer{WalletService: service}
}

func (s *WalletServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	balance, err := s.WalletService.GetBalance(req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetBalanceResponse{
		Balance:  balance.Amount,
		Bonus:    balance.Bonus,
		Currency: balance.Currency,
	}, nil
}

func (s *WalletServer) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	tx, err := s.WalletService.Deposit(req.UserId, req.Amount, req.Method)
	if err != nil {
		return &pb.DepositResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.DepositResponse{
		Success:       true,
		TransactionId: tx.ID,
		NewBalance:    tx.BalanceAfter,
	}, nil
}

func (s *WalletServer) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	tx, err := s.WalletService.Withdraw(req.UserId, req.Amount, req.BankAccountId)
	if err != nil {
		return &pb.WithdrawResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.WithdrawResponse{
		Success:       true,
		TransactionId: tx.ID,
		NewBalance:    tx.BalanceAfter,
	}, nil
}

func (s *WalletServer) Debit(ctx context.Context, req *pb.DebitRequest) (*pb.DebitResponse, error) {
	// Use idempotency key if provided
	if req.IdempotencyKey != "" {
		// Check idempotency (simplified for this implementation)
		// In prod, check Redis/DB for key
	}

	tx, err := s.WalletService.Debit(req.UserId, req.Amount, req.ReferenceId, req.ReferenceType)
	if err != nil {
		if errors.Is(err, wallet.ErrInsufficientFunds) {
			return &pb.DebitResponse{Success: false, Message: "Insufficient funds"}, nil
		}
		return &pb.DebitResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.DebitResponse{
		Success:       true,
		TransactionId: tx.ID,
		NewBalance:    tx.BalanceAfter,
	}, nil
}

func (s *WalletServer) Credit(ctx context.Context, req *pb.CreditRequest) (*pb.CreditResponse, error) {
	tx, err := s.WalletService.Credit(req.UserId, req.Amount, req.ReferenceId, req.ReferenceType)
	if err != nil {
		return &pb.CreditResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.CreditResponse{
		Success:       true,
		TransactionId: tx.ID,
		NewBalance:    tx.BalanceAfter,
	}, nil
}
