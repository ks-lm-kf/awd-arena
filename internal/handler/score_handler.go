package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

var ScoreHandler *scoreHandler
type scoreHandler struct{}

func init() {
	ScoreHandler = &scoreHandler{}
}

type AdjustScoreRequest struct {
	GameID      int64  `json:"game_id"`
	TeamID      int64  `json:"team_id"`
	AdjustValue int    `json:"adjust_value"`
	Reason      string `json:"reason"`
}

func (h *scoreHandler) AdjustScore(c fiber.Ctx) error {
	var req AdjustScoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	userID, _ := c.Locals("user_id").(int64)
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"error": "database not available"})
	}
	adj := &model.ScoreAdjustment{
		GameID: req.GameID, TeamID: req.TeamID,
		AdjustValue: req.AdjustValue, Reason: req.Reason, OperatorID: userID,
	}
	if err := db.Create(adj).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ok", "adjustment": adj})
}

func (h *scoreHandler) GetScoreAdjustments(c fiber.Ctx) error {
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"error": "database not available"})
	}
	gameID, _ := strconv.ParseInt(c.Query("game_id", "0"), 10, 64)
	teamID, _ := strconv.ParseInt(c.Query("team_id", "0"), 10, 64)
	var adjs []model.ScoreAdjustment
	q := db.Model(&model.ScoreAdjustment{})
	if gameID > 0 { q = q.Where("game_id = ?", gameID) }
	if teamID > 0 { q = q.Where("team_id = ?", teamID) }
	q.Order("created_at desc").Find(&adjs)
	return c.JSON(fiber.Map{"adjustments": adjs})
}

func (h *scoreHandler) GetMyContainers(c fiber.Ctx) error {
	// 添加 panic 恢复
	defer func() {
		if r := recover(); r != nil {
			logger.Error("GetMyContainers panic", "error", r)
		}
	}()

	// 获取 user_id - 注意 Fiber v3 的 Locals 返回 any
	userIDVal := c.Locals("user_id")
	logger.Info("GetMyContainers", "user_id_raw", userIDVal, "type", userIDVal)
	
	var userID int64
	switch v := userIDVal.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		logger.Warn("GetMyContainers: invalid user_id type", "value", userIDVal)
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - invalid user_id"})
	}
	
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid game ID"})
	}
	
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"error": "database not available"})
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	
	logger.Info("GetMyContainers", "user_team_id", user.TeamID, "game_id", gameID)
	
	if user.TeamID == nil {
		return c.JSON(fiber.Map{"code": 0, "data": []interface{}{}})
	}
	if *user.TeamID == 0 {
		return c.JSON(fiber.Map{"code": 0, "data": []interface{}{}})
	}

	var containers []model.TeamContainer
	db.Where("team_id = ? AND game_id = ?", *user.TeamID, gameID).Find(&containers)

	challengeMap := make(map[int64]string)
	var challenges []model.Challenge
	db.Find(&challenges)
	for _, ch := range challenges { challengeMap[ch.ID] = ch.Name }

	type CI struct {
		ID            int64  `json:"id"`
		ContainerID   string `json:"container_id"`
		IPAddress     string `json:"ip_address"`
		PortMapping   string `json:"port_mapping"`
		SSHUser       string `json:"ssh_user"`
		SSHPassword   string `json:"ssh_password"`
		ChallengeName string `json:"challenge_name"`
		Status        string `json:"status"`
	}

	result := make([]CI, 0, len(containers))
	for _, cont := range containers {
		ci := CI{
			ID: cont.ID, ContainerID: cont.ContainerID,
			IPAddress: cont.IPAddress, PortMapping: cont.PortMapping,
			SSHUser: cont.SSHUser, SSHPassword: cont.SSHPassword,
			Status: cont.Status,
		}
		ci.ChallengeName = challengeMap[cont.ChallengeID]
		result = append(result, ci)
	}
	return c.JSON(fiber.Map{"code": 0, "data": result})
}
