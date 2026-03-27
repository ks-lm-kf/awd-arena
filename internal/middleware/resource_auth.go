package middleware

import (
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/repo"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"strconv"
)

// ResourceAuth handles resource-level authorization
type ResourceAuth struct {
	db *gorm.DB
}

// NewResourceAuth creates a new ResourceAuth instance
func NewResourceAuth(db *gorm.DB) *ResourceAuth {
	return &ResourceAuth{db: db}
}

// CanAccessGame checks if user can access a specific game
func (ra *ResourceAuth) CanAccessGame(c fiber.Ctx, gameID int64) error {
	userID, _, role, teamID := GetCurrentUser(c)
	
	// Admin can access all games
	if role == model.RoleAdmin {
		return nil
	}
	
	// Organizer can access games they created or all games (depends on requirements)
	if role == model.RoleOrganizer {
		return nil
	}
	
	// Player can only access games they've joined
	var gameTeam repo.GameTeam
	result := ra.db.Where("game_id = ? AND team_id = ?", gameID, teamID).First(&gameTeam)
	if result.Error != nil {
		return c.Status(403).JSON(fiber.Map{
			"code":    403,
			"message": "not authorized to access this game",
		})
	}
	
	_ = userID // Avoid unused variable error
	return nil
}

// CanModifyGame checks if user can modify a specific game
func (ra *ResourceAuth) CanModifyGame(c fiber.Ctx, gameID int64) error {
	_, _, role, _ := GetCurrentUser(c)
	
	// Only admin and organizer can modify games
	if role != model.RoleAdmin && role != model.RoleOrganizer {
		return c.Status(403).JSON(fiber.Map{
			"code":    403,
			"message": "not authorized to modify this game",
		})
	}
	
	// TODO: If needed, check if organizer created this specific game
	// For now, any organizer can modify any game
	
	_ = gameID // Avoid unused variable error
	return nil
}

// CanAccessChallenge checks if user can access a specific challenge
func (ra *ResourceAuth) CanAccessChallenge(c fiber.Ctx, gameID, challengeID int64) error {
	// First check game access
	if err := ra.CanAccessGame(c, gameID); err != nil {
		return err
	}
	
	// Verify challenge belongs to this game
	var challenge model.Challenge
	result := ra.db.Where("id = ? AND game_id = ?", challengeID, gameID).First(&challenge)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "challenge not found in this game",
		})
	}
	
	return nil
}

// CanModifyChallenge checks if user can modify a specific challenge
func (ra *ResourceAuth) CanModifyChallenge(c fiber.Ctx, gameID, challengeID int64) error {
	_, _, role, _ := GetCurrentUser(c)
	
	// Only admin and organizer can modify challenges
	if role != model.RoleAdmin && role != model.RoleOrganizer {
		return c.Status(403).JSON(fiber.Map{
			"code":    403,
			"message": "not authorized to modify this challenge",
		})
	}
	
	// Verify challenge belongs to this game
	var challenge model.Challenge
	result := ra.db.Where("id = ? AND game_id = ?", challengeID, gameID).First(&challenge)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "challenge not found in this game",
		})
	}
	
	return nil
}

// CanAccessContainer checks if user can access a specific container
func (ra *ResourceAuth) CanAccessContainer(c fiber.Ctx, gameID, containerID int64) error {
	_, _, role, _ := GetCurrentUser(c)
	
	// Admin and Organizer can access all containers in games they manage
	if role == model.RoleAdmin || role == model.RoleOrganizer {
		// Verify container belongs to this game
		var container model.TeamContainer
		result := ra.db.Where("id = ? AND game_id = ?", containerID, gameID).First(&container)
		if result.Error != nil {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "container not found in this game",
			})
		}
		return nil
	}
	
	// Players cannot directly access containers
	return c.Status(403).JSON(fiber.Map{
		"code":    403,
		"message": "not authorized to access containers",
	})
}

// GameMiddleware creates a middleware that checks game access
func (ra *ResourceAuth) GameMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "invalid game id",
			})
		}
		
		return ra.CanAccessGame(c, int64(gameID))
	}
}

// ChallengeMiddleware creates a middleware that checks challenge access
func (ra *ResourceAuth) ChallengeMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "invalid game id",
			})
		}
		
		challengeID, err := strconv.Atoi(c.Params("challenge_id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "invalid challenge id",
			})
		}
		
		return ra.CanAccessChallenge(c, int64(gameID), int64(challengeID))
	}
}

