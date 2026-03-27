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

	// Get all teams
	var teams []model.Team
	if err := db.Find(&teams).Error; err != nil {
		return err
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

		// Update team total score
		db.Model(&model.Team{}).Where("id = ?", team.ID).Update("score", total)
	}

	return sc.UpdateRankings(ctx, round)
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

	// Get round scores ordered by total
	var scores []model.RoundScore
	db.Where("game_id = ? AND round = ?", sc.game.ID, round).Order("total_score desc").Find(&scores)

	for i, s := range scores {
		db.Model(&model.RoundScore{}).Where("id = ?", s.ID).Update("rank", i+1)
	}
	return nil
}
