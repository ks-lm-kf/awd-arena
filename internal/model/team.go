package model

import "time"

type Team struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"uniqueIndex;not null"`
    Token       string    `json:"token" gorm:"uniqueIndex"`
    Description string    `json:"description"`
    AvatarURL   string    `json:"avatar_url"`
    Score       float64   `json:"score" gorm:"default:0"`
    CreatedAt   time.Time `json:"created_at"`
}

type GameTeam struct {
    ID     int64 `json:"id" gorm:"primaryKey"`
    GameID int64 `json:"game_id" gorm:"index"`
    TeamID int64 `json:"team_id" gorm:"index"`
    Team   Team  `json:"team" gorm:"foreignKey:TeamID"`
}
