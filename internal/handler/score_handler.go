package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/gofiber/fiber/v3"
)

// 全局 ScoreHandler 变量
var ScoreHandler *scoreHandler

// scoreHandler 分数调整Handler
type scoreHandler struct {
	scoreService *service.ScoreService
}

func init() {
	// 初始化 ScoreHandler
	ScoreHandler = NewScoreHandler(service.NewScoreService(database.GetDB()))
}

// NewScoreHandler 创建ScoreHandler
func NewScoreHandler(scoreService *service.ScoreService) *scoreHandler {
	return &scoreHandler{
		scoreService: scoreService,
	}
}

// AdjustScoreRequest 分数调整请求
type AdjustScoreRequest struct {
	GameID      int64  `json:"game_id"`
	TeamID      int64  `json:"team_id"`
	AdjustValue int    `json:"adjust_value"`
	Reason      string `json:"reason"`
}

// AdjustScore 手动调整分数
func (h *scoreHandler) AdjustScore(c fiber.Ctx) error {
	var req AdjustScoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 验证必填字段
	if req.GameID == 0 || req.TeamID == 0 || req.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "game_id, team_id and reason are required"})
	}

	// 获取当前用户（操作人）
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 调用服务层
	adjustment, err := h.scoreService.AdjustScore(
		req.GameID,
		req.TeamID,
		req.AdjustValue,
		req.Reason,
		userID.(int64),
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Score adjusted successfully",
		"adjustment": adjustment,
	})
}

// GetScoreAdjustments 获取分数调整历史
func (h *scoreHandler) GetScoreAdjustments(c fiber.Ctx) error {
	gameIDStr := c.Query("game_id", "0")
	teamIDStr := c.Query("team_id", "0")

	gameID, _ := strconv.ParseInt(gameIDStr, 10, 64)
	teamID, _ := strconv.ParseInt(teamIDStr, 10, 64)

	adjustments, err := h.scoreService.GetAdjustments(gameID, teamID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"adjustments": adjustments})
}

// GetMyContainers 获取当前用户的容器信息
func (h *scoreHandler) GetMyContainers(c fiber.Ctx) error {
	// 获取当前用户
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 获取比赛ID
	gameIDStr := c.Params("id")
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid game ID"})
	}

	// 查询容器信息
	containers, err := h.scoreService.GetUserContainers(userID.(int64), gameID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"code": 0,
		"data": containers,
	})
}

