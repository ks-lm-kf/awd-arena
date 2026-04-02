package engine

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"

	"github.com/docker/docker/client"
)

// HealthChecker periodically checks container health for a running game.
type HealthChecker struct {
	mu           sync.Mutex
	gameID       int64
	dockerClient *client.Client
	cancelFunc   context.CancelFunc
	running      bool

	// Track previous statuses to detect changes
	prevStatus map[int64]string // container DB ID -> status

	// Consecutive failure counts for each container
	failCount map[int64]int
}

// NewHealthChecker creates a new health checker for a game.
func NewHealthChecker(gameID int64, dockerClient *client.Client) *HealthChecker {
	return &HealthChecker{
		gameID:       gameID,
		dockerClient: dockerClient,
		prevStatus:   make(map[int64]string),
		failCount:    make(map[int64]int),
	}
}

// Start begins the health check loop (every 30 seconds).
func (hc *HealthChecker) Start(ctx context.Context) {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	ctx, hc.cancelFunc = context.WithCancel(ctx)
	hc.running = true
	hc.mu.Unlock()

	go hc.run(ctx)
	logger.Info("health checker started", "game_id", hc.gameID)
}

// Stop stops the health check loop.
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	if hc.cancelFunc != nil {
		hc.cancelFunc()
	}
	hc.running = false
	logger.Info("health checker stopped", "game_id", hc.gameID)
}

func (hc *HealthChecker) run(ctx context.Context) {
	// Do first check immediately
	hc.checkAll(ctx)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkAll(ctx)
		}
	}
}

func (hc *HealthChecker) checkAll(ctx context.Context) {
	db := database.GetDB()
	if db == nil {
		return
	}

	var containers []model.TeamContainer
	if err := db.Where("game_id = ?", hc.gameID).Find(&containers).Error; err != nil {
		logger.Error("health check: failed to query containers", "error", err)
		return
	}

	for i := range containers {
		hc.checkContainer(ctx, &containers[i])
	}
}

func (hc *HealthChecker) checkContainer(ctx context.Context, tc *model.TeamContainer) {
	db := database.GetDB()

	newStatus := "running"
	var responseTime int64
	var errMsg string

	// 1. Check via Docker API
	if hc.dockerClient != nil && tc.ContainerID != "" {
		inspect, err := hc.dockerClient.ContainerInspect(ctx, tc.ContainerID)
		if err != nil {
			// Docker inspect failed - container might be removed
			newStatus = "error"
			errMsg = err.Error()
			logger.Warn("health check: container inspect failed",
				"container_id", tc.ContainerID,
				"team_id", tc.TeamID,
				"error", err)
		} else if !inspect.State.Running {
			newStatus = "stopped"
			if inspect.State.Status != "" {
				errMsg = "container state: " + inspect.State.Status
			}
		} else {
			// Container is running, try optional TCP/HTTP probe
			if tc.IPAddress != "" && tc.PortMapping != "" {
				rt, probeErr := hc.probeContainer(tc)
				responseTime = rt
				if probeErr != nil {
					// Container running but port not responding - still mark as running
					// but log the probe failure
					logger.Debug("health check: port probe failed",
						"container_id", tc.ContainerID,
						"ip", tc.IPAddress,
						"error", probeErr)
				}
			}
		}
	} else if tc.ContainerID == "" {
		newStatus = "pending"
	}

	// 2. Track failure counts
	containerDBID := tc.ID
	hc.mu.Lock()
	if newStatus != "running" {
		hc.failCount[containerDBID]++
	} else {
		hc.failCount[containerDBID] = 0
	}
	hc.mu.Unlock()

	// 3. Detect status change
	hc.mu.Lock()
	prevStatus, hadPrev := hc.prevStatus[containerDBID]
	hc.mu.Unlock()
	statusChanged := !hadPrev || prevStatus != newStatus

	// 4. Update TeamContainer.Status in DB
	if statusChanged && db != nil {
		db.Model(&model.TeamContainer{}).Where("id = ?", tc.ID).Update("status", newStatus)
		logger.Info("health check: container status changed",
			"container_id", tc.ContainerID,
			"team_id", tc.TeamID,
			"old_status", prevStatus,
			"new_status", newStatus)
	}
	hc.mu.Lock()
	hc.prevStatus[containerDBID] = newStatus
	hc.mu.Unlock()

	// 5. Record to ServiceHealth table
	if db != nil {
		healthRecord := model.ServiceHealth{
			ServiceID:    uint(tc.ID),
			Status:       mapStatus(newStatus),
			CheckedAt:    time.Now(),
			ResponseTime: responseTime,
			ErrorMsg:     errMsg,
			CreatedAt:    time.Now(),
		}
		hc.mu.Lock()
		shouldStore := newStatus != "running" || hc.failCount[containerDBID] > 0
		hc.mu.Unlock()
		if shouldStore {
			db.Create(&healthRecord)
		}
	}

	// 6. Container went down → publish event + write EventLog
	if statusChanged && newStatus != "running" && newStatus != "pending" {
		// Publish container:status event via EventBus
		bus := eventbus.GetBus()
		_ = bus.Publish(context.Background(), "container:status", map[string]interface{}{
			"game_id":      hc.gameID,
			"team_id":      tc.TeamID,
			"container_id": tc.ContainerID,
			"status":       newStatus,
			"previous":     prevStatus,
			"error":        errMsg,
		})

		// Write EventLog
		if db != nil {
			eventLog := model.EventLog{
				GameID:    &hc.gameID,
				EventType: "container_down",
				Level:     "warning",
				TeamID:    &tc.TeamID,
				Detail:    "Container " + tc.ContainerID + " status changed to " + newStatus + ": " + errMsg,
				CreatedAt: time.Now(),
			}
			db.Create(&eventLog)
		}
	}
}

