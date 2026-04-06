package handler

import (
	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

var DashboardHandler *dashboardHandler

func init() {
	DashboardHandler = &dashboardHandler{}
}

type dashboardHandler struct{}

// GetDashboard returns dashboard statistics
// GET /api/v1/dashboard
func (h *dashboardHandler) GetDashboard(c fiber.Ctx) error {
	db := database.GetDB()
	if db == nil {
		return c.JSON(fiber.Map{"code": 0, "data": fiber.Map{}})
	}

	var stats struct {
		TotalGames   int64 `json:"total_games"`
		ActiveGames  int64 `json:"active_games"`
		TotalTeams   int64 `json:"total_teams"`
		TotalUsers   int64 `json:"total_users"`
		TotalAttacks int64 `json:"total_attacks"`
		TotalFlags   int64 `json:"total_flags"`
	}

	db.Model(&model.Game{}).Count(&stats.TotalGames)
	db.Model(&model.Game{}).Where("status IN ?", []string{"running", "paused"}).Count(&stats.ActiveGames)
	db.Model(&model.Team{}).Count(&stats.TotalTeams)
	db.Model(&model.User{}).Count(&stats.TotalUsers)
	db.Table("flag_submissions").Where("is_correct = ?", true).Count(&stats.TotalFlags)

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": stats})
}

// GetRecentActivity returns recent activity for the dashboard
// GET /api/v1/dashboard/activity
func (h *dashboardHandler) GetRecentActivity(c fiber.Ctx) error {
	db := database.GetDB()
	if db == nil {
		return c.JSON(fiber.Map{"code": 0, "data": fiber.Map{}})
	}

	limit := 20

	var flagSubmissions []model.FlagSubmission
	db.Where("is_correct = ?", true).Order("submitted_at DESC").Limit(limit).Find(&flagSubmissions)

	var adminLogs []model.AdminLog
	db.Order("created_at DESC").Limit(limit).Find(&adminLogs)

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": fiber.Map{
			"flag_submissions": flagSubmissions,
			"admin_logs":       adminLogs,
		},
	})
}
