package middleware

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// ResourceAuthorizer provides fine-grained resource access control
type ResourceAuthorizer struct {
	db *gorm.DB
}

// NewResourceAuthorizer creates a new resource authorizer
func NewResourceAuthorizer(db *gorm.DB) *ResourceAuthorizer {
	return &ResourceAuthorizer{db: db}
}

// CanAccessGame checks if user can access a specific game
func (ra *ResourceAuthorizer) CanAccessGame(c fiber.Ctx, gameID int64) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return fiber.NewError(401, "user not authenticated")
	}

	// Admin can access all games
	if user.Role == string(model.RoleAdmin) {
		return nil
	}

	// Organizer can access games they created or all games
	if user.Role == string(model.RoleOrganizer) {
		// TODO: Check if organizer owns this game
		return nil
	}

	// Player can only access public games or games they're participating in
	// For now, allow all players to access all games (simplified)
	var game model.Game
	result := ra.db.First(&game, gameID)
	if result.Error != nil {
		return fiber.NewError(404, "game not found")
	}

	return nil
}

// CanModifyGame checks if user can modify a specific game
func (ra *ResourceAuthorizer) CanModifyGame(c fiber.Ctx, gameID int64) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return fiber.NewError(401, "user not authenticated")
	}

	// Admin can modify all games
	if user.Role == string(model.RoleAdmin) {
		return nil
	}

	// Organizer can modify games they created
	if user.Role == string(model.RoleOrganizer) {
		var game model.Game
		result := ra.db.First(&game, gameID)
		if result.Error != nil {
			return fiber.NewError(404, "game not found")
		}

		// TODO: Check if organizer is the creator
		return nil
	}

	return fiber.NewError(403, "not authorized to modify this game")
}

// CanAccessChallenge checks if user can access a specific challenge
func (ra *ResourceAuthorizer) CanAccessChallenge(c fiber.Ctx, challengeID int64) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return fiber.NewError(401, "user not authenticated")
	}

	// Admin can access all challenges
	if user.Role == string(model.RoleAdmin) {
		return nil
	}

	// Organizer can access all challenges
	if user.Role == string(model.RoleOrganizer) {
		return nil
	}

	// Player can only access challenges in games they can access
	var challenge model.Challenge
	result := ra.db.First(&challenge, challengeID)
	if result.Error != nil {
		return fiber.NewError(404, "challenge not found")
	}

	// Check if player can access the game
	return ra.CanAccessGame(c, challenge.GameID)
}

// CanModifyChallenge checks if user can modify a specific challenge
func (ra *ResourceAuthorizer) CanModifyChallenge(c fiber.Ctx, challengeID int64) error {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		return fiber.NewError(401, "user not authenticated")
	}

	// Admin can modify all challenges
	if user.Role == string(model.RoleAdmin) {
		return nil
	}

	// Organizer can modify challenges in games they can modify
	if user.Role == string(model.RoleOrganizer) {
		var challenge model.Challenge
		result := ra.db.First(&challenge, challengeID)
		if result.Error != nil {
			return fiber.NewError(404, "challenge not found")
		}

		return ra.CanModifyGame(c, challenge.GameID)
	}

	return fiber.NewError(403, "not authorized to modify this challenge")
}

// GameAccessMiddleware creates middleware for game access control
func (ra *ResourceAuthorizer) GameAccessMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameIDStr := c.Params("id")
		if gameIDStr == "" {
			gameIDStr = c.Params("gameId")
		}

		if gameIDStr == "" {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "game id required",
			})
		}

		gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "invalid game id",
			})
		}

		if err := ra.CanAccessGame(c, gameID); err != nil {
			// Check if it's a fiber.Error
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{
					"code":    e.Code,
					"message": e.Message,
				})
			}
			// Generic error
			return c.Status(500).JSON(fiber.Map{
				"code":    500,
				"message": err.Error(),
			})
		}

		return c.Next()
	}
}

// ChallengeAccessMiddleware creates middleware for challenge access control
func (ra *ResourceAuthorizer) ChallengeAccessMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		challengeIDStr := c.Params("id")
		if challengeIDStr == "" {
			challengeIDStr = c.Params("challengeId")
		}

		if challengeIDStr == "" {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "challenge id required",
			})
		}

		challengeID, err := strconv.ParseInt(challengeIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "invalid challenge id",
			})
		}

		if err := ra.CanAccessChallenge(c, challengeID); err != nil {
			// Check if it's a fiber.Error
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{
					"code":    e.Code,
					"message": e.Message,
				})
			}
			// Generic error
			return c.Status(500).JSON(fiber.Map{
				"code":    500,
				"message": err.Error(),
			})
		}

		return c.Next()
	}
}
