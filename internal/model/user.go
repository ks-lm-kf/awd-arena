package model

import "time"

type User struct {
    ID                int64      `json:"id" gorm:"primaryKey"`
    Username          string     `json:"username" gorm:"uniqueIndex;not null"`
    Password          string     `json:"-" gorm:"not null"`
    Email             string     `json:"email"`
    Role              string     `json:"role" gorm:"default:player"`
    TeamID            *int64     `json:"team_id" gorm:"index"`
    PasswordChangedAt *time.Time `json:"password_changed_at"`
    MustChangePassword bool      `json:"must_change_password" gorm:"default:false"`
    CreatedAt         time.Time  `json:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at"`
}
