package model

import (
	"gorm.io/gorm"
)

// TargetService 轮询服务（竞技场目标）
type TargetService struct {
	gorm.Model
	Name      string `gorm:"size:100;not null" json:"name"`
	Protocol  string `gorm:"size:20;not null;default:'tcp'" json:"protocol"` // http, https, tcp
	Host      string `gorm:"size:255;not null" json:"host"`
	Port      int    `gorm:"not null" json:"port"`
	Path      string `gorm:"size:255;default:'/'" json:"path"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	GameID    int64  `gorm:"index" json:"game_id"` // 所属的赛场ID
	TeamID    int64  `gorm:"index" json:"team_id"` // 所属的队伍ID
}
