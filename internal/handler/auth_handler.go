package handler

import (
	"context"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/middleware"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

var (
	AuthHandler            *authHandler
	GameHandler            *gameHandler
	TeamHandler            *teamHandler
	FlagHandler            *flagHandler
	FlagRefreshHandlerObj  *FlagRefreshHandler
	ContainerHandler       *containerHandler
	RankingHandler         *rankingHandler
	ChallengeHandler       *challengeHandler
	UserHandler            *userHandler
	DockerImageHandlerObj  *DockerImageHandler
	RoundHandler          *roundHandler
	TemplateHandlerObj      *TemplateHandler
)

func init() {
	AuthHandler = &authHandler{svc: &service.AuthService{}}
	GameHandler = &gameHandler{svc: &service.GameService{}}
	TeamHandler = &teamHandler{svc: &service.TeamService{}}
	FlagHandler = &flagHandler{svc: &service.FlagService{}}
	ContainerHandler = &containerHandler{svc: &service.ContainerService{}}
	RankingHandler = &rankingHandler{svc: &service.RankingService{}}
	ChallengeHandler = &challengeHandler{svc: service.NewChallengeService()}
	UserHandler = &userHandler{svc: &service.AuthService{}}
	DockerImageHandlerObj = NewDockerImageHandler()
	FlagRefreshHandlerObj = NewFlagRefreshHandler(nil) // Docker client will be set later if needed
	RoundHandler = NewRoundHandler()
	TemplateHandlerObj = NewTemplateHandler()
}

func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// containsHTML checks if a string contains HTML tags
func containsHTML(s string) bool {
	// Match common HTML tags
	htmlTagRegex := regexp.MustCompile(`<(script|iframe|object|embed|link|meta|style|img|video|audio|source|track|canvas|svg|math|form|input|button|select|textarea|label|div|span|p|a|br|hr|table|tr|td|th|thead|tbody|ul|ol|li|h[1-6])[^>]*>`)
	return htmlTagRegex.MatchString(strings.ToLower(s))
}

// sanitizeInput sanitizes user input by escaping HTML special characters
func sanitizeInput(s string) string {
	return html.EscapeString(s)
}

// validateAndSanitizeInput validates input doesn't contain HTML and sanitizes it
func validateAndSanitizeInput(input string, fieldName string) (string, error) {
	if containsHTML(input) {
		return "", fmt.Errorf("%s contains invalid characters (HTML tags not allowed)", fieldName)
	}
	// Additional sanitization for extra safety
	return sanitizeInput(input), nil
}

// --- Auth ---

type authHandler struct {
	svc *service.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *authHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "username and password required"})
	}
	
	// Validate username doesn't contain HTML tags
	if containsHTML(req.Username) {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "username contains invalid characters"})
	}
	
	// Create context with timeout for login operation (10 seconds)
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()
	
	token, userInfo, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return c.Status(504).JSON(fiber.Map{"code": 504, "message": "login request timeout"})
		}
		return c.Status(401).JSON(fiber.Map{"code": 401, "message": "invalid credentials"})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": fiber.Map{"token": token, "user": userInfo}})
}

func (h *authHandler) Logout(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

// ChangePassword handles password change requests
func (h *authHandler) ChangePassword(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"code": 401, "message": "unauthorized"})
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "old_password and new_password required"})
	}

	if err := h.svc.ChangePassword(c.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "password changed successfully"})
}


// Register handles user registration with XSS protection

