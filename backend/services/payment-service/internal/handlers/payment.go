package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/payment-service/internal/fraud"
	"github.com/playkaro/payment-service/internal/gateways/razorpay"
	"github.com/playkaro/payment-service/internal/models"
)

type PaymentHandler struct {
	DB              *sql.DB
	RazorpayClient  *razorpay.Client
	FraudDetector   *fraud.Detector
}

type DepositRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency"`
	Gateway     string  `json:"gateway"`
	CallbackURL string  `json:"callback_url"`
}

type DepositResponse struct {
	OrderID        string `json:"order_id"`
	GatewayOrderID string `json:"gateway_order_id"`
	PaymentURL     string `json:"payment_url"`
	Status         string `json:"status"`
}

func NewPaymentHandler(db *sql.DB, razorpayClient *razorpay.Client) *PaymentHandler {
	return &PaymentHandler{
		DB:             db,
		RazorpayClient: razorpayClient,
		FraudDetector:  fraud.NewDetector(db),
	}
}

// InitiateDeposit creates a new deposit order
func (h *PaymentHandler) InitiateDeposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = "INR"
	}
	if req.Gateway == "" {
		req.Gateway = models.GatewayRazorpay
	}

	// Run fraud detection
	fraudCheck, err := h.FraudDetector.RunAllChecks(c.Request.Context(), userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fraud check failed"})
		return
	}

	if !fraudCheck.Passed {
		// Log fraud attempt
		h.logFraudCheck(userID, "DEPOSIT_BLOCKED", fraudCheck.RiskScore, fraudCheck.Reason)
		c.JSON(http.StatusForbidden, gin.H{"error": fraudCheck.Reason})
		return
	}

	// Log fraud check even if passed (for monitoring)
	if fraudCheck.RiskScore > 50 {
		h.logFraudCheck(userID, "DEPOSIT_HIGH_RISK", fraudCheck.RiskScore, fraudCheck.Reason)
	}

	// Generate unique order ID
	orderID := fmt.Sprintf("ord_%d", time.Now().UnixNano())

	// Create order with gateway
	var gatewayOrderID string
	var paymentURL string

	if req.Gateway == models.GatewayRazorpay {
		rzpOrder, err := h.RazorpayClient.CreateOrder(req.Amount, req.Currency, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment order"})
			return
		}
		gatewayOrderID = rzpOrder.ID
		paymentURL = fmt.Sprintf("https://checkout.razorpay.com/v1/checkout.html?order_id=%s&callback_url=%s",
			rzpOrder.ID, req.CallbackURL)

		// Note: We rely on the frontend/checkout to pass 'notes' with user_id to Razorpay
		// OR we should store the mapping of order_id -> user_id in our DB (which we do in payment_orders)
		// and look it up in the webhook.
		// For this implementation, let's look it up from DB in the webhook handler.
	}

	// Save to database
	_, err = h.DB.Exec(`
		INSERT INTO payment_orders
		(user_id, order_id, gateway, gateway_order_id, amount, currency, type, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, userID, orderID, req.Gateway, gatewayOrderID, req.Amount, req.Currency,
		models.TypeDeposit, models.StatusInitiated)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order"})
		return
	}

	c.JSON(http.StatusOK, DepositResponse{
		OrderID:        orderID,
		GatewayOrderID: gatewayOrderID,
		PaymentURL:     paymentURL,
		Status:         models.StatusInitiated,
	})
}

// HandleRazorpayWebhook processes Razorpay webhooks
func (h *PaymentHandler) HandleRazorpayWebhook(c *gin.Context) {
	// Read payload
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Razorpay-Signature")
	if !h.RazorpayClient.VerifyWebhookSignature(payload, signature) {
		// Log invalid signature
		h.logWebhook("razorpay", "signature_invalid", payload, signature, false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Log valid webhook
	h.logWebhook("razorpay", "payment.success", payload, signature, true)

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	event := webhookData["event"].(string)

	if event == "payment.captured" {
		// Extract payment details
		paymentData := webhookData["payload"].(map[string]interface{})["payment"].(map[string]interface{})["entity"].(map[string]interface{})

		orderID := paymentData["order_id"].(string)
		paymentID := paymentData["id"].(string)
		status := paymentData["status"].(string)

		// Lookup UserID from our DB
		var userID string
		err = h.DB.QueryRow("SELECT user_id FROM payment_orders WHERE gateway_order_id = $1", orderID).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order not found"})
			return
		}

		// Update order status
		if status == "captured" {
			_, err = h.DB.Exec(`
				UPDATE payment_orders
				SET status = $1, gateway_order_id = $2, completed_at = $3, updated_at = $3
				WHERE gateway_order_id = $4
			`, models.StatusSuccess, paymentID, time.Now(), orderID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
				return
			}

			// Credit Points to Wallet (1 INR = 1 Point)
			// We use the internal transaction API logic directly here for efficiency
			// or call a helper function. For now, let's do a direct DB update for simplicity
			// but ideally we should use the Ledger system.

			// Let's use a helper to ensure ledger consistency
			err = h.creditDepositToWallet(userID,
				paymentData["amount"].(float64)/100, // Amount is in paise
				orderID)

			if err != nil {
				// Log error but don't fail the webhook response (idempotency needed)
				fmt.Printf("Failed to credit wallet: %v\n", err)
			}

			// TODO: Publish Kafka event `payment.deposit.success`
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// GetOrderStatus retrieves the status of a payment order
func (h *PaymentHandler) GetOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	userID := c.GetString("userID")

	var order models.PaymentOrder
	err := h.DB.QueryRow(`
		SELECT id, user_id, order_id, gateway, gateway_order_id, amount, currency,
		       type, status, created_at, completed_at
		FROM payment_orders
		WHERE order_id = $1 AND user_id = $2
	`, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.OrderID, &order.Gateway, &order.GatewayOrderID,
		&order.Amount, &order.Currency, &order.Type, &order.Status, &order.CreatedAt, &order.CompletedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, order)
}
func (h *PaymentHandler) creditDepositToWallet(userID string, amount float64, orderID string) error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Check if already processed
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM ledger WHERE transaction_id = $1)", orderID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already processed
	}

	// 2. Lock Wallet
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO wallets (user_id, balance, currency) VALUES ($1, 0, 'PTS')", userID)
		currentBalance = 0
	} else if err != nil {
		return err
	}

	// 3. Update Balance
	newBalance := currentBalance + amount
	_, err = tx.Exec("UPDATE wallets SET balance = $1, updated_at = $2 WHERE user_id = $3", newBalance, time.Now(), userID)
	if err != nil {
		return err
	}

	// 4. Create Ledger Entry
	_, err = tx.Exec(`
		INSERT INTO ledger (transaction_id, user_id, type, amount, reference_id, reference_type, balance_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, orderID, userID, models.TxTypeDeposit, amount, orderID, "PAYMENT_GATEWAY", newBalance)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Helper functions
func (h *PaymentHandler) logFraudCheck(userID, checkType string, riskScore int, details string) {
	h.DB.Exec(`
		INSERT INTO fraud_checks (user_id, check_type, risk_score, flagged, details)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, checkType, riskScore, riskScore > 70, fmt.Sprintf(`{"reason": "%s"}`, details))
}

func (h *PaymentHandler) logWebhook(gateway, eventType string, payload []byte, signature string, valid bool) {
	h.DB.Exec(`
		INSERT INTO webhook_logs (gateway, event_type, payload, signature, signature_valid)
		VALUES ($1, $2, $3, $4, $5)
	`, gateway, eventType, string(payload), signature, valid)
}
