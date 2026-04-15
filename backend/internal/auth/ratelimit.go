package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimit throttles bad login attempts per source IP. Keeps the last N
// attempts; if the trailing N fit inside `window`, the next one is rejected
// for `cooldown`. In-memory only — adequate for a single-replica deploy.
func LoginRateLimit(maxAttempts int, window, cooldown time.Duration) gin.HandlerFunc {
	type bucket struct {
		attempts   []time.Time
		blockUntil time.Time
	}
	var (
		mu    sync.Mutex
		state = map[string]*bucket{}
	)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b := state[ip]
		if b == nil {
			b = &bucket{}
			state[ip] = b
		}
		if now.Before(b.blockUntil) {
			retry := int(b.blockUntil.Sub(now).Seconds()) + 1
			mu.Unlock()
			c.Header("Retry-After", intStr(retry))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many attempts — try again shortly",
			})
			return
		}
		// drop stale attempts
		cutoff := now.Add(-window)
		kept := b.attempts[:0]
		for _, t := range b.attempts {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		b.attempts = kept
		mu.Unlock()

		c.Next()

		// Only count failures (401/403); successful logins reset the bucket.
		status := c.Writer.Status()
		mu.Lock()
		defer mu.Unlock()
		if status == http.StatusOK {
			b.attempts = nil
			return
		}
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			b.attempts = append(b.attempts, now)
			if len(b.attempts) >= maxAttempts {
				b.blockUntil = now.Add(cooldown)
				b.attempts = nil
			}
		}
	}
}

func intStr(n int) string {
	if n <= 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
