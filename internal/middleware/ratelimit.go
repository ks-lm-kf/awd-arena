package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

// MultiKeyRateLimiter supports rate limiting by multiple keys (IP, user, endpoint).
type MultiKeyRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewMultiKeyRateLimiter(limit int, window time.Duration) *MultiKeyRateLimiter {
	return &MultiKeyRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *MultiKeyRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.requests[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= rl.limit {
		rl.requests[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	rl.requests[key] = filtered
	return true
}

// RateLimit limits by IP (with team awareness).
func RateLimit(limit int, window time.Duration) fiber.Handler {
	limiter := NewMultiKeyRateLimiter(limit, window)
	return func(c fiber.Ctx) error {
		key := c.IP()
		if teamID := c.Locals("team_id"); teamID != nil {
			key = "team:" + teamID.(string)
		}
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "rate limit exceeded"})
		}
		return c.Next()
	}
}

// LoginRateLimit limits login attempts per IP (prevent brute force).
// Stricter: 10 attempts per minute per IP.
func LoginRateLimit() fiber.Handler {
	limiter := NewMultiKeyRateLimiter(10, 1*time.Minute)
	return func(c fiber.Ctx) error {
		key := "login:" + c.IP()
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "too many login attempts, try again later"})
		}
		return c.Next()
	}
}

// FlagSubmitRateLimit limits flag submissions.
func FlagSubmitRateLimit(limit int, window time.Duration) fiber.Handler {
	limiter := NewMultiKeyRateLimiter(limit, window)
	return func(c fiber.Ctx) error {
		key := "flag:" + c.IP()
		if teamID := c.Locals("team_id"); teamID != nil {
			key = "flag:team:" + teamID.(string)
		}
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "flag submit rate limit exceeded"})
		}
		return c.Next()
	}
}

// PerUserRateLimit limits per authenticated user.
func PerUserRateLimit(limit int, window time.Duration) fiber.Handler {
	limiter := NewMultiKeyRateLimiter(limit, window)
	return func(c fiber.Ctx) error {
		key := c.IP()
		if userID := c.Locals("user_id"); userID != nil {
			key = "user:" + userID.(string)
		}
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "user rate limit exceeded"})
		}
		return c.Next()
	}
}

// GlobalAPIRateLimit applies to all API routes.
func GlobalAPIRateLimit(limit int, window time.Duration) fiber.Handler {
	limiter := NewMultiKeyRateLimiter(limit, window)
	return func(c fiber.Ctx) error {
		key := "api:" + c.IP()
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "global API rate limit exceeded"})
		}
		return c.Next()
	}
}
