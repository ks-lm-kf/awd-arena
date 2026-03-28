package engine

import (
	"context"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/awd-platform/awd-arena/pkg/logger"

	"github.com/docker/docker/client"
)

// CompetitionEngine is the main competition engine.
type CompetitionEngine struct {
	mu             sync.Mutex
	game           *model.Game
	roundDuration  time.Duration
	breakDuration  time.Duration
	totalRounds    int
	currentRound   int
	currentPhase   string
	flagSvc        *service.FlagService
	gameSvc        *service.GameService
	scorer         *ScoreCalculator
	roundScheduler *RoundScheduler
	healthChecker  *HealthChecker
	cancelFunc     context.CancelFunc
	running        bool
	flagWriter     *FlagWriter
	dockerClient   *client.Client
}

// NewCompetitionEngine creates a new engine instance.
func NewCompetitionEngine(game *model.Game) *CompetitionEngine {
	// Initialize Docker client for flag writing
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("failed to create docker client for engine", "error", err)
	}

	var flagWriter *FlagWriter
	if dockerClient != nil {
		flagWriter = NewFlagWriter(dockerClient)
	}

	return &CompetitionEngine{
		game:          game,
		flagSvc:       &service.FlagService{},
		gameSvc:       &service.GameService{},
		scorer:        NewScoreCalculator(game),
		roundDuration: time.Duration(game.RoundDuration) * time.Second,
		breakDuration: time.Duration(game.BreakDuration) * time.Second,
		totalRounds:   game.TotalRounds,
		currentPhase:  "preparation",
		dockerClient:  dockerClient,
		flagWriter:    flagWriter,
	}
}

func (e *CompetitionEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return nil
	}
	ctx, e.cancelFunc = context.WithCancel(ctx)
	e.running = true

	// Generate flags for round 1 immediately
	if err := e.onRoundStart(ctx, 1); err != nil {
		logger.Error("initial round start error", "error", err)
	}

	// Start round scheduler
	e.roundScheduler = NewRoundScheduler(e)
	go e.roundScheduler.Run(ctx)

	// Start health checker
	e.healthChecker = NewHealthChecker(e.game.ID, e.dockerClient)
	e.healthChecker.Start(ctx)

	logger.Info("competition engine started", "game", e.game.Title)
	return nil
}

func (e *CompetitionEngine) Pause() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return nil
	}
	// Pause the round scheduler (freezes timer without killing it)
	if e.roundScheduler != nil {
		e.roundScheduler.Pause()
	}
	// Pause health checker
	if e.healthChecker != nil {
		e.healthChecker.Stop()
	}
	e.running = false
	e.currentPhase = "break"
	logger.Info("competition engine paused", "round", e.currentRound)
	return nil
}

func (e *CompetitionEngine) Resume(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return nil
	}
	ctx, e.cancelFunc = context.WithCancel(ctx)
	e.running = true

	// If we haven't started any round yet, trigger round 1
	if e.currentRound == 0 {
		if err := e.onRoundStart(ctx, 1); err != nil {
			logger.Error("round start error on resume", "error", err)
		}
	}

	// Resume or create new round scheduler
	if e.roundScheduler != nil {
		e.roundScheduler.Resume(ctx)
	} else {
		e.roundScheduler = NewRoundScheduler(e)
		go e.roundScheduler.Run(ctx)
	}

	// Resume health checker
	e.healthChecker = NewHealthChecker(e.game.ID, e.dockerClient)
	e.healthChecker.Start(ctx)

	logger.Info("competition engine resumed", "round", e.currentRound)
	return nil
}

func (e *CompetitionEngine) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancelFunc != nil {
		e.cancelFunc()
	}
	e.running = false
	e.currentPhase = "finished"

	// Stop health checker
	if e.healthChecker != nil {
		e.healthChecker.Stop()
	}

	// Update game in DB
	db := database.GetDB()
	if db != nil {
		db.Model(&model.Game{}).Where("id = ?", e.game.ID).Updates(map[string]interface{}{
			"status":        "finished",
			"current_phase": "finished",
		})
	}

	// Cleanup all containers for this game
	csvc := &service.ContainerService{}
	if err := csvc.TeardownContainers(ctx, e.game.ID); err != nil {
		logger.Error("failed to teardown containers on stop", "game_id", e.game.ID, "error", err)
	}

	// Broadcast game finished
	bus := eventbus.GetBus()
	_ = bus.Publish(ctx, "game:finished", map[string]interface{}{
		"game_id": e.game.ID,
		"round":   e.currentRound,
	})

	logger.Info("competition engine stopped", "round", e.currentRound)
	return nil
}

