package model

import "time"

type TeamContainer struct {
    ID           int64     `json:"id" gorm:"primaryKey"`
    GameID       int64     `json:"game_id" gorm:"index"`
    TeamID       int64     `json:"team_id" gorm:"index"`
    ChallengeID  int64     `json:"challenge_id" gorm:"index"`
    ContainerID  string    `json:"container_id"`
    IPAddress    string    `json:"ip_address"`
    PortMapping  string    `json:"port_mapping"`
    SSHUser      string    `json:"ssh_user" gorm:"default:awd"`
    SSHPassword  string    `json:"ssh_password"`
    Status       string    `json:"status" gorm:"default:creating"`
    CreatedAt    time.Time `json:"created_at"`
}
