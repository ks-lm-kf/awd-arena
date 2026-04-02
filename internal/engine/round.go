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
	engine     *CompetitionEngine
	mu         sync.Mutex
	paused     bool
	pauseTime  time.Time
	roundStart time.Time

	// Remaining time tracking for pause/resume
	remainingRoundTime time.Duration
	remainingBreakTime time.Duration
	inBreak            bool
	breakStartTime     time.Time

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

	// Save remaining time for the current active timer
	if rs.inBreak {
		// We're in break phase - save remaining break time
		elapsed := time.Since(rs.breakStartTime)
		if rs.remainingBreakTime > elapsed {
			rs.remainingBreakTime = rs.remainingBreakTime - elapsed
		} else {
			rs.remainingBreakTime = 0
		}
	} else {
		// We're in round phase - save remaining round time
		elapsed := time.Since(rs.roundStart)
		if rs.remainingRoundTime > elapsed {
			rs.remainingRoundTime = rs.remainingRoundTime - elapsed
		} else {
			rs.remainingRoundTime = 0
		}
	}

	if rs.pauseCancel != nil {
		rs.pauseCancel()
	}
	logger.Info("round scheduler paused", "round", rs.engine.currentRound,
		"remaining_round", rs.remainingRoundTime, "remaining_break", rs.remainingBreakTime,
		"in_break", rs.inBreak)
}

// Resume unfreezes the timer from where it left off.
func (rs *RoundScheduler) Resume(ctx context.Context) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if !rs.paused {
		return
	}
	rs.paused = false
	rs.pauseTime = time.Time{}
	// Create new pause context
	rs.pauseCtx, rs.pauseCancel = context.WithCancel(ctx)

	// Adjust start times so elapsed calculation is correct
	now := time.Now()
	if rs.inBreak {
		rs.breakStartTime = now
	} else {
		rs.roundStart = now
	}

	logger.Info("round scheduler resumed", "round", rs.engine.currentRound,
		"remaining_round", rs.remainingRoundTime, "remaining_break", rs.remainingBreakTime,
		"in_break", rs.inBreak)
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

	// Initialize remaining round time
	rs.mu.Lock()
	rs.remainingRoundTime = roundDuration
	rs.inBreak = false
	rs.mu.Unlock()

	roundTimer := time.NewTimer(roundDuration)
	defer roundTimer.Stop()

	logger.Info("round scheduler running", "current_round", currentRound)

	for {
		select {
		case <-ctx.Done():
			return
		case <-roundTimer.C:
			// Check if paused - wait for resume
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

				rs.engine.finishGame(ctx)
				return
			}

			rs.engine.mu.Lock()
			rs.engine.currentPhase = "break"
			breakDuration := rs.engine.breakDuration
			rs.engine.mu.Unlock()

			// Set up break phase with pause/resume support
			rs.mu.Lock()
			rs.inBreak = true
			rs.remainingBreakTime = breakDuration
			rs.breakStartTime = time.Now()
			rs.mu.Unlock()

			breakTimer := time.NewTimer(breakDuration)
		breakLoop:
			for {
				select {
				case <-ctx.Done():
					breakTimer.Stop()
					return
				case <-breakTimer.C:
					break breakLoop
				}
			}

			// Check if paused during break - wait for resume
			if !rs.waitForResume(ctx) {
				return
			}

			// Use saved remaining break time if we were paused during break
			rs.mu.Lock()
			wasInBreak := rs.inBreak
			remainingBreak := rs.remainingBreakTime
			rs.mu.Unlock()

			if wasInBreak && remainingBreak > 0 {
				// We were paused during break and have remaining time
				continuedBreak := time.NewTimer(remainingBreak)
				select {
				case <-ctx.Done():
					continuedBreak.Stop()
					return
				case <-continuedBreak.C:
				}
			}

			rs.engine.mu.Lock()
			currentRound = rs.engine.currentRound
			rs.engine.mu.Unlock()
			nextRound := currentRound + 1

			logger.Info("round starting", "round", nextRound)

			rs.mu.Lock()
			rs.roundStart = time.Now()
			rs.inBreak = false
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

			// Reset remaining round time for new round
			rs.mu.Lock()
			rs.remainingRoundTime = roundDuration
			rs.mu.Unlock()

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
