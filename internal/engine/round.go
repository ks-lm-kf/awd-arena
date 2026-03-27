package engine

import (
	"context"
	"time"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// RoundScheduler manages round timing.
type RoundScheduler struct {
	engine *CompetitionEngine
}

func NewRoundScheduler(engine *CompetitionEngine) *RoundScheduler {
	return &RoundScheduler{engine: engine}
}

func (rs *RoundScheduler) Run(ctx context.Context) {
	roundTimer := time.NewTimer(rs.engine.roundDuration)
	defer roundTimer.Stop()

	// If we're resuming mid-round, the first timer fires at the remaining time
	// otherwise round 1 already started in engine.Start()
	logger.Info("round scheduler running", "current_round", rs.engine.currentRound)

	for {
		select {
		case <-ctx.Done():
			return
		case <-roundTimer.C:
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
				return
			case <-breakTimer.C:
			}

			nextRound := rs.engine.currentRound + 1
			logger.Info("round starting", "round", nextRound)
			if err := rs.engine.onRoundStart(ctx, nextRound); err != nil {
				logger.Error("round start error", "error", err)
			}
			rs.broadcastRoundStart(nextRound)
			roundTimer.Reset(rs.engine.roundDuration)
		}
	}
}

func (rs *RoundScheduler) broadcastRoundStart(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:start", map[string]interface{}{"round": round, "phase": "running"})
}

func (rs *RoundScheduler) broadcastRoundEnd(round int) {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "round:end", map[string]interface{}{"round": round})
}

func (rs *RoundScheduler) broadcastGameFinished() {
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "game:finished", map[string]interface{}{"rounds": rs.engine.totalRounds})
}
