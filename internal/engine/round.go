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

	roundTimer := time.NewTimer(rs.engine.roundDuration)
	defer roundTimer.Stop()

	logger.Info("round scheduler running", "current_round", rs.engine.currentRound)

	for {
		select {
		case <-ctx.Done():
			return
		case <-roundTimer.C:
			// Wait if paused
			if !rs.waitForResume(ctx) {
				return
			}

			logger.Info("round ended", "round", rs.engine.currentRound)
			if err := rs.engine.onRoundEnd(ctx, rs.engine.currentRound); err != nil {
				logger.Error("round end error", "error", err)
			}
			rs.broadcastRoundEnd(rs.engine.currentRound)

			if rs.engine.currentRound >= rs.engine.totalRounds {
				rs.engine.currentPhase = "finished"
				logger.Info("game finished", "rounds", rs.engine.totalRounds)
				rs.broadcastGameFinished()
				return
			}

			rs.engine.currentPhase = "break"
			breakTimer := time.NewTimer(rs.engine.breakDuration)
			select {
			case <-ctx.Done():
				breakTimer.Stop()
				return
			case <-breakTimer.C:
			}

			nextRound := rs.engine.currentRound + 1
			logger.Info("round starting", "round", nextRound)

			rs.mu.Lock()
			rs.roundStart = time.Now()
			rs.mu.Unlock()

			if err := rs.engine.onRoundStart(ctx, nextRound); err != nil {
				logger.Error("round start error", "error", err)
			}
			rs.broadcastRoundStart(nextRound)
			roundTimer.Reset(rs.engine.roundDuration)
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
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "game:finished", map[string]interface{}{
		"rounds": rs.engine.totalRounds,
	})
}
