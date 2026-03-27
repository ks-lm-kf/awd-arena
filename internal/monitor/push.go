package monitor

import (
	"encoding/json"

	"github.com/awd-platform/awd-arena/internal/server"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// PushService pushes real-time updates via WebSocket.
type PushService struct{}

func NewPushService() *PushService {
	return &PushService{}
}

// PushRankingUpdate pushes ranking changes.
func (p *PushService) PushRankingUpdate(gameID string, rankings interface{}) {
	msg := server.WSMessage{
		Type: "ranking:update",
		Data: map[string]interface{}{
			"game_id":  gameID,
			"rankings": rankings,
		},
	}
	p.send(gameID, msg)
}

// PushFlagCaptured pushes a flag capture event.
func (p *PushService) PushFlagCaptured(gameID, attacker, target string, points float64) {
	msg := server.WSMessage{
		Type: "flag:captured",
		Data: map[string]interface{}{
			"game_id":  gameID,
			"attacker": attacker,
			"target":   target,
			"points":   points,
		},
	}
	p.send(gameID, msg)
}

// PushAlert pushes a security alert.
func (p *PushService) PushAlert(gameID, level, team, message string) {
	msg := server.WSMessage{
		Type: "alert:new",
		Data: map[string]interface{}{
			"game_id": gameID,
			"level":   level,
			"team":    team,
			"message": message,
		},
	}
	p.send(gameID, msg)
}

// PushContainerStatus pushes container status updates.
func (p *PushService) PushContainerStatus(gameID, teamID, status string, cpu float64) {
	msg := server.WSMessage{
		Type: "container:status",
		Data: map[string]interface{}{
			"game_id": gameID,
			"team_id": teamID,
			"status":  status,
			"cpu":     cpu,
		},
	}
	p.send(gameID, msg)
}

// PushGameStatus pushes game status changes.
func (p *PushService) PushGameStatus(gameID, status, reason string) {
	msg := server.WSMessage{
		Type: "game:status",
		Data: map[string]interface{}{
			"game_id": gameID,
			"status":  status,
			"reason":  reason,
		},
	}
	p.send(gameID, msg)
}

func (p *PushService) send(gameID string, msg server.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("push marshal error", "error", err)
		return
	}
	server.Hub.BroadcastToGame("game:"+gameID, data)
}
