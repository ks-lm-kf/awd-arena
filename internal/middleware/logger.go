package middleware

import (
	"time"

	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		logger.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency", latency.String(),
			"ip", c.IP(),
		)
		return err
	}
}
