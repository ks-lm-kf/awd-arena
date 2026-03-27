package middleware

import (
	"os"

	"github.com/gofiber/fiber/v3"
)

func CORS() fiber.Handler {
	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")

		// Check if CORS_ORIGINS env var or config is set
		allowedOrigins := os.Getenv("CORS_ORIGINS")
		if allowedOrigins == "" {
			// Default: allow all (set CORS_ORIGINS env var in production)
			allowedOrigins = "*"
		}

		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Set("Access-Control-Allow-Origin", origin)
		} else if allowedOrigins == "*" {
			c.Set("Access-Control-Allow-Origin", "*")
		} else {
			c.Set("Access-Control-Allow-Origin", allowedOrigins)
		}

		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Set("Access-Control-Max-Age", "86400")
		c.Set("Access-Control-Allow-Credentials", "true")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}
		return c.Next()
	}
}

func isOriginAllowed(origin, allowed string) bool {
	if allowed == "*" {
		return true
	}
	// Simple check - in production you'd want proper origin matching
	return origin == allowed
}
