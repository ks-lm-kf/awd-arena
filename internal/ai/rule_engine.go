package ai

import (
	"context"
	"strings"
)

// RuleEngine implements rule-based analysis.
type RuleEngine struct {
	rules []AnalysisRule
}

// AnalysisRule defines a single analysis rule.
type AnalysisRule struct {
	Name     string   `json:"name"`
	Category string   `json:"category"` // attack, vulnerability, hardening
	Patterns []string `json:"patterns"`
	Severity string   `json:"severity"`
	Action   string   `json:"action"`
}

// NewRuleEngine creates a new rule engine with default rules.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules: []AnalysisRule{
			{
				Name:     "sql_injection_detect",
				Category: "attack",
				Patterns: []string{"UNION", "SELECT.*FROM", "OR 1=1", "SLEEP", "BENCHMARK"},
				Severity: "high",
				Action:   "detect",
			},
			{
				Name:     "xss_detect",
				Category: "attack",
				Patterns: []string{"<script", "onerror=", "document.cookie", "alert("},
				Severity: "medium",
				Action:   "detect",
			},
			{
				Name:     "rce_detect",
				Category: "attack",
				Patterns: []string{"; ls", "&& cat", "system(", "exec(", "passthru("},
				Severity: "critical",
				Action:   "detect",
			},
			{
				Name:     "path_traversal",
				Category: "vulnerability",
				Patterns: []string{"../", "..\\", "/etc/passwd", "/proc/self"},
				Severity: "high",
				Action:   "detect",
			},
		},
	}
}

// AnalyzeAttackPatterns identifies attack patterns in logs.
func (re *RuleEngine) AnalyzeAttackPatterns(ctx context.Context, gameID, teamID string) []AttackPattern {
	var patterns []AttackPattern
	for _, rule := range re.rules {
		if rule.Category == "attack" {
			patterns = append(patterns, AttackPattern{
				Type:       rule.Name,
				Severity:   rule.Severity,
				Confidence: 0.8,
			})
		}
	}
	return patterns
}

// DetectVulnerabilities identifies potential vulnerabilities.
func (re *RuleEngine) DetectVulnerabilities(ctx context.Context, gameID, teamID string) []Vulnerability {
	var vulns []Vulnerability
	for _, rule := range re.rules {
		if rule.Category == "vulnerability" {
			vulns = append(vulns, Vulnerability{
				Type:       rule.Name,
				Severity:   rule.Severity,
				Confidence: 0.7,
			})
		}
	}
	return vulns
}

// GenerateHardeningTips generates recommendations.
func (re *RuleEngine) GenerateHardeningTips(ctx context.Context, gameID, teamID string) []HardeningTip {
	return []HardeningTip{
		{Priority: 1, Service: "web", Action: "input_validation", Detail: "Implement strict input validation and parameterized queries"},
		{Priority: 2, Service: "web", Action: "output_encoding", Detail: "Encode all user-controlled output to prevent XSS"},
		{Priority: 3, Service: "web", Action: "waf_config", Detail: "Configure WAF rules to block common attack patterns"},
	}
}

// MatchRule checks input against rules.
func (re *RuleEngine) MatchRule(input string) *AnalysisRule {
	lower := strings.ToLower(input)
	for i := range re.rules {
		for _, pattern := range re.rules[i].Patterns {
			if strings.Contains(lower, strings.ToLower(pattern)) {
				return &re.rules[i]
			}
		}
	}
	return nil
}

// AddRule adds a custom analysis rule.
func (re *RuleEngine) AddRule(rule AnalysisRule) {
	re.rules = append(re.rules, rule)
}
