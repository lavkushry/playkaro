package ludo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// AntiCheatAlert severity levels
const (
	SeverityLow    = "LOW"
	SeverityMedium = "MEDIUM"
	SeverityHigh   = "HIGH"
)

// Alert types
const (
	AlertTypeTiming      = "TIMING"
	AlertTypeWinRate     = "WIN_RATE"
	AlertTypeInvalidMove = "INVALID_MOVE"
	AlertTypeStatistical = "STATISTICAL"
)

// AntiCheatDetector monitors gameplay for suspicious patterns
type AntiCheatDetector struct {
	DB *sql.DB
}

// NewAntiCheatDetector creates a new anti-cheat detector
func NewAntiCheatDetector(db *sql.DB) *AntiCheatDetector {
	return &AntiCheatDetector{DB: db}
}

// AntiCheatAlert represents a suspicious activity alert
type AntiCheatAlert struct {
	UserID    string
	SessionID string
	AlertType string
	Severity  string
	Details   map[string]interface{}
}

// CheckMove validates a move and detects suspicious patterns
func (a *AntiCheatDetector) CheckMove(sessionID, userID string, move MoveRecord) []AntiCheatAlert {
	var alerts []AntiCheatAlert

	// Check 1: Timing anomalies
	if timingAlert := a.checkTiming(sessionID, userID); timingAlert != nil {
		alerts = append(alerts, *timingAlert)
	}

	// Check 2: Invalid move distance
	if move.ToPos - move.FromPos != move.DiceRoll && move.FromPos > 0 {
		alerts = append(alerts, AntiCheatAlert{
			UserID:    userID,
			SessionID: sessionID,
			AlertType: AlertTypeInvalidMove,
			Severity:  SeverityHigh,
			Details: map[string]interface{}{
				"expected_distance": move.DiceRoll,
				"actual_distance":   move.ToPos - move.FromPos,
				"move":              move,
			},
		})
	}

	return alerts
}

// CheckSession performs session-level checks after game ends
func (a *AntiCheatDetector) CheckSession(sessionID, userID string, won bool) []AntiCheatAlert {
	var alerts []AntiCheatAlert

	// Check win rate if player won
	if won {
		if winRateAlert := a.checkWinRate(userID); winRateAlert != nil {
			alerts = append(alerts, *winRateAlert)
		}
	}

	return alerts
}

// checkTiming analyzes move timing patterns
func (a *AntiCheatDetector) checkTiming(sessionID, userID string) *AntiCheatAlert {
	// Get recent moves from this session
	rows, err := a.DB.Query(`
		SELECT moves
		FROM game_replays
		WHERE session_id = $1
		LIMIT 1
	`, sessionID)

	if err != nil || !rows.Next() {
		return nil
	}
	defer rows.Close()

	var movesJSON []byte
	rows.Scan(&movesJSON)

	var moves []MoveRecord
	json.Unmarshal(movesJSON, &moves)

	// Filter moves by this user
	var userMoves []MoveRecord
	for _, move := range moves {
		if move.PlayerID == userID {
			userMoves = append(userMoves, move)
		}
	}

	if len(userMoves) < 5 {
		return nil // Not enough data
	}

	// Calculate average time between moves
	var totalTime time.Duration
	for i := 1; i < len(userMoves); i++ {
		totalTime += userMoves[i].Timestamp.Sub(userMoves[i-1].Timestamp)
	}
	avgTime := totalTime / time.Duration(len(userMoves)-1)

	// Flag if consistently < 200ms (bot-like) or > 60s (AFK)
	if avgTime < 200*time.Millisecond {
		return &AntiCheatAlert{
			UserID:    userID,
			SessionID: sessionID,
			AlertType: AlertTypeTiming,
			Severity:  SeverityHigh,
			Details: map[string]interface{}{
				"avg_move_time_ms": avgTime.Milliseconds(),
				"threshold_ms":     200,
				"reason":           "Suspiciously fast moves (possible bot)",
			},
		}
	}

	return nil
}

// checkWinRate analyzes historical win rate
func (a *AntiCheatDetector) checkWinRate(userID string) *AntiCheatAlert {
	// Get last 100 games
	var totalGames, wins int
	err := a.DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE winner = $1) as wins
		FROM game_replays
		WHERE $1 = ANY(players)
		AND completed_at > NOW() - INTERVAL '30 days'
		LIMIT 100
	`, userID).Scan(&totalGames, &wins)

	if err != nil || totalGames < 20 {
		return nil // Not enough data
	}

	winRate := float64(wins) / float64(totalGames)

	// Flag if win rate > 80% (statistically improbable)
	if winRate > 0.80 {
		return &AntiCheatAlert{
			UserID:    userID,
			AlertType: AlertTypeWinRate,
			Severity:  SeverityMedium,
			Details: map[string]interface{}{
				"win_rate":     winRate,
				"total_games":  totalGames,
				"wins":         wins,
				"threshold":    0.80,
				"reason":       "Suspiciously high win rate",
			},
		}
	}

	return nil
}

// LogAlert saves an alert to the database
func (a *AntiCheatDetector) LogAlert(alert AntiCheatAlert) error {
	detailsJSON, _ := json.Marshal(alert.Details)

	_, err := a.DB.Exec(`
		INSERT INTO anticheat_alerts (user_id, session_id, alert_type, severity, details)
		VALUES ($1, $2, $3, $4, $5)
	`, alert.UserID, alert.SessionID, alert.AlertType, alert.Severity, string(detailsJSON))

	return err
}

// GetUserAlerts retrieves all alerts for a user
func (a *AntiCheatDetector) GetUserAlerts(userID string, limit int) ([]AntiCheatAlert, error) {
	rows, err := a.DB.Query(`
		SELECT user_id, session_id, alert_type, severity, details
		FROM anticheat_alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AntiCheatAlert
	for rows.Next() {
		var alert AntiCheatAlert
		var detailsJSON string

		rows.Scan(&alert.UserID, &alert.SessionID, &alert.AlertType, &alert.Severity, &detailsJSON)
		json.Unmarshal([]byte(detailsJSON), &alert.Details)

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetSuspiciousUsers returns users with high-severity alerts
func (a *AntiCheatDetector) GetSuspiciousUsers(limit int) ([]string, error) {
	rows, err := a.DB.Query(`
		SELECT DISTINCT user_id
		FROM anticheat_alerts
		WHERE severity = $1
		AND created_at > NOW() - INTERVAL '7 days'
		ORDER BY created_at DESC
		LIMIT $2
	`, SeverityHigh, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userID string
		rows.Scan(&userID)
		users = append(users, userID)
	}

	return users, nil
}

// AutoBan automatically bans users with multiple high-severity alerts
func (a *AntiCheatDetector) AutoBan(userID string) (bool, error) {
	var highSeverityCount int
	err := a.DB.QueryRow(`
		SELECT COUNT(*)
		FROM anticheat_alerts
		WHERE user_id = $1
		AND severity = $2
		AND created_at > NOW() - INTERVAL '7 days'
	`, userID, SeverityHigh).Scan(&highSeverityCount)

	if err != nil {
		return false, err
	}

	// Auto-ban if 3+ high-severity alerts in 7 days
	if highSeverityCount >= 3 {
		// TODO: Integrate with user management service to actually ban
		fmt.Printf("AUTO-BAN: User %s has %d high-severity alerts\n", userID, highSeverityCount)
		return true, nil
	}

	return false, nil
}
