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
	// Store active round managers by game ID
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
	RoundDuration  int    `json:"round_duration"`  // seconds
	BreakDuration  int    `json:"break_duration"`  // seconds
	ElapsedTime    int    `json:"elapsed_time"`    // seconds
	RemainingTime  int    `json:"remaining_time"`  // seconds
	IsPaused       bool   `json:"is_paused"`
	RoundStartTime string `json:"round_start_time,omitempty"`
	RoundEndTime   string `json:"round_end_time,omitempty"`
	BreakStartTime string `json:"break_start_time,omitempty"`
	BreakEndTime   string `json:"break_end_time,omitempty"`
}

// RoundControlRequest represents a request to control rounds
type RoundControlRequest struct {
	Action        string `json:"action" validate:"required,oneof=pause resume stop start"`
	RoundDuration *int   `json:"round_duration,omitempty"`  // optional: update round duration
	BreakDuration *int   `json:"break_duration,omitempty"`  // optional: update break duration
}

// RoundControlResponse represents the response for round control
type RoundControlResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	State   *RoundInfoResponse  `json:"state,omitempty"`
}

// GetRounds handles GET /api/v1/games/:id/rounds
// Returns current round information
func (rh *roundHandler) GetRounds(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid game ID",
		})
	}
	
	// Get game from database
	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "game not found",
		})
	}
	
	// Check if we have an active manager for this game
	manager, exists := rh.managers[gameID]
	if !exists {
		// Return game state from database
		return c.JSON(RoundInfoResponse{
			GameID:        game.ID,
			GameTitle:     game.Title,
			CurrentRound:  game.CurrentRound,
			TotalRounds:   game.TotalRounds,
			Phase:         game.CurrentPhase,
			RoundDuration: game.RoundDuration,
			BreakDuration: game.BreakDuration,
			IsPaused:      game.Status == "paused",
		})
	}
	
	// Get live state from manager
	state := manager.GetState()
	
	response := RoundInfoResponse{
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
	
	if !state.RoundStartTime.IsZero() {
		response.RoundStartTime = state.RoundStartTime.Format("2006-01-02T15:04:05Z07:00")
		response.RoundEndTime = state.RoundEndTime.Format("2006-01-02T15:04:05Z07:00")
	}
	
	if !state.BreakStartTime.IsZero() {
		response.BreakStartTime = state.BreakStartTime.Format("2006-01-02T15:04:05Z07:00")
		response.BreakEndTime = state.BreakEndTime.Format("2006-01-02T15:04:05Z07:00")
	}
	
	return c.JSON(response)
}

// ControlRounds handles POST /api/v1/games/:id/rounds
// Controls round state (pause/resume/stop/start)
func (rh *roundHandler) ControlRounds(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid game ID",
		})
	}
	
	// Parse request
	var req RoundControlRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	
	// Get game from database
	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "game not found",
		})
	}
	
	ctx := context.Background()
	
	// Get or create manager
	manager, exists := rh.managers[gameID]
	if !exists {
		// Create new manager
		eng := engine.NewCompetitionEngine(&game)
		manager = engine.NewRoundManager(&game, eng)
		manager.SetCallbacks(eng.GetRoundCallbacks())
		rh.managers[gameID] = manager
	}
	
	// Update durations if provided
	if req.RoundDuration != nil {
		if req.BreakDuration == nil {
			req.BreakDuration = &game.BreakDuration
		}
		manager.UpdateDurations(*req.RoundDuration, *req.BreakDuration)
	}
	
	// Execute action
	var response RoundControlResponse
	switch req.Action {
	case "start":
		if manager.IsRunning() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds already running",
			}
		} else {
			if err := manager.Start(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "failed to start rounds",
					"details": err.Error(),
				})
			}
			response = RoundControlResponse{
				Success: true,
				Message: "rounds started successfully",
			}
		}
		
	case "pause":
		if !manager.IsRunning() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds not running",
			}
		} else if manager.IsPaused() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds already paused",
			}
		} else {
			if err := manager.Pause(); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "failed to pause rounds",
					"details": err.Error(),
				})
			}
			response = RoundControlResponse{
				Success: true,
				Message: "rounds paused successfully",
			}
		}
		
	case "resume":
		if !manager.IsRunning() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds not running",
			}
		} else if !manager.IsPaused() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds not paused",
			}
		} else {
			if err := manager.Resume(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "failed to resume rounds",
					"details": err.Error(),
				})
			}
			response = RoundControlResponse{
				Success: true,
				Message: "rounds resumed successfully",
			}
		}
		
	case "stop":
		if !manager.IsRunning() {
			response = RoundControlResponse{
				Success: false,
				Message: "rounds not running",
			}
		} else {
			if err := manager.Stop(ctx); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "failed to stop rounds",
					"details": err.Error(),
				})
			}
			// Remove from active managers
			delete(rh.managers, gameID)
			response = RoundControlResponse{
				Success: true,
				Message: "rounds stopped successfully",
			}
		}
		
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid action",
			"valid_actions": []string{"start", "pause", "resume", "stop"},
		})
	}
	
	// Add current state to response
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
// Returns history of all rounds (optional endpoint)
func (rh *roundHandler) GetRoundHistory(c fiber.Ctx) error {
	gameID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid game ID",
		})
	}
	
	// Get game from database
	db := database.GetDB()
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "game not found",
		})
	}
	
	// TODO: Query round history from database
	// For now, return current state
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
		"game_id":        game.ID,
		"game_title":     game.Title,
		"current_round":  state.CurrentRound,
		"total_rounds":   state.TotalRounds,
		"phase":          string(state.Phase),
	})
}
