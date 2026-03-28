package model

import "time"

type DockerImage struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    Name         string    `json:"name"`
    Tag          string    `json:"tag"`
    ImageID      string    `json:"image_id"`
    Description  string    `json:"description"`
    Category     string    `json:"category"`
    Difficulty   string    `json:"difficulty"`
    Ports        string    `json:"ports"`
    MemoryLimit  int       `json:"memory_limit"`
    CPULimit     float64   `json:"cpu_limit"`
    Flag         string    `json:"flag"`
    InitialScore int       `json:"initial_score"`
    Status       string    `json:"status" gorm:"default:active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
