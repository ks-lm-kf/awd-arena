package model

import "time"

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

// Game represents a competition game.
type Game struct {
	ID             int64      `json:"id" gorm:"primaryKey"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Mode           string     `json:"mode" gorm:"default:awd_score"`
	Status         string     `json:"status" gorm:"default:draft"`           // draft, active, finished
	CurrentPhase   string     `json:"current_phase" gorm:"default:preparation"` // preparation, running, break, finished
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

// IsDraft returns true if the game is in draft status
func (g *Game) IsDraft() bool {
	return g.Status == GameStatusDraft
}

// IsActive returns true if the game is active
func (g *Game) IsActive() bool {
	return g.Status == GameStatusActive
}

// IsFinished returns true if the game is finished
func (g *Game) IsFinished() bool {
	return g.Status == GameStatusFinished
}

// IsPreparing returns true if the game is in preparation phase
func (g *Game) IsPreparing() bool {
	return g.CurrentPhase == GamePhasePreparation
}

// IsRunning returns true if the game is in running phase
func (g *Game) IsRunning() bool {
	return g.CurrentPhase == GamePhaseRunning
}

// IsBreak returns true if the game is in break phase
func (g *Game) IsBreak() bool {
	return g.CurrentPhase == GamePhaseBreak
}

// CanStart returns true if the game can be started
func (g *Game) CanStart() bool {
	return g.Status == GameStatusDraft && g.CurrentPhase == GamePhasePreparation
}

// CanPause returns true if the game can be paused
func (g *Game) CanPause() bool {
	return g.Status == GameStatusActive && g.CurrentPhase == GamePhaseRunning
}

// CanResume returns true if the game can be resumed
func (g *Game) CanResume() bool {
	return g.Status == GameStatusActive && g.CurrentPhase == GamePhaseBreak
}

// CanFinish returns true if the game can be finished
func (g *Game) CanFinish() bool {
	return g.Status == GameStatusActive && (g.CurrentPhase == GamePhaseRunning || g.CurrentPhase == GamePhaseBreak)
}
