package fantasy

import (
	"net/http"
	"time"
)

// CricketAPIClient handles external API communication
type CricketAPIClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewCricketAPIClient creates a new client
func NewCricketAPIClient(apiKey string) *CricketAPIClient {
	return &CricketAPIClient{
		BaseURL: "https://api.cricapi.com/v1", // Example API
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetMatchStats fetches live stats for a match
func (c *CricketAPIClient) GetMatchStats(matchID string) (map[string]PlayerStats, error) {
	// In a real implementation, this would make an HTTP request
	// resp, err := c.Client.Get(fmt.Sprintf("%s/match_score?apikey=%s&id=%s", c.BaseURL, c.APIKey, matchID))

	// For now, return mock data
	return c.getMockStats(matchID), nil
}

// getMockStats returns dummy data for testing
func (c *CricketAPIClient) getMockStats(matchID string) map[string]PlayerStats {
	stats := make(map[string]PlayerStats)

	// Mock stats for a few players
	stats["player_1"] = PlayerStats{
		Runs:       45,
		BallsFaced: 30,
		Fours:      4,
		Sixes:      1,
	}

	stats["player_2"] = PlayerStats{
		Wickets:     2,
		OversBowled: 4.0,
		RunsConceded: 28,
		Maidens:     0,
	}

	return stats
}

// GetSquads fetches players for a match
func (c *CricketAPIClient) GetSquads(matchID string) ([]FantasyPlayer, error) {
	// Mock squad data
	players := []FantasyPlayer{
		{PlayerID: "player_1", Name: "Virat Kohli", Team: "IND", Role: RoleBatsman, Cost: 10.5},
		{PlayerID: "player_2", Name: "Jasprit Bumrah", Team: "IND", Role: RoleBowler, Cost: 9.5},
		{PlayerID: "player_3", Name: "Steve Smith", Team: "AUS", Role: RoleBatsman, Cost: 10.0},
		{PlayerID: "player_4", Name: "Pat Cummins", Team: "AUS", Role: RoleAllRounder, Cost: 9.0},
		// Add more mock players...
	}

	return players, nil
}
