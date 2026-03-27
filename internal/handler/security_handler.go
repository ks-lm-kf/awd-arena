package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/security"
	"github.com/gofiber/fiber/v3"
)

var (
	AlertManager *security.AlertManager
)

func init() {
	AlertManager = security.NewAlertManager(1000)
}

// GetGameAlerts returns security alerts for a specific game.
func GetGameAlerts(c fiber.Ctx) error {
	gameID := c.Params("id")
	alerts := AlertManager.GetAlertsByGame(gameID)
	return c.JSON(fiber.Map{"code": 0, "data": alerts})
}

// GetGameAttacks returns attack logs for a specific game.
// Currently returns empty array as attack logs are stored in ClickHouse (not yet integrated).
func GetGameAttacks(c fiber.Ctx) error {
	// Placeholder: attack logs will be fetched from ClickHouse in the future
	// For now return empty to avoid 404
	gameID := c.Params("id")
	_ = gameID // suppress unused warning
	return c.JSON(fiber.Map{"code": 0, "data": []struct{}{}})
}

func parseGameID(c fiber.Ctx) int64 {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	return id
}
