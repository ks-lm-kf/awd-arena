package middleware

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

func localToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case int64:
		return strconv.FormatInt(t, 10)
	case *int64:
		if t != nil {
			return strconv.FormatInt(*t, 10)
		}
		return ""
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

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
	filtered := make([]time.Time, 0, len(times))
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
			if s := localToString(teamID); s != "" {
				key = "team:" + s
			}
		}
		if !limiter.Allow(key) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "rate limit exceeded"})
		}
		return c.Next()
	}
}

// LoginRateLimit limits login attempts per IP and per account (prevent brute force).
// Stricter: 10 attempts per minute per IP and per username.
func LoginRateLimit() fiber.Handler {
	ipLimiter := NewMultiKeyRateLimiter(10, 1*time.Minute)
	accountLimiter := NewMultiKeyRateLimiter(10, 1*time.Minute)
	return func(c fiber.Ctx) error {
		ipKey := "login:" + c.IP()
		if !ipLimiter.Allow(ipKey) {
			return c.Status(429).JSON(fiber.Map{"code": 429, "message": "too many login attempts, try again later"})
		}
		username := c.FormValue("username")
		if username == "" {
			// Try JSON body
			var body struct {
				Username string `json:"username"`
			}
			_ = c.Bind().Body(&body)
			username = body.Username
		}
		if username != "" {
			accountKey := fmt.Sprintf("login:account:%s", username)
			if !accountLimiter.Allow(accountKey) {
				return c.Status(429).JSON(fiber.Map{"code": 429, "message": "too many login attempts for this account, try again later"})
			}
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
			if s := localToString(teamID); s != "" {
				key = "flag:team:" + s
			}
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
			if s := localToString(userID); s != "" {
				key = "user:" + s
			}
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
