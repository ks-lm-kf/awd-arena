package mode

import (
	"context"
	"fmt"

	"github.com/awd-platform/awd-arena/internal/engine/scoring"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/repo"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// ScoringConfig holds configuration for scoring modes.
type ScoringConfig struct {
	InitialTotal     float64 // Initial total points pool
	FlagValue        float64 // Points per flag
	FirstBloodBonus  float64 // First blood bonus ratio (e.g., 0.1 = 10% bonus)
	DefenseValue     float64 // Defense points per round
}

// DefaultScoringConfig returns default scoring configuration.
func DefaultScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		InitialTotal:    10000.0,
		FlagValue:       100.0,
		FirstBloodBonus: 0.1, // 10% bonus
		DefenseValue:    50.0,
	}
}

// AWDScoreMode implements classic AWD attack-defense scoring with zero-sum.
type AWDScoreMode struct {
	game       *model.Game
	config     *ScoringConfig
	scorer     *scoring.ZeroSumScorer
	scoreRepo  repo.ScoreRepo
	teamIDs    []int64
}

// NewAWDScoreMode creates a new AWD score mode with default config.
func NewAWDScoreMode() *AWDScoreMode {
	return NewAWDScoreModeWithConfig(DefaultScoringConfig(), nil)
}

// NewAWDScoreModeWithConfig creates a new AWD score mode with custom config.
func NewAWDScoreModeWithConfig(config *ScoringConfig, scoreRepo repo.ScoreRepo) *AWDScoreMode {
	if config == nil {
		config = DefaultScoringConfig()
	}
	
	return &AWDScoreMode{
		config:    config,
		scoreRepo: scoreRepo,
	}
}

// Start initializes the scoring mode for a game.
func (m *AWDScoreMode) Start(ctx context.Context, game *model.Game) error {
	m.game = game

	// Initialize zero-sum scorer
	m.scorer = scoring.NewZeroSumScorer(
		m.config.InitialTotal,
		m.config.FlagValue,
		m.config.FirstBloodBonus,
		m.scoreRepo,
	)

	// Get team IDs from game
	// TODO: Load teams from game.Teams or database
	// For now, we'll initialize when teams are registered
	
	logger.Info("AWD score mode started",
		"game", game.Title,
		"initial_total", m.config.InitialTotal,
		"flag_value", m.config.FlagValue,
		"first_blood_bonus", m.config.FirstBloodBonus,
	)

	return nil
}

// InitializeTeams sets up teams for scoring.
func (m *AWDScoreMode) InitializeTeams(teamIDs []int64) {
	m.teamIDs = teamIDs
	m.scorer.Initialize(teamIDs)
}

// OnRoundStart handles round start events.
func (m *AWDScoreMode) OnRoundStart(ctx context.Context, round int) error {
	logger.Info("AWD round start", "round", round)
	
	// Reset round-specific state if needed
	// First blood detector should track per-flag, so no reset needed
	
	return nil
}

// OnRoundEnd handles round end events and calculates defense scores.
func (m *AWDScoreMode) OnRoundEnd(ctx context.Context, round int) error {
	logger.Info("AWD round end", "round", round)

	if m.scorer == nil {
		return fmt.Errorf("scorer not initialized")
	}

	// Calculate defense scores for this round
	totalRounds := m.game.TotalRounds
	if totalRounds <= 0 {
		totalRounds = 10 // Default to 10 rounds
	}

	if err := m.scorer.OnDefense(ctx, round, totalRounds); err != nil {
		logger.Error("failed to calculate defense scores", "error", err)
		return err
	}

	// Validate zero-sum invariant
	if !m.scorer.ValidateZeroSum() {
		logger.Warn("zero-sum validation failed", "round", round)
		// Don't return error - just log warning
	}

	return nil
}

// OnAttack processes an attack event.
func (m *AWDScoreMode) OnAttack(ctx context.Context, attack *model.FlagSubmission) error {
	logger.Info("AWD attack recorded",
		"attacker", attack.AttackerTeam,
		"target", attack.TargetTeam,
		"correct", attack.IsCorrect,
	)

	if m.scorer == nil {
		return fmt.Errorf("scorer not initialized")
	}

	// Process attack through zero-sum scorer
	if err := m.scorer.OnAttack(ctx, attack); err != nil {
		logger.Error("failed to process attack", "error", err)
		return err
	}

	return nil
}

// OnDefense handles defense events (when a team's flag is captured).
func (m *AWDScoreMode) OnDefense(ctx context.Context, teamID int64, flag string) error {
	// Defense is handled in OnRoundEnd through zero-sum scorer
	// This method exists for interface compatibility
	logger.Debug("defense event recorded", "team", teamID, "flag", flag)
	return nil
}

