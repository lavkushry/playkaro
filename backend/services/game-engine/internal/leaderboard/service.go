package leaderboard

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type LeaderboardService struct {
	Redis *redis.Client
}

func NewLeaderboardService(redisURL string) (*LeaderboardService, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &LeaderboardService{Redis: redis.NewClient(opt)}, nil
}

// UpdateScore adds points to a user's score (Weekly Leaderboard)
func (s *LeaderboardService) UpdateScore(userID string, points float64) error {
	ctx := context.Background()
	key := fmt.Sprintf("leaderboard:weekly:%d", time.Now().ISOWeek())

	return s.Redis.ZIncrBy(ctx, key, points, userID).Err()
}

// GetTopPlayers returns the top N players
func (s *LeaderboardService) GetTopPlayers(limit int64) ([]map[string]interface{}, error) {
	ctx := context.Background()
	key := fmt.Sprintf("leaderboard:weekly:%d", time.Now().ISOWeek())

	// Get top users with scores
	results, err := s.Redis.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	var leaderboard []map[string]interface{}
	for i, z := range results {
		leaderboard = append(leaderboard, map[string]interface{}{
			"rank": i + 1,
			"user_id": z.Member,
			"score": z.Score,
		})
	}

	return leaderboard, nil
}
