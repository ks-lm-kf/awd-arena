package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// EngineManager manages competition engines for all games.
type EngineManager struct {
	mu      sync.RWMutex
	engines map[int64]*CompetitionEngine
}

var Manager = &EngineManager{
	engines: make(map[int64]*CompetitionEngine),
}

func (m *EngineManager) StartGame(game *model.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.engines[game.ID]; exists {
		return nil
	}

	eng, err := NewCompetitionEngine(game)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	eng.SetCancelFunc(cancel)

	if err := eng.Start(ctx); err != nil {
		cancel()
		return err
	}

	m.engines[game.ID] = eng
	logger.Info("engine started", "game_id", game.ID)
	return nil
}

func (m *EngineManager) PauseGame(gameID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eng, exists := m.engines[gameID]
	if !exists {
		return nil
	}

	if err := eng.Pause(); err != nil {
		return err
	}
	logger.Info("engine paused", "game_id", gameID)
	return nil
}

func (m *EngineManager) ResumeGame(game *model.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if eng, exists := m.engines[game.ID]; exists {
		return eng.Resume(context.Background())
	}

	eng, err := NewCompetitionEngine(game)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	eng.SetCancelFunc(cancel)
	eng.currentRound = game.CurrentRound

	if err := eng.Resume(ctx); err != nil {
		cancel()
		return err
	}

	m.engines[game.ID] = eng
	logger.Info("engine resumed", "game_id", game.ID)
	return nil
}

func (m *EngineManager) StopGame(gameID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eng, exists := m.engines[gameID]
	if !exists {
		return nil
	}

	ctx := context.Background()
	eng.Stop(ctx)
	eng.Close()
	delete(m.engines, gameID)
	logger.Info("engine stopped", "game_id", gameID)
	return nil
}

func (m *EngineManager) GetEngine(gameID int64) *CompetitionEngine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engines[gameID]
}

func (m *EngineManager) IsRunning(gameID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	eng, exists := m.engines[gameID]
	return exists && eng.IsRunning()
}

func (m *EngineManager) ShutdownAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, eng := range m.engines {
		ctx := context.Background()
		eng.Stop(ctx)
		eng.Close()
		delete(m.engines, id)
	}
}

// BroadcastRankingUpdate sends ranking update via WebSocket.
func BroadcastRankingUpdate(gameID int64, round int) {
	db := database.GetDB()
	if db == nil {
		return
	}

	var scores []model.RoundScore
	db.Where("game_id = ? AND round = ?", gameID, round).
		Order("total_score desc").Find(&scores)

	eventbus.BroadcastJSON("ranking:update", map[string]interface{}{
		"game_id": gameID,
		"round":   round,
		"scores":  scores,
	})
}
