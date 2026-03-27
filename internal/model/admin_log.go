package model

import "time"

// AdminLog represents an administrative action log
type AdminLog struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    UserID      int64     `json:"user_id" gorm:"index"`
    Username    string    `json:"username"`
    Action      string    `json:"action" gorm:"index"`        // create, update, delete, start, pause, stop, reset, adjust_score, import
    ResourceType string   `json:"resource_type" gorm:"index"` // game, team, user, score
    ResourceID   int64    `json:"resource_id" gorm:"index"`
    Description string    `json:"description"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    Details     string    `json:"details" gorm:"type:text"` // JSON string for additional details
    CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// TableName returns the table name for AdminLog
func (AdminLog) TableName() string {
    return "admin_logs"
}
