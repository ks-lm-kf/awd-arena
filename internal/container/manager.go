package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// ContainerManager manages Docker container lifecycle.
type ContainerManager struct {
	client DockerClient
	store  ContainerStore
	mu     sync.Mutex
}

// DockerClient abstracts Docker SDK operations.
type DockerClient interface {
	CreateContainer(ctx context.Context, opts CreateOptions) (string, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string) error
	InspectContainer(ctx context.Context, id string) (*ContainerInfo, error)
	ListContainers(ctx context.Context, filters map[string]string) ([]ContainerInfo, error)
	Stats(ctx context.Context, id string) (*ContainerStats, error)
	PauseContainer(ctx context.Context, id string) error
	UnpauseContainer(ctx context.Context, id string) error
	Logs(ctx context.Context, id string, opts LogOptions) (string, error)
	LogsStream(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)
	ExecCreate(ctx context.Context, id string, cmd []string) (string, error)
	ExecStart(ctx context.Context, execID string) (string, error)
}

// ContainerStore abstracts container persistence.
type ContainerStore interface {
	Save(ctx context.Context, tc *model.TeamContainer) error
	Delete(ctx context.Context, id string) error
	FindByGame(ctx context.Context, gameID int64) ([]model.TeamContainer, error)
	FindByTeam(ctx context.Context, gameID, teamID int64) ([]model.TeamContainer, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

// CreateOptions holds container creation parameters.
type CreateOptions struct {
	Image       string
	Name        string
	Cmd         []string
	Env         []string
	NetworkID   string
	IPAddress   string
	Ports       map[int]int // container port -> host port
	Resources   ResourceLimit
	AutoRemove  bool
}

// ContainerInfo holds container inspection data.
type ContainerInfo struct {
	ID     string
	Name   string
	Status string
	IP     string
	Image  string
}

// ContainerStats holds resource usage.
type ContainerStats struct {
	ID      string  `json:"id"`
	CPU     float64 `json:"cpu_percent"`
	Memory  uint64  `json:"memory_bytes"`
	Network uint64  `json:"network_bytes"`
	BlockIO uint64  `json:"block_io_bytes"`
	PIDs    int     `json:"pids"`
}

// NewContainerManager creates a new manager.
func NewContainerManager(client DockerClient, store ContainerStore) *ContainerManager {
	return &ContainerManager{
		client: client,
		store:  store,
	}
}

// CreateChallengeContainer creates a container for a team's challenge.
func (m *ContainerManager) CreateChallengeContainer(ctx context.Context, teamID int64, teamName string, challenge *model.Challenge, gameID int64, networkID string, ipAddress string, hostPortBase int) (*model.TeamContainer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	imageRef := challenge.ImageName
	if challenge.ImageTag != "" && challenge.ImageTag != "latest" {
		imageRef = imageRef + ":" + challenge.ImageTag
	} else {
		imageRef = imageRef + ":latest"
	}

	containerName := fmt.Sprintf("awd-game%d-team%d-chal%d", gameID, teamID, challenge.ID)

	limits := ResourceLimit{
		CPUCores:   challenge.CPULimit,
		MemoryMB:   int64(challenge.MemLimit),
		DiskGB:     1,
		NetworkBPS: 10 * 1024 * 1024,
		PidsLimit:  64,
	}
	if limits.CPUCores <= 0 {
		limits.CPUCores = 0.5
	}
	if limits.MemoryMB <= 0 {
		limits.MemoryMB = 256
	}

	// Parse exposed ports
	ports := make(map[int]int)
	portStr := challenge.ExposedPorts
	if portStr != "" {
		// Try JSON array parse or comma-separated
		for i, p := range splitPorts(portStr) {
			ports[p] = hostPortBase + i
		}
	}

	opts := CreateOptions{
		Image:      imageRef,
		Name:       containerName,
		NetworkID:  networkID,
		IPAddress:  ipAddress,
		Ports:      ports,
		Resources:  limits,
		AutoRemove: false,
		Env: []string{
			fmt.Sprintf("AWD_TEAM_ID=%d", teamID),
			fmt.Sprintf("AWD_GAME_ID=%d", gameID),
			fmt.Sprintf("AWD_CHALLENGE_ID=%d", challenge.ID),
			fmt.Sprintf("AWD_TEAM_NAME=%s", teamName),
		},
	}

	containerID, err := m.client.CreateContainer(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("create container %s: %w", containerName, err)
	}

	if err := m.client.StartContainer(ctx, containerID); err != nil {
		m.client.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("start container %s: %w", containerName, err)
	}

	// Get actual IP from Docker
	info, err := m.client.InspectContainer(ctx, containerID)
	if err == nil && info.IP != "" {
		ipAddress = info.IP
	}

	// Generate SSH password before saving
	sshPassword := generateRandomPassword(12)

	// Build port mapping JSON
	portMappingJSON := "{}"
	if len(ports) > 0 {
		pm := make(map[string]int)
		for cp, hp := range ports {
			pm[fmt.Sprintf("%d", cp)] = hp
		}
		if pmBytes, err := json.Marshal(pm); err == nil {
			portMappingJSON = string(pmBytes)
		}
	}

	tc := &model.TeamContainer{
		GameID:      gameID,
		TeamID:      teamID,
		ChallengeID: challenge.ID,
		ContainerID: containerID,
		IPAddress:   ipAddress,
		PortMapping: portMappingJSON,
		Status:      "running",
		SSHUser:     "awd",
		SSHPassword: sshPassword,
	}

	if err := m.store.Save(ctx, tc); err != nil {
		m.client.StopContainer(ctx, containerID)
		m.client.RemoveContainer(ctx, containerID)
		return nil, fmt.Errorf("save container record: %w", err)
	}


	// === 创建SSH用户 ===
	sshCreateCmd := []string{
		"sh", "-c",
		fmt.Sprintf("useradd -m -s /bin/bash awd && echo 'awd:%s' | chpasswd && usermod -aG sudo awd", sshPassword),
	}
	
	execID, sshErr := m.client.ExecCreate(ctx, containerID, sshCreateCmd)
	if sshErr != nil {
		logger.Error("failed to create exec for SSH user", "container", containerID, "error", sshErr)
	} else {
		_, sshErr := m.client.ExecStart(ctx, execID)
		if sshErr != nil {
			logger.Error("failed to execute SSH user creation", "container", containerID, "error", sshErr)
		} else {
			logger.Info("SSH user created successfully", "container", containerID, "user", "awd", "password", sshPassword)
		}
	}
	// === SSH用户创建结束 ===

	logger.Info("container created", "name", containerName, "ip", ipAddress, "team", teamID, "challenge", challenge.ID)
	return tc, nil
}

// DestroyContainer stops and removes a container.
func (m *ContainerManager) DestroyContainer(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.client.StopContainer(ctx, containerID); err != nil {
		logger.Error("failed to stop container", "id", containerID, "error", err)
		// 继续尝试移除容器
	}

	if err := m.client.RemoveContainer(ctx, containerID); err != nil {
		logger.Error("failed to remove container", "id", containerID, "error", err)
		return fmt.Errorf("remove container %s: %w", containerID, err)
	}

	if err := m.store.Delete(ctx, containerID); err != nil {
		logger.Error("failed to delete container record from store", "id", containerID, "error", err)
		// 不返回错误，因为容器已经被移除
	}

	logger.Info("container destroyed", "id", containerID[:12])
	return nil
}

// RestartContainer restarts a single container.
func (m *ContainerManager) RestartContainer(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.store.UpdateStatus(ctx, containerID, "restarting")
	if err := m.client.StopContainer(ctx, containerID); err != nil {
		logger.Error("stop container failed", "id", containerID, "error", err)
	}
	if err := m.client.StartContainer(ctx, containerID); err != nil {
		_ = m.store.UpdateStatus(ctx, containerID, "error")
		return fmt.Errorf("start container: %w", err)
	}
	_ = m.store.UpdateStatus(ctx, containerID, "running")
	return nil
}

// BulkRestart restarts all containers for a game.
func (m *ContainerManager) BulkRestart(ctx context.Context, gameID int64) error {
	containers, err := m.store.FindByGame(ctx, gameID)
	if err != nil {
		return err
	}
	for _, tc := range containers {
		if err := m.RestartContainer(ctx, tc.ContainerID); err != nil {
			logger.Error("bulk restart failed for container", "id", tc.ContainerID, "error", err)
		}
	}
	return nil
}

// PauseAll pauses all containers for a game.
func (m *ContainerManager) PauseAll(ctx context.Context, gameID int64) error {
	containers, err := m.store.FindByGame(ctx, gameID)
	if err != nil {
		return err
	}
	for _, tc := range containers {
		if err := m.client.PauseContainer(ctx, tc.ContainerID); err != nil {
			logger.Error("pause container failed", "id", tc.ContainerID, "error", err)
			continue
		}
		_ = m.store.UpdateStatus(ctx, tc.ContainerID, "paused")
	}
	logger.Info("all containers paused", "game", gameID, "count", len(containers))
	return nil
}

// UnpauseAll resumes all containers for a game.
func (m *ContainerManager) UnpauseAll(ctx context.Context, gameID int64) error {
	containers, err := m.store.FindByGame(ctx, gameID)
	if err != nil {
		return err
	}
	for _, tc := range containers {
		if err := m.client.UnpauseContainer(ctx, tc.ContainerID); err != nil {
			logger.Error("unpause container failed", "id", tc.ContainerID, "error", err)
			continue
		}
		_ = m.store.UpdateStatus(ctx, tc.ContainerID, "running")
	}
	logger.Info("all containers resumed", "game", gameID, "count", len(containers))
	return nil
}

// CleanupAll stops and removes all containers for a game.
func (m *ContainerManager) CleanupAll(ctx context.Context, gameID int64) error {
	containers, err := m.store.FindByGame(ctx, gameID)
	if err != nil {
		return err
	}
	for _, tc := range containers {
		_ = m.DestroyContainer(ctx, tc.ContainerID)
	}
	logger.Info("all containers cleaned up", "game", gameID, "count", len(containers))
	return nil
}

// MonitorStats returns resource usage for all containers of a game.
func (m *ContainerManager) MonitorStats(ctx context.Context, gameID int64) ([]ContainerStats, error) {
	containers, err := m.store.FindByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	var stats []ContainerStats
	for _, tc := range containers {
		s, err := m.client.Stats(ctx, tc.ContainerID)
		if err != nil {
			logger.Error("stats failed", "container", tc.ContainerID, "error", err)
			continue
		}
		s.ID = tc.ContainerID
		stats = append(stats, *s)
	}
	return stats, nil
}

// GetContainerLogs retrieves logs for a container.
func (m *ContainerManager) GetContainerLogs(ctx context.Context, containerID string, opts LogOptions) (string, error) {
	return m.client.Logs(ctx, containerID, opts)
}

// StreamContainerLogs streams logs for a container.
func (m *ContainerManager) StreamContainerLogs(ctx context.Context, containerID string, opts LogOptions) (io.ReadCloser, error) {
	return m.client.LogsStream(ctx, containerID, opts)
}

// ExecInContainer executes a command in a running container.
func (m *ContainerManager) ExecInContainer(ctx context.Context, containerID string, cmd []string) (string, error) {
	execID, err := m.client.ExecCreate(ctx, containerID, cmd)
	if err != nil {
		return "", fmt.Errorf("create exec: %w", err)
	}
	output, err := m.client.ExecStart(ctx, execID)
	if err != nil {
		return "", fmt.Errorf("start exec: %w", err)
	}
	return output, nil
}

// splitPorts parses port configuration from string.
func splitPorts(s string) []int {
	// Handle JSON array or comma-separated integers
	var ports []int
	// Simple comma-separated parsing
	for _, part := range splitString(s, ',') {
		p, err := strconv.Atoi(trimSpace(part))
		if err == nil && p > 0 && p < 65536 {
			ports = append(ports, p)
		}
	}
	return ports
}

func splitString(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}


// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
