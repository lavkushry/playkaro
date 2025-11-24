package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	KeyID     string
	KeySecret string
	BaseURL   string
}

type OrderRequest struct {
	Amount   int64  `json:"amount"`   // Amount in paise (₹1 = 100 paise)
	Currency string `json:"currency"` // INR
	Receipt  string `json:"receipt"`  // Unique receipt ID
}

type OrderResponse struct {
	ID       string `json:"id"`
	Entity   string `json:"entity"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
	Status   string `json:"status"`
}

type PaymentResponse struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	Status  string `json:"status"`
	Method  string `json:"method"`
}

func NewClient(keyID, keySecret string) *Client {
	return &Client{
		KeyID:     keyID,
		KeySecret: keySecret,
		BaseURL:   "https://api.razorpay.com/v1",
	}
}

// CreateOrder creates a payment order with Razorpay
func (c *Client) CreateOrder(amount float64, currency, receipt string) (*OrderResponse, error) {
	amountPaise := int64(amount * 100) // Convert to paise

	orderReq := OrderRequest{
		Amount:   amountPaise,
		Currency: currency,
		Receipt:  receipt,
	}

	payload, err := json.Marshal(orderReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/orders", strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.KeyID, c.KeySecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay error: %s", string(body))
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

// VerifySignature verifies the webhook signature from Razorpay
func (c *Client) VerifySignature(orderID, paymentID, signature string) bool {
	message := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(c.KeySecret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}

// VerifyWebhookSignature verifies the webhook signature
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) bool {
	h := hmac.New(sha256.New, []byte(c.KeySecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}

// GetPayment fetches payment details
func (c *Client) GetPayment(paymentID string) (*PaymentResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.KeyID, c.KeySecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay error: %s", string(body))
	}

	var paymentResp PaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}
