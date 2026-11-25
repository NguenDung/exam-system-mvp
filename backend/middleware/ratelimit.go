package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRateLimiter struct {
	mu      sync.Mutex
	history map[string][]time.Time
	limit   int
	window  time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		history: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
	// Background cleanup to prevent memory leaks
	go func() {
		ticker := time.NewTicker(2 * window)
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *ipRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for ip, timestamps := range rl.history {
		var valid []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.history, ip)
		} else {
			rl.history[ip] = valid
		}
	}
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	timestamps := rl.history[ip]
	valid := make([]time.Time, 0, len(timestamps))
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.history[ip] = valid
		return false
	}

	rl.history[ip] = append(valid, now)
	return true
}

// IPRateLimit returns a middleware that limits requests per IP.
// It is exempted when running in test mode or if RATE_LIMIT_DISABLED="true".
func IPRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newIPRateLimiter(limit, window)

	return func(c *gin.Context) {
		if gin.Mode() == gin.TestMode || os.Getenv("RATE_LIMIT_DISABLED") == "true" {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if !limiter.allow(clientIP) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Quá nhiều yêu cầu. Vui lòng thử lại sau.",
			})
			return
		}
		c.Next()
	}
}
