package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/playkaro/social-service/internal/models"
)

type FriendHandler struct {
	DB *sql.DB
}

func NewFriendHandler(db *sql.DB) *FriendHandler {
	return &FriendHandler{DB: db}
}

// SendFriendRequest sends a friend request
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	requesterID := c.GetString("userID")
	if requesterID == "" {
		requesterID = c.GetHeader("X-User-ID")
	}

	var req struct {
		AddresseeID string `json:"addressee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requesterID == req.AddresseeID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot friend yourself"})
		return
	}

	// Check existing relationship
	var existingStatus string
	err := h.DB.QueryRow(`
		SELECT status FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
	`, requesterID, req.AddresseeID).Scan(&existingStatus)

	if err == nil {
		if existingStatus == models.FriendStatusBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot send request"})
			return
		}
		if existingStatus == models.FriendStatusAccepted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Already friends"})
			return
		}
		if existingStatus == models.FriendStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request already pending"})
			return
		}
	}

	// Create request
	id := uuid.New().String()
	_, err = h.DB.Exec(`
		INSERT INTO friendships (id, requester_id, addressee_id, status)
		VALUES ($1, $2, $3, $4)
	`, id, requesterID, req.AddresseeID, models.FriendStatusPending)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Friend request sent"})
}

// AcceptFriendRequest accepts a pending request
func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	var req struct {
		RequesterID string `json:"requester_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update status to ACCEPTED
	result, err := h.DB.Exec(`
		UPDATE friendships
		SET status = $1, updated_at = $2
		WHERE requester_id = $3 AND addressee_id = $4 AND status = $5
	`, models.FriendStatusAccepted, time.Now(), req.RequesterID, userID, models.FriendStatusPending)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// GetFriends returns the user's friend list
func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	rows, err := h.DB.Query(`
		SELECT
			CASE WHEN requester_id = $1 THEN addressee_id ELSE requester_id END as friend_id,
			status,
			created_at
		FROM friendships
		WHERE (requester_id = $1 OR addressee_id = $1)
		AND status = $2
	`, userID, models.FriendStatusAccepted)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var friends []map[string]interface{}
	for rows.Next() {
		var friendID, status string
		var since time.Time
		rows.Scan(&friendID, &status, &since)
		friends = append(friends, map[string]interface{}{
			"friend_id": friendID,
			"since":     since,
		})
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

// GetPendingRequests returns incoming friend requests
func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	rows, err := h.DB.Query(`
		SELECT requester_id, created_at
		FROM friendships
		WHERE addressee_id = $1 AND status = $2
	`, userID, models.FriendStatusPending)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var requesterID string
		var sentAt time.Time
		rows.Scan(&requesterID, &sentAt)
		requests = append(requests, map[string]interface{}{
			"requester_id": requesterID,
			"sent_at":      sentAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}
