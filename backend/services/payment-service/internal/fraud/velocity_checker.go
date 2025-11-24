package fraud

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type VelocityChecker struct {
	Redis *redis.Client
}

func NewVelocityChecker(redisClient *redis.Client) *VelocityChecker {
	return &VelocityChecker{Redis: redisClient}
}

// CheckDepositVelocity checks if user has exceeded deposit limits
func (v *VelocityChecker) CheckDepositVelocity(userID string, amount float64) error {
	ctx := context.Background()

	// Rule 1: Max 3 deposits per hour
	hourKey := fmt.Sprintf("deposit:count:hour:%s:%d", userID, time.Now().Hour())
	count, _ := v.Redis.Incr(ctx, hourKey).Result()
	if count == 1 {
		v.Redis.Expire(ctx, hourKey, time.Hour)
	}
	if count > 3 {
		return fmt.Errorf("deposit limit exceeded: max 3 deposits per hour")
	}

	// Rule 2: Max total deposit of ₹50,000 per day
	dayKey := fmt.Sprintf("deposit:amount:day:%s:%s", userID, time.Now().Format("2006-01-02"))
	totalAmount, _ := v.Redis.Get(ctx, dayKey).Float64()
	newTotal := totalAmount + amount
	if newTotal > 50000 {
		return fmt.Errorf("daily deposit limit exceeded: max ₹50,000 per day")
	}

	// Update daily amount
	v.Redis.IncrByFloat(ctx, dayKey, amount)
	v.Redis.Expire(ctx, dayKey, 24*time.Hour)

	return nil
}

// CheckWithdrawalVelocity checks if user has exceeded withdrawal limits
func (v *VelocityChecker) CheckWithdrawalVelocity(userID string) error {
	ctx := context.Background()

	// Rule: Max 10 withdrawals per day
	dayKey := fmt.Sprintf("withdrawal:count:day:%s:%s", userID, time.Now().Format("2006-01-02"))
	count, _ := v.Redis.Incr(ctx, dayKey).Result()
	if count == 1 {
		v.Redis.Expire(ctx, dayKey, 24*time.Hour)
	}
	if count > 10 {
		return fmt.Errorf("withdrawal limit exceeded: max 10 withdrawals per day")
	}

	return nil
}
