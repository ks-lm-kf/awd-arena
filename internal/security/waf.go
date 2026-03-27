package security

import (
	"net"
	"regexp"
	"strings"
	"sync"
)

// GlobalWAF is the global WAF instance
var GlobalWAF *WAFEngine

// WAFEngine implements WAF rule checking with regex support.
type WAFEngine struct {
	mu       sync.RWMutex
	rules    []WAFFilterRule
	logStore *AttackLogStore
}

// WAFFilterRule represents a single WAF filter rule.
type WAFFilterRule struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // sql_injection, xss, command_injection, path_traversal
	Patterns []string `json:"patterns"`
	Severity string   `json:"severity"` // low, medium, high, critical
	Action   string   `json:"action"`   // block, log, allow
	compiled []*regexp.Regexp
}

// WAFResult holds the result of a WAF check.
type WAFResult struct {
	Blocked  bool   `json:"blocked"`
	Reason   string `json:"reason,omitempty"`
	Rule     string `json:"rule,omitempty"`
	RuleType string `json:"rule_type,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// NewWAFEngine creates a new WAF engine with default rules.
func NewWAFEngine(logStore *AttackLogStore) *WAFEngine {
	w := &WAFEngine{
		logStore: logStore,
	}
	
	// Enhanced SQL Injection Rules
	sqlInjectionRules := []string{
		// Union-based injection
		`(?i)(\bUNION\b\s+\bSELECT\b)`,
		`(?i)(\bUNION\b\s+\bALL\b\s+\bSELECT\b)`,
		
		// Boolean-based injection
		`(?i)(\bOR\b\s+1\s*=\s*1)`,
		`(?i)(\bOR\b\s+['"]?\d+['"]?\s*=\s*['"]?\d+)`,
		`(?i)(\bAND\b\s+1\s*=\s*1)`,
		`(?i)(\bOR\b\s+['"]true['"]\s*=\s*['"]true['"])`,
		`(?i)(\bAND\b\s+['"]true['"]\s*=\s*['"]true['"])`,
		
		// Time-based injection
		`(?i)(\bSLEEP\s*\()`,
		`(?i)(\bBENCHMARK\s*\()`,
		`(?i)(\bWAITFOR\b\s+\bDELAY\b)`,
		`(?i)(\bPG_SLEEP\s*\()`,
		
		// Destructive operations
		`(?i)(;\s*DROP\s+\bTABLE\b)`,
		`(?i)(;\s*DELETE\s+\bFROM\b)`,
		`(?i)(;\s*INSERT\s+\bINTO\b)`,
		`(?i)(;\s*UPDATE\b.*\bSET\b)`,
		`(?i)(;\s*TRUNCATE\s+\bTABLE?\b)`,
		
		// SQL functions
		`(?i)(\bEXTRACTVALUE\s*\()`,
		`(?i)(\bUPDATEXML\s*\()`,
		`(?i)(\bLOAD_FILE\s*\()`,
		`(?i)(\bINTO\s+\bOUTFILE\b)`,
		`(?i)(\bINTO\s+\bDUMPFILE\b)`,
		
		// Comment injection
		`(?i)(--\s*$)`,
		`(?i)(/\*.*\*/)`,
		`(?i)(#.*$)`,
		
		// String escape
		`(?i)('\s*OR\s+')`,
		`(?i)('\s*AND\s+')`,
		`(?i)('\s*;\s*--)`,
		
		// Stacked queries
		`(?i)(;\s*SELECT\b)`,
		`(?i)(;\s*EXEC\b)`,
		`(?i)(;\s*EXECUTE\b)`,
		
		// Information schema
		`(?i)(\binformation_schema\b)`,
		`(?i)(\bsys\.tables\b)`,
		`(?i)(\bsysobjects\b)`,
		`(?i)(\bsyscolumns\b)`,
		
		// Hex encoding
		`(?i)(\b0x[0-9a-f]+\b)`,
		
		// CHAR encoding
		`(?i)(\bCHAR\s*\(\s*\d+\s*\))`,
		
		// CONCAT
		`(?i)(\bCONCAT\s*\(\s*['"]|%27)`,
		
		// HAVING
		`(?i)(\bHAVING\s+\d+\s*=\s*\d+)`,
		
		// ORDER BY injection
		`(?i)(\bORDER\s+\bBY\s+\d+)`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "sql_injection",
		Type:     "sql_injection",
		Severity: "high",
		Action:   "block",
		Patterns: sqlInjectionRules,
	})
	
	// Enhanced XSS Rules
	xssRules := []string{
		// Script tags
		`(?i)<\s*script[^>]*>`,
		`(?i)<\s*/\s*script\s*>`,
		
		// JavaScript protocol
		`(?i)javascript\s*:`,
		`(?i)vbscript\s*:`,
		`(?i)livescript\s*:`,
		
		// Event handlers
		`(?i)\bon\w+\s*=`,
		`(?i)\bon\w+\s*=\s*["']`,
		
		// DOM manipulation
		`(?i)document\.(cookie|domain|write|location)`,
		`(?i)window\.(location|open|navigate)`,
		
		// JavaScript functions
		`(?i)eval\s*\(`,
		`(?i)alert\s*\(`,
		`(?i)prompt\s*\(`,
		`(?i)confirm\s*\(`,
		`(?i)setTimeout\s*\(`,
		`(?i)setInterval\s*\(`,
		`(?i)Function\s*\(`,
		
		// Image and iframe injection
		`(?i)<\s*img[^>]+onerror`,
		`(?i)<\s*img[^>]+src\s*=\s*["']?javascript:`,
		`(?i)<\s*svg[^>]+onload`,
		`(?i)<\s*svg[^>]+onerror`,
		`(?i)<\s*iframe[^>]*>`,
		`(?i)<\s*iframe[^>]+src\s*=`,
		`(?i)<\s*embed[^>]*>`,
		`(?i)<\s*object[^>]*>`,
		
		// CSS expressions
		`(?i)expression\s*\(`,
		`(?i)behavior\s*:`,
		`(?i)-moz-binding\s*:`,
		
		// HTML5 sources
		`(?i)<\s*video[^>]+onerror`,
		`(?i)<\s*audio[^>]+onerror`,
		`(?i)<\s*body[^>]+onload`,
		`(?i)<\s*input[^>]+onfocus`,
		`(?i)<\s*marquee[^>]+onstart`,
		
		// Data URI
		`(?i)data\s*:\s*text/html`,
		
		// SVG/XML injection
		`(?i)<\s*\?xml`,
		`(?i)<!DOCTYPE`,
		`(?i)<\s*entity`,
		
		// Unicode encoding
		`(?i)&#x?[0-9a-f]+;?`,
		
		// HTML entities bypass
		`(?i)&lt;script`,
		`(?i)&#60;script`,
		
		// Meta refresh
		`(?i)<\s*meta[^>]+http-equiv\s*=\s*["']?refresh`,
		
		// Base tag hijacking
		`(?i)<\s*base[^>]+href`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "xss",
		Type:     "xss",
		Severity: "medium",
		Action:   "block",
		Patterns: xssRules,
	})
	
	// Command Injection Rules
	commandInjectionRules := []string{
		// Basic commands
		`;\s*(ls|cat|id|whoami|uname|pwd|wget|curl|nc|netcat)\b`,
		`\|\s*(cat|nc|bash|sh|wget|curl|ls|id)\b`,
		
		// Backticks
		"`[^`]+`",
		
		// Command substitution
		`\$\([^)]+\)`,
		`\$\{[^}]+\}`,
		
		// Chained commands
		`&&\s*(rm|wget|curl|nc|bash|sh|chmod|chown|kill|ps)\b`,
		`\|\|\s*(rm|wget|curl|nc|bash|sh)\b`,
		
		// PHP functions
		`(?i)(system|exec|passthru|popen|proc_open|shell_exec|pcntl_exec)\s*\(`,
		
		// Sensitive files
		`(?i)/etc/(passwd|shadow|hosts|crontab|group|sudoers)`,
		`(?i)/proc/self/`,
		`(?i)/var/log/`,
		
		// Network utilities
		`(?i)\b(nc|netcat|telnet|ftp|tftp)\s+.*-e`,
		`(?i)\bbash\s+-[ci]`,
		`(?i)\bpython\s+-c`,
		`(?i)\bperl\s+-e`,
		`(?i)\bruby\s+-e`,
		
		// Redirection
		`>\s*/`,
		`>>\s*/`,
		
		// Wildcard abuse
		`\*\s+/`,
		`\?\s+/`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "command_injection",
		Type:     "command_injection",
		Severity: "critical",
		Action:   "block",
		Patterns: commandInjectionRules,
	})
	
	// Path Traversal Rules
	pathTraversalRules := []string{
		// Basic traversal
		`\.\./`,
		`\.\.\\`,
		
		// Encoded traversal
		`(?i)\.\.%2[fF]`,
		`(?i)\.\.%5[cC]`,
		`(?i)%2[eE]%2[eE]%2[fF]`,
		`(?i)%2[eE]%2[eE]%5[cC]`,
		
		// Null byte
		`(?i)%00`,
		
		// Sensitive files
		`(?i)/etc/passwd`,
		`(?i)/etc/shadow`,
		`(?i)/proc/self`,
		`(?i)/windows/system32`,
		`(?i)boot\.ini`,
		`(?i)win\.ini`,
		
		// Double encoding
		`(?i)%252[eE]`,
		
		// UTF-8 encoding
		`(?i)\.\./\.\./`,
		`(?i)\.\.%c0%af`,
		`(?i)\.\.%c1%9c`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "path_traversal",
		Type:     "path_traversal",
		Severity: "high",
		Action:   "block",
		Patterns: pathTraversalRules,
	})
	
	// LDAP Injection Rules
	ldapInjectionRules := []string{
		`(?i)\(\|`,
		`(?i)\(&`,
		`(?i)\(\!`,
		`(?i)\)\(`,
		`(?i)=\*`,
		`(?i)\*\)`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "ldap_injection",
		Type:     "ldap_injection",
		Severity: "high",
		Action:   "block",
		Patterns: ldapInjectionRules,
	})
	
	// XML/XXE Injection Rules
	xxeInjectionRules := []string{
		`(?i)<!ENTITY`,
		`(?i)SYSTEM\s+"`,
		`(?i)PUBLIC\s+"`,
		`(?i)<!DOCTYPE.*\[`,
		`(?i)<!ATTLIST`,
		`(?i)<!ELEMENT`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "xxe_injection",
		Type:     "xxe_injection",
		Severity: "critical",
		Action:   "block",
		Patterns: xxeInjectionRules,
	})
	
	// SSRF Protection Rules
	ssrfRules := []string{
		`(?i)^(http|https|ftp|file)://127\.0\.0\.1`,
		`(?i)^(http|https|ftp|file)://localhost`,
		`(?i)^(http|https|ftp|file)://0\.0\.0\.0`,
		`(?i)^(http|https|ftp|file)://192\.168\.`,
		`(?i)^(http|https|ftp|file)://10\.`,
		`(?i)^(http|https|ftp|file)://172\.(1[6-9]|2[0-9]|3[0-1])\.`,
		`(?i)^(http|https|ftp|file)://\[::1\]`,
		`(?i)^(http|https|ftp|file)://\[fe80:`,
		`(?i)@.*\.(internal|local|lan)`,
	}
	
	w.AddRule(WAFFilterRule{
		Name:     "ssrf",
		Type:     "ssrf",
		Severity: "high",
		Action:   "block",
		Patterns: ssrfRules,
	})
	
	return w
}

