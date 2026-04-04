package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/engine"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/awd-platform/awd-arena/pkg/crypto"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

var AdminHandler *adminHandler

func init() {
	AdminHandler = &adminHandler{
		gameSvc:      &service.GameService{},
		teamSvc:      &service.TeamService{},
		containerSvc: service.NewContainerService(),
	}
}

type adminHandler struct {
	gameSvc      *service.GameService
	teamSvc      *service.TeamService
	containerSvc *service.ContainerService
}

// logAdminAction logs an administrative action
func (h *adminHandler) logAdminAction(c fiber.Ctx, action, resourceType string, resourceID int64, description string, details interface{}) {
	userID, _ := c.Locals("user_id").(int64)
	username, _ := c.Locals("username").(string)

	detailsJSON := ""
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			detailsJSON = string(b)
		}
	}

	log := &model.AdminLog{
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Description:  description,
		IPAddress:    c.IP(),
		UserAgent:    string(c.Request().Header.UserAgent()),
		Details:      detailsJSON,
		CreatedAt:    time.Now(),
	}

	db := database.GetDB()
	db.Create(log)
}

// GetAdminLogs returns all admin action logs
func (h *adminHandler) GetAdminLogs(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))
	action := c.Query("action")
	resourceType := c.Query("resource_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	db := database.GetDB()
	var logs []model.AdminLog
	var total int64

	query := db.Model(&model.AdminLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": fiber.Map{
			"items":     logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// --- Game Management (Enhanced with Logging) ---

func (h *adminHandler) CreateGame(c fiber.Ctx) error {
	var req createGameRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "title is required"})
	}

	userID, _ := c.Locals("user_id").(int64)
	game := &model.Game{
		Title:         req.Title,
		Description:   req.Description,
		Mode:          req.Mode,
		Status:        "draft",
		RoundDuration: req.RoundDuration,
		BreakDuration: req.BreakDuration,
		TotalRounds:   req.TotalRounds,
		FlagFormat:    req.FlagFormat,
		AttackWeight:  req.AttackWeight,
		DefenseWeight: req.DefenseWeight,
		CreatedBy:     userID,
	}

	if game.Mode == "" {
		game.Mode = "awd_score"
	}
	if game.RoundDuration == 0 {
		game.RoundDuration = 300
	}
	if game.BreakDuration == 0 {
		game.BreakDuration = 120
	}
	if game.TotalRounds == 0 {
		game.TotalRounds = 20
	}
	if game.FlagFormat == "" {
		game.FlagFormat = "flag{%s}"
	}
	if game.AttackWeight == 0 {
		game.AttackWeight = 1.0
	}
	if game.DefenseWeight == 0 {
		game.DefenseWeight = 0.5
	}

	if err := h.gameSvc.CreateGame(c.Context(), game); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "create", "game", game.ID, "Created game: "+game.Title, req)

	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": game})
}

func (h *adminHandler) UpdateGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	var req createGameRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	oldTitle := game.Title
	if req.Title != "" {
		game.Title = html.EscapeString(req.Title)
	}
	if req.Description != "" {
		game.Description = html.EscapeString(req.Description)
	}
	if req.Mode != "" {
		game.Mode = req.Mode
	}
	if req.RoundDuration > 0 {
		game.RoundDuration = req.RoundDuration
	}
	if req.BreakDuration > 0 {
		game.BreakDuration = req.BreakDuration
	}
	if req.TotalRounds > 0 {
		game.TotalRounds = req.TotalRounds
	}
	if req.FlagFormat != "" {
		game.FlagFormat = req.FlagFormat
	}
	if req.AttackWeight > 0 {
		game.AttackWeight = req.AttackWeight
	}
	if req.DefenseWeight > 0 {
		game.DefenseWeight = req.DefenseWeight
	}

	if err := h.gameSvc.UpdateGame(c.Context(), game); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "update", "game", game.ID, "Updated game: "+oldTitle, req)

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": game})
}

func (h *adminHandler) DeleteGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	db := database.GetDB()
	if err := db.Delete(&model.Game{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "delete", "game", id, "Deleted game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

func (h *adminHandler) StartGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	if err := h.gameSvc.StartGame(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "start", "game", id, "Started game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "game started"})
}

func (h *adminHandler) PauseGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	if err := h.gameSvc.PauseGame(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "pause", "game", id, "Paused game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "game paused"})
}

func (h *adminHandler) ResumeGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	if err := h.gameSvc.ResumeGame(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "resume", "game", id, "Resumed game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "game resumed"})
}

func (h *adminHandler) StopGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	if err := h.gameSvc.StopGame(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "stop", "game", id, "Stopped game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "game stopped"})
}

func (h *adminHandler) ResetGame(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	game, err := h.gameSvc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	if err := h.gameSvc.ResetGame(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "reset", "game", id, "Reset game: "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "game reset"})
}

// --- Team Management (Enhanced with Batch Import) ---

type adminTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
	Token       string `json:"token"`
}

