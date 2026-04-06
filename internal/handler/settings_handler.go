package handler

import (
	"strings"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

var SettingsHandler *settingsHandler

func init() {
	SettingsHandler = &settingsHandler{}
}

type settingsHandler struct{}

// SystemSettings represents system-wide settings
type SystemSettings struct {
	ID            int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteName      string  `json:"site_name"`
	Announcement  string  `json:"announcement"`
	FlagFormat    string  `json:"flag_format"`
	InitialScore  int     `json:"initial_score"`
	AttackWeight  float64 `json:"attack_weight"`
	DefenseWeight float64 `json:"defense_weight"`
	MaxTeamSize   int     `json:"max_team_size"`
	RoundDuration int     `json:"round_duration"`
	BreakDuration int     `json:"break_duration"`
}

// TableName overrides the table name.
func (SystemSettings) TableName() string {
	return "system_settings"
}

// GetSettings returns current system settings
// GET /api/v1/settings
func (h *settingsHandler) GetSettings(c fiber.Ctx) error {
	db := database.GetDB()
	settings := defaultSettings()

	if db != nil {
		db.AutoMigrate(&SystemSettings{})
		var s SystemSettings
		if err := db.First(&s).Error; err == nil {
			settings = &s
		}
	}

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": settings})
}

// UpdateSettings updates system settings
// PUT /api/v1/settings
func (h *settingsHandler) UpdateSettings(c fiber.Ctx) error {
	var req SystemSettings
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid request body"})
	}

	// Validate
	if req.SiteName == "" {
		req.SiteName = "AWD Arena"
	}
	if len(req.SiteName) > 50 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "site_name must be at most 50 characters"})
	}
	if req.AttackWeight <= 0 {
		req.AttackWeight = 1.0
	}
	if req.AttackWeight > 10 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "attack_weight must be at most 10"})
	}
	if req.DefenseWeight < 0 {
		req.DefenseWeight = 0.5
	}
	if req.DefenseWeight > 10 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "defense_weight must be at most 10"})
	}
	if req.MaxTeamSize < 1 {
		req.MaxTeamSize = 5
	}
	if req.MaxTeamSize > 20 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "max_team_size must be at most 20"})
	}
	if req.RoundDuration < 30 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "round_duration must be at least 30 seconds"})
	}
	if req.RoundDuration > 3600 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "round_duration must be at most 3600 seconds (1 hour)"})
	}
	if req.BreakDuration < 0 {
		req.BreakDuration = 60
	}
	if req.BreakDuration > 600 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "break_duration must be at most 600 seconds"})
	}
	if req.FlagFormat == "" {
		req.FlagFormat = "flag{%s}"
	}
	if !strings.Contains(req.FlagFormat, "%s") {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "flag_format must contain %s placeholder"})
	}

	db := database.GetDB()
	if db != nil {
		// Auto migrate to ensure table exists
		db.AutoMigrate(&SystemSettings{})

		var existing SystemSettings
		if err := db.First(&existing).Error; err == gorm.ErrRecordNotFound {
			db.Create(&req)
		} else {
			req.ID = existing.ID
			db.Save(&req)
		}
	}

	return c.JSON(fiber.Map{"code": 0, "message": "Settings updated successfully", "data": req})
}

func defaultSettings() *SystemSettings {
	return &SystemSettings{
		SiteName:      "AWD Arena",
		FlagFormat:    "flag{%s}",
		InitialScore:  100,
		AttackWeight:  1.0,
		DefenseWeight: 0.5,
		MaxTeamSize:   5,
		RoundDuration: 300,
		BreakDuration: 60,
	}
}
