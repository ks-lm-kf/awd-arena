package service

import (
	"context"
	"errors"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
)

// RankingItem represents a team in the ranking list.
type RankingItem struct {
	Rank       int     `json:"rank"`
	TeamID     int64   `json:"team_id"`
	TeamName   string  `json:"team_name"`
	Score      float64 `json:"score"`
	AttackScore float64 `json:"attack_score"`
	DefenseLoss float64 `json:"defense_loss"`
}

// RankingService handles ranking calculations.
type RankingService struct{}

// GetRankings returns current game rankings.
func (s *RankingService) GetRankings(ctx context.Context, gameID int64) ([]RankingItem, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var teams []model.Team
	if err := db.Order("score desc").Find(&teams).Error; err != nil {
		return nil, err
	}

	// Calculate attack and defense scores from submissions
	attackScores := make(map[int64]float64)
	defenseLosses := make(map[int64]float64)
	var subs []model.FlagSubmission
	db.Where("game_id = ? AND is_correct = ?", gameID, true).Find(&subs)
	for _, sub := range subs {
		attackScores[sub.AttackerTeam] += sub.PointsEarned
		defenseLosses[sub.TargetTeam] += sub.PointsEarned
	}

	items := make([]RankingItem, len(teams))
	for i, t := range teams {
		items[i] = RankingItem{
			Rank:        i + 1,
			TeamID:      t.ID,
			TeamName:    t.Name,
			Score:       t.Score,
			AttackScore: attackScores[t.ID],
			DefenseLoss: defenseLosses[t.ID],
		}
	}
	return items, nil
}

// GetRoundRankings returns rankings for a specific round.
func (s *RankingService) GetRoundRankings(ctx context.Context, gameID int64, round int) ([]RankingItem, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var roundScores []model.RoundScore
	query := db.Where("game_id = ?", gameID)
	if round > 0 {
		query = query.Where("round = ?", round)
	}
	if err := query.Order("total_score desc").Find(&roundScores).Error; err != nil {
		return nil, err
	}

	items := make([]RankingItem, len(roundScores))
	for i, rs := range roundScores {
		var team model.Team
		db.First(&team, rs.TeamID)
		items[i] = RankingItem{
			Rank:        rs.Rank,
			TeamID:      rs.TeamID,
			TeamName:    team.Name,
			Score:       rs.TotalScore,
			AttackScore: rs.AttackScore,
			DefenseLoss: rs.DefenseScore,
		}
	}
	return items, nil
}