// Check inspects input against WAF rules.
func (w *WAFEngine) Check(input string) *WAFResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	lowerInput := strings.ToLower(input)
	for _, rule := range w.rules {
		for i, pattern := range rule.Patterns {
			var matched bool
			if i < len(rule.compiled) && rule.compiled[i] != nil {
				matched = rule.compiled[i].MatchString(input)
			} else {
				matched = strings.Contains(lowerInput, strings.ToLower(pattern))
			}
			if matched {
				return &WAFResult{
					Blocked:  rule.Action == "block",
					Reason:   "matched pattern: " + pattern,
					Rule:     rule.Name,
					RuleType: rule.Type,
					Severity: rule.Severity,
				}
			}
		}
	}
	return &WAFResult{Blocked: false}
}

// CheckRequest inspects a full HTTP request's common attack vectors.
func (w *WAFEngine) CheckRequest(query, body, path string) *WAFResult {
	// Check path
	if result := w.Check(path); result.Blocked {
		return result
	}
	// Check query string
	if result := w.Check(query); result.Blocked {
		return result
	}
	// Check body
	if result := w.Check(body); result.Blocked {
		return result
	}
	return &WAFResult{Blocked: false}
}

// AddRule adds a custom WAF rule with compiled regex.
func (w *WAFEngine) AddRule(rule WAFFilterRule) {
	w.mu.Lock()
	defer w.mu.Unlock()
	compiled := make([]*regexp.Regexp, len(rule.Patterns))
	for i, pattern := range rule.Patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled[i] = re
		}
	}
	rule.compiled = compiled
	w.rules = append(w.rules, rule)
}

// GetRules returns all configured rules.
func (w *WAFEngine) GetRules() []WAFFilterRule {
	w.mu.RLock()
	defer w.mu.RUnlock()
	// Return copy without compiled regex
	result := make([]WAFFilterRule, len(w.rules))
	copy(result, w.rules)
	for i := range result {
		result[i].compiled = nil
	}
	return result
}

// IsIPBlocked checks if an IP is in the blocklist.
func (w *WAFEngine) IsIPBlocked(ip string) bool {
	// Placeholder: could integrate with a persistent blocklist
	return false
}

// SanitizeIP extracts IP from X-Forwarded-For or RemoteAddr.
func SanitizeIP(remoteAddr string) string {
	// Handle RemoteAddr which includes port
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
