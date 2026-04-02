package model

import "time"

type FlagRecord struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	GameID    int64     `json:"game_id" gorm:"index:idx_game_round_team"`
	Round     int       `json:"round" gorm:"index:idx_game_round_team"`
	TeamID    int64     `json:"team_id" gorm:"index:idx_game_round_team"`
	FlagHash  string    `json:"flag_hash"`
	Service   string    `json:"service"`
	CreatedAt time.Time `json:"created_at"`
}

type FlagSubmission struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	GameID       int64     `json:"game_id" gorm:"index"`
	Round        int       `json:"round"`
	AttackerTeam int64     `json:"attacker_team"`
	TargetTeam   int64     `json:"target_team"`
	FlagHash     string    `json:"flag_hash" gorm:"index"`
	FlagValue    string    `json:"flag_value"`
	IsCorrect    bool      `json:"is_correct"`
	PointsEarned float64   `json:"points_earned"`
	SubmittedAt  time.Time `json:"submitted_at"`
}