// validateUsername validates the username format to prevent XSS
func validateUsername(username string) error {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\x{4e00}-\x{9fa5}]{2,50}$`, username)
	if !matched {
		return errors.New("用户名只能包含字母、数字、下划线、中文 (2-50字符)")
	}
	return nil
}

func (h *authHandler) Register(c fiber.Ctx) error {
	var req struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		TeamToken string `json:"team_token"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "username and password required"})
	}

	// Validate username format
	if err := validateUsername(req.Username); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	// XSS Protection: Validate and sanitize username
	sanitizedUsername, err := validateAndSanitizeInput(req.Username, "username")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	// XSS Protection: Validate and sanitize email if provided
	var sanitizedEmail string
	if req.Email != "" {
		sanitizedEmail, err = validateAndSanitizeInput(req.Email, "email")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}

	// XSS Protection: Validate role if provided
	var sanitizedRole string
	if req.Role != "" {
		sanitizedRole, err = validateAndSanitizeInput(req.Role, "role")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}

	// Default role is player if not specified
	role := sanitizedRole
	if role == "" {
		role = "player"
	}

	// Use RegisterWithTokenAndRole for full registration with sanitized username
	token, userInfo, err := h.svc.RegisterWithTokenAndRole(c.Context(), sanitizedUsername, req.Password, req.TeamToken, role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	// Update email if provided (using sanitized email)
	if sanitizedEmail != "" {
		if err := h.svc.UpdateUser(c.Context(), userInfo.ID, nil, &sanitizedEmail, nil, nil); err != nil {
			logger.Warn("failed to update user email", "user_id", userInfo.ID, "error", err)
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"code":    0,
		"message": "registration successful",
		"data": fiber.Map{
			"token": token,
			"user":  userInfo,
		},
	})
}

// RefreshToken handles token refresh
func (h *authHandler) RefreshToken(c fiber.Ctx) error {
	// Get user info from current token (middleware already validated it)
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"code": 401, "message": "unauthorized"})
	}

	// Get current user info
	userInfo, err := h.svc.GetUser(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "user not found"})
	}

	// Generate new token
	token, err := middleware.GenerateToken(userID, userInfo.Username, userInfo.Role, userInfo.TeamID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "token refreshed",
		"data": fiber.Map{
			"token": token,
			"user":  userInfo,
		},
	})
}


func (h *authHandler) Me(c fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	userID, ok := userIDRaw.(int64)
	if !ok {
		uid, err := middleware.ValidateToken(c, middleware.GetJWTSecret())
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"code": 401, "message": "unauthorized"})
		}
		userID = uid
	}
	userInfo, err := h.svc.GetUser(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "user not found"})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": userInfo})
}

// --- Game ---

type gameHandler struct {
	svc *service.GameService
}

func (h *gameHandler) List(c fiber.Ctx) error {
	games, err := h.svc.ListGames(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": games})
}

func (h *gameHandler) Get(c fiber.Ctx) error {
	game, err := h.svc.GetGame(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": game})
}

type createGameRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Mode           string  `json:"mode"`
	RoundDuration  int     `json:"round_duration"`
	BreakDuration  int     `json:"break_duration"`
	TotalRounds    int     `json:"total_rounds"`
	FlagFormat     string  `json:"flag_format"`
	AttackWeight   float64 `json:"attack_weight"`
	DefenseWeight  float64 `json:"defense_weight"`
}

func (h *gameHandler) Create(c fiber.Ctx) error {
	// Permission check - only admin and organizer can create games
	roleStr, _ := c.Locals("role").(string)
	role := model.Role(roleStr)
	if role != model.RoleAdmin && role != model.RoleOrganizer {
		return c.Status(403).JSON(fiber.Map{"code": 403, "message": "permission denied: only admin and organizer can create games"})
	}

	var req createGameRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "title is required"})
	}
	
	// XSS Protection: Validate and sanitize title
	sanitizedTitle, err := validateAndSanitizeInput(req.Title, "title")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	
	// XSS Protection: Sanitize description
	var sanitizedDesc string
	if req.Description != "" {
		sanitizedDesc, err = validateAndSanitizeInput(req.Description, "description")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}
	
	userID, _ := c.Locals("user_id").(int64)

	game := &model.Game{
		Title:         sanitizedTitle,
		Description:   sanitizedDesc,
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

	if err := h.svc.CreateGame(c.Context(), game); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": game})
}

