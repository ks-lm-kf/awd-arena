package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/crypto"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

const BaseAttackPoints = 10

type FlagService struct{}

type FlagHistoryItem struct {
	ID           int64     `json:"id"`
	AttackerTeam int64     `json:"attacker_team"`
	TargetTeam   int64     `json:"target_team"`
	FlagValue    string    `json:"flag_value"`
	IsCorrect    bool      `json:"is_correct"`
	PointsEarned float64   `json:"points_earned"`
	Round        int       `json:"round"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

func (s *FlagService) GenerateFlags(ctx context.Context, gameID int64, round int) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var game model.Game
	if err := db.First(&game, gameID).Error; err != nil {
		return err
	}

	var challenges []model.Challenge
	if err := db.Where("game_id = ?", gameID).Find(&challenges).Error; err != nil {
		return err
	}

	// Get teams for this game via GameTeam table
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Find(&gameTeams).Error; err != nil {
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

	for _, team := range teams {
		for _, ch := range challenges {
			// Use consistent format: flag{gameID_round_teamID_challengeID_random}
			randomHex, _ := crypto.GenerateRandomHex(16)
			flagValue := fmt.Sprintf("flag{%d_%d_%d_%d_%s}", gameID, round, team.ID, ch.ID, randomHex)
			record := model.FlagRecord{
				GameID:    gameID,
				Round:     round,
				TeamID:    team.ID,
				FlagHash:  crypto.SHA256Hex(flagValue),
				FlagValue: flagValue,
				Service:   ch.Name,
				CreatedAt: time.Now(),
			}
			db.Create(&record)
			logger.Info("flag generated", "team", team.ID, "service", ch.Name, "round", round)
		}
	}
	return nil
}

func (s *FlagService) SubmitFlag(ctx context.Context, gameID, round, attackerTeam int64, flagValue string) (bool, float64, error) {
	db := database.GetDB()
	if db == nil {
		return false, 0, errors.New("database not initialized")
	}

	flagHash := crypto.SHA256Hex(flagValue)

	var flagRecord model.FlagRecord
	err := db.Where("game_id = ? AND round = ? AND flag_hash = ? AND flag_value = ?",
		gameID, round, flagHash, flagValue).First(&flagRecord).Error
	if err != nil {
		submission := model.FlagSubmission{
			GameID:       gameID,
			Round:        int(round),
			AttackerTeam: attackerTeam,
			TargetTeam:   0,
			FlagValue:    flagValue,
			IsCorrect:    false,
			PointsEarned: 0,
			SubmittedAt:  time.Now(),
		}
		db.Create(&submission)
		return false, 0, nil
	}

	if flagRecord.TeamID == attackerTeam {
		return false, 0, nil
	}

	var existing model.FlagSubmission
	err = db.Where("game_id = ? AND round = ? AND attacker_team = ? AND target_team = ? AND is_correct = ?",
		gameID, round, attackerTeam, flagRecord.TeamID, true).First(&existing).Error
	if err == nil {
		return true, 0, nil
	}

	var game model.Game
	db.First(&game, gameID)
	points := float64(BaseAttackPoints) * game.AttackWeight

	submission := model.FlagSubmission{
		GameID:       gameID,
		Round:        int(round),
		AttackerTeam: attackerTeam,
		TargetTeam:   flagRecord.TeamID,
		FlagValue:    flagValue,
		IsCorrect:    true,
		PointsEarned: points,
		SubmittedAt:  time.Now(),
	}
	db.Create(&submission)

	db.Model(&model.Team{}).Where("id = ?", attackerTeam).
		Update("score", db.Raw("COALESCE(score, 0) + ?", points))

	// WebSocket push
	eventbus.BroadcastJSON("flag:captured", map[string]interface{}{
		"game_id":       gameID,
		"round":         round,
		"attacker_team": attackerTeam,
		"target_team":   flagRecord.TeamID,
		"points":        points,
	})

	return true, points, nil
}

func (s *FlagService) GetFlagHistory(ctx context.Context, gameID int64, round int) ([]FlagHistoryItem, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	query := db.Model(&model.FlagSubmission{}).Where("game_id = ?", gameID)
	if round > 0 {
		query = query.Where("round = ?", round)
	}

	var submissions []model.FlagSubmission
	if err := query.Order("submitted_at desc").Find(&submissions).Error; err != nil {
		return nil, err
	}

	items := make([]FlagHistoryItem, len(submissions))
	for i, sub := range submissions {
		items[i] = FlagHistoryItem{
			ID:           sub.ID,
			AttackerTeam: sub.AttackerTeam,
			TargetTeam:   sub.TargetTeam,
			FlagValue:    sub.FlagValue,
			IsCorrect:    sub.IsCorrect,
			PointsEarned: sub.PointsEarned,
			Round:        sub.Round,
			SubmittedAt:  sub.SubmittedAt,
		}
	}
	return items, nil
}

func (s *FlagService) GetCurrentRoundFlags(ctx context.Context, gameID int64, round int) ([]model.FlagRecord, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var records []model.FlagRecord
	err := db.Where("game_id = ? AND round = ?", gameID, round).Find(&records).Error
	return records, err
}
