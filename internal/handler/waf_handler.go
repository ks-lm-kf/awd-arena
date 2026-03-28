package handler

import (
	"github.com/awd-platform/awd-arena/internal/security"
	"github.com/gofiber/fiber/v3"
)

// GetWAFRules returns all WAF rules.
func GetWAFRules(c fiber.Ctx) error {
	if security.GlobalWAF == nil {
		return c.JSON(fiber.Map{"code": 0, "data": []interface{}{}})
	}
	rules := security.GlobalWAF.GetRules()
	return c.JSON(fiber.Map{"code": 0, "data": rules})
}

// AddWAFRule adds a new WAF rule.
func AddWAFRule(c fiber.Ctx) error {
	var req struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Patterns []string `json:"patterns"`
		Severity string   `json:"severity"`
		Action   string   `json:"action"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request"})
	}
	if req.Name == "" || len(req.Patterns) == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "name and patterns required"})
	}
	if security.GlobalWAF == nil {
		security.GlobalWAF = security.NewWAFEngine(security.NewAttackLogStore(1000))
	}

	rule := security.WAFFilterRule{
		Name:     req.Name,
		Type:     req.Type,
		Patterns: req.Patterns,
		Severity: req.Severity,
		Action:   req.Action,
	}
	if rule.Severity == "" {
		rule.Severity = "medium"
	}
	if rule.Action == "" {
		rule.Action = "block"
	}

	security.GlobalWAF.AddRule(rule)
	return c.JSON(fiber.Map{"code": 0, "message": "rule added", "data": rule})
}

// CheckWAFInput tests input against WAF rules.
func CheckWAFInput(c fiber.Ctx) error {
	var req struct {
		Input string `json:"input"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request"})
	}
	if security.GlobalWAF == nil {
		return c.JSON(fiber.Map{"code": 0, "data": fiber.Map{"blocked": false}})
	}
	result := security.GlobalWAF.Check(req.Input)
	return c.JSON(fiber.Map{"code": 0, "data": result})
}

// GetWAFStats returns WAF statistics.
func GetWAFStats(c fiber.Ctx) error {
	if security.GlobalWAF == nil {
		return c.JSON(fiber.Map{"code": 0, "data": fiber.Map{"total_rules": 0}})
	}
	rules := security.GlobalWAF.GetRules()
	return c.JSON(fiber.Map{
		"code": 0,
		"data": fiber.Map{
			"total_rules": len(rules),
			"rule_types":  countRuleTypes(rules),
		},
	})
}

func countRuleTypes(rules []security.WAFFilterRule) map[string]int {
	counts := make(map[string]int)
	for _, r := range rules {
		counts[r.Type]++
	}
	return counts
}
