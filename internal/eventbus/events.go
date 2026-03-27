package eventbus

// Event represents a domain event.
type Event struct {
	Type      string      `json:"type"`
	GameID    int64       `json:"game_id"`
	Round     int         `json:"round"`
	TeamID    int64       `json:"team_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Event types
const (
	EventRoundStart   = "round_start"
	EventRoundEnd     = "round_end"
	EventFlagCaptured = "flag_captured"
	EventAttack       = "attack"
	EventAlert        = "alert"
	EventContainerUp  = "container_up"
	EventContainerDown = "container_down"
)
