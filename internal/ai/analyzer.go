package ai

import (
	"context"
)

// AIAnalyzer is the main AI analysis entry point.
type AIAnalyzer struct {
	ruleEngine    *RuleEngine
	statsAnalyzer *StatsAnalyzer
}

// StatsAnalyzer provides statistical analysis.
type StatsAnalyzer struct{}

// NewAIAnalyzer creates a new AI analyzer.
func NewAIAnalyzer(ruleEngine *RuleEngine) *AIAnalyzer {
	return &AIAnalyzer{
		ruleEngine:    ruleEngine,
		statsAnalyzer: &StatsAnalyzer{},
	}
}

// AnalysisReport holds the analysis result.
type AnalysisReport struct {
	TeamID          string           `json:"team_id"`
	AttackPatterns  []AttackPattern  `json:"attack_patterns"`
	Vulnerabilities []Vulnerability  `json:"vulnerabilities"`
	HardeningTips   []HardeningTip   `json:"hardening_tips"`
	RiskScore       float64          `json:"risk_score"`
}

// AttackPattern describes an identified attack pattern.
type AttackPattern struct {
	Type       string  `json:"type"` // sql_injection, xss, rce, etc.
	Count      int     `json:"count"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
}

// Vulnerability describes a found vulnerability.
type Vulnerability struct {
	Service    string  `json:"service"`
	Type       string  `json:"type"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
}

// HardeningTip provides a hardening recommendation.
type HardeningTip struct {
	Priority  int    `json:"priority"`
	Service   string `json:"service"`
	Action    string `json:"action"`
	Detail    string `json:"detail"`
}

// AnalyzeTeam generates an analysis report for a team.
func (a *AIAnalyzer) AnalyzeTeam(ctx context.Context, gameID string, teamID string) (*AnalysisReport, error) {
	report := &AnalysisReport{TeamID: teamID}
	// Use rule engine to analyze attack patterns
	report.AttackPatterns = a.ruleEngine.AnalyzeAttackPatterns(ctx, gameID, teamID)
	report.Vulnerabilities = a.ruleEngine.DetectVulnerabilities(ctx, gameID, teamID)
	report.HardeningTips = a.ruleEngine.GenerateHardeningTips(ctx, gameID, teamID)
	// Calculate risk score
	return report, nil
}

// AnalyzeGame generates a game-wide analysis.
func (a *AIAnalyzer) AnalyzeGame(ctx context.Context, gameID string) (*AnalysisReport, error) {
	return &AnalysisReport{TeamID: "game_summary"}, nil
}
