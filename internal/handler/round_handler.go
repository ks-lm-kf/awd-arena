package handler

import (
	"context"
	"strconv"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/engine"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

// roundHandler handles round-related API endpoints
type roundHandler struct {
	managers map[int64]*engine.RoundManager
}

// NewRoundHandler creates a new round handler
func NewRoundHandler() *roundHandler {
	return &roundHandler{
		managers: make(map[int64]*engine.RoundManager),
	}
}

// RoundInfoResponse represents the response for round information
type RoundInfoResponse struct {
	GameID         int64  `json:"game_id"`
	GameTitle      string `json:"game_title"`
	CurrentRound   int    `json:"current_round"`
	TotalRounds    int    `json:"total_rounds"`
	Phase          string `json:"phase"`
	RoundDuration  int    `json:"round_duration"`
	BreakDuration  int    `json:"break_duration"`
	ElapsedTime    int    `json:"elapsed_time"`
	RemainingTime  int    `json:"remaining_time"`
	IsPaused       bool   `json:"is_paused"`
	RoundStartTime string `json:"round_start_time,omitempty"`
	RoundEndTime   string `json:"round_end_time,omitempty"`
	BreakStartTime string `json:"break_start_time,omitempty"`
	BreakEndTime   string `json:"break_end_time,omitempty"`
}

// RoundControlRequest represents a request to control rounds
type RoundControlRequest struct {
	Action        string `json:"action" validate:"required,oneof=pause resume stop start"`
	RoundDuration *int   `json:"round_duration,omitempty"`
	BreakDuration *int   `json:"break_duration,omitempty"`
}

// RoundControlResponse represents the response for round control
type RoundControlResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	State   *RoundInfoResponse `json:"state,omitempty"`
}

// RoundHistoryItem represents a single round in history
type RoundHistoryItem struct {
	Round           int     `json:"round"`
	GameID          int64   `json:"game_id"`
	Phase           string  `json:"phase"`
	StartTime       string  `json:"start_time,omitempty"`
	EndTime         string  `json:"end_time,omitempty"`
	SubmissionCount int     `json:"submission_count"`
	TopAttacker     string  `json:"top_attacker,omitempty"`
	TopScore        float64 `json:"top_score,omitempty"`
}

// GetRounds handles GET /api/v1/games/:id/rounds
// Returns round history from database
func (rh *roundHandler) GetRounds(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid game ID"})
	}

	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "game not found"})
	}

	// Try to get round history from flag_submissions grouped by round
	type RoundStat struct {
		Round           int
		SubmissionCount int
	}
	var roundStats []RoundStat
	db.Table("flag_submissions").
		Select("round, count(*) as submission_count").
		Where("game_id = ?", gameID).
		Group("round").
		Order("round ASC").
		Find(&roundStats)

	// Build round history
	var rounds []RoundHistoryItem
	for i := 1; i <= game.CurrentRound; i++ {
		item := RoundHistoryItem{
			Round:  i,
			GameID: gameID,
			Phase:  "finished",
		}

		// Find stats for this round
		for _, rs := range roundStats {
			if rs.Round == i {
				item.SubmissionCount = rs.SubmissionCount
				break
			}
		}

		// If this is the current round and game is still running, mark as running
		if i == game.CurrentRound && (game.Status == "running" || game.Status == "active") {
			item.Phase = game.CurrentPhase
			if game.StartTime != nil {
				item.StartTime = game.StartTime.Format("2006-01-02T15:04:05Z07:00")
			}
		}

		rounds = append(rounds, item)
	}

	// If no rounds yet, return current state info
	if len(rounds) == 0 {
		// Check if there's an active manager
		manager, exists := rh.managers[gameID]
		if exists {
			state := manager.GetState()
			return c.JSON(fiber.Map{
				"game_id":        game.ID,
				"game_title":     game.Title,
				"current_round":  state.CurrentRound,
				"total_rounds":   state.TotalRounds,
				"phase":          string(state.Phase),
				"round_duration": state.RoundDuration,
				"break_duration": state.BreakDuration,
				"elapsed_time":   state.ElapsedTime,
				"remaining_time": state.RemainingTime,
				"is_paused":      state.IsPaused,
			})
		}

		return c.JSON(fiber.Map{
			"game_id":       game.ID,
			"game_title":    game.Title,
			"current_round": game.CurrentRound,
			"total_rounds":  game.TotalRounds,
			"phase":         game.CurrentPhase,
			"rounds":        []interface{}{},
		})
	}

	return c.JSON(rounds)
}

