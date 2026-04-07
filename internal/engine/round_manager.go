package engine

import (
	"context"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// RoundPhase represents the current phase of a round
type RoundPhase string

const (
	PhasePreparation RoundPhase = "preparation"
	PhaseRunning     RoundPhase = "running"
	PhaseBreak       RoundPhase = "break"
	PhasePaused      RoundPhase = "paused"
	PhaseFinished    RoundPhase = "finished"
	PhaseScoring     RoundPhase = "scoring"
)

// RoundState represents the state of the current round
type RoundState struct {
	CurrentRound   int        `json:"current_round"`
	TotalRounds    int        `json:"total_rounds"`
	Phase          RoundPhase `json:"phase"`
	RoundStartTime time.Time  `json:"round_start_time"`
	RoundEndTime   time.Time  `json:"round_end_time"`
	BreakStartTime time.Time  `json:"break_start_time,omitempty"`
	BreakEndTime   time.Time  `json:"break_end_time,omitempty"`
	PauseTime      time.Time  `json:"pause_time,omitempty"`
	ElapsedTime    int        `json:"elapsed_time"`   // seconds elapsed in current round
	RemainingTime  int        `json:"remaining_time"` // seconds remaining in current round
	RoundDuration  int        `json:"round_duration"` // seconds
	BreakDuration  int        `json:"break_duration"` // seconds
	IsPaused       bool       `json:"is_paused"`
}

// RoundManager manages round timing and state transitions
type RoundManager struct {
	mu           sync.RWMutex
	game         *model.Game
	engine       *CompetitionEngine
	currentRound int
	phase        RoundPhase

	// Timing
	roundDuration  time.Duration
	breakDuration  time.Duration
	roundStartTime time.Time
	breakStartTime time.Time
	pauseTime      time.Time

	// State management
	running    bool
	paused     bool
	cancelFunc context.CancelFunc

	// Callbacks
	onRoundStart func(ctx context.Context, round int) error
	onRoundEnd   func(ctx context.Context, round int) error
}

// NewRoundManager creates a new round manager
func NewRoundManager(game *model.Game, engine *CompetitionEngine) *RoundManager {
	return &RoundManager{
		game:          game,
		engine:        engine,
		currentRound:  0,
		phase:         PhasePreparation,
		roundDuration: time.Duration(game.RoundDuration) * time.Second,
		breakDuration: time.Duration(game.BreakDuration) * time.Second,
		running:       false,
		paused:        false,
	}
}

// SetCallbacks sets the round start and end callbacks
func (rm *RoundManager) SetCallbacks(onRoundStart, onRoundEnd func(ctx context.Context, round int) error) {
	rm.onRoundStart = onRoundStart
	rm.onRoundEnd = onRoundEnd
}

// Start begins the round management
func (rm *RoundManager) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return nil
	}

	ctx, rm.cancelFunc = context.WithCancel(ctx)
	rm.running = true
	rm.paused = false

	// Start round 1
	if err := rm.startRound(ctx, 1); err != nil {
		logger.Error("failed to start round 1", "error", err)
		return err
	}

	go rm.runScheduler(ctx)

	logger.Info("round manager started", "game_id", rm.game.ID)
	return nil
}

// Pause pauses the round timer
func (rm *RoundManager) Pause() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running || rm.paused {
		return nil
	}

	rm.paused = true
	rm.pauseTime = time.Now()
	rm.phase = PhasePaused

	logger.Info("round manager paused", "round", rm.currentRound, "phase", rm.phase)

	// Broadcast pause event
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:paused", map[string]interface{}{
		"round":      rm.currentRound,
		"phase":      rm.phase,
		"pause_time": rm.pauseTime,
	})

	return nil
}

// Resume resumes the round timer
func (rm *RoundManager) Resume(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running || !rm.paused {
		return nil
	}

	// Calculate paused duration and adjust start time
	pauseDuration := time.Since(rm.pauseTime)
	rm.roundStartTime = rm.roundStartTime.Add(pauseDuration)

	rm.paused = false
	rm.phase = PhaseRunning

	logger.Info("round manager resumed", "round", rm.currentRound, "phase", rm.phase)

	// Broadcast resume event
	bus := eventbus.GetBus()
	_ = bus.Publish(ctx, "round:resumed", map[string]interface{}{
		"round":       rm.currentRound,
		"phase":       rm.phase,
		"resume_time": time.Now(),
	})

	return nil
}

// Stop stops the round manager
func (rm *RoundManager) Stop(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.cancelFunc != nil {
		rm.cancelFunc()
	}

	rm.running = false
	rm.paused = false
	rm.phase = PhaseFinished

	logger.Info("round manager stopped", "round", rm.currentRound)

	// Broadcast stop event
	bus := eventbus.GetBus()
	_ = bus.Publish(ctx, "round:stopped", map[string]interface{}{
		"round": rm.currentRound,
	})

	return nil
}

