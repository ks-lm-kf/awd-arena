package middleware

import (
	"encoding/json"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

// AuditLogger records user actions to audit logs
func AuditLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Continue with request
		err := c.Next()
		
		// Only log if user is authenticated
		userID, ok := c.Locals("user_id").(int64)
		if !ok {
			return err
		}
		
		username, _ := c.Locals("username").(string)
		path := c.Path()
		method := c.Method()
		
		// Determine action and resource type based on path
		action := getActionFromMethod(method)
		resourceType := getResourceTypeFromPath(path)
		
		// Skip certain paths
		if resourceType == "" || path == "/api/v1/auth/me" {
			return err
		}
		
		// Create audit log
		log := &model.AdminLog{
			UserID:       userID,
			Username:     username,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   0,
			Description:  method + " " + path,
			IPAddress:    c.IP(),
			UserAgent:    string(c.Request().Header.UserAgent()),
			Details:      getRequestDetails(c),
			CreatedAt:    time.Now(),
		}
		
		db := database.GetDB()
		db.Create(log)
		
		return err
	}
}

func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "view"
	default:
		return method
	}
}

func getResourceTypeFromPath(path string) string {
	if len(path) >= 8 && path[:8] == "/api/v1/" {
		rest := path[8:]
		// Get first segment
		for i, c := range rest {
			if c == '/' {
				return rest[:i]
			}
		}
		return rest
	}
	return ""
}

func getRequestDetails(c fiber.Ctx) string {
	details := map[string]interface{}{
		"method": c.Method(),
		"path":   c.Path(),
	}
	
	if c.Method() != "GET" && len(c.Body()) > 0 && len(c.Body()) < 1000 {
		details["body"] = string(c.Body())
	}
	
	if b, err := json.Marshal(details); err == nil {
		return string(b)
	}
	return ""
}

