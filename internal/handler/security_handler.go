package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/security"
	"github.com/gofiber/fiber/v3"
)

var (
	AlertManager = security.NewAlertManager(1000)
)

// GetGameAlerts returns security alerts for a specific game.
func GetGameAlerts(c fiber.Ctx) error {
	gameID := c.Params("id")
	alerts := AlertManager.GetAlertsByGame(gameID)
	return c.JSON(fiber.Map{"code": 0, "data": alerts})
}

// GetGameAttacks returns attack logs for a specific game.
func GetGameAttacks(c fiber.Ctx) error {
	gameID := parseGameID(c)
	if gameID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid game id"})
	}

	// Try to get from event log (SQLite)
	db := database.GetDB()
	if db != nil {
		var events []model.EventLog
		query := db.Where("event_type IN ?", []string{"attack", "flag_submit"}).Order("created_at desc")
		if gameID > 0 {
			query = query.Where("game_id = ?", gameID)
		}
		query = query.Limit(100)
		if err := query.Find(&events).Error; err == nil {
			return c.JSON(fiber.Map{"code": 0, "data": events})
		}
	}

	// Fallback to in-memory attack store
	if AttackStore != nil {
		attackLogs := AttackStore.GetRecent(100)
		return c.JSON(fiber.Map{"code": 0, "data": attackLogs})
	}

	return c.JSON(fiber.Map{"code": 0, "data": []interface{}{}})
}

// GetAttackStats returns aggregated attack statistics for a game.
func GetAttackStats(c fiber.Ctx) error {
	gameID := parseGameID(c)
	db := database.GetDB()
	if db == nil {
		return c.JSON(fiber.Map{"code": 0, "data": fiber.Map{}})
	}

	// Count flag submissions by correctness
	var correct, incorrect int64
	db.Model(&model.FlagSubmission{}).Where("game_id = ? AND is_correct = ?", gameID, true).Count(&correct)
	db.Model(&model.FlagSubmission{}).Where("game_id = ? AND is_correct = ?", gameID, false).Count(&incorrect)

	// Count by round
	type RoundCount struct {
		Round int   `json:"round"`
		Count int64 `json:"count"`
	}
	var roundCounts []RoundCount
	db.Model(&model.FlagSubmission{}).
		Select("round, count(*) as count").
		Where("game_id = ? AND is_correct = ?", gameID, true).
		Group("round").
		Order("round").
		Find(&roundCounts)

	return c.JSON(fiber.Map{
		"code": 0,
		"data": fiber.Map{
			"total_attacks":    correct + incorrect,
			"successful":       correct,
			"failed":           incorrect,
			"success_rate":     floatRatio(correct, correct+incorrect),
			"attacks_by_round": roundCounts,
		},
	})
}

func floatRatio(a, b int64) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func parseGameID(c fiber.Ctx) int64 {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	return id
}