// ControlRounds handles POST /api/v1/games/:id/rounds
func (rh *roundHandler) ControlRounds(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid game ID"})
	}

	var req RoundControlRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "game not found"})
	}

	ctx := context.Background()

	manager, exists := rh.managers[gameID]
	if !exists {
		eng := engine.NewCompetitionEngine(&game)
		manager = engine.NewRoundManager(&game, eng)
		manager.SetCallbacks(eng.GetRoundCallbacks())
		rh.managers[gameID] = manager
	}

	if req.RoundDuration != nil {
		if req.BreakDuration == nil {
			req.BreakDuration = &game.BreakDuration
		}
		manager.UpdateDurations(*req.RoundDuration, *req.BreakDuration)
	}

	var response RoundControlResponse
	switch req.Action {
	case "start":
		if manager.IsRunning() {
			response = RoundControlResponse{Success: false, Message: "rounds already running"}
		} else {
			if err := manager.Start(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to start rounds", "details": err.Error()})
			}
			response = RoundControlResponse{Success: true, Message: "rounds started successfully"}
		}
	case "pause":
		if !manager.IsRunning() {
			response = RoundControlResponse{Success: false, Message: "rounds not running"}
		} else if manager.IsPaused() {
			response = RoundControlResponse{Success: false, Message: "rounds already paused"}
		} else {
			if err := manager.Pause(); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to pause rounds", "details": err.Error()})
			}
			response = RoundControlResponse{Success: true, Message: "rounds paused successfully"}
		}
	case "resume":
		if !manager.IsRunning() {
			response = RoundControlResponse{Success: false, Message: "rounds not running"}
		} else if !manager.IsPaused() {
			response = RoundControlResponse{Success: false, Message: "rounds not paused"}
		} else {
			if err := manager.Resume(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to resume rounds", "details": err.Error()})
			}
			response = RoundControlResponse{Success: true, Message: "rounds resumed successfully"}
		}
	case "stop":
		if !manager.IsRunning() {
			response = RoundControlResponse{Success: false, Message: "rounds not running"}
		} else {
			if err := manager.Stop(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to stop rounds", "details": err.Error()})
			}
			delete(rh.managers, gameID)
			response = RoundControlResponse{Success: true, Message: "rounds stopped successfully"}
		}
	default:
		return c.Status(400).JSON(fiber.Map{"error": "invalid action", "valid_actions": []string{"start", "pause", "resume", "stop"}})
	}

	state := manager.GetState()
	response.State = &RoundInfoResponse{
		GameID:        game.ID,
		GameTitle:     game.Title,
		CurrentRound:  state.CurrentRound,
		TotalRounds:   state.TotalRounds,
		Phase:         string(state.Phase),
		RoundDuration: state.RoundDuration,
		BreakDuration: state.BreakDuration,
		ElapsedTime:   state.ElapsedTime,
		RemainingTime: state.RemainingTime,
		IsPaused:      state.IsPaused,
	}

	return c.JSON(response)
}

// GetRoundHistory handles GET /api/v1/games/:id/rounds/history
func (rh *roundHandler) GetRoundHistory(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid game ID"})
	}

	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "game not found"})
	}

	manager, exists := rh.managers[gameID]
	if !exists {
		return c.JSON(fiber.Map{
			"game_id":    game.ID,
			"game_title": game.Title,
			"rounds":     []interface{}{},
			"message":    "no active rounds",
		})
	}

	state := manager.GetState()
	return c.JSON(fiber.Map{
		"game_id":       game.ID,
		"game_title":    game.Title,
		"current_round": state.CurrentRound,
		"total_rounds":  state.TotalRounds,
		"phase":         string(state.Phase),
	})
}
