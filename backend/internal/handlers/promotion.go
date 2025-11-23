package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
)

type ClaimBonusRequest struct {
	BonusType string `json:"bonus_type" binding:"required"` // WELCOME, DAILY
}

type ApplyReferralRequest struct {
	ReferralCode string `json:"referral_code" binding:"required"`
}

// GetBonuses returns user's active bonuses
func GetBonuses(c *gin.Context) {
	userID := c.GetString("userID")

	rows, err := db.DB.Query(`
		SELECT id, type, amount, wagering_requirement, wagered, status, expires_at, created_at
		FROM bonuses WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type BonusInfo struct {
		ID                  string    `json:"id"`
		Type                string    `json:"type"`
		Amount              float64   `json:"amount"`
		WageringRequirement float64   `json:"wagering_requirement"`
		Wagered             float64   `json:"wagered"`
		Status              string    `json:"status"`
		ExpiresAt           time.Time `json:"expires_at"`
		CreatedAt           time.Time `json:"created_at"`
	}

	var bonuses []BonusInfo
	for rows.Next() {
		var b BonusInfo
		rows.Scan(&b.ID, &b.Type, &b.Amount, &b.WageringRequirement, &b.Wagered, &b.Status, &b.ExpiresAt, &b.CreatedAt)
		bonuses = append(bonuses, b)
	}

	c.JSON(http.StatusOK, bonuses)
}

// ClaimBonus allows user to claim available bonuses
func ClaimBonus(c *gin.Context) {
	userID := c.GetString("userID")
	var req ClaimBonusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already claimed this bonus type
	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM bonuses WHERE user_id=$1 AND type=$2)", userID, req.BonusType).Scan(&exists)
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bonus already claimed"})
		return
	}

	var amount float64
	var wageringMultiplier float64

	switch req.BonusType {
	case "WELCOME":
		amount = 100.0 // ₹100 welcome bonus
		wageringMultiplier = 5.0 // Must wager 5x
	case "DAILY":
		amount = 20.0 // ₹20 daily bonus
		wageringMultiplier = 3.0
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bonus type"})
		return
	}

	wageringRequirement := amount * wageringMultiplier
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	_, err := db.DB.Exec(`
		INSERT INTO bonuses (user_id, type, amount, wagering_requirement, wagered, status, expires_at)
		VALUES ($1, $2, $3, $4, 0, 'ACTIVE', $5)
	`, userID, req.BonusType, amount, wageringRequirement, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to claim bonus"})
		return
	}

	// Add bonus to wallet
	db.DB.Exec("UPDATE wallets SET bonus = bonus + $1 WHERE user_id=$2", amount, userID)

	c.JSON(http.StatusOK, gin.H{
		"message":              "Bonus claimed successfully",
		"amount":               amount,
		"wagering_requirement": wageringRequirement,
	})
}

// GenerateReferralCode creates a unique referral code for the user
func GenerateReferralCode(c *gin.Context) {
	userID := c.GetString("userID")

	// Check if user already has a code
	var existingCode string
	err := db.DB.QueryRow("SELECT code FROM referrals WHERE referrer_id=$1 LIMIT 1", userID).Scan(&existingCode)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"referral_code": existingCode})
		return
	}

	// Generate new code
	code := generateRandomCode(8)

	_, err = db.DB.Exec("INSERT INTO referrals (referrer_id, code) VALUES ($1, $2)", userID, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"referral_code": code})
}

// ApplyReferralCode applies a referral code during registration
func ApplyReferralCode(c *gin.Context) {
	userID := c.GetString("userID")
	var req ApplyReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already used a referral
	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM referrals WHERE referred_id=$1)", userID).Scan(&exists)
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Referral code already used"})
		return
	}

	// Find referrer
	var referrerID string
	err := db.DB.QueryRow("SELECT referrer_id FROM referrals WHERE code=$1", req.ReferralCode).Scan(&referrerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid referral code"})
		return
	}

	if referrerID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use own referral code"})
		return
	}

	bonusAmount := 50.0 // ₹50 for both referrer and referred

	tx, _ := db.DB.Begin()
	defer tx.Rollback()

	// Update referral record
	tx.Exec("UPDATE referrals SET referred_id=$1, bonus_awarded=$2 WHERE code=$3", userID, bonusAmount, req.ReferralCode)

	// Give bonus to referrer
	tx.Exec("UPDATE wallets SET bonus = bonus + $1 WHERE user_id=$2", bonusAmount, referrerID)
	tx.Exec(`
		INSERT INTO bonuses (user_id, type, amount, wagering_requirement, wagered, status, expires_at)
		VALUES ($1, 'REFERRAL', $2, $3, 0, 'ACTIVE', $4)
	`, referrerID, bonusAmount, bonusAmount*3, time.Now().Add(30*24*time.Hour))

	// Give bonus to referred user
	tx.Exec("UPDATE wallets SET bonus = bonus + $1 WHERE user_id=$2", bonusAmount, userID)
	tx.Exec(`
		INSERT INTO bonuses (user_id, type, amount, wagering_requirement, wagered, status, expires_at)
		VALUES ($1, 'REFERRAL', $2, $3, 0, 'ACTIVE', $4)
	`, userID, bonusAmount, bonusAmount*3, time.Now().Add(30*24*time.Hour))

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Referral code applied successfully",
		"bonus":   bonusAmount,
	})
}

// GetLeaderboard returns top users by total bets
func GetLeaderboard(c *gin.Context) {
	period := c.Query("period") // daily, weekly, monthly

	var timeFilter string
	switch period {
	case "daily":
		timeFilter = "AND created_at >= NOW() - INTERVAL '1 day'"
	case "weekly":
		timeFilter = "AND created_at >= NOW() - INTERVAL '7 days'"
	case "monthly":
		timeFilter = "AND created_at >= NOW() - INTERVAL '30 days'"
	default:
		timeFilter = ""
	}

	query := `
		SELECT u.username, SUM(b.amount) as total_wagered, COUNT(b.id) as bet_count
		FROM bets b
		JOIN users u ON b.user_id = u.id
		WHERE b.status IN ('WON', 'LOST') ` + timeFilter + `
		GROUP BY u.id, u.username
		ORDER BY total_wagered DESC
		LIMIT 10
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type LeaderboardEntry struct {
		Rank         int     `json:"rank"`
		Username     string  `json:"username"`
		TotalWagered float64 `json:"total_wagered"`
		BetCount     int     `json:"bet_count"`
	}

	var leaderboard []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		rows.Scan(&entry.Username, &entry.TotalWagered, &entry.BetCount)
		entry.Rank = rank
		leaderboard = append(leaderboard, entry)
		rank++
	}

	c.JSON(http.StatusOK, leaderboard)
}

// Helper: Generate random code
func generateRandomCode(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
