package model

import "time"

type Game struct {
    ID             int64      `json:"id" gorm:"primaryKey"`
    Title          string     `json:"title"`
    Description    string     `json:"description"`
    Mode           string     `json:"mode" gorm:"default:awd_score"`
    Status         string     `json:"status" gorm:"default:draft"`
    CurrentPhase   string     `json:"current_phase" gorm:"default:preparation"`
    RoundDuration  int        `json:"round_duration" gorm:"default:300"`
    BreakDuration  int        `json:"break_duration" gorm:"default:120"`
    TotalRounds    int        `json:"total_rounds" gorm:"default:20"`
    CurrentRound   int        `json:"current_round" gorm:"default:0"`
    FlagFormat     string     `json:"flag_format" gorm:"default:flag{%s}"`
    AttackWeight   float64    `json:"attack_weight" gorm:"default:1.0"`
    DefenseWeight  float64    `json:"defense_weight" gorm:"default:0.5"`
    StartTime      *time.Time `json:"start_time"`
    EndTime        *time.Time `json:"end_time"`
    CreatedBy      int64      `json:"created_by"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

func (g *Game) IsDraft() bool    { return g.Status == "draft" }
func (g *Game) IsActive() bool   { return g.Status == "active" }
func (g *Game) IsFinished() bool { return g.Status == "finished" }
func (g *Game) CanStart() bool   { return g.Status == "draft" || g.Status == "active" && g.CurrentPhase == "preparation" }
func (g *Game) CanPause() bool   { return g.Status == "active" && g.CurrentPhase == "running" }
func (g *Game) CanResume() bool  { return g.Status == "active" && g.CurrentPhase == "break" }
func (g *Game) CanFinish() bool  { return g.Status == "active" }

// Game status constants
const (
    GameStatusDraft    = "draft"
    GameStatusActive   = "active"
    GameStatusFinished = "finished"
)

// Game phase constants
const (
    GamePhasePreparation = "preparation"
    GamePhaseRunning     = "running"
    GamePhaseBreak       = "break"
    GamePhaseFinished    = "finished"
)
