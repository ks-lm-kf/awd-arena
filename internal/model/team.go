package model

import "time"

// Team represents a competition team.
type Team struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	Token       string    `json:"-" gorm:"uniqueIndex"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatar_url"`
	Score       float64   `json:"score" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
}

// GameTeam represents the association between a game and a team.
type GameTeam struct {
	ID     int64 `json:"id" gorm:"primaryKey"`
	GameID int64 `json:"game_id" gorm:"index"`
	TeamID int64 `json:"team_id" gorm:"index"`
	Team   Team  `json:"team" gorm:"foreignKey:TeamID"`
}
