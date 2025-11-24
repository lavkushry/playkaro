package models

import (
	"time"
)

// Friendship Status
const (
	FriendStatusPending  = "PENDING"
	FriendStatusAccepted = "ACCEPTED"
	FriendStatusBlocked  = "BLOCKED"
)

// Chat Types
const (
	ChatTypeGlobal  = "GLOBAL"
	ChatTypePrivate = "PRIVATE"
	ChatTypeSystem  = "SYSTEM"
)

// Friendship represents a relationship between two users
type Friendship struct {
	ID          string    `json:"id" db:"id"`
	RequesterID string    `json:"requester_id" db:"requester_id"`
	AddresseeID string    `json:"addressee_id" db:"addressee_id"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ChatMessage represents a message in the system
type ChatMessage struct {
	ID          string    `json:"id" db:"id"`
	SenderID    string    `json:"sender_id" db:"sender_id"`
	RecipientID *string   `json:"recipient_id,omitempty" db:"recipient_id"` // Null for global chat
	Content     string    `json:"content" db:"content"`
	Type        string    `json:"type" db:"type"`
	Timestamp   time.Time `json:"timestamp" db:"created_at"`
}

// UserProfile represents enhanced user data
type UserProfile struct {
	UserID      string                 `json:"user_id" db:"user_id"`
	Bio         string                 `json:"bio" db:"bio"`
	AvatarFrame string                 `json:"avatar_frame" db:"avatar_frame"`
	Title       string                 `json:"title" db:"title"`
	Stats       map[string]interface{} `json:"stats" db:"stats"` // JSONB
	IsOnline    bool                   `json:"is_online"`
	LastActive  time.Time              `json:"last_active" db:"last_active"`
}
