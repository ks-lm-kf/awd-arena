package middleware

import (
	"fmt"
	"time"

	"github.com/awd-platform/awd-arena/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TeamID   *int64 `json:"team_id"`
	jwt.RegisteredClaims
}

// getJWTSecret returns the JWT secret from config.
func getJWTSecret() string {
	return config.C.Server.JWTSecret
}

// parseJWTToken parses and validates a JWT token string.
func parseJWTToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// JWTAuth returns a JWT authentication middleware.
func JWTAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"code": 401, "message": "missing authorization header"})
		}

		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		secret := getJWTSecret()
		token, err := parseJWTToken(tokenString, secret)
		if err != nil || !token.Valid {
			logger.Warn("JWT validation failed", "error", err)
			return c.Status(401).JSON(fiber.Map{"code": 401, "message": "invalid or expired token"})
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"code": 401, "message": "invalid token claims"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("team_id", claims.TeamID)
		return c.Next()
	}
}

// AdminOnly returns a middleware that requires admin role.
func AdminOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{"code": 403, "message": "admin access required"})
		}
		return c.Next()
	}
}

// GenerateToken generates a JWT token.
func GenerateToken(userID int64, username, role string, teamID *int64) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TeamID:   teamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.C.Security.JWTExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.Server.JWTSecret))
}

// GetJWTSecret returns the JWT secret from config.
func GetJWTSecret() string {
	return config.C.Server.JWTSecret
}

// ValidateToken parses and validates a JWT token from the request.
func ValidateToken(c fiber.Ctx, secret string) (int64, error) {
	logger.Info("[JWTAuth DEBUG]", "path", c.Path(), "method", c.Method())
		authHeader := c.Get("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("missing authorization header")
	}
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}
	token, err := parseJWTToken(tokenString, secret)
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid or expired token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}
	return claims.UserID, nil
}

// JudgeOnly returns a middleware that requires judge or admin role.
func JudgeOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" && role != "judge" {
			return c.Status(403).JSON(fiber.Map{"code": 403, "message": "judge or admin access required"})
		}
		return c.Next()
	}
}
