package engine

import (
	"context"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// RoundScheduler manages round timing with pause/resume support.
type RoundScheduler struct {
	engine      *CompetitionEngine
	mu          sync.Mutex
	paused      bool
	pauseTime   time.Time
	roundStart  time.Time
	pauseCtx    context.Context
	pauseCancel context.CancelFunc
}

func NewRoundScheduler(engine *CompetitionEngine) *RoundScheduler {
	return &RoundScheduler{engine: engine}
}

// Pause freezes the round timer without killing the goroutine.
func (rs *RoundScheduler) Pause() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.paused {
		return
	}
	rs.paused = true
	rs.pauseTime = time.Now()
	// Signal the wait loop to stop
	if rs.pauseCancel != nil {
		rs.pauseCancel()
	}
	logger.Info("round scheduler paused", "round", rs.engine.currentRound)
}

// Resume unfreezes the timer from where it left off.
func (rs *RoundScheduler) Resume(ctx context.Context) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if !rs.paused {
		return
	}
	// Adjust round start time by the paused duration
	if !rs.pauseTime.IsZero() && !rs.roundStart.IsZero() {
		pausedDuration := time.Since(rs.pauseTime)
		rs.roundStart = rs.roundStart.Add(pausedDuration)
	}
	rs.paused = false
	rs.pauseTime = time.Time{}
	// Create new pause context
	rs.pauseCtx, rs.pauseCancel = context.WithCancel(ctx)
	logger.Info("round scheduler resumed", "round", rs.engine.currentRound)
}

func (rs *RoundScheduler) Run(ctx context.Context) {
	rs.mu.Lock()
	rs.roundStart = time.Now()
	rs.pauseCtx, rs.pauseCancel = context.WithCancel(ctx)
	rs.mu.Unlock()

	rs.engine.mu.Lock()
	roundDuration := rs.engine.roundDuration
	currentRound := rs.engine.currentRound
	rs.engine.mu.Unlock()

	roundTimer := time.NewTimer(roundDuration)
	defer roundTimer.Stop()

	logger.Info("round scheduler running", "current_round", currentRound)

	for {
		select {
		case <-ctx.Done():
			return
		case <-roundTimer.C:
			// Wait if paused
			if !rs.waitForResume(ctx) {
				return
			}

			rs.engine.mu.Lock()
			currentRound = rs.engine.currentRound
			rs.engine.mu.Unlock()

			logger.Info("round ended", "round", currentRound)

			rs.engine.mu.Lock()
			if err := rs.engine.onRoundEnd(ctx, currentRound); err != nil {
				logger.Error("round end error", "error", err)
			}
			rs.engine.mu.Unlock()

			rs.broadcastRoundEnd(currentRound)

			rs.engine.mu.Lock()
			currentRound = rs.engine.currentRound
			totalRounds := rs.engine.totalRounds
			rs.engine.mu.Unlock()

			if currentRound >= totalRounds {
				rs.engine.mu.Lock()
				rs.engine.currentPhase = "finished"
				rs.engine.mu.Unlock()
				logger.Info("game finished", "rounds", totalRounds)
				rs.broadcastGameFinished()
				return
			}

			rs.engine.mu.Lock()
			rs.engine.currentPhase = "break"
			breakDuration := rs.engine.breakDuration
			rs.engine.mu.Unlock()

			breakTimer := time.NewTimer(breakDuration)
			select {
			case <-ctx.Done():
				breakTimer.Stop()
				return
			case <-breakTimer.C:
			}

			rs.engine.mu.Lock()
			currentRound = rs.engine.currentRound
			rs.engine.mu.Unlock()
			nextRound := currentRound + 1

			logger.Info("round starting", "round", nextRound)

			rs.mu.Lock()
			rs.roundStart = time.Now()
			rs.mu.Unlock()

			rs.engine.mu.Lock()
			if err := rs.engine.onRoundStart(ctx, nextRound); err != nil {
				logger.Error("round start error", "error", err)
			}
			rs.engine.mu.Unlock()

			rs.broadcastRoundStart(nextRound)

			rs.engine.mu.Lock()
			roundDuration = rs.engine.roundDuration
			rs.engine.mu.Unlock()
			roundTimer.Reset(roundDuration)
		}
	}
}

// waitForResume blocks while paused, returns false if context is done.
func (rs *RoundScheduler) waitForResume(ctx context.Context) bool {
	for {
		rs.mu.Lock()
		if !rs.paused {
			rs.mu.Unlock()
			return true
		}
		pCtx := rs.pauseCtx
		rs.mu.Unlock()

		select {
		case <-ctx.Done():
			return false
		case <-pCtx.Done():
			// Pause was triggered, keep waiting
			// Re-create pause context for next iteration
			rs.mu.Lock()
			if rs.paused {
				rs.pauseCtx, rs.pauseCancel = context.WithCancel(ctx)
			}
			rs.mu.Unlock()
		case <-time.After(100 * time.Millisecond):
			// Check again
		}
	}
}

func (rs *RoundScheduler) broadcastRoundStart(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:start", map[string]interface{}{
		"round": round, "phase": "running",
	})
}

func (rs *RoundScheduler) broadcastRoundEnd(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:end", map[string]interface{}{
		"round": round,
	})
}

func (rs *RoundScheduler) broadcastGameFinished() {
	rs.engine.mu.Lock()
	totalRounds := rs.engine.totalRounds
	rs.engine.mu.Unlock()
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "game:finished", map[string]interface{}{
		"rounds": totalRounds,
	})
}
