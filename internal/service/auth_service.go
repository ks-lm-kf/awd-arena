package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/middleware"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/crypto"
)

// AuthService handles authentication logic.
type AuthService struct{}

// UserInfo represents user info returned by login.
type UserInfo struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	TeamID             *int64 `json:"team_id"`
	MustChangePassword bool   `json:"must_change_password"`
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *UserInfo, error) {
	db := database.GetDB()
	if db == nil {
		return "", nil, errors.New("database not initialized")
	}

	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if !crypto.CheckPassword(password, user.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, user.TeamID)
	if err != nil {
		return "", nil, err
	}

	info := &UserInfo{
		ID:                 user.ID,
		Username:           user.Username,
		Role:               user.Role,
		TeamID:             user.TeamID,
		MustChangePassword: user.MustChangePassword,
	}
	return token, info, nil
}

// Register creates a new user.
func (s *AuthService) Register(ctx context.Context, username, password, role string, teamID *int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	if err := validatePasswordStrength(password, username); err != nil {
		return err
	}

	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	userRole := "player"
	if role != "" {
		switch role {
		case "admin", "organizer", "player":
			userRole = role
		}
	}

	user := model.User{
		Username:           username,
		Password:           hashed,
		Role:               userRole,
		TeamID:             teamID,
		MustChangePassword: true,
	}
	return db.Create(&user).Error
}

// GetUser retrieves a user by ID.
func (s *AuthService) GetUser(ctx context.Context, userID int64) (*UserInfo, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:                 user.ID,
		Username:           user.Username,
		Role:               user.Role,
		TeamID:             user.TeamID,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// GetUserModel retrieves the full user model by ID.
func (s *AuthService) GetUserModel(ctx context.Context, userID int64) (*model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByTeamID retrieves users belonging to a team.
func (s *AuthService) GetUserByTeamID(ctx context.Context, teamID int64) ([]model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var users []model.User
	err := db.Where("team_id = ?", teamID).Find(&users).Error
	return users, err
}

// ListUsers returns all users.
func (s *AuthService) ListUsers(ctx context.Context) ([]model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var users []model.User
	err := db.Find(&users).Error
	return users, err
}

// UpdateUser updates a user's fields.
func (s *AuthService) UpdateUser(ctx context.Context, userID int64, password *string, email *string, role *string, teamID *int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	updates := map[string]interface{}{}
	if password != nil && *password != "" {
		hashed, err := crypto.HashPassword(*password)
		if err != nil {
			return err
		}
		updates["password"] = hashed
	}
	if email != nil {
		updates["email"] = *email
	}
	if role != nil {
		updates["role"] = *role
	}
	if teamID != nil {
		updates["team_id"] = *teamID
	}
	if len(updates) == 0 {
		return nil
	}
	return db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// DeleteUser deletes a user by ID.
func (s *AuthService) DeleteUser(ctx context.Context, userID int64) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Delete(&model.User{}, userID).Error
}

func (s *AuthService) RegisterWithToken(ctx context.Context, username, password, teamToken string) (string, *UserInfo, error) {
	db := database.GetDB()
	if db == nil {
		return "", nil, errors.New("database not initialized")
	}

	// Check duplicate username
	var count int64
	db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return "", nil, errors.New("username already exists")
	}

	var teamID *int64
	if teamToken != "" {
		var team model.Team
		if err := db.Where("token = ?", crypto.SHA256Hex(teamToken)).First(&team).Error; err != nil {
			return "", nil, fmt.Errorf("invalid team token")
		}
		teamID = &team.ID
	}

	if err := validatePasswordStrength(password, username); err != nil {
		return "", nil, err
	}

	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return "", nil, err
	}

	user := model.User{
		Username:           username,
		Password:           hashed,
		Role:               "player",
		TeamID:             teamID,
		MustChangePassword: true,
	}
	if err := db.Create(&user).Error; err != nil {
		return "", nil, err
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, user.TeamID)
	if err != nil {
		return "", nil, err
	}

	info := &UserInfo{
		ID:                 user.ID,
		Username:           user.Username,
		Role:               user.Role,
		TeamID:             user.TeamID,
		MustChangePassword: user.MustChangePassword,
	}
	return token, info, nil
}

// JoinTeam adds the current user to a team via team_token.
func (s *AuthService) JoinTeam(ctx context.Context, userID int64, teamToken string) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if user.TeamID != nil {
		return errors.New("already in a team")
	}

	var team model.Team
	if err := db.Where("token = ?", crypto.SHA256Hex(teamToken)).First(&team).Error; err != nil {
		return errors.New("invalid team token")
	}

	// Update user's team
	if err := db.Model(&user).Update("team_id", team.ID).Error; err != nil {
		return errors.New("failed to join team")
	}

	return nil
}

// RegisterWithTokenAndRole creates a new user with specified role.
func (s *AuthService) RegisterWithTokenAndRole(ctx context.Context, username, password, teamToken, role string) (string, *UserInfo, error) {
	db := database.GetDB()
	if db == nil {
		return "", nil, errors.New("database not initialized")
	}

	// Check duplicate username
	var count int64
	db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return "", nil, errors.New("username already exists")
	}

	var teamID *int64
	if teamToken != "" {
		var team model.Team
		if err := db.Where("token = ?", crypto.SHA256Hex(teamToken)).First(&team).Error; err != nil {
			return "", nil, fmt.Errorf("invalid team token")
		}
		teamID = &team.ID
	}

	if err := validatePasswordStrength(password, username); err != nil {
		return "", nil, err
	}

	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return "", nil, err
	}

	// Default to player if role is empty
	if role == "" {
		role = "player"
	}

	user := model.User{
		Username:           username,
		Password:           hashed,
		Role:               role,
		TeamID:             teamID,
		MustChangePassword: true,
	}
	if err := db.Create(&user).Error; err != nil {
		return "", nil, err
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, user.TeamID)
	if err != nil {
		return "", nil, err
	}

	info := &UserInfo{
		ID:                 user.ID,
		Username:           user.Username,
		Role:               user.Role,
		TeamID:             user.TeamID,
		MustChangePassword: user.MustChangePassword,
	}
	return token, info, nil
}

// ChangePassword changes a user's password with validation.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !crypto.CheckPassword(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	// Validate new password strength
	if err := validatePasswordStrength(newPassword, user.Username); err != nil {
		return err
	}

	// Hash new password
	hashed, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and clear must_change_password flag
	now := time.Now()
	updates := map[string]interface{}{
		"password":             hashed,
		"password_changed_at":  &now,
		"must_change_password": false,
	}

	return db.Model(&user).Updates(updates).Error
}

// validatePasswordStrength validates password meets security requirements
func validatePasswordStrength(password, username string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if password == username {
		return errors.New("password cannot be the same as username")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}
