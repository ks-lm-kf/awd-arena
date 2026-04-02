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

	fields := map[string]interface{}{}
	if updates.Name != "" {
		fields["name"] = updates.Name
	}
	if updates.Description != "" {
		fields["description"] = updates.Description
	}
	if updates.ImageName != "" {
		fields["image_name"] = updates.ImageName
	}
	if updates.ImageTag != "" {
		fields["image_tag"] = updates.ImageTag
	}
	if updates.Difficulty != "" {
		fields["difficulty"] = updates.Difficulty
	}
	if updates.BaseScore != 0 {
		fields["base_score"] = updates.BaseScore
	}
	if updates.ExposedPorts != "" {
		fields["exposed_ports"] = updates.ExposedPorts
	}
	if updates.CPULimit != 0 {
		fields["cpu_limit"] = updates.CPULimit
	}
	if updates.MemLimit != 0 {
		fields["mem_limit"] = updates.MemLimit
	}

	if len(fields) == 0 {
		return &ch, nil
	}

	if err := db.Model(&ch).Updates(fields).Error; err != nil {
		return nil, err
	}

	if err := db.First(&ch, id).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}
