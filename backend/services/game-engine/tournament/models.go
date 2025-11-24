package tournament

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Tournament Status
const (
	StatusRegistration = "REGISTRATION"
	StatusActive       = "ACTIVE"
	StatusCompleted    = "COMPLETED"
	StatusCancelled    = "CANCELLED"
)

// Participant Status
const (
	ParticipantRegistered = "REGISTERED"
	ParticipantEliminated = "ELIMINATED"
	ParticipantWinner     = "WINNER"
)

// Match Status
const (
	MatchScheduled  = "SCHEDULED"
	MatchInProgress = "IN_PROGRESS"
	MatchCompleted  = "COMPLETED"
)

// Prize Types
const (
	PrizeWinnerTakesAll = "WINNER_TAKES_ALL"
	PrizeTop3           = "TOP_3"
	PrizeTiered         = "TIERED"
)

// TournamentConfig defines rules for the tournament
type TournamentConfig struct {
	BracketType     string                 `json:"bracket_type"` // "SINGLE_ELIMINATION"
	PrizeStrategy   string                 `json:"prize_strategy"`
	PrizeDistribution map[string]float64 `json:"prize_distribution,omitempty"` // Rank -> Percentage
	MinPlayers      int                    `json:"min_players"`
}

// Value implements driver.Valuer for JSON storage
func (c TournamentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for JSON storage
func (c *TournamentConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

// Tournament represents a tournament instance
type Tournament struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	GameType    string           `json:"game_type"` // "LUDO", "CHESS"
	EntryFee    float64          `json:"entry_fee"`
	PrizePool   float64          `json:"prize_pool"`
	MaxPlayers  int              `json:"max_players"`
	CurrentPlayers int           `json:"current_players"`
	Status      string           `json:"status"`
	Config      TournamentConfig `json:"config"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// Participant represents a user in the tournament
type Participant struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	Rank         int       `json:"rank"`
	PrizeAmount  float64   `json:"prize_amount"`
	RegisteredAt time.Time `json:"registered_at"`
}

// TournamentMatch represents a match node in the bracket
type TournamentMatch struct {
	ID           string     `json:"id"`
	TournamentID string     `json:"tournament_id"`
	Round        int        `json:"round"`       // 1 = Final, 2 = Semis, etc. (or reverse)
	MatchIndex   int        `json:"match_index"` // Position in the round
	Player1ID    *string    `json:"player_1_id"` // UserID (nullable for Bye)
	Player2ID    *string    `json:"player_2_id"` // UserID
	WinnerID     *string    `json:"winner_id"`
	Status       string     `json:"status"`
	NextMatchID  *string    `json:"next_match_id"` // Pointer to parent node
	Metadata     string     `json:"metadata"`      // Game session ID etc.
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
