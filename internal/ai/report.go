package ai

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateReport formats an analysis report into a readable string.
func GenerateReport(report *AnalysisReport) string {
	if report == nil {
		return "No report data available."
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Analysis Report - Team %s\n\n", report.TeamID))

	// Risk Score
	sb.WriteString(fmt.Sprintf("## Overall Risk Score: %.1f/100\n\n", report.RiskScore))

	// Attack Patterns
	if len(report.AttackPatterns) > 0 {
		sb.WriteString("## Attack Patterns Detected\n\n")
		// Sort by count descending
		sort.Slice(report.AttackPatterns, func(i, j int) bool {
			return report.AttackPatterns[i].Count > report.AttackPatterns[j].Count
		})
		for _, p := range report.AttackPatterns {
			sb.WriteString(fmt.Sprintf("- **%s** (Severity: %s): %d occurrences, confidence %.0f%%\n",
				p.Type, p.Severity, p.Count, p.Confidence*100))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("## Attack Patterns: None detected\n\n")
	}

	// Vulnerabilities
	if len(report.Vulnerabilities) > 0 {
		sb.WriteString("## Vulnerabilities Found\n\n")
		for _, v := range report.Vulnerabilities {
			sb.WriteString(fmt.Sprintf("- **%s** on %s (Severity: %s, Confidence: %.0f%%)\n",
				v.Type, v.Service, v.Severity, v.Confidence*100))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("## Vulnerabilities: None detected\n\n")
	}

	// Hardening Tips
	if len(report.HardeningTips) > 0 {
		sb.WriteString("## Hardening Recommendations\n\n")
		sort.Slice(report.HardeningTips, func(i, j int) bool {
			return report.HardeningTips[i].Priority < report.HardeningTips[j].Priority
		})
		for _, t := range report.HardeningTips {
			sb.WriteString(fmt.Sprintf("%d. **[%s]** %s: %s\n", t.Priority, t.Service, t.Action, t.Detail))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("## Hardening: No specific recommendations\n\n")
	}

	return sb.String()
}

// GenerateJSONReport formats the report as a map suitable for JSON output.
func GenerateJSONReport(report *AnalysisReport) map[string]interface{} {
	if report == nil {
		return map[string]interface{}{"error": "no report data"}
	}

	attackTypes := make([]string, 0, len(report.AttackPatterns))
	for _, p := range report.AttackPatterns {
		attackTypes = append(attackTypes, p.Type)
	}

	vulnServices := make([]string, 0, len(report.Vulnerabilities))
	for _, v := range report.Vulnerabilities {
		vulnServices = append(vulnServices, v.Service)
	}

	highPriorityTips := 0
	for _, t := range report.HardeningTips {
		if t.Priority <= 2 {
			highPriorityTips++
		}
	}

	return map[string]interface{}{
		"team_id":               report.TeamID,
		"risk_score":            report.RiskScore,
		"attack_pattern_count":  len(report.AttackPatterns),
		"vulnerability_count":   len(report.Vulnerabilities),
		"hardening_tip_count":   len(report.HardeningTips),
		"high_priority_tips":    highPriorityTips,
		"attack_types":          attackTypes,
		"vulnerable_services":   vulnServices,
	}
}
