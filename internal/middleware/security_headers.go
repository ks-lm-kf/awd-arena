package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// SecurityHeaders adds security-related HTTP headers to responses.
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Set("X-Frame-Options", "DENY")
		
		// Enable XSS protection in browsers
		c.Set("X-XSS-Protection", "1; mode=block")
		
		// Enforce HTTPS (adjust max-age as needed)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy (basic policy, adjust as needed)
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
		
		// Permissions Policy (formerly Feature Policy)
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
		
		// Cache control for sensitive pages
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		
		// Remove server information
		c.Set("Server", "")
		
		return c.Next()
	}
}

// CORSSecurityHeaders adds CORS-specific security headers.
func CORSSecurityHeaders(allowOrigins string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", allowOrigins)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400") // 24 hours
		
		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		
		return c.Next()
	}
}
