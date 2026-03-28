package middleware

import (
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

// HasPermission checks if a role has a specific permission.
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

// HasRoleOrHigher checks if user has the required role or a higher one.
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

// OptionalAuth extracts user info if present but doesn't require it.
func OptionalAuth(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}
	// Use the real JWTAuth logic
	secret := getJWTSecret()
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}
	token, err := parseJWTToken(tokenString, secret)
	if err != nil || !token.Valid {
		return c.Next() // Silently skip for optional auth
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return c.Next()
	}
	c.Locals("user_id", claims.UserID)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	c.Locals("team_id", claims.TeamID)
	return c.Next()
}