func (e *CompetitionEngine) GetCurrentRound() int    { return e.currentRound }
func (e *CompetitionEngine) GetCurrentPhase() string { return e.currentPhase }
func (e *CompetitionEngine) GetGame() *model.Game    { return e.game }
func (e *CompetitionEngine) IsRunning() bool         { return e.running }

func (e *CompetitionEngine) onRoundStart(ctx context.Context, round int) error {
	e.currentRound = round
	e.currentPhase = "running"

	// Update game current_round in DB
	db := database.GetDB()
	if db != nil {
		db.Model(&model.Game{}).Where("id = ?", e.game.ID).Updates(map[string]interface{}{
			"current_round":  round,
			"current_phase": "running",
		})
	}

	// Generate flags
	err := e.flagSvc.GenerateFlags(ctx, e.game.ID, round)
	if err != nil {
		logger.Error("failed to generate flags", "round", round, "error", err)
		return err
	}

	// Write flags to containers
	if err := e.writeFlagsToContainers(ctx, round); err != nil {
		logger.Error("failed to write flags to containers", "round", round, "error", err)
		// Don't return error here, flags were generated successfully
	}

	// Publish round:start event
	bus := eventbus.GetBus()
	_ = bus.Publish(ctx, "round:start", map[string]interface{}{
		"game_id": e.game.ID,
		"round":   round,
		"phase":   "running",
	})

	logger.Info("round started", "round", round, "game_id", e.game.ID)
	return nil
}

// writeFlagsToContainers writes the generated flags to each team's container.
func (e *CompetitionEngine) writeFlagsToContainers(ctx context.Context, round int) error {
	if e.flagWriter == nil || e.dockerClient == nil {
		logger.Warn("flag writer or docker client not initialized, skipping flag write")
		return nil
	}

	db := database.GetDB()
	if db == nil {
		return nil
	}

	// Get all team containers for this game
	var containers []model.TeamContainer
	if err := db.Where("game_id = ?", e.game.ID).Find(&containers).Error; err != nil {
		logger.Error("failed to get team containers", "error", err)
		return err
	}

	if len(containers) == 0 {
		logger.Info("no containers found for game", "game_id", e.game.ID)
		return nil
	}

	// Get flag records for this round
	var flagRecords []model.FlagRecord
	if err := db.Where("game_id = ? AND round = ?", e.game.ID, round).Find(&flagRecords).Error; err != nil {
		logger.Error("failed to get flag records", "error", err)
		return err
	}

	// Build map of teamID -> flag value
	flagMap := make(map[int64]string)
	for _, record := range flagRecords {
		flagMap[record.TeamID] = record.FlagValue
	}

	// Write flags to containers
	successCount := 0
	for _, container := range containers {
		flagValue, ok := flagMap[container.TeamID]
		if !ok {
			logger.Warn("no flag found for team", "team_id", container.TeamID, "container_id", container.ContainerID)
			continue
		}

		if err := e.flagWriter.WriteFlag(ctx, container.ContainerID, flagValue); err != nil {
			logger.Error("failed to write flag to container",
				"container_id", container.ContainerID,
				"team_id", container.TeamID,
				"error", err)
		} else {
			successCount++
			logger.Info("flag written to container",
				"container_id", container.ContainerID,
				"team_id", container.TeamID)
		}
	}

	logger.Info("flag writing completed",
		"round", round,
		"total_containers", len(containers),
		"success_count", successCount)

	return nil
}

func (e *CompetitionEngine) onRoundEnd(ctx context.Context, round int) error {
	e.currentPhase = "scoring"

	// Publish round:end event
	bus := eventbus.GetBus()
	_ = bus.Publish(ctx, "round:end", map[string]interface{}{
		"game_id": e.game.ID,
		"round":   round,
		"phase":   "scoring",
	})

	return e.scorer.CalculateRoundScores(ctx, round)
}

// SetCancelFunc sets the cancel function for the engine.
func (e *CompetitionEngine) SetCancelFunc(cancel context.CancelFunc) {
	e.cancelFunc = cancel
}

// GetRoundCallbacks returns the round start and end callbacks for RoundManager.
func (e *CompetitionEngine) GetRoundCallbacks() (func(ctx context.Context, round int) error, func(ctx context.Context, round int) error) {
	return e.onRoundStart, e.onRoundEnd
}