func (h *gameHandler) Update(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	game, err := h.svc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}
	var req createGameRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Title != "" {
		// XSS Protection: Validate and sanitize title
		sanitizedTitle, err := validateAndSanitizeInput(req.Title, "title")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
		game.Title = sanitizedTitle
	}
	if req.Description != "" {
		// XSS Protection: Sanitize description
		sanitizedDesc, err := validateAndSanitizeInput(req.Description, "description")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
		game.Description = sanitizedDesc
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
	if err := h.svc.UpdateGame(c.Context(), game); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": game})
}

func (h *gameHandler) Start(c fiber.Ctx) error {
	if err := h.svc.StartGame(c.Context(), parseID(c.Params("id"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "game started"})
}

func (h *gameHandler) Pause(c fiber.Ctx) error {
	if err := h.svc.PauseGame(c.Context(), parseID(c.Params("id"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "game paused"})
}

func (h *gameHandler) Stop(c fiber.Ctx) error {
	if err := h.svc.StopGame(c.Context(), parseID(c.Params("id"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "game stopped"})
}

func (h *gameHandler) Reset(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	if id == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid game id"})
	}
	
	// Check if game exists
	game, err := h.svc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}
	
	// Check if game can be reset (should not be running)
	if game.Status == "running" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "cannot reset a running game, please stop it first"})
	}
	
	if err := h.svc.ResetGame(c.Context(), id); err != nil {
		logger.Error("failed to reset game", "game_id", id, "error", err)
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": fmt.Sprintf("reset failed: %s", err.Error())})
	}
	
	logger.Info("game reset successfully", "game_id", id)
	return c.JSON(fiber.Map{"code": 0, "message": "game reset successfully"})
}

func (h *gameHandler) Delete(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	if id == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}
	
	// 检查游戏是否存在
	_, err := h.svc.GetGame(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}
	
	// 删除游戏
	db := database.GetDB()
	if err := db.Delete(&model.Game{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	
	return c.JSON(fiber.Map{"code": 0, "message": "deleted"})
}

// --- Team ---

type teamHandler struct {
	svc *service.TeamService
}

func (h *teamHandler) List(c fiber.Ctx) error {
	teams, err := h.svc.ListTeams(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": teams})
}

func (h *teamHandler) Get(c fiber.Ctx) error {
	team, err := h.svc.GetTeam(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "team not found"})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": team})
}

type createTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

func (h *teamHandler) Create(c fiber.Ctx) error {
	var req createTeamRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "team name is required"})
	}
	
	// XSS Protection: Validate and sanitize team name
	sanitizedName, err := validateAndSanitizeInput(req.Name, "team name")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	
	team, err := h.svc.CreateTeam(c.Context(), sanitizedName)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	if req.Description != "" {
		// XSS Protection: Sanitize description
		sanitizedDesc, err := validateAndSanitizeInput(req.Description, "description")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
		team.Description = sanitizedDesc
	}
	if req.AvatarURL != "" {
		team.AvatarURL = req.AvatarURL
	}
	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": team})
}

