package redis

import (
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewClusterClient creates a Redis Cluster client
func NewClusterClient() *redis.ClusterClient {
	addrs := os.Getenv("REDIS_CLUSTER_ADDRS")
	if addrs == "" {
		addrs = ":7000,:7001,:7002" // Default local cluster ports
	}

	password := os.Getenv("REDIS_PASSWORD")

	opts := &redis.ClusterOptions{
		Addrs:    strings.Split(addrs, ","),
		Password: password,

		// Connection Pool
		PoolSize:     100,
		MinIdleConns: 10,

		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Retry
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}

	// TLS Support
	if os.Getenv("REDIS_TLS_ENABLED") == "true" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return redis.NewClusterClient(opts)
}

// NewStandaloneClient creates a standard Redis client (fallback)
func NewStandaloneClient() *redis.Client {
	addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}

// GetClient returns either a cluster or standalone client based on config
// Note: In a real app, we'd use an interface, but go-redis v8 UniversalClient covers both.
func GetClient() redis.UniversalClient {
	if os.Getenv("REDIS_MODE") == "cluster" {
		return NewClusterClient()
	}
	return NewStandaloneClient()
}
