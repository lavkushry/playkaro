package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type IdempotencyMiddleware struct {
	Redis *redis.Client
}

func NewIdempotencyMiddleware(redisClient *redis.Client) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{Redis: redisClient}
}

// IdempotencyHandler ensures requests with same Idempotency-Key return same response
func (m *IdempotencyMiddleware) IdempotencyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idempotencyKey := c.GetHeader("Idempotency-Key")

		// Skip if no idempotency key provided
		if idempotencyKey == "" {
			c.Next()
			return
		}

		ctx := context.Background()
		cacheKey := fmt.Sprintf("idempotency:%s", idempotencyKey)

		// Check if we've seen this key before
		cachedResponse, err := m.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			// Return cached response
			c.Header("X-Idempotency-Replay", "true")
			c.Data(http.StatusOK, "application/json", []byte(cachedResponse))
			c.Abort()
			return
		}

		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &[]byte{},
		}
		c.Writer = writer

		c.Next()

		// Cache successful responses (2xx status codes)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			m.Redis.Set(ctx, cacheKey, *writer.body, 24*time.Hour)
		}
	}
}

// responseWriter captures response body
type responseWriter struct {
	gin.ResponseWriter
	body *[]byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	*w.body = append(*w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}
