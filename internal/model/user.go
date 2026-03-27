package model

import "time"

// User represents a platform user.
type User struct {
	ID                 int64      `json:"id" gorm:"primaryKey"`
	Username           string     `json:"username" gorm:"uniqueIndex"`
	Password           string     `json:"-"`
	Email              string     `json:"email"`
	Role               string     `json:"role" gorm:"default:player"`
	TeamID             *int64     `json:"team_id"`
	PasswordChangedAt  *time.Time `json:"password_changed_at"`
	MustChangePassword bool       `json:"must_change_password" gorm:"default:true"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

