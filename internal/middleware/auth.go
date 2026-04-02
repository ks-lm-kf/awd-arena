package middleware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/config"
	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var tokenBlacklist sync.Map

func BlacklistToken(tokenString string) {
	expiry := time.Now().Add(24 * time.Hour)
	secret := getJWTSecret()
	if token, err := parseJWTToken(tokenString, secret); err == nil && token.Valid {
		if claims, ok := token.Claims.(*Claims); ok {
			if claims.ExpiresAt != nil {
				expiry = claims.ExpiresAt.Time
			}
		}
	}
	tokenBlacklist.Store(tokenString, expiry)
	cleanupBlacklist()
}

func cleanupBlacklist() {
	now := time.Now()
	tokenBlacklist.Range(func(key, value interface{}) bool {
		if exp, ok := value.(time.Time); ok && now.After(exp) {
			tokenBlacklist.Delete(key)
		}
		return true
	})
}

func isTokenBlacklisted(token string) bool {
	expiry, ok := tokenBlacklist.Load(token)
	if !ok {
		return false
	}
	if exp, ok := expiry.(time.Time); ok && time.Now().After(exp) {
		tokenBlacklist.Delete(token)
		return false
	}
	return true
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TeamID   *int64 `json:"team_id"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	return config.C.Server.JWTSecret
}

func parseJWTToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
}

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

		if isTokenBlacklisted(tokenString) {
			return c.Status(401).JSON(fiber.Map{"code": 401, "message": "token has been revoked"})
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

func AdminOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{"code": 403, "message": "admin access required"})
		}
		return c.Next()
	}
}

func EnforcePasswordChange() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int64)
		if !ok || userID == 0 {
			return c.Next()
		}

		path := c.Path()
		allowedPrefixes := []string{
			"/api/v1/auth/change-password",
			"/api/v1/auth/password",
			"/api/v1/auth/logout",
			"/api/v1/auth/me",
		}
		for _, p := range allowedPrefixes {
			if strings.HasPrefix(path, p) {
				return c.Next()
			}
		}

		db := database.GetDB()
		if db == nil {
			return c.Next()
		}

		var mustChange bool
		if err := db.Table("users").Select("must_change_password").
			Where("id = ?", userID).Scan(&mustChange).Error; err == nil {
			if mustChange {
				return c.Status(403).JSON(fiber.Map{
					"code":    403,
					"message": "password change required before accessing this resource",
				})
			}
		}

		return c.Next()
	}
}

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

func GetJWTSecret() string {
	return config.C.Server.JWTSecret
}

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

func JudgeOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" && role != "judge" {
			return c.Status(403).JSON(fiber.Map{"code": 403, "message": "judge or admin access required"})
		}
		return c.Next()
	}
}
