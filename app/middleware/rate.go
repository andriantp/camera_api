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

// rateLimitStore holds the counters of a single limiter.
// Every call to RateLimiter.Middleware gets its own store, so limiters with
// different limits/windows never share keys or wipe each other's entries.
type rateLimitStore struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	clients   map[string]*rateLimitEntry
	lastSweep time.Time
}

type RateLimiter struct{}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (r *RateLimiter) Middleware(limit int, window time.Duration, keyFunc func(echo.Context) string) echo.MiddlewareFunc {
	store := &rateLimitStore{
		limit:   limit,
		window:  window,
		clients: make(map[string]*rateLimitEntry),
	}

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

			if !store.allow(key, time.Now()) {
				return response.APIError{
					StatusCode: http.StatusTooManyRequests,
					Message:    "too many requests",
				}
			}

			return next(c)
		}
	}
}

func (s *rateLimitStore) allow(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sweep expired entries at most once per window instead of on every
	// request, keeping the hot path O(1) while bounding memory usage.
	if now.Sub(s.lastSweep) >= s.window {
		for clientKey, entry := range s.clients {
			if now.Sub(entry.startedAt) >= s.window {
				delete(s.clients, clientKey)
			}
		}
		s.lastSweep = now
	}

	entry, exists := s.clients[key]
	if !exists || now.Sub(entry.startedAt) >= s.window {
		s.clients[key] = &rateLimitEntry{
			count:     1,
			startedAt: now,
		}

		return true
	}

	if entry.count >= s.limit {
		return false
	}

	entry.count++

	return true
}
