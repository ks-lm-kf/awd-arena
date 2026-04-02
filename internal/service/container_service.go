package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/container"
	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/network"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"gorm.io/gorm"
)

var (
	globalOnce   sync.Once
	globalMgr    *container.ContainerManager
	globalNetMgr *network.NetworkManager
	initErr      error
)

// initContainerManager initializes the global Docker client, container manager, and network manager.
func initContainerManager() {
	dockerClient, err := container.NewDockerClientImpl()
	if err != nil {
		initErr = err
		return
	}
	store := container.NewGormContainerStore()
	globalMgr = container.NewContainerManager(dockerClient, store)
	globalNetMgr = network.NewNetworkManager(dockerClient)
}

// getManager returns the global ContainerManager.
func getManager() (*container.ContainerManager, error) {
	globalOnce.Do(initContainerManager)
	if initErr != nil {
		return nil, initErr
	}
	if globalMgr == nil {
		return nil, errors.New("container manager not initialized")
	}
	return globalMgr, nil
}

// getNetManager returns the global NetworkManager.
func getNetManager() (*network.NetworkManager, error) {
	globalOnce.Do(initContainerManager)
	if initErr != nil {
		return nil, initErr
	}
	if globalNetMgr == nil {
		return nil, errors.New("network manager not initialized")
	}
	return globalNetMgr, nil
}

// ContainerService handles container lifecycle management.
type ContainerService struct{}

// ContainerInfo represents container status info.
type ContainerInfo struct {
	ID            int64  `json:"id"`
	TeamID        int64  `json:"team_id"`
	TeamName      string `json:"team_name"`
	ChallengeID   int64  `json:"challenge_id"`
	ChallengeName string `json:"challenge_name"`
	ContainerID   string `json:"container_id"`
	IPAddress     string `json:"ip_address"`
	PortMapping   string `json:"port_mapping"`
	Status        string `json:"status"`
}