// CalculateScore calculates and persists final scores for current round.
func (m *AWDScoreMode) CalculateScore(ctx context.Context) error {
	if m.scorer == nil {
		return fmt.Errorf("scorer not initialized")
	}

	if m.game == nil {
		return fmt.Errorf("game not set")
	}

	// Get current round
	round := m.game.CurrentRound
	if round <= 0 {
		round = 1
	}

	// Calculate and persist scores
	if err := m.scorer.CalculateScore(ctx, m.game.ID, round); err != nil {
		logger.Error("failed to calculate scores", "error", err)
		return err
	}

	// Log final scores
	scores := m.scorer.GetTeamScores()
	logger.Info("scores calculated", "round", round, "teams", len(scores))

	return nil
}

// Stop stops the scoring mode.
func (m *AWDScoreMode) Stop(ctx context.Context) error {
	logger.Info("AWD score mode stopped")

	// Final validation
	if m.scorer != nil && !m.scorer.ValidateZeroSum() {
		logger.Warn("final zero-sum validation failed")
	}

	return nil
}

// GetTeamScores returns current team scores.
func (m *AWDScoreMode) GetTeamScores() map[int64]float64 {
	if m.scorer == nil {
		return make(map[int64]float64)
	}
	return m.scorer.GetTeamScores()
}

// GetTeamScoreDetails returns detailed score breakdown for a team.
func (m *AWDScoreMode) GetTeamScoreDetails(teamID int64) *scoring.TeamScore {
	if m.scorer == nil {
		return nil
	}
	return m.scorer.GetTeamScoreDetails(teamID)
}

// ValidateZeroSum checks if zero-sum invariant holds.
func (m *AWDScoreMode) ValidateZeroSum() bool {
	if m.scorer == nil {
		return false
	}
	return m.scorer.ValidateZeroSum()
}

// AWDMixMode implements attack-defense + challenge solving.
type AWDMixMode struct {
	game       *model.Game
	config     *ScoringConfig
	awdScorer  *AWDScoreMode
	// TODO: Add challenge scoring component
}

func NewAWDMixMode() *AWDMixMode {
	return &AWDMixMode{
		config:    DefaultScoringConfig(),
		awdScorer: NewAWDScoreMode(),
	}
}

func (m *AWDMixMode) Start(ctx context.Context, game *model.Game) error {
	m.game = game
	if err := m.awdScorer.Start(ctx, game); err != nil {
		return err
	}
	logger.Info("AWD mix mode started", "game", game.Title)
	return nil
}

func (m *AWDMixMode) OnRoundStart(ctx context.Context, round int) error {
	return m.awdScorer.OnRoundStart(ctx, round)
}

func (m *AWDMixMode) OnRoundEnd(ctx context.Context, round int) error {
	return m.awdScorer.OnRoundEnd(ctx, round)
}

func (m *AWDMixMode) OnAttack(ctx context.Context, attack *model.FlagSubmission) error {
	return m.awdScorer.OnAttack(ctx, attack)
}

func (m *AWDMixMode) OnDefense(ctx context.Context, teamID int64, flag string) error {
	return m.awdScorer.OnDefense(ctx, teamID, flag)
}

func (m *AWDMixMode) CalculateScore(ctx context.Context) error {
	// TODO: Add challenge scoring calculation
	return m.awdScorer.CalculateScore(ctx)
}

func (m *AWDMixMode) Stop(ctx context.Context) error {
	return m.awdScorer.Stop(ctx)
}

// KingOfHillMode implements king-of-the-hill mode.
type KingOfHillMode struct {
	game *model.Game
}

func NewKingOfHillMode() *KingOfHillMode {
	return &KingOfHillMode{}
}

func (m *KingOfHillMode) Start(ctx context.Context, game *model.Game) error {
	m.game = game
	logger.Info("King of Hill mode started", "game", game.Title)
	return nil
}

func (m *KingOfHillMode) OnRoundStart(ctx context.Context, round int) error  { return nil }
func (m *KingOfHillMode) OnRoundEnd(ctx context.Context, round int) error    { return nil }
func (m *KingOfHillMode) OnAttack(ctx context.Context, attack *model.FlagSubmission) error { return nil }
func (m *KingOfHillMode) OnDefense(ctx context.Context, teamID int64, flag string) error  { return nil }
func (m *KingOfHillMode) CalculateScore(ctx context.Context) error { return nil }
func (m *KingOfHillMode) Stop(ctx context.Context) error           { return nil }
