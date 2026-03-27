package service

import (
	"context"
	"errors"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// EngineCallbacks holds callbacks to start/stop/pause/resume the competition engine
// This avoids cyclic dependency between service and engine packages
var EngineCallbacks struct {
	StartGame  func(game *model.Game) error
	PauseGame  func(gameID int64) error
	ResumeGame func(game *model.Game) error
	StopGame   func(gameID int64) error
}


// GameService handles game management logic.
type GameService struct{}

// CreateGame creates a new game.
func (s *GameService) CreateGame(ctx context.Context, game *model.Game) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Create(game).Error
}

// UpdateGame updates a game.
func (s *GameService) UpdateGame(ctx context.Context, game *model.Game) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Save(game).Error
}

// StartGame starts a game and provisions containers.
func (s *GameService) StartGame(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	
	// Get the game object
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}
	
	now := time.Now()
	if err := db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":       "running",
		"current_phase": "running",
		"start_time":   now,
	}).Error; err != nil {
		return err
	}

	// Provision containers asynchronously
	go func() {
		csvc := NewContainerService()
		if err := csvc.ProvisionContainers(context.Background(), gameID); err != nil {
			logger.Error("container provisioning failed", "game", gameID, "error", err)
		}
	}()

	// Start the competition engine to manage rounds
	game.Status = "running"
	game.CurrentPhase = "running"
	game.StartTime = &now
	if EngineCallbacks.StartGame != nil {
		if err := EngineCallbacks.StartGame(&game); err != nil {
			logger.Error("failed to start competition engine", "game", gameID, "error", err)
			// Don't return error here as containers are already being provisioned
		} else {
			logger.Info("competition engine started successfully", "game", gameID)
		}
	} else {
		logger.Warn("EngineCallbacks.StartGame not set, competition engine not started")
	}

	logger.Info("game started, provisioning containers", "game", gameID)
	return nil
}

// PauseGame pauses a game and pauses all containers.
func (s *GameService) PauseGame(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	if err := db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":        "paused",
		"current_phase": "break",
	}).Error; err != nil {
		return err
	}

	// Pause the competition engine
	if EngineCallbacks.PauseGame != nil {
		if err := EngineCallbacks.PauseGame(gameID); err != nil {
			logger.Error("failed to pause competition engine", "game", gameID, "error", err)
		}
	}

	go func() {
		csvc := NewContainerService()
		if err := csvc.PauseContainers(context.Background(), gameID); err != nil {
			logger.Error("pause containers failed", "game", gameID, "error", err)
		}
	}()

	logger.Info("game paused", "game", gameID)
	return nil
}

// ResumeGame resumes a paused game and unpauses containers.
func (s *GameService) ResumeGame(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	
	// Get the game object
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}
	
	if err := db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":        "running",
		"current_phase": "running",
	}).Error; err != nil {
		return err
	}
	
	// Resume the competition engine
	game.Status = "running"
	game.CurrentPhase = "running"
	if EngineCallbacks.ResumeGame != nil {
		if err := EngineCallbacks.ResumeGame(&game); err != nil {
			logger.Error("failed to resume competition engine", "game", gameID, "error", err)
		}
	}
	
	go func() {
		csvc := NewContainerService()
		if err := csvc.ResumeContainers(context.Background(), gameID); err != nil {
			logger.Error("resume containers failed", "game", gameID, "error", err)
		}
	}()
	
	logger.Info("game resumed", "game", gameID)
	return nil
}

// StopGame stops a game and tears down all containers.
func (s *GameService) StopGame(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	now := time.Now()
	if err := db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":        "finished",
		"current_phase": "finished",
		"end_time":      now,
	}).Error; err != nil {
		return err
	}

	// Stop the competition engine
	if EngineCallbacks.StopGame != nil {
		if err := EngineCallbacks.StopGame(gameID); err != nil {
			logger.Error("failed to stop competition engine", "game", gameID, "error", err)
		}
	}

	go func() {
		csvc := NewContainerService()
		if err := csvc.TeardownContainers(context.Background(), gameID); err != nil {
			logger.Error("container teardown failed", "game", gameID, "error", err)
		}
	}()

	logger.Info("game stopped", "game", gameID)
	return nil
}

// ResetGame resets a game and cleans up containers.
func (s *GameService) ResetGame(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// Cleanup containers first
	go func() {
		csvc := NewContainerService()
		if err := csvc.TeardownContainers(context.Background(), gameID); err != nil {
			logger.Error("reset cleanup failed", "game", gameID, "error", err)
		}
	}()

	return db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":        "draft",
		"current_round": 0,
		"current_phase": "preparation",
		"start_time":    nil,
		"end_time":      nil,
	}).Error
}

// ListGames returns all games.
func (s *GameService) ListGames(ctx context.Context) ([]*model.Game, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var games []*model.Game
	err := db.Order("id desc").Find(&games).Error
	return games, err
}

// GetGame returns a game by ID.
func (s *GameService) GetGame(ctx context.Context, id int64) (*model.Game, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var game model.Game
	err := db.First(&game, id).Error
	return &game, err
}