func (h *teamHandler) Members(c fiber.Ctx) error {
	members, err := h.svc.GetTeamMembers(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": members})
}

// --- Flag ---

type flagHandler struct {
	svc *service.FlagService
}

type submitFlagRequest struct {
	FlagValue string `json:"flag_value"`
}

func (h *flagHandler) Submit(c fiber.Ctx) error {
	var req submitFlagRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	gameID := parseID(c.Params("id"))

	// Get team_id from JWT claims
	teamIDVal := c.Locals("team_id")
	var attackerTeamID int64
	if teamIDPtr, ok := teamIDVal.(*int64); ok && teamIDPtr != nil {
		attackerTeamID = *teamIDPtr
	}

	// Get current round from game
	gameSvc := &service.GameService{}
	game, err := gameSvc.GetGame(c.Context(), gameID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
	}
	round := int64(game.CurrentRound)

	correct, points, err := h.svc.SubmitFlag(c.Context(), gameID, round, attackerTeamID, req.FlagValue)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	// Broadcast leaderboard update if flag was correct
	if correct {
		db := database.GetDB()
		go func() {
			if err := LeaderboardHandler.BroadcastUpdate(gameID, db); err != nil {
				logger.Error("failed to broadcast leaderboard update", "game_id", gameID, "error", err)
			}
		}()
	}

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": fiber.Map{"correct": correct, "points": points}})
}

func (h *flagHandler) History(c fiber.Ctx) error {
	roundStr := c.Query("round", "0")
	round, _ := strconv.Atoi(roundStr)
	history, err := h.svc.GetFlagHistory(c.Context(), parseID(c.Params("id")), round)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": history})
}

// --- Container ---

type containerHandler struct {
	svc *service.ContainerService
}

func (h *containerHandler) List(c fiber.Ctx) error {
	containers, err := h.svc.GetContainers(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": containers})
}

func (h *containerHandler) Restart(c fiber.Ctx) error {
	if err := h.svc.RestartContainer(c.Context(), parseID(c.Params("id")), parseID(c.Params("cid"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

func (h *containerHandler) BulkRestart(c fiber.Ctx) error {
	if err := h.svc.RestartAll(c.Context(), parseID(c.Params("id"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

func (h *containerHandler) Stats(c fiber.Ctx) error {
	stats, err := h.svc.GetStats(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": stats})
}

// --- Ranking ---

type rankingHandler struct {
	svc *service.RankingService
}

func (h *rankingHandler) Get(c fiber.Ctx) error {
	rankings, err := h.svc.GetRankings(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": rankings})
}

func (h *rankingHandler) GetRound(c fiber.Ctx) error {
	round, _ := strconv.Atoi(c.Params("round"))
	rankings, err := h.svc.GetRoundRankings(c.Context(), parseID(c.Params("id")), round)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": rankings})
}

// --- Challenge ---

type challengeHandler struct {
	svc *service.ChallengeService
}

func (h *challengeHandler) List(c fiber.Ctx) error {
	challenges, err := h.svc.ListChallenges(c.Context(), parseID(c.Params("id")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": challenges})
}

func (h *challengeHandler) Create(c fiber.Ctx) error {
	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		ImageName    string  `json:"image_name"`
		ImageTag     string  `json:"image_tag"`
		Difficulty   string  `json:"difficulty"`
		BaseScore    int     `json:"base_score"`
		ExposedPorts string  `json:"exposed_ports"`
		CPULimit     float64 `json:"cpu_limit"`
		MemLimit     int     `json:"mem_limit"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Name == "" || req.ImageName == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "name and image_name are required"})
	}
	
	// XSS Protection: Validate and sanitize challenge name
	sanitizedName, err := validateAndSanitizeInput(req.Name, "challenge name")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	
	// XSS Protection: Sanitize description
	var sanitizedDesc string
	if req.Description != "" {
		sanitizedDesc, err = validateAndSanitizeInput(req.Description, "description")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}
	
	ch, err := h.svc.CreateChallenge(c.Context(), &model.Challenge{
		GameID:       parseID(c.Params("id")),
		Name:         sanitizedName,
		Description:  sanitizedDesc,
		ImageName:    req.ImageName,
		ImageTag:     req.ImageTag,
		Difficulty:   req.Difficulty,
		BaseScore:    req.BaseScore,
		ExposedPorts: req.ExposedPorts,
		CPULimit:     req.CPULimit,
		MemLimit:     req.MemLimit,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": ch})
}

// --- User ---

func (h *challengeHandler) Update(c fiber.Ctx) error {
	challengeID := parseID(c.Params("challengeId"))
	if challengeID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid challenge id"})
	}

	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		ImageName    string  `json:"image_name"`
		ImageTag     string  `json:"image_tag"`
		Difficulty   string  `json:"difficulty"`
		BaseScore    int     `json:"base_score"`
		ExposedPorts string  `json:"exposed_ports"`
		CPULimit     float64 `json:"cpu_limit"`
		MemLimit     int     `json:"mem_limit"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}

	// XSS Protection
	sanitizedName, err := validateAndSanitizeInput(req.Name, "challenge name")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}

	var sanitizedDesc string
	if req.Description != "" {
		sanitizedDesc, err = validateAndSanitizeInput(req.Description, "description")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}

	ch, err := h.svc.UpdateChallenge(c.Context(), challengeID, &model.Challenge{
		Name:         sanitizedName,
		Description:  sanitizedDesc,
		ImageName:    req.ImageName,
		ImageTag:     req.ImageTag,
		Difficulty:   req.Difficulty,
		BaseScore:    req.BaseScore,
		ExposedPorts: req.ExposedPorts,
		CPULimit:     req.CPULimit,
		MemLimit:     req.MemLimit,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": ch})
}

func (h *challengeHandler) Delete(c fiber.Ctx) error {
	challengeID := parseID(c.Params("challengeId"))
	if challengeID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid challenge id"})
	}

	if err := h.svc.DeleteChallenge(c.Context(), challengeID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}


type userHandler struct {
	svc *service.AuthService
}

func (h *userHandler) List(c fiber.Ctx) error {
	users, err := h.svc.ListUsers(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": users})
}

func (h *userHandler) Create(c fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		TeamID   *int64 `json:"team_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "username and password required"})
	}
	
	// XSS Protection: Validate and sanitize username
	sanitizedUsername, err := validateAndSanitizeInput(req.Username, "username")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	
	// XSS Protection: Sanitize role if provided
	var sanitizedRole string
	if req.Role != "" {
		sanitizedRole, err = validateAndSanitizeInput(req.Role, "role")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
	}
	
	if err := h.svc.Register(c.Context(), sanitizedUsername, req.Password, sanitizedRole, req.TeamID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok"})
}

func (h *userHandler) Update(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	var req struct {
		Password *string `json:"password"`
		Email    *string `json:"email"`
		Role     *string `json:"role"`
		TeamID   *int64  `json:"team_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	
	// XSS Protection: Sanitize email if provided
	var sanitizedEmail *string
	if req.Email != nil {
		sanitized, err := validateAndSanitizeInput(*req.Email, "email")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
		sanitizedEmail = &sanitized
	}
	
	// XSS Protection: Sanitize role if provided
	var sanitizedRole *string
	if req.Role != nil {
		sanitized, err := validateAndSanitizeInput(*req.Role, "role")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
		}
		sanitizedRole = &sanitized
	}
	
	if err := h.svc.UpdateUser(c.Context(), id, req.Password, sanitizedEmail, sanitizedRole, req.TeamID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

func (h *userHandler) Delete(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	if err := h.svc.DeleteUser(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

// GetUser returns a single user by ID.
func (h *userHandler) GetUser(c fiber.Ctx) error {
	id := parseID(c.Params("id"))
	if id == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	var user model.User
	if err := database.GetDB().First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// Resume resumes a paused game.
func (h *gameHandler) Resume(c fiber.Ctx) error {
	if err := h.svc.ResumeGame(c.Context(), parseID(c.Params("id"))); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "game resumed"})
}




func (h *teamHandler) AddMember(c fiber.Ctx) error {
	teamID := parseID(c.Params("id"))
	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.UserID == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "user_id is required"})
	}

	if err := h.svc.AddMember(c.Context(), teamID, req.UserID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "member added successfully"})
}

func (h *teamHandler) RemoveMember(c fiber.Ctx) error {
	teamID := parseID(c.Params("id"))
	userID := parseID(c.Params("userId"))

	if err := h.svc.RemoveMember(c.Context(), teamID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "member removed successfully"})
}
