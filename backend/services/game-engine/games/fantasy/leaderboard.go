package fantasy

import (
	"sort"
)

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	TeamID      string  `json:"team_id"`
	UserID      string  `json:"user_id"`
	TotalPoints float64 `json:"total_points"`
}

// Leaderboard manages team rankings
type Leaderboard struct {
	ContestID string
	Entries   []LeaderboardEntry
}

// NewLeaderboard creates a new leaderboard
func NewLeaderboard(contestID string) *Leaderboard {
	return &Leaderboard{
		ContestID: contestID,
		Entries:   []LeaderboardEntry{},
	}
}

// UpdateRankings recalculates ranks based on points
func (l *Leaderboard) UpdateRankings(teams []*FantasyTeam) {
	// Convert teams to entries
	l.Entries = make([]LeaderboardEntry, len(teams))
	for i, team := range teams {
		l.Entries[i] = LeaderboardEntry{
			TeamID:      team.ID,
			UserID:      team.UserID,
			TotalPoints: team.TotalPoints,
		}
	}

	// Sort by points descending
	sort.Slice(l.Entries, func(i, j int) bool {
		return l.Entries[i].TotalPoints > l.Entries[j].TotalPoints
	})

	// Assign ranks (handling ties)
	currentRank := 1
	for i := 0; i < len(l.Entries); i++ {
		if i > 0 && l.Entries[i].TotalPoints < l.Entries[i-1].TotalPoints {
			currentRank = i + 1
		}
		l.Entries[i].Rank = currentRank

		// Update original team object rank if needed
		// (In a real system, we'd persist this back to DB)
	}
}

// GetTopTeams returns the top N teams
func (l *Leaderboard) GetTopTeams(n int) []LeaderboardEntry {
	if n > len(l.Entries) {
		n = len(l.Entries)
	}
	return l.Entries[:n]
}

// GetUserRank returns the rank of a specific user
func (l *Leaderboard) GetUserRank(userID string) *LeaderboardEntry {
	for _, entry := range l.Entries {
		if entry.UserID == userID {
			return &entry
		}
	}
	return nil
}
