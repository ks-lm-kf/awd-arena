package model

import "time"

// DockerImage represents a Docker image available for challenges.
type DockerImage struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Tag          string    `json:"tag" gorm:"default:latest"`
	ImageID      string    `json:"image_id"`
	Description  string    `json:"description"`
	Category     string    `json:"category" gorm:"default:general"`
	Difficulty   string    `json:"difficulty" gorm:"default:medium"`
	Ports        string    `json:"ports"`
	MemoryLimit  int       `json:"memory_limit" gorm:"default:256"`
	CPULimit     float64   `json:"cpu_limit" gorm:"default:1.0"`
	Flag         string    `json:"flag"`
	InitialScore int       `json:"initial_score" gorm:"default:100"`
	Status       string    `json:"status" gorm:"default:active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
