// Package middleware provides HTTP middleware for the GoFlow API.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goflow-atom/goflow-service/internal/api/dto"
)

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // requests per window
	window   time.Duration // time window
	cleanupT *time.Ticker
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter.
//
// Parameters:
//   - rate: Maximum number of requests per window
//   - window: Time window duration
//
// Returns:
//   - *RateLimiter: New rate limiter instance
//
// Example:
//
//	limiter := middleware.NewRateLimiter(100, time.Minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		window:   window,
		cleanupT: time.NewTicker(window),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old buckets periodically.
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupT.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			if now.Sub(b.lastSeen) > rl.window {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:   rl.rate - 1,
			lastSeen: now,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(b.lastSeen)
	if elapsed > rl.window {
		b.tokens = rl.rate
		b.lastSeen = now
	}

	b.lastSeen = now

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates a rate limiting middleware.
//
// Parameters:
//   - rate: Maximum number of requests per window
//   - window: Time window duration
//
// Returns:
//   - gin.HandlerFunc: Rate limiting middleware
//
// Example:
//
//	router.Use(middleware.RateLimitMiddleware(100, time.Minute))
func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// Use client IP as the key
		key := c.ClientIP()

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
