package model

import "time"

type Team struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	Token       string    `json:"token" gorm:"uniqueIndex"`
	RawToken    string    `json:"raw_token,omitempty" gorm:"-"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatar_url"`
	Score       float64   `json:"score" gorm:"default:0"`
	MemberCount int       `json:"member_count" gorm:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type GameTeam struct {
	ID     int64 `json:"id" gorm:"primaryKey"`
	GameID int64 `json:"game_id" gorm:"index"`
	TeamID int64 `json:"team_id" gorm:"index"`
	Team   Team  `json:"team" gorm:"foreignKey:TeamID"`
}
