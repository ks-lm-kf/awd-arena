package service

import (
	"context"
	"errors"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/crypto"
)

// TeamService handles team management logic.
type TeamService struct{}

func (s *TeamService) CreateTeam(ctx context.Context, name string) (*model.Team, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	token, err := crypto.GenerateToken()
	if err != nil {
		return nil, err
	}

	team := &model.Team{Name: name, Token: token}
	if err := db.Create(team).Error; err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, id int64) (*model.Team, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var team model.Team
	err := db.First(&team, id).Error
	return &team, err
}

func (s *TeamService) ListTeams(ctx context.Context) ([]*model.Team, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var teams []*model.Team
	err := db.Order("score desc").Find(&teams).Error
	return teams, err
}

func (s *TeamService) GetTeamMembers(ctx context.Context, teamID int64) ([]model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var users []model.User
	err := db.Where("team_id = ?", teamID).Find(&users).Error
	return users, err
}

func (s *TeamService) UpdateScore(ctx context.Context, teamID int64, delta float64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Model(&model.Team{}).Where("id = ?", teamID).
		Update("score", db.Raw("COALESCE(score, 0) + ?", delta)).Error
}


func (s *TeamService) AddMember(ctx context.Context, teamID int64, userID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// Check if team exists
	var team model.Team
	if err := db.First(&team, teamID).Error; err != nil {
		return errors.New("team not found")
	}

	// Check if user exists
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Update user's team_id
	return db.Model(&user).Update("team_id", teamID).Error
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID int64, userID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	// Check if user exists and belongs to this team
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if user.TeamID == nil || *user.TeamID != teamID {
		return errors.New("user is not a member of this team")
	}

	// Remove user from team by setting team_id to NULL
	return db.Model(&user).Update("team_id", nil).Error
}

