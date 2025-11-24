package models

import (
	"time"
)

type StreamStatus string

const (
	StreamStatusOffline StreamStatus = "OFFLINE"
	StreamStatusLive    StreamStatus = "LIVE"
)

type Stream struct {
	ID          string       `json:"id" db:"id"`
	MatchID     string       `json:"match_id" db:"match_id"` // Can be MatchID or TableID
	StreamKey   string       `json:"-" db:"stream_key"`      // Secret, never return in JSON
	Status      StreamStatus `json:"status" db:"status"`
	ViewerCount int          `json:"viewer_count" db:"viewer_count"`
	PlaybackURL string       `json:"playback_url" db:"playback_url"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

type CreateStreamRequest struct {
	MatchID string `json:"match_id" binding:"required"`
}

type StreamResponse struct {
	ID          string       `json:"id"`
	MatchID     string       `json:"match_id"`
	StreamKey   string       `json:"stream_key,omitempty"` // Only returned on creation
	Status      StreamStatus `json:"status"`
	PlaybackURL string       `json:"playback_url"`
}
