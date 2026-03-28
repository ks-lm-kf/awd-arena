package model

import "gorm.io/gorm"

type TargetService struct {
    gorm.Model
    Name     string `json:"name"`
    Protocol string `json:"protocol"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Path     string `json:"path"`
    Enabled  bool   `json:"enabled" gorm:"default:true"`
    GameID   int64  `json:"game_id"`
    TeamID   int64  `json:"team_id"`
}
