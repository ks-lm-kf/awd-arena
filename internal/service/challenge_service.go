package service

import (
	"context"
	"errors"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
)

// ChallengeService handles challenge operations.
type ChallengeService struct{}

func NewChallengeService() *ChallengeService {
	return &ChallengeService{}
}

func (s *ChallengeService) ListChallenges(ctx context.Context, gameID int64) ([]model.Challenge, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var challenges []model.Challenge
	err := db.Where("game_id = ?", gameID).Find(&challenges).Error
	return challenges, err
}

func (s *ChallengeService) CreateChallenge(ctx context.Context, ch *model.Challenge) (*model.Challenge, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	if ch.ImageTag == "" {
		ch.ImageTag = "latest"
	}
	if ch.Difficulty == "" {
		ch.Difficulty = "medium"
	}
	if ch.BaseScore == 0 {
		ch.BaseScore = 100
	}
	if ch.CPULimit == 0 {
		ch.CPULimit = 0.5
	}
	if ch.MemLimit == 0 {
		ch.MemLimit = 256
	}
	if err := db.Create(ch).Error; err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *ChallengeService) GetChallenge(ctx context.Context, id int64) (*model.Challenge, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var ch model.Challenge
	err := db.First(&ch, id).Error
	return &ch, err
}

func (s *ChallengeService) DeleteChallenge(ctx context.Context, id int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Delete(&model.Challenge{}, id).Error
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, id int64, updates *model.Challenge) (*model.Challenge, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	
	var ch model.Challenge
	if err := db.First(&ch, id).Error; err != nil {
		return nil, err
	}
	
	// Update fields
	if updates.Name != "" {
		ch.Name = updates.Name
	}
	if updates.Description != "" {
		ch.Description = updates.Description
	}
	if updates.ImageName != "" {
		ch.ImageName = updates.ImageName
	}
	if updates.ImageTag != "" {
		ch.ImageTag = updates.ImageTag
	}
	if updates.Difficulty != "" {
		ch.Difficulty = updates.Difficulty
	}
	if updates.BaseScore != 0 {
		ch.BaseScore = updates.BaseScore
	}
	if updates.ExposedPorts != "" {
		ch.ExposedPorts = updates.ExposedPorts
	}
	if updates.CPULimit != 0 {
		ch.CPULimit = updates.CPULimit
	}
	if updates.MemLimit != 0 {
		ch.MemLimit = updates.MemLimit
	}
	
	if err := db.Save(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

