package model

import "time"

// RoundScore stores per-round team scores.
type RoundScore struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	GameID       int64     `json:"game_id"`
	Round        int       `json:"round"`
	TeamID       int64     `json:"team_id"`
	AttackScore  float64   `json:"attack_score"`
	DefenseScore float64   `json:"defense_score"`
	TotalScore   float64   `json:"total_score"`
	Rank         int       `json:"rank"`
	CalculatedAt time.Time `json:"calculated_at"`
}
