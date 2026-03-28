package model

import "time"

type RoundScore struct {
    ID           int64     `json:"id" gorm:"primaryKey"`
    GameID       int64     `json:"game_id" gorm:"index"`
    Round        int       `json:"round"`
    TeamID       int64     `json:"team_id" gorm:"index"`
    AttackScore  float64   `json:"attack_score"`
    DefenseScore float64   `json:"defense_score"`
    TotalScore   float64   `json:"total_score"`
    Rank         int       `json:"rank"`
    CalculatedAt time.Time `json:"calculated_at"`
}

type ScoreAdjustment struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    GameID      int64     `json:"game_id" gorm:"index"`
    TeamID      int64     `json:"team_id" gorm:"index"`
    AdjustValue int       `json:"adjust_value"`
    Reason      string    `json:"reason"`
    OperatorID  int64     `json:"operator_id"`
    Round       int       `json:"round"`
    CreatedAt   time.Time `json:"created_at"`
}
