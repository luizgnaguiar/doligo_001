package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"doligo_001/internal/infrastructure/logger"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	limiters map[string]*clientLimiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter instance.
func NewRateLimiter(requestsPerSecond int, burst int) *rateLimiter {
	// Start a cleanup goroutine to remove stale limiters
	rl := &rateLimiter{
		limiters: make(map[string]*clientLimiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
	rl.cleanupStale()
	return rl
}

func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.limiters[ip]
	if !exists {
		client = &clientLimiter{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.limiters[ip] = client
	}

	client.lastSeen = time.Now()
	return client.limiter
}

// Middleware returns the Echo middleware function.
func (rl *rateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				// Log blocked request
				// Retrieve logger from context, or use default if not present
				log := logger.FromContext(c.Request().Context())
				
				log.Warn("Rate limit exceeded",
					"ip", ip,
					"path", c.Request().URL.Path,
				)

				c.Response().Header().Set("Retry-After", "60")
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}

// cleanupStale periodically removes limiters that haven't been used recently.
func (rl *rateLimiter) cleanupStale() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for ip, client := range rl.limiters {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(rl.limiters, ip)
				}
			}
			slog.Info("Rate limiter cleanup completed", "active_ips", len(rl.limiters))
			rl.mu.Unlock()
		}
	}()
}
