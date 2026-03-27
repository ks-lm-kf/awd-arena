package handler

import (
"github.com/awd-platform/awd-arena/internal/security"
"github.com/gofiber/fiber/v3"
)

// GetWAFRules returns WAF rules (placeholder).
func GetWAFRules(c fiber.Ctx) error {
return c.JSON(fiber.Map{
"rules": []interface{}{},
"message": "WAF rules not implemented yet",
})
}

// AttackStore is the global attack log store.
var AttackStore = security.NewAttackLogStore(1000)
