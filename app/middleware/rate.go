package middleware

import (
	"net/http"
	"sync"
	"time"

	"camera_api/pkg/api/response"

	"github.com/labstack/echo/v4"
)

type rateLimitEntry struct {
	count     int
	startedAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*rateLimitEntry
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*rateLimitEntry),
	}
}

func (r *RateLimiter) Middleware(limit int, window time.Duration, keyFunc func(echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if limit <= 0 || window <= 0 {
				return response.APIError{
					StatusCode: http.StatusInternalServerError,
					Message:    "invalid rate limiter configuration",
				}
			}

			key := keyFunc(c)
			if key == "" {
				return response.APIError{
					StatusCode: http.StatusBadRequest,
					Message:    "invalid rate limit key",
				}
			}

			now := time.Now()

			allowed := r.allow(key, limit, window, now)
			if !allowed {
				return response.APIError{
					StatusCode: http.StatusTooManyRequests,
					Message:    "too many requests",
				}
			}

			return next(c)
		}
	}
}

func (r *RateLimiter) allow(key string, limit int, window time.Duration, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove expired entries.
	for clientKey, entry := range r.clients {
		if now.Sub(entry.startedAt) >= window {
			delete(r.clients, clientKey)
		}
	}

	entry, exists := r.clients[key]

	if !exists {
		r.clients[key] = &rateLimitEntry{
			count:     1,
			startedAt: now,
		}

		return true
	}

	if entry.count >= limit {
		return false
	}

	entry.count++

	return true
}
