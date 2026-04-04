package service

import (
	"context"
	"errors"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
)

// RankingItem represents a team in the ranking list.
type RankingItem struct {
	Rank         int     `json:"rank"`
	TeamID       int64   `json:"team_id"`
	TeamName     string  `json:"team_name"`
	TotalScore   float64 `json:"total_score"`
	AttackScore  float64 `json:"attack_score"`
	DefenseScore float64 `json:"defense_score"`
	FlagCount    int     `json:"flag_count"`
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
	if err := db.Joins("JOIN game_teams ON game_teams.team_id = teams.id").
		Where("game_teams.game_id = ?", gameID).
		Order("score desc").Find(&teams).Error; err != nil {
		return nil, err
	}

	// Calculate attack and defense scores from submissions
	attackScores := make(map[int64]float64)
	defenseLosses := make(map[int64]float64)
	flagCounts := make(map[int64]int)
	var subs []model.FlagSubmission
	db.Where("game_id = ? AND is_correct = ?", gameID, true).Find(&subs)
	for _, sub := range subs {
		attackScores[sub.AttackerTeam] += sub.PointsEarned
		defenseLosses[sub.TargetTeam] += sub.PointsEarned
		flagCounts[sub.AttackerTeam]++
	}

	items := make([]RankingItem, len(teams))
	rank := 1
	for i, t := range teams {
		if i > 0 && t.Score < teams[i-1].Score {
			rank = i + 1
		}
		items[i] = RankingItem{
			Rank:         rank,
			TeamID:       t.ID,
			TeamName:     t.Name,
			TotalScore:   t.Score,
			AttackScore:  attackScores[t.ID],
			DefenseScore: defenseLosses[t.ID],
			FlagCount:    flagCounts[t.ID],
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
			Rank:         rs.Rank,
			TeamID:       rs.TeamID,
			TeamName:     team.Name,
			TotalScore:   rs.TotalScore,
			AttackScore:  rs.AttackScore,
			DefenseScore: rs.DefenseScore,
			FlagCount:    0,
		}
	}
	return items, nil
}