// ContainerStatsInfo represents container stats with team/challenge info.
type ContainerStatsInfo struct {
	ContainerID  string  `json:"container_id"`
	TeamID       int64   `json:"team_id"`
	ChallengeID  int64   `json:"challenge_id"`
	IPAddress    string  `json:"ip_address"`
	Status       string  `json:"status"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemoryBytes  uint64  `json:"memory_bytes"`
	NetworkBytes uint64  `json:"network_bytes"`
}

// NewContainerService creates a new ContainerService instance.
func NewContainerService() *ContainerService {
	return &ContainerService{}
}

// ProvisionContainers creates all containers for a game's teams and challenges.
func (s *ContainerService) ProvisionContainers(ctx context.Context, gameID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

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

	var challenges []model.Challenge
	if err := db.Where("game_id = ?", gameID).Find(&challenges).Error; err != nil {
		return err
	}

	if len(teams) == 0 || len(challenges) == 0 {
		logger.Info("no teams or challenges, skipping provisioning", "teams", len(teams), "challenges", len(challenges))
		return nil
	}

	mgr, err := getManager()
	if err != nil {
		return err
	}

	netMgr, err := getNetManager()
	if err != nil {
		return err
	}

	for _, team := range teams {
		if _, err := netMgr.CreateTeamNetwork(ctx, team.ID); err != nil {
			logger.Error("create team network failed", "team", team.ID, "error", err)
			continue
		}
	}

	hostPortBase := 31000
	totalContainers := len(teams) * len(challenges)
	var failedCount int
	for _, team := range teams {
		netName, _ := netMgr.GetTeamNetwork(team.ID)
		chalIdx := 0
		for _, chal := range challenges {
			ip := netMgr.GetTeamIP(team.ID, chalIdx)
			portBase := hostPortBase + (len(challenges)*(int(team.ID)-1)+chalIdx)*10

			tc, err := mgr.CreateChallengeContainer(ctx, team.ID, team.Name, &chal, gameID, netName, ip, portBase)
			if err != nil {
				failedCount++
				logger.Error("create container failed", "team", team.ID, "challenge", chal.ID, "error", err)
				chalIdx++
				continue
			}
			logger.Info("container provisioned", "team", team.ID, "challenge", chal.ID, "ip", tc.IPAddress)
			chalIdx++
		}
	}

	var isoTeamIDs []int64
	for _, t := range teams {
		isoTeamIDs = append(isoTeamIDs, t.ID)
	}
	_ = netMgr.IsolateTeams(ctx, isoTeamIDs)

	if failedCount > 0 {
		_ = eventbus.GetBus().Publish(ctx, "container:creation_failed", map[string]interface{}{
			"game_id": gameID,
			"failed":  failedCount,
			"total":   totalContainers,
		})
		if failedCount >= totalContainers {
			return errors.New("all container creations failed")
		}
		logger.Warn("partial container creation failure", "failed", failedCount, "total", totalContainers)
	}

	logger.Info("provisioning complete", "teams", len(teams), "challenges", len(challenges))
	return nil
}

// TeardownContainers stops and removes all containers for a game.
func (s *ContainerService) TeardownContainers(ctx context.Context, gameID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	if err := mgr.CleanupAll(ctx, gameID); err != nil {
		return err
	}

	var teamIDs []int64
	db.Model(&model.TeamContainer{}).Where("game_id = ?", gameID).Distinct("team_id").Pluck("team_id", &teamIDs)

	netMgr, err := getNetManager()
	if err != nil {
		return err
	}
	for _, tid := range teamIDs {
		_ = netMgr.RemoveTeamNetwork(ctx, tid)
	}

	return nil
}

// CleanupTeamContainers removes all Docker containers for a specific team.
func (s *ContainerService) CleanupTeamContainers(ctx context.Context, teamID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var containers []model.TeamContainer
	if err := db.Where("team_id = ?", teamID).Find(&containers).Error; err != nil {
		return err
	}

	for _, tc := range containers {
		if err := mgr.DestroyContainer(ctx, tc.ContainerID); err != nil {
			logger.Error("failed to destroy container during team cleanup", "container_id", tc.ContainerID, "error", err)
		}
	}

	if len(containers) > 0 {
		logger.Info("cleaned up containers for team", "team_id", teamID, "count", len(containers))
	}

	return nil
}

// PauseContainers pauses all containers for a game.
func (s *ContainerService) PauseContainers(ctx context.Context, gameID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	return mgr.PauseAll(ctx, gameID)
}

// ResumeContainers resumes all paused containers for a game.
func (s *ContainerService) ResumeContainers(ctx context.Context, gameID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	return mgr.UnpauseAll(ctx, gameID)
}

// RestartAll restarts all containers for a game.
func (s *ContainerService) RestartAll(ctx context.Context, gameID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	return mgr.BulkRestart(ctx, gameID)
}

// RestartContainer restarts a specific container and deducts score from the team.
func (s *ContainerService) RestartContainer(ctx context.Context, gameID, containerID, operatorID int64) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}

	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var tc model.TeamContainer
	if err := db.Where("game_id = ? AND id = ?", gameID, containerID).First(&tc).Error; err != nil {
		return errors.New("container not found")
	}

	// Restart the container
	if err := mgr.RestartContainer(ctx, tc.ContainerID); err != nil {
		return err
	}

	// Get current round from game model
	currentRound := 0
	var game model.Game
	if err := db.Select("current_round").Where("id = ?", gameID).First(&game).Error; err == nil {
		currentRound = game.CurrentRound
	}

	// Deduct score: -50 points for container restart
	adjustment := model.ScoreAdjustment{
		GameID:      gameID,
		TeamID:      tc.TeamID,
		AdjustValue: -50,
		Reason:      "容器重启",
		OperatorID:  operatorID,
		Round:       currentRound,
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&adjustment).Error; err != nil {
		logger.Error("failed to create score adjustment for container restart", "error", err)
	}

	penalty := -50
	db.Model(&model.Team{}).Where("id = ?", tc.TeamID).Update("score", gorm.Expr("score + ?", penalty))

	return nil
}

// GetContainers returns container list for a game with team and challenge names.
func (s *ContainerService) GetContainers(ctx context.Context, gameID int64) ([]ContainerInfo, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var containers []model.TeamContainer
	if err := db.Where("game_id = ?", gameID).Find(&containers).Error; err != nil {
		return nil, err
	}

	teamIDs := make([]int64, len(containers))
	challengeIDs := make([]int64, len(containers))
	for i, c := range containers {
		teamIDs[i] = c.TeamID
		challengeIDs[i] = c.ChallengeID
	}

	var teams []model.Team
	teamMap := make(map[int64]string)
	if err := db.Where("id IN ?", teamIDs).Find(&teams).Error; err == nil {
		for _, t := range teams {
			teamMap[t.ID] = t.Name
		}
	}

	var challenges []model.Challenge
	challengeMap := make(map[int64]string)
	if err := db.Where("id IN ?", challengeIDs).Find(&challenges).Error; err == nil {
		for _, c := range challenges {
			challengeMap[c.ID] = c.Name
		}
	}

	items := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		items[i] = ContainerInfo{
			ID:            c.ID,
			TeamID:        c.TeamID,
			TeamName:      teamMap[c.TeamID],
			ChallengeID:   c.ChallengeID,
			ChallengeName: challengeMap[c.ChallengeID],
			ContainerID:   c.ContainerID,
			IPAddress:     c.IPAddress,
			PortMapping:   c.PortMapping,
			Status:        c.Status,
		}
	}
	return items, nil
}

// GetStats returns container statistics for a game with real Docker stats.
func (s *ContainerService) GetStats(ctx context.Context, gameID int64) ([]ContainerStatsInfo, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var containers []model.TeamContainer
	if err := db.Where("game_id = ?", gameID).Find(&containers).Error; err != nil {
		return nil, err
	}

	mgr, err := getManager()
	if err != nil {
		result := make([]ContainerStatsInfo, len(containers))
		for i, c := range containers {
			result[i] = ContainerStatsInfo{
				ContainerID: c.ContainerID,
				TeamID:      c.TeamID,
				ChallengeID: c.ChallengeID,
				IPAddress:   c.IPAddress,
				Status:      c.Status,
			}
		}
		return result, nil
	}

	stats, err := mgr.MonitorStats(ctx, gameID)
	if err != nil {
		logger.Error("monitor stats failed", "error", err)
	}

	statsMap := make(map[string]*container.ContainerStats)
	if stats != nil {
		for i := range stats {
			statsMap[stats[i].ID] = &stats[i]
		}
	}

	result := make([]ContainerStatsInfo, len(containers))
	for i, c := range containers {
		info := ContainerStatsInfo{
			ContainerID: c.ContainerID,
			TeamID:      c.TeamID,
			ChallengeID: c.ChallengeID,
			IPAddress:   c.IPAddress,
			Status:      c.Status,
		}
		if st, ok := statsMap[c.ContainerID]; ok {
			info.CPUPercent = st.CPU
			info.MemoryBytes = st.Memory
			info.NetworkBytes = st.Network
		}
		result[i] = info
	}
	return result, nil
}
