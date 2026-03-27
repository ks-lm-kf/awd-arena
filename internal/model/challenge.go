package model

import "time"

// Challenge represents an AWD challenge target.
type Challenge struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	GameID      int64     `json:"game_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageName   string    `json:"image_name"`
	ImageTag    string    `json:"image_tag" gorm:"default:latest"`
	Difficulty  string    `json:"difficulty" gorm:"default:medium"`
	BaseScore   int       `json:"base_score" gorm:"default:100"`
	ExposedPorts string   `json:"exposed_ports" gorm:"type:jsonb"`
	CPULimit    float64   `json:"cpu_limit" gorm:"default:0.5"`
	MemLimit    int       `json:"mem_limit" gorm:"default:256"`
	CreatedAt   time.Time `json:"created_at"`
}
