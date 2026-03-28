package middleware

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// ResourceAuth handles resource-level authorization.
type ResourceAuth struct {
	db *gorm.DB
}

// NewResourceAuth creates a new ResourceAuth instance.
func NewResourceAuth(db *gorm.DB) *ResourceAuth {
	return &ResourceAuth{db: db}
}

// CanAccessGame checks if user can access a specific game.
func (ra *ResourceAuth) CanAccessGame(c fiber.Ctx, gameID int64) error {
	userID, _, role, teamID := GetCurrentUser(c)

	if role == model.RoleAdmin {
		return nil
	}

	if role == model.RoleOrganizer {
		return nil
	}

	// Player: check if they're in the game
	if teamID != nil {
		var count int64
		ra.db.Model(&model.GameTeam{}).Where("game_id = ? AND team_id = ?", gameID, *teamID).Count(&count)
		if count > 0 {
			return nil
		}
	}

	_ = userID
	return c.Status(403).JSON(fiber.Map{"code": 403, "message": "not authorized to access this game"})
}

// CanModifyGame checks if user can modify a specific game.
func (ra *ResourceAuth) CanModifyGame(c fiber.Ctx, gameID int64) error {
	userID, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin {
		return nil
	}

	if role == model.RoleOrganizer {
		// Check if organizer created this game
		var game model.Game
		if err := ra.db.First(&game, gameID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"code": 404, "message": "game not found"})
		}
		if game.CreatedBy == userID {
			return nil
		}
		return c.Status(403).JSON(fiber.Map{"code": 403, "message": "not the creator of this game"})
	}

	_ = gameID
	return c.Status(403).JSON(fiber.Map{"code": 403, "message": "not authorized to modify this game"})
}

// CanAccessChallenge checks if user can access a specific challenge.
func (ra *ResourceAuth) CanAccessChallenge(c fiber.Ctx, gameID, challengeID int64) error {
	if err := ra.CanAccessGame(c, gameID); err != nil {
		return err
	}
	var challenge model.Challenge
	result := ra.db.Where("id = ? AND game_id = ?", challengeID, gameID).First(&challenge)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "challenge not found in this game"})
	}
	return nil
}

// CanModifyChallenge checks if user can modify a specific challenge.
func (ra *ResourceAuth) CanModifyChallenge(c fiber.Ctx, gameID, challengeID int64) error {
	if err := ra.CanModifyGame(c, gameID); err != nil {
		return err
	}
	var challenge model.Challenge
	result := ra.db.Where("id = ? AND game_id = ?", challengeID, gameID).First(&challenge)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "challenge not found in this game"})
	}
	return nil
}

// CanAccessContainer checks if user can access a specific container.
func (ra *ResourceAuth) CanAccessContainer(c fiber.Ctx, gameID, containerID int64) error {
	_, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin || role == model.RoleOrganizer {
		var container model.TeamContainer
		result := ra.db.Where("id = ? AND game_id = ?", containerID, gameID).First(&container)
		if result.Error != nil {
			return c.Status(404).JSON(fiber.Map{"code": 404, "message": "container not found in this game"})
		}
		return nil
	}
	return c.Status(403).JSON(fiber.Map{"code": 403, "message": "not authorized to access containers"})
}

// GameMiddleware creates a middleware that checks game access.
func (ra *ResourceAuth) GameMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid game id"})
		}
		return ra.CanAccessGame(c, int64(gameID))
	}
}

// ChallengeMiddleware creates a middleware that checks challenge access.
func (ra *ResourceAuth) ChallengeMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid game id"})
		}
		challengeID, err := strconv.Atoi(c.Params("challenge_id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid challenge id"})
		}
		return ra.CanAccessChallenge(c, int64(gameID), int64(challengeID))
	}
}
