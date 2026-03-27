package security

import (
	"sync"
	"time"
)

// Alert represents a security alert.
type Alert struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`   // info, warning, critical
	TeamID    string    `json:"team_id"`
	Type      string    `json:"type"`    // port_scan, brute_force, exploit, suspicious
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
	GameID    string    `json:"game_id"`
}

// AlertManager manages security alerts.
type AlertManager struct {
	mu      sync.RWMutex
	alerts  []Alert
	maxSize int
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(maxSize int) *AlertManager {
	return &AlertManager{
		alerts:  make([]Alert, 0, maxSize),
		maxSize: maxSize,
	}
}

// AddAlert adds a new alert.
func (am *AlertManager) AddAlert(alert Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()
	alert.Timestamp = time.Now()
	am.alerts = append(am.alerts, alert)
	if len(am.alerts) > am.maxSize {
		am.alerts = am.alerts[len(am.alerts)-am.maxSize:]
	}
}

// GetAlerts returns all alerts, optionally filtered.
func (am *AlertManager) GetAlerts(level, teamID string, limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	var result []Alert
	for i := len(am.alerts) - 1; i >= 0 && len(result) < limit; i-- {
		a := am.alerts[i]
		if level != "" && a.Level != level {
			continue
		}
		if teamID != "" && a.TeamID != teamID {
			continue
		}
		result = append(result, a)
	}
	return result
}

// GetAlertsByGame returns alerts for a specific game.
func (am *AlertManager) GetAlertsByGame(gameID string) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	var result []Alert
	for _, a := range am.alerts {
		if a.GameID == gameID {
			result = append(result, a)
		}
	}
	return result
}

// Clear clears all alerts.
func (am *AlertManager) Clear() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.alerts = am.alerts[:0]
}
