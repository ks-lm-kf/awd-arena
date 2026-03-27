package middleware

import (
	"github.com/awd-platform/awd-arena/internal/security"
	"github.com/gofiber/fiber/v3"
)

// GlobalWAF is the global WAF middleware instance.
var GlobalWAF *security.WAFEngine

// InitWAF initializes the global WAF engine.
func InitWAF(logStore *security.AttackLogStore) {
	GlobalWAF = security.NewWAFEngine(logStore)
}

// WAFMiddleware returns a Fiber middleware that checks requests against WAF rules.
func WAFMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if GlobalWAF == nil {
			return c.Next()
		}

		// Collect all checkable content
		path := c.Path()
		query := c.Context().QueryArgs().String()
		body := string(c.Body())

		result := GlobalWAF.CheckRequest(query, body, path)

		if result.Blocked {
			// Log the attack
			if GlobalWAF != nil {
				// We'll use the log store directly via the security handler
				// Store result in context for logging middleware
				c.Locals("waf_result", result)
				c.Locals("waf_blocked", true)
			}

			return c.Status(403).JSON(fiber.Map{
				"code":    403,
				"message": "request blocked by WAF",
				"detail":  result.Reason,
				"rule":    result.Rule,
				"severity": result.Severity,
			})
		}

		return c.Next()
	}
}
