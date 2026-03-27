package middleware

import (
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// RequirePermission returns a middleware that checks if the user has a specific permission.
func RequirePermission(perm model.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		roleStr, _ := c.Locals("role").(string)
		role := model.Role(roleStr)
		
		logger.Info("[Permission Check]", "path", c.Path(), "role", role, "required_permission", perm)
		
		// Check if role has the required permission
		permissions, exists := model.RolePermissions[role]
		if !exists {
			logger.Error("[Permission Check]", "error", "invalid role", "role", role)
			return c.Status(403).JSON(fiber.Map{
				"code":    403,
				"message": "invalid role",
			})
		}
		
		logger.Info("[Permission Check]", "role_permissions", permissions)
		
		// Check if the permission is in the role's permission list
		hasPermission := false
		for _, p := range permissions {
			if p == perm {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			logger.Error("[Permission Check]", "error", "permission denied", "role", role, "required", perm)
			return c.Status(403).JSON(fiber.Map{
				"code":    403,
				"message": "permission denied",
			})
		}
		
		logger.Info("[Permission Check]", "result", "allowed")
		return c.Next()
	}
}

// RequireAnyPermission returns a middleware that checks if the user has ANY of the specified permissions.
func RequireAnyPermission(perms ...model.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		roleStr, _ := c.Locals("role").(string)
		role := model.Role(roleStr)
		
		permissions, exists := model.RolePermissions[role]
		if !exists {
			return c.Status(403).JSON(fiber.Map{
				"code":    403,
				"message": "invalid role",
			})
		}
		
		// Check if any of the required permissions are in the role's permission list
		for _, requiredPerm := range perms {
			for _, p := range permissions {
				if p == requiredPerm {
					return c.Next()
				}
			}
		}
		
		return c.Status(403).JSON(fiber.Map{
			"code":    403,
			"message": "permission denied",
		})
	}
}

// RequireRole returns a middleware that checks if the user has one of the specified roles.
func RequireRole(roles ...model.Role) fiber.Handler {
	return func(c fiber.Ctx) error {
		roleStr, _ := c.Locals("role").(string)
		userRole := model.Role(roleStr)
		
		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}
		
		return c.Status(403).JSON(fiber.Map{
			"code":    403,
			"message": "insufficient role",
		})
	}
}

// GetCurrentUser returns the current user's information from context.
func GetCurrentUser(c fiber.Ctx) (userID int64, username string, role model.Role, teamID *int64) {
	userID, _ = c.Locals("user_id").(int64)
	username, _ = c.Locals("username").(string)
	roleStr, _ := c.Locals("role").(string)
	role = model.Role(roleStr)
	teamID, _ = c.Locals("team_id").(*int64)
	return
}

