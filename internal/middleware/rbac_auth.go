package middleware

import (
	"strings"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware extracts user information from JWT
func AuthMiddleware(c fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{
			"code":    401,
			"message": "missing authorization header",
		})
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(401).JSON(fiber.Map{
			"code":    401,
			"message": "invalid authorization format",
		})
	}

	token := parts[1]

	// TODO: Validate JWT token and extract user info
	// For now, we'll store the token in context for later use
	c.Locals("token", token)

	// Mock user for development - in production, this should come from JWT validation
	// You'll need to implement proper JWT validation here
	user := &model.User{
		ID:   1,
		Role: string(model.RoleAdmin), // Default to admin for testing
	}
	c.Locals("user", user)

	// Also set individual fields for compatibility with permission.go
	c.Locals("user_id", user.ID)
	c.Locals("username", user.Username)
	c.Locals("role", user.Role)
	c.Locals("team_id", user.TeamID)

	return c.Next()
}

// HasPermission checks if a role has a specific permission
func HasPermission(role model.Role, permission model.Permission) bool {
	permissions, exists := model.RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRoleOrHigher checks if user has the required role or a higher one
func HasRoleOrHigher(userRole, requiredRole model.Role) bool {
	roleHierarchy := map[model.Role]int{
		model.RolePlayer:    1,
		model.RoleOrganizer: 2,
		model.RoleAdmin:     3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// OptionalAuth extracts user info if present but doesn't require it
func OptionalAuth(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	// Same logic as AuthMiddleware but don't return error if invalid
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		token := parts[1]
		c.Locals("token", token)

		// Mock user - replace with actual JWT validation
		user := &model.User{
			ID:   1,
			Role: string(model.RolePlayer),
		}
		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("role", user.Role)
		c.Locals("team_id", user.TeamID)
	}

	return c.Next()
}
