package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type WalletClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type TransactionRequest struct {
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	TransactionID string  `json:"transaction_id"`
	ReferenceID   string  `json:"reference_id"`
	ReferenceType string  `json:"reference_type"`
}

type TransactionResponse struct {
	Status     string  `json:"status"`
	NewBalance float64 `json:"new_balance"`
	Error      string  `json:"error,omitempty"`
}

func NewWalletClient() *WalletClient {
	baseURL := os.Getenv("PAYMENT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://payment-service:8081"
	}
	return &WalletClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Debit deducts money from user's wallet (Bet)
func (c *WalletClient) Debit(userID string, amount float64, refID, refType string) error {
	return c.sendTransaction(userID, amount, "BET", refID, refType)
}

// Credit adds money to user's wallet (Win)
func (c *WalletClient) Credit(userID string, amount float64, refID, refType string) error {
	return c.sendTransaction(userID, amount, "WIN", refID, refType)
}

func (c *WalletClient) sendTransaction(userID string, amount float64, txType, refID, refType string) error {
	txID := fmt.Sprintf("tx_%s_%s_%d", txType, refID, time.Now().UnixNano())

	reqBody := TransactionRequest{
		UserID:        userID,
		Amount:        amount,
		Type:          txType,
		TransactionID: txID,
		ReferenceID:   refID,
		ReferenceType: refType,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/payments/internal/transaction", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("transaction failed: " + resp.Status)
	}

	var txResp TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResp); err != nil {
		return err
	}

	if txResp.Error != "" {
		return errors.New(txResp.Error)
	}

	return nil
}