func (h *adminHandler) CreateTeam(c fiber.Ctx) error {
	var req adminTeamRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "team name is required"})
	}

	team, err := h.teamSvc.CreateTeam(c.Context(), req.Name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	if req.Description != "" {
		team.Description = req.Description
	}
	if req.AvatarURL != "" {
		team.AvatarURL = req.AvatarURL
	}
	if req.Token != "" {
		team.Token = crypto.SHA256Hex(req.Token)
	}

	// Log the action
	h.logAdminAction(c, "create", "team", team.ID, "Created team: "+team.Name, req)

	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": team})
}

func (h *adminHandler) UpdateTeam(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	team, err := h.teamSvc.GetTeam(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	var req adminTeamRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	oldName := team.Name
	if req.Name != "" {
		team.Name = req.Name
	}
	if req.Description != "" {
		team.Description = req.Description
	}
	if req.AvatarURL != "" {
		team.AvatarURL = req.AvatarURL
	}
	if req.Token != "" {
		team.Token = crypto.SHA256Hex(req.Token)
	}

	db := database.GetDB()
	if err := db.Save(team).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "update", "team", team.ID, "Updated team: "+oldName, req)

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": team})
}

func (h *adminHandler) DeleteTeam(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	team, err := h.teamSvc.GetTeam(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	// Clean up Docker containers for this team
	if err := h.containerSvc.CleanupTeamContainers(c.Context(), id); err != nil {
		logger.Error("failed to cleanup containers for team", "team_id", id, "error", err)
	}

	teamID := id
	db := database.GetDB()
	// Remove from game associations
	db.Where("team_id = ?", teamID).Delete(&model.GameTeam{})
	// Nullify user references
	db.Model(&model.User{}).Where("team_id = ?", teamID).Update("team_id", nil)
	// Remove flag records
	db.Where("team_id = ?", teamID).Delete(&model.FlagRecord{})
	// Remove flag submissions
	db.Where("attacker_team = ? OR target_team = ?", teamID, teamID).Delete(&model.FlagSubmission{})
	// Then delete the team
	if err := db.Delete(&model.Team{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "delete", "team", id, "Deleted team: "+team.Name, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

type batchImportTeamsRequest struct {
	Teams []adminTeamRequest `json:"teams"`
}

func (h *adminHandler) BatchImportTeams(c fiber.Ctx) error {
	var req batchImportTeamsRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if len(req.Teams) == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "no teams to import"})
	}
	if len(req.Teams) > 100 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "maximum 100 teams per batch"})
	}

	var imported []model.Team
	var errors []string

	for i, teamReq := range req.Teams {
		if teamReq.Name == "" {
			errors = append(errors, "Row "+strconv.Itoa(i+1)+": team name is required")
			continue
		}

		team, err := h.teamSvc.CreateTeam(c.Context(), teamReq.Name)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(i+1)+": "+err.Error())
			continue
		}

		if teamReq.Description != "" {
			team.Description = teamReq.Description
		}
		if teamReq.AvatarURL != "" {
			team.AvatarURL = teamReq.AvatarURL
		}
		if teamReq.Token != "" {
			team.Token = crypto.SHA256Hex(teamReq.Token)
		}

		db := database.GetDB()
		if err := db.Save(team).Error; err != nil {
			errors = append(errors, fmt.Sprintf("failed to update team %s: %v", team.Name, err))
			continue
		}

		imported = append(imported, *team)
	}

	// Log the action
	h.logAdminAction(c, "import", "team", 0, "Batch imported teams", fiber.Map{
		"total":   len(req.Teams),
		"success": len(imported),
		"errors":  len(errors),
	})

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": fiber.Map{
			"imported": imported,
			"errors":   errors,
			"total":    len(req.Teams),
			"success":  len(imported),
		},
	})
}