// probeContainer does a TCP dial to the first port in PortMapping.
// Returns response time in ms and error if probe fails.
func (hc *HealthChecker) probeContainer(tc *model.TeamContainer) (int64, error) {
	// Parse first port from PortMapping (format like "8080:80" or "8080:80,8443:443")
	portStr := tc.PortMapping
	if idx := len(portStr); idx > 0 {
		for i, c := range portStr {
			if c == ':' {
				// Extract host port part after ':'
				hostPort := ""
				for j := i + 1; j < len(portStr); j++ {
					if portStr[j] == ',' || portStr[j] == ' ' {
						break
					}
					hostPort += string(portStr[j])
				}
				if hostPort != "" {
					addr := tc.IPAddress + ":" + hostPort
					start := time.Now()
					conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
					elapsed := time.Since(start).Milliseconds()
					if err != nil {
						return elapsed, err
					}
					conn.Close()
					return elapsed, nil
				}
				break
			}
		}
	}

	// Fallback: HTTP probe
	if tc.IPAddress != "" {
		url := "http://" + tc.IPAddress + "/health"
		start := time.Now()
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		elapsed := time.Since(start).Milliseconds()
		if err != nil {
			return elapsed, err
		}
		resp.Body.Close()
		return elapsed, nil
	}

	return 0, nil
}

// mapStatus maps engine status to health status constants.
func mapStatus(status string) string {
	switch status {
	case "running":
		return model.HealthStatusHealthy
	case "stopped", "error":
		return model.HealthStatusUnhealthy
	default:
		return model.HealthStatusUnknown
	}
}

// parseFirstPort extracts the first container port from a port mapping string.
func parseFirstPort(mapping string) int {
	for i := 0; i < len(mapping); i++ {
		if mapping[i] == ':' {
			portStr := ""
			for j := i + 1; j < len(mapping); j++ {
				if mapping[j] == ',' || mapping[j] == ' ' {
					break
				}
				portStr += string(mapping[j])
			}
			p, _ := strconv.Atoi(portStr)
			return p
		}
	}
	return 0
}
