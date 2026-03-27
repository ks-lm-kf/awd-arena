package container

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
)

// GormContainerStore implements ContainerStore using GORM.
type GormContainerStore struct{}

// NewGormContainerStore creates a new GORM store.
func NewGormContainerStore() *GormContainerStore {
	return &GormContainerStore{}
}

// Save persists a TeamContainer record.
func (s *GormContainerStore) Save(ctx context.Context, tc *model.TeamContainer) error {
	return database.GetDB().Create(tc).Error
}

// Delete removes a TeamContainer by container Docker ID.
func (s *GormContainerStore) Delete(ctx context.Context, id string) error {
	return database.GetDB().Where("container_id = ?", id).Delete(&model.TeamContainer{}).Error
}

// FindByGame returns all containers for a game.
func (s *GormContainerStore) FindByGame(ctx context.Context, gameID int64) ([]model.TeamContainer, error) {
	var containers []model.TeamContainer
	err := database.GetDB().Where("game_id = ?", gameID).Find(&containers).Error
	return containers, err
}

// FindByTeam returns all containers for a team in a game.
func (s *GormContainerStore) FindByTeam(ctx context.Context, gameID, teamID int64) ([]model.TeamContainer, error) {
	var containers []model.TeamContainer
	err := database.GetDB().Where("game_id = ? AND team_id = ?", gameID, teamID).Find(&containers).Error
	return containers, err
}

// UpdateStatus updates the status of a container.
func (s *GormContainerStore) UpdateStatus(ctx context.Context, id string, status string) error {
	return database.GetDB().Model(&model.TeamContainer{}).Where("container_id = ?", id).Update("status", status).Error
}

// UpdateIPAddress updates the IP address of a container.
func (s *GormContainerStore) UpdateIPAddress(ctx context.Context, containerID, ip string) error {
	return database.GetDB().Model(&model.TeamContainer{}).Where("container_id = ?", containerID).Update("ip_address", ip).Error
}

// FindByID returns a container by database ID.
func (s *GormContainerStore) FindByID(ctx context.Context, id int64) (*model.TeamContainer, error) {
	var tc model.TeamContainer
	err := database.GetDB().First(&tc, id).Error
	return &tc, err
}
