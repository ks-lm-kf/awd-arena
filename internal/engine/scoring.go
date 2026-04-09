package engine

import (
	"context"
	"fmt"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/engine/scoring"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// Default first blood bonus multiplier (e.g., 1.5x attack score for first blood)
const firstBloodBonusPoints = 50.0

// ScoreCalculator handles round scoring.
type ScoreCalculator struct {
	game               *model.Game
	firstBloodDetector *scoring.FirstBloodDetector
}

func NewScoreCalculator(game *model.Game) *ScoreCalculator {
	return &ScoreCalculator{
		game:               game,
		firstBloodDetector: scoring.NewFirstBloodDetector(),
	}
}

func (sc *ScoreCalculator) CalculateRoundScores(ctx context.Context, round int) error {
	logger.Info("calculating round scores", "round", round, "game_id", sc.game.ID)

	db := database.GetDB()
	if db == nil {
		return nil
	}

	// Get teams for this game via GameTeam table
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", sc.game.ID).Find(&gameTeams).Error; err != nil {
		return err
	}
	var teamIDs []int64
	for _, gt := range gameTeams {
		teamIDs = append(teamIDs, gt.TeamID)
	}
	var teams []model.Team
	if len(teamIDs) > 0 {
		if err := db.Where("id IN ?", teamIDs).Find(&teams).Error; err != nil {
			return err
		}
	}

	// Get correct submissions for this round
	var subs []model.FlagSubmission
	if err := db.Where("game_id = ? AND round = ? AND is_correct = ?", sc.game.ID, round, true).Find(&subs).Error; err != nil {
		logger.Error("failed to query correct flag submissions", "game_id", sc.game.ID, "round", round, "error", err)
	}

	// Calculate attack scores per team
	attackScores := make(map[int64]float64)
	defenseLosses := make(map[int64]float64)
	firstBloodBonus := make(map[int64]float64)
	for _, sub := range subs {
		attackScores[sub.AttackerTeam] += sub.PointsEarned
		defenseLosses[sub.TargetTeam] += sub.PointsEarned

		// Check for first blood: unique key per target team's flag
		flagKey := fmt.Sprintf("game%d_target%d_flag%s", sc.game.ID, sub.TargetTeam, sub.FlagHash)
		if sc.firstBloodDetector.CheckAndRecord(flagKey, sub.AttackerTeam, round) {
			firstBloodBonus[sub.AttackerTeam] += firstBloodBonusPoints
			logger.Info("first blood bonus awarded",
				"game_id", sc.game.ID,
				"round", round,
				"attacker_team", sub.AttackerTeam,
				"target_team", sub.TargetTeam,
				"bonus", firstBloodBonusPoints)
		}
	}

	attackWeight := sc.game.AttackWeight
	defenseWeight := sc.game.DefenseWeight

	// Create/update round scores
	for _, team := range teams {
		attack := attackScores[team.ID] * attackWeight
		defense := defenseLosses[team.ID] * defenseWeight
		bonus := firstBloodBonus[team.ID]
		total := attack - defense + bonus

		// Zero-sum base with first blood bonus on top

		roundScore := model.RoundScore{
			GameID:       sc.game.ID,
			Round:        round,
			TeamID:       team.ID,
			AttackScore:  attack,
			DefenseScore: defense,
			TotalScore:   total,
		}

		var existing model.RoundScore
		err := db.Where("game_id = ? AND round = ? AND team_id = ?", sc.game.ID, round, team.ID).First(&existing).Error
		if err != nil {
			if err := db.Create(&roundScore).Error; err != nil {
				logger.Error("failed to create round score", "game_id", sc.game.ID, "round", round, "team_id", team.ID, "error", err)
			}
		} else {
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"attack_score":  attack,
				"defense_score": defense,
				"total_score":   total,
			}).Error; err != nil {
				logger.Error("failed to update round score", "game_id", sc.game.ID, "round", round, "team_id", team.ID, "error", err)
			}
		}
	}

	return sc.UpdateCumulativeTeamScores(ctx)
}

// UpdateCumulativeTeamScores recalculates each team's total score from all rounds.
func (sc *ScoreCalculator) UpdateCumulativeTeamScores(ctx context.Context) error {
	db := database.GetDB()
	if db == nil {
		return nil
	}

	// Sum all round scores per team for this game
	type TeamTotal struct {
		TeamID     int64
		TotalScore float64
	}
	var totals []TeamTotal
	if err := db.Model(&model.RoundScore{}).
		Select("team_id, SUM(total_score) as total_score").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Find(&totals).Error; err != nil {
		logger.Error("failed to query cumulative team scores", "game_id", sc.game.ID, "error", err)
	}

	// Add score adjustments
	type AdjTotal struct {
		TeamID   int64
		AdjTotal int
	}
	var adjTotals []AdjTotal
	if err := db.Model(&model.ScoreAdjustment{}).
		Select("team_id, SUM(adjust_value) as adj_total").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Find(&adjTotals).Error; err != nil {
		logger.Error("failed to query score adjustments", "game_id", sc.game.ID, "error", err)
	}

	adjMap := make(map[int64]int)
	for _, a := range adjTotals {
		adjMap[a.TeamID] = a.AdjTotal
	}

	// Update each team's cumulative score
	for _, t := range totals {
		cumulative := t.TotalScore + float64(adjMap[t.TeamID])
		if err := db.Model(&model.Team{}).Where("id = ?", t.TeamID).Update("score", cumulative).Error; err != nil {
			logger.Error("failed to update team cumulative score", "team_id", t.TeamID, "score", cumulative, "error", err)
		}
	}

	return sc.UpdateRankings(ctx, 0)
}

func (sc *ScoreCalculator) CalculateTeamScore(teamID int64, round int) (*model.RoundScore, error) {
	db := database.GetDB()
	if db == nil {
		return &model.RoundScore{
			GameID: sc.game.ID,
			Round:  round,
			TeamID: teamID,
		}, nil
	}

	var score model.RoundScore
	err := db.Where("game_id = ? AND round = ? AND team_id = ?", sc.game.ID, round, teamID).First(&score).Error
	if err != nil {
		return &model.RoundScore{
			GameID: sc.game.ID,
			Round:  round,
			TeamID: teamID,
		}, nil
	}
	return &score, nil
}

func (sc *ScoreCalculator) UpdateRankings(ctx context.Context, round int) error {
	logger.Info("updating rankings", "round", round)
	db := database.GetDB()
	if db == nil {
		return nil
	}

	// Get cumulative scores for ranking (all rounds)
	type TeamTotal struct {
		TeamID     int64
		TotalScore float64
	}
	var totals []TeamTotal
	if err := db.Model(&model.RoundScore{}).
		Select("team_id, SUM(total_score) as total_score").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Order("total_score desc").
		Find(&totals).Error; err != nil {
		logger.Error("failed to query scores for ranking", "game_id", sc.game.ID, "error", err)
	}

	// Update each team's rank based on cumulative score
	for i, t := range totals {
		rank := i + 1
		var latestScore model.RoundScore
		if err := db.Where("game_id = ? AND team_id = ?", sc.game.ID, t.TeamID).Order("round desc").First(&latestScore).Error; err == nil {
			if err := db.Model(&latestScore).Update("rank", rank).Error; err != nil {
				logger.Error("failed to update team rank", "team_id", t.TeamID, "rank", rank, "error", err)
			}
		}
	}

	return nil
}
