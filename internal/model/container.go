package model

import "time"

// TeamContainer represents a team's challenge container.
type TeamContainer struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	GameID       int64     `json:"game_id"`
	TeamID       int64     `json:"team_id"`
	ChallengeID  int64     `json:"challenge_id"`
	ContainerID  string    `json:"container_id"`
	IPAddress    string    `json:"ip_address"`
	PortMapping  string    `json:"port_mapping" gorm:"type:jsonb"`
	SSHUser      string    `json:"ssh_user" gorm:"default:awd"`
	SSHPassword  string    `json:"ssh_password"`
	Status       string    `json:"status" gorm:"default:creating"`
	CreatedAt    time.Time `json:"created_at"`
}
