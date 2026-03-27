package model

import "time"

// ScoreAdjustment 分数调整记录
type ScoreAdjustment struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	GameID      int64     `json:"game_id" gorm:"index"`
	TeamID      int64     `json:"team_id" gorm:"index"`
	AdjustValue int       `json:"adjust_value"` // 正数为加分，负数为减分
	Reason      string    `json:"reason"`
	OperatorID  int64     `json:"operator_id"` // 操作人ID
	Round       int       `json:"round"`       // 调整时的轮次
	CreatedAt   time.Time `json:"created_at"`
}

func (ScoreAdjustment) TableName() string {
	return "score_adjustments"
}

