package engine

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// ScoreCalculator handles round scoring.
type ScoreCalculator struct {
	game *model.Game
}

func NewScoreCalculator(game *model.Game) *ScoreCalculator {
	return &ScoreCalculator{game: game}
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
	db.Where("game_id = ? AND round = ? AND is_correct = ?", sc.game.ID, round, true).Find(&subs)

	// Calculate attack scores per team
	attackScores := make(map[int64]float64)
	defenseLosses := make(map[int64]float64)
	for _, sub := range subs {
		attackScores[sub.AttackerTeam] += sub.PointsEarned
		defenseLosses[sub.TargetTeam] += sub.PointsEarned
	}

	attackWeight := sc.game.AttackWeight
	defenseWeight := sc.game.DefenseWeight

	// Create/update round scores
	for _, team := range teams {
		attack := attackScores[team.ID] * attackWeight
		defense := defenseLosses[team.ID] * defenseWeight
		total := attack - defense

		roundScore := model.RoundScore{
			GameID:       sc.game.ID,
			Round:        round,
			TeamID:       team.ID,
			AttackScore:  attack,
			DefenseScore: defense,
			TotalScore:   total,
		}

		// Upsert
		var existing model.RoundScore
		err := db.Where("game_id = ? AND round = ? AND team_id = ?", sc.game.ID, round, team.ID).First(&existing).Error
		if err != nil {
			db.Create(&roundScore)
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"attack_score":  attack,
				"defense_score": defense,
				"total_score":   total,
			})
		}
	}

	// Update cumulative team scores (sum of all rounds)
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
	db.Model(&model.RoundScore{}).
		Select("team_id, SUM(total_score) as total_score").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Find(&totals)

	// Add score adjustments
	type AdjTotal struct {
		TeamID     int64
		AdjTotal   int
	}
	var adjTotals []AdjTotal
	db.Model(&model.ScoreAdjustment{}).
		Select("team_id, SUM(adjust_value) as adj_total").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Find(&adjTotals)

	adjMap := make(map[int64]int)
	for _, a := range adjTotals {
		adjMap[a.TeamID] = a.AdjTotal
	}

	// Update each team's cumulative score
	for _, t := range totals {
		cumulative := t.TotalScore + float64(adjMap[t.TeamID])
		db.Model(&model.Team{}).Where("id = ?", t.TeamID).Update("score", cumulative)
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
	db.Model(&model.RoundScore{}).
		Select("team_id, SUM(total_score) as total_score").
		Where("game_id = ?", sc.game.ID).
		Group("team_id").
		Order("total_score desc").
		Find(&totals)

	// Update each team's rank based on cumulative score
	for i, t := range totals {
		rank := i + 1
		// Update the latest round score's rank
		var latestScore model.RoundScore
		if err := db.Where("game_id = ? AND team_id = ? AND round = ?", sc.game.ID, t.TeamID, sc.game.CurrentRound).First(&latestScore).Error; err == nil {
			db.Model(&latestScore).Update("rank", rank)
		}
	}

	return nil
}
