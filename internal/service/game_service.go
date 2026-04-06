package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

var EngineCallbacks struct {
	StartGame  func(game *model.Game) error
	PauseGame  func(gameID int64) error
	ResumeGame func(game *model.Game) error
	StopGame   func(gameID int64) error
}

var gameMutexes = struct {
	sync.RWMutex
	m map[int64]*sync.Mutex
}{m: make(map[int64]*sync.Mutex)}

func getGameMutex(gameID int64) *sync.Mutex {
	gameMutexes.Lock()
	defer gameMutexes.Unlock()
	mu, ok := gameMutexes.m[gameID]
	if !ok {
		mu = &sync.Mutex{}
		gameMutexes.m[gameID] = mu
	}
	return mu
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
	mu := getGameMutex(gameID)
	mu.Lock()
	defer mu.Unlock()

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// Get the game object
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}

	// State machine validation: only draft → running is allowed
	if !game.CanStart() {
		return fmt.Errorf("cannot start game in %q status, must be in draft", game.Status)
	}

	now := time.Now()
	if err := db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"status":        "running",
		"current_phase": "running",
		"start_time":    now,
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
			db.Model(&model.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
				"status":        "draft",
				"current_phase": "preparation",
			})
			return fmt.Errorf("engine start failed: %w", err)
		}
		logger.Info("competition engine started successfully", "game", gameID)
	} else {
		logger.Warn("EngineCallbacks.StartGame not set, competition engine not started")
	}

	logger.Info("game started, provisioning containers", "game", gameID)
	return nil
}

// PauseGame pauses a game and pauses all containers.
func (s *GameService) PauseGame(ctx context.Context, gameID int64) error {
	mu := getGameMutex(gameID)
	mu.Lock()
	defer mu.Unlock()

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// State machine validation: only running → paused is allowed
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}
	if !game.CanPause() {
		return fmt.Errorf("cannot pause game in %q status", game.Status)
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
	mu := getGameMutex(gameID)
	mu.Lock()
	defer mu.Unlock()

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// Get the game object
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}

	// State machine validation: only paused → running is allowed
	if !game.CanResume() {
		return fmt.Errorf("cannot resume game in %q status, must be paused", game.Status)
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
	mu := getGameMutex(gameID)
	mu.Lock()
	defer mu.Unlock()

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// State machine validation: only running/paused → finished is allowed
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}
	if !game.CanFinish() {
		return fmt.Errorf("cannot stop game in %q status", game.Status)
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

// ResetGame resets a game and cleans up containers and scoring data.
func (s *GameService) ResetGame(ctx context.Context, gameID int64) error {
	mu := getGameMutex(gameID)
	mu.Lock()
	defer mu.Unlock()

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// State machine validation: only finished/stopped → draft is allowed
	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return errors.New("game not found")
	}
	if game.Status != model.GameStatusFinished && game.Status != model.GameStatusDraft {
		return fmt.Errorf("cannot reset game in %q status, must be finished or draft", game.Status)
	}

	// Cleanup containers first
	go func() {
		csvc := NewContainerService()
		if err := csvc.TeardownContainers(context.Background(), gameID); err != nil {
			logger.Error("reset cleanup failed", "game", gameID, "error", err)
		}
	}()

	// Clean up scoring data
	db.Where("game_id = ?", gameID).Delete(&model.FlagSubmission{})
	db.Where("game_id = ?", gameID).Delete(&model.RoundScore{})
	db.Where("game_id = ?", gameID).Delete(&model.FlagRecord{})

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
