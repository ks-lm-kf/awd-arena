package model

import "time"

type AdminLog struct {
    ID            int64     `json:"id" gorm:"primaryKey"`
    UserID        int64     `json:"user_id" gorm:"index"`
    Username      string    `json:"username"`
    Action        string    `json:"action"`
    ResourceType  string    `json:"resource_type"`
    ResourceID    int64     `json:"resource_id"`
    Description   string    `json:"description"`
    IPAddress     string    `json:"ip_address"`
    UserAgent     string    `json:"user_agent"`
    Details       string    `json:"details"`
    CreatedAt     time.Time `json:"created_at"`
}
