package handler

import (
"github.com/awd-platform/awd-arena/internal/model"
"github.com/gofiber/fiber/v3"
"gorm.io/gorm"
)

// DashboardHandler handles dashboard-related requests
var DashboardHandler *dashboardHandler

func init() {
DashboardHandler = &dashboardHandler{}
}

type dashboardHandler struct{}

// GetDashboard returns dashboard statistics
// GET /api/dashboard
func (h *dashboardHandler) GetDashboard(c fiber.Ctx) error {
db := c.Locals("db").(*gorm.DB)

// Get basic statistics
var stats struct {
TotalGames     int64 `json:"total_games"`
ActiveGames    int64 `json:"active_games"`
TotalTeams     int64 `json:"total_teams"`
TotalUsers     int64 `json:"total_users"`
TotalAttacks   int64 `json:"total_attacks"`
TotalFlags     int64 `json:"total_flags"`
}

// Count games
db.Model(&model.Game{}).Count(&stats.TotalGames)
db.Model(&model.Game{}).Where("status = ?", model.GameStatusActive).Count(&stats.ActiveGames)

// Count teams
db.Model(&model.Team{}).Count(&stats.TotalTeams)

// Count users
db.Model(&model.User{}).Count(&stats.TotalUsers)

// Count attacks (from AttackLog if exists)
db.Table("attack_logs").Count(&stats.TotalAttacks)

// Count flag submissions
db.Table("flag_submissions").Where("is_correct = ?", true).Count(&stats.TotalFlags)

return c.JSON(fiber.Map{
"code":    0,
"message": "ok",
"data":    stats,
})
}

// GetRecentActivity returns recent activity for the dashboard
// GET /api/dashboard/activity
func (h *dashboardHandler) GetRecentActivity(c fiber.Ctx) error {
db := c.Locals("db").(*gorm.DB)
limit := 20

// Get recent flag submissions
var flagSubmissions []model.FlagSubmission
db.Where("is_correct = ?", true).
Order("submitted_at DESC").
Limit(limit).
Find(&flagSubmissions)

// Get recent admin logs
var adminLogs []model.AdminLog
db.Order("created_at DESC").
Limit(limit).
Find(&adminLogs)

return c.JSON(fiber.Map{
"code":    0,
"message": "ok",
"data": fiber.Map{
"flag_submissions": flagSubmissions,
"admin_logs":       adminLogs,
},
})
}
