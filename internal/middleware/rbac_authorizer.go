package middleware

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// ResourceAuthorizer provides fine-grained resource access control.
type ResourceAuthorizer struct {
	db *gorm.DB
}

// NewResourceAuthorizer creates a new resource authorizer.
func NewResourceAuthorizer(db *gorm.DB) *ResourceAuthorizer {
	return &ResourceAuthorizer{db: db}
}

// CanAccessGame checks if user can access a specific game.
func (ra *ResourceAuthorizer) CanAccessGame(c fiber.Ctx, gameID int64) error {
	_, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin {
		return nil
	}

	if role == model.RoleOrganizer {
		// Check if organizer owns this game
		var game model.Game
		if err := ra.db.First(&game, gameID).Error; err != nil {
			return fiber.NewError(404, "game not found")
		}
		if game.CreatedBy != 0 {
			userID, _, _, _ := GetCurrentUser(c)
			if game.CreatedBy == userID {
				return nil
			}
		}
		return nil // Allow organizers to access all games for now
	}

	// Player: check if they're in the game
	var count int64
	ra.db.Model(&model.GameTeam{}).Where("game_id = ? AND team_id = ?", gameID, c.Locals("team_id")).Count(&count)
	if count == 0 {
		return fiber.NewError(403, "not participating in this game")
	}
	return nil
}

// CanModifyGame checks if user can modify a specific game.
func (ra *ResourceAuthorizer) CanModifyGame(c fiber.Ctx, gameID int64) error {
	userID, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin {
		return nil
	}

	if role == model.RoleOrganizer {
		var game model.Game
		if err := ra.db.First(&game, gameID).Error; err != nil {
			return fiber.NewError(404, "game not found")
		}
		if game.CreatedBy == userID {
			return nil
		}
		return fiber.NewError(403, "not the creator of this game")
	}

	return fiber.NewError(403, "not authorized to modify this game")
}

// CanAccessChallenge checks if user can access a specific challenge.
func (ra *ResourceAuthorizer) CanAccessChallenge(c fiber.Ctx, challengeID int64) error {
	_, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin || role == model.RoleOrganizer {
		return nil
	}

	var challenge model.Challenge
	if err := ra.db.First(&challenge, challengeID).Error; err != nil {
		return fiber.NewError(404, "challenge not found")
	}
	return ra.CanAccessGame(c, challenge.GameID)
}

// CanModifyChallenge checks if user can modify a specific challenge.
func (ra *ResourceAuthorizer) CanModifyChallenge(c fiber.Ctx, challengeID int64) error {
	_, _, role, _ := GetCurrentUser(c)

	if role == model.RoleAdmin {
		return nil
	}

	var challenge model.Challenge
	if err := ra.db.First(&challenge, challengeID).Error; err != nil {
		return fiber.NewError(404, "challenge not found")
	}
	return ra.CanModifyGame(c, challenge.GameID)
}

// GameAccessMiddleware creates middleware for game access control.
func (ra *ResourceAuthorizer) GameAccessMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		gameIDStr := c.Params("id")
		if gameIDStr == "" {
			gameIDStr = c.Params("gameId")
		}
		if gameIDStr == "" {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "game id required"})
		}
		gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid game id"})
		}
		if err := ra.CanAccessGame(c, gameID); err != nil {
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{"code": e.Code, "message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
		return c.Next()
	}
}

// ChallengeAccessMiddleware creates middleware for challenge access control.
func (ra *ResourceAuthorizer) ChallengeAccessMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		challengeIDStr := c.Params("id")
		if challengeIDStr == "" {
			challengeIDStr = c.Params("challengeId")
		}
		if challengeIDStr == "" {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "challenge id required"})
		}
		challengeID, err := strconv.ParseInt(challengeIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid challenge id"})
		}
		if err := ra.CanAccessChallenge(c, challengeID); err != nil {
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{"code": e.Code, "message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
		return c.Next()
	}
}