// GetState returns the current round state
func (rm *RoundManager) GetState() *RoundState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	state := &RoundState{
		CurrentRound:  rm.currentRound,
		TotalRounds:   rm.game.TotalRounds,
		Phase:         rm.phase,
		RoundDuration: int(rm.roundDuration.Seconds()),
		BreakDuration: int(rm.breakDuration.Seconds()),
		IsPaused:      rm.paused,
	}

	if !rm.roundStartTime.IsZero() {
		state.RoundStartTime = rm.roundStartTime
		state.RoundEndTime = rm.roundStartTime.Add(rm.roundDuration)

		elapsed := int(time.Since(rm.roundStartTime).Seconds())
		remaining := int(rm.roundDuration.Seconds()) - elapsed

		if rm.paused {
			pauseElapsed := int(time.Since(rm.pauseTime).Seconds())
			elapsed -= pauseElapsed
			remaining += pauseElapsed
			state.PauseTime = rm.pauseTime
		}

		state.ElapsedTime = elapsed
		state.RemainingTime = remaining
	}

	if !rm.breakStartTime.IsZero() {
		state.BreakStartTime = rm.breakStartTime
		state.BreakEndTime = rm.breakStartTime.Add(rm.breakDuration)
	}

	return state
}

// runScheduler runs the round timing scheduler
func (rm *RoundManager) runScheduler(ctx context.Context) {
	roundTimer := time.NewTimer(rm.calculateRemainingTime())
	defer roundTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-roundTimer.C:
			// Wait if paused
			rm.waitForResume(ctx)

			rm.mu.Lock()

			// Round ended
			logger.Info("round ended", "round", rm.currentRound)

			if rm.onRoundEnd != nil {
				if err := rm.onRoundEnd(ctx, rm.currentRound); err != nil {
					logger.Error("round end callback error", "error", err)
				}
			}

			rm.broadcastRoundEnd(rm.currentRound)

			// Check if game finished
			if rm.currentRound >= rm.game.TotalRounds {
				rm.phase = PhaseFinished
				rm.running = false
				logger.Info("game finished", "total_rounds", rm.game.TotalRounds)
				rm.broadcastGameFinished()
				rm.mu.Unlock()
				return
			}

			// Start break
			rm.phase = PhaseBreak
			rm.breakStartTime = time.Now()
			rm.mu.Unlock()

			rm.broadcastBreakStart(rm.currentRound)

			// Wait for break duration with pause support
			breakStart := time.Now()
		breakLoop:
			for {
				rm.mu.RLock()
				isPaused := rm.paused
				rm.mu.RUnlock()

				if isPaused {
					rm.waitForResume(ctx)
					breakStart = time.Now()
					continue
				}

				if time.Since(breakStart) >= rm.breakDuration {
					break breakLoop
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}

			// Start next round
			rm.mu.Lock()
			nextRound := rm.currentRound + 1

			if err := rm.startRound(ctx, nextRound); err != nil {
				logger.Error("failed to start round", "round", nextRound, "error", err)
			}

			roundTimer.Reset(rm.roundDuration)
			rm.mu.Unlock()
		}
	}
}

// startRound starts a new round
func (rm *RoundManager) startRound(ctx context.Context, round int) error {
	rm.currentRound = round
	rm.phase = PhaseRunning
	rm.roundStartTime = time.Now()

	if rm.onRoundStart != nil {
		if err := rm.onRoundStart(ctx, round); err != nil {
			return err
		}
	}

	logger.Info("round started", "round", round)
	rm.broadcastRoundStart(round)

	return nil
}

// calculateRemainingTime calculates remaining time in current round
func (rm *RoundManager) calculateRemainingTime() time.Duration {
	if rm.roundStartTime.IsZero() {
		return rm.roundDuration
	}

	elapsed := time.Since(rm.roundStartTime)
	remaining := rm.roundDuration - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// waitForResume waits while the manager is paused
func (rm *RoundManager) waitForResume(ctx context.Context) {
	for {
		rm.mu.RLock()
		if !rm.paused {
			rm.mu.RUnlock()
			return
		}
		rm.mu.RUnlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			// Check again
		}
	}
}

// Broadcast methods
func (rm *RoundManager) broadcastRoundStart(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:start", map[string]interface{}{
		"round":      round,
		"phase":      PhaseRunning,
		"start_time": time.Now(),
		"duration":   int(rm.roundDuration.Seconds()),
	})
}

func (rm *RoundManager) broadcastRoundEnd(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:end", map[string]interface{}{
		"round": round,
		"phase": PhaseBreak,
	})
}

func (rm *RoundManager) broadcastBreakStart(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:break", map[string]interface{}{
		"round":          round,
		"phase":          PhaseBreak,
		"break_start":    time.Now(),
		"break_duration": int(rm.breakDuration.Seconds()),
	})
}

func (rm *RoundManager) broadcastGameFinished() {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "game:finished", map[string]interface{}{
		"total_rounds": rm.game.TotalRounds,
		"phase":        PhaseFinished,
	})
}

// GetCurrentRound returns the current round number
func (rm *RoundManager) GetCurrentRound() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.currentRound
}

// GetPhase returns the current phase
func (rm *RoundManager) GetPhase() RoundPhase {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.phase
}

// IsRunning returns whether the manager is running
func (rm *RoundManager) IsRunning() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.running
}

// IsPaused returns whether the manager is paused
func (rm *RoundManager) IsPaused() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.paused
}

// UpdateDurations updates round and break durations
func (rm *RoundManager) UpdateDurations(roundDuration, breakDuration int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.roundDuration = time.Duration(roundDuration) * time.Second
	rm.breakDuration = time.Duration(breakDuration) * time.Second

	logger.Info("round durations updated",
		"round_duration", roundDuration,
		"break_duration", breakDuration)
}