func (h *adminHandler) ResetTeam(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	team, err := h.teamSvc.GetTeam(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	// Reset team score to 0
	team.Score = 0
	db := database.GetDB()
	if err := db.Save(team).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "reset", "team", team.ID, "Reset team: "+team.Name, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": team})
}

// --- Score Adjustment ---

type adjustScoreRequest struct {
	GameID     int64   `json:"game_id"`
	TeamID     int64   `json:"team_id"`
	Amount     float64 `json:"amount"`
	Adjustment float64 `json:"adjustment"`
	Reason     string  `json:"reason"`
}

func (h *adminHandler) AdjustScore(c fiber.Ctx) error {
	var req adjustScoreRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Amount == 0 && req.Adjustment != 0 {
		req.Amount = req.Adjustment
	}
	if req.Amount == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "amount or adjustment is required and must be non-zero"})
	}
	if req.GameID == 0 || req.TeamID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "game_id and team_id are required"})
	}
	if req.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "reason is required for score adjustment"})
	}

	db := database.GetDB()

	// Verify game and team exist
	var game model.Game
	if err := db.First(&game, req.GameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	var team model.Team
	if err := db.First(&team, req.TeamID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	oldScore := team.Score

	adjustment := model.ScoreAdjustment{
		GameID:      req.GameID,
		TeamID:      req.TeamID,
		AdjustValue: int(req.Amount),
		Reason:      req.Reason,
	}
	if err := db.Create(&adjustment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to create score adjustment"})
	}

	// Immediately recalculate cumulative scores
	sc := engine.NewScoreCalculator(&game)
	if err := sc.UpdateCumulativeTeamScores(c.Context()); err != nil {
		logger.Error("failed to update cumulative scores after adjustment", "error", err)
	}

	// Log the action
	h.logAdminAction(c, "adjust_score", "score", 0, "Adjusted score for team: "+team.Name, fiber.Map{
		"game_id": req.GameID,
		"team_id": req.TeamID,
		"amount":  req.Amount,
		"reason":  req.Reason,
	})

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": fiber.Map{
			"team_id":    team.ID,
			"team_name":  team.Name,
			"old_score":  oldScore,
			"adjustment": req.Amount,
			"new_score":  oldScore + req.Amount,
		},
	})
}

// AddTeamToGame adds a team to a specific game
func (h *adminHandler) AddTeamToGame(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	var req struct {
		TeamID int64 `json:"team_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	db := database.GetDB()

	// Verify game and team exist
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	var team model.Team
	if err := db.First(&team, req.TeamID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	// Check if already added
	var existing model.GameTeam
	if err := db.Where("game_id = ? AND team_id = ?", gameID, req.TeamID).First(&existing).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "team already in game"})
	}

	// Add team to game
	gameTeam := &model.GameTeam{
		GameID: gameID,
		TeamID: req.TeamID,
	}
	if err := db.Create(gameTeam).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Reload with Team association
	db.Preload("Team").First(gameTeam, gameTeam.ID)

	// Log the action
	h.logAdminAction(c, "add_team", "game", gameID, "Added team "+team.Name+" to game "+game.Title, req)

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": gameTeam})
}

// RemoveTeamFromGame removes a team from a specific game
func (h *adminHandler) RemoveTeamFromGame(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	teamID, err := parseID(c.Params("team_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	db := database.GetDB()

	// Verify game and team exist
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	var team model.Team
	if err := db.First(&team, teamID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}

	// Remove team from game
	if err := db.Where("game_id = ? AND team_id = ?", gameID, teamID).Delete(&model.GameTeam{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Log the action
	h.logAdminAction(c, "remove_team", "game", gameID, "Removed team "+team.Name+" from game "+game.Title, nil)

	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

// GetGameTeams gets the list of teams for a specific game
func (h *adminHandler) GetGameTeams(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	db := database.GetDB()

	// Verify game exists
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}

	// Get teams for the game
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Preload("Team").Find(&gameTeams).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "internal server error"})
	}

	// Transform to response format with member count
	type TeamResponse struct {
		ID          int64   `json:"id"`
		GameTeamID  int64   `json:"game_team_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		AvatarURL   string  `json:"avatar_url"`
		Score       float64 `json:"score"`
		MemberCount int64   `json:"member_count"`
	}

	response := make([]TeamResponse, 0)
	for _, gt := range gameTeams {
		// Skip if team doesn't exist (ID is 0 means preload failed)
		if gt.Team.ID == 0 {
			// Delete invalid game_team record
			db.Delete(&model.GameTeam{}, gt.ID)
			continue
		}

		// Count team members
		var memberCount int64
		db.Model(&model.User{}).Where("team_id = ?", gt.TeamID).Count(&memberCount)

		response = append(response, TeamResponse{
			ID:          gt.TeamID,
			GameTeamID:  gt.ID,
			Name:        gt.Team.Name,
			Description: gt.Team.Description,
			AvatarURL:   gt.Team.AvatarURL,
			Score:       gt.Team.Score,
			MemberCount: memberCount,
		})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": response})
}
