package handler

import (
	"github.com/gofiber/fiber/v3"
)

// SettingsHandler handles system settings requests
var SettingsHandler *settingsHandler

func init() {
	SettingsHandler = &settingsHandler{}
}

type settingsHandler struct{}

// SystemSettings represents system-wide settings
type SystemSettings struct {
	SiteName       string  `json:"site_name"`
	Announcement   string  `json:"announcement"`
	FlagFormat     string  `json:"flag_format"`
	InitialScore   int     `json:"initial_score"`
	AttackWeight   float64 `json:"attack_weight"`
	DefenseWeight  float64 `json:"defense_weight"`
	MaxTeamSize    int     `json:"max_team_size"`
	RoundDuration  int     `json:"round_duration"`
	BreakDuration  int     `json:"break_duration"`
}

// GetSettings returns current system settings
// GET /api/v1/settings
func (h *settingsHandler) GetSettings(c fiber.Ctx) error {
	// For now, return default settings
	// TODO: Load from database or config file
	settings := SystemSettings{
		SiteName:       "AWD Arena",
		Announcement:   "",
		FlagFormat:     "flag{...}",
		InitialScore:   100,
		AttackWeight:   1.0,
		DefenseWeight:  0.5,
		MaxTeamSize:    5,
		RoundDuration:  300,  // 5 minutes
		BreakDuration:  60,   // 1 minute
	}

	return c.JSON(fiber.Map{
		"code": 0,
		"msg":  "success",
		"data": settings,
	})
}

// UpdateSettings updates system settings
// PUT /api/v1/settings
func (h *settingsHandler) UpdateSettings(c fiber.Ctx) error {
	var req SystemSettings
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code": 400,
			"msg":  "Invalid request body: " + err.Error(),
		})
	}

	// TODO: Save to database or config file
	// For now, just validate and return success
	
	// Validate settings
	if req.SiteName == "" {
		req.SiteName = "AWD Arena"
	}
	if req.InitialScore < 0 {
		req.InitialScore = 0
	}
	if req.AttackWeight < 0 {
		req.AttackWeight = 1.0
	}
	if req.DefenseWeight < 0 {
		req.DefenseWeight = 0.5
	}
	if req.MaxTeamSize < 1 {
		req.MaxTeamSize = 5
	}
	if req.RoundDuration < 60 {
		req.RoundDuration = 300
	}
	if req.BreakDuration < 0 {
		req.BreakDuration = 60
	}

	// In a real implementation, you would save to database:
	// db := c.Locals("db").(*gorm.DB)
	// db.Save(&settings)

	return c.JSON(fiber.Map{
		"code": 0,
		"msg":  "Settings updated successfully",
		"data": req,
	})
}
