package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/playkaro/match-service/internal/models"
)

type MatchCache struct {
	client *redis.Client
}

func NewMatchCache(redisURL string) (*MatchCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &MatchCache{client: client}, nil
}

// CacheMatch stores match in Redis with 60s TTL
func (c *MatchCache) CacheMatch(ctx context.Context, match *models.Match) error {
	key := fmt.Sprintf("match:%s", match.MatchID)
	data, err := json.Marshal(match)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, 60*time.Second).Err()
}

// GetMatch retrieves match from Redis
func (c *MatchCache) GetMatch(ctx context.Context, matchID string) (*models.Match, error) {
	key := fmt.Sprintf("match:%s", matchID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var match models.Match
	if err := json.Unmarshal([]byte(data), &match); err != nil {
		return nil, err
	}

	return &match, nil
}

// CacheMatches stores multiple matches (for list endpoints)
func (c *MatchCache) CacheMatches(ctx context.Context, key string, matches []*models.Match) error {
	data, err := json.Marshal(matches)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, 30*time.Second).Err()
}

// GetMatches retrieves cached match list
func (c *MatchCache) GetMatches(ctx context.Context, key string) ([]*models.Match, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var matches []*models.Match
	if err := json.Unmarshal([]byte(data), &matches); err != nil {
		return nil, err
	}

	return matches, nil
}

// InvalidateMatch removes match from cache
func (c *MatchCache) InvalidateMatch(ctx context.Context, matchID string) error {
	key := fmt.Sprintf("match:%s", matchID)
	return c.client.Del(ctx, key).Err()
}

// InvalidateAll removes all match caches
func (c *MatchCache) InvalidateAll(ctx context.Context) error {
	// Delete all match:* keys
	keys, err := c.client.Keys(ctx,  "match:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// PublishOddsUpdate publishes odds update to Redis Pub/Sub
func (c *MatchCache) PublishOddsUpdate(ctx context.Context, matchID string, oddsA, oddsB, oddsDraw float64) error {
	update := map[string]interface{}{
		"match_id":  matchID,
		"odds_a":    oddsA,
		"odds_b":    oddsB,
		"odds_draw": oddsDraw,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return c.client.Publish(ctx, "odds_updates", data).Err()
}

// SubscribeOddsUpdates returns a channel for odds updates
func (c *MatchCache) SubscribeOddsUpdates(ctx context.Context) <-chan *redis.Message {
	pubsub := c.client.Subscribe(ctx, "odds_updates")
	return pubsub.Channel()
}
