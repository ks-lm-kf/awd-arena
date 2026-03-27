package security

import (
	"strings"
	"unicode"
)

// ValidatePassword checks password strength.
// Requirements: min 8 chars, at least 1 uppercase, 1 lowercase, 1 digit.
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters"
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasUpper {
		return false, "password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "password must contain at least one lowercase letter"
	}
	if !hasDigit {
		return false, "password must contain at least one digit"
	}
	return true, ""
}

// SanitizeInput performs basic input sanitization.
func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}

// SeverityForType returns default severity for an attack type.
func SeverityForType(attackType string) string {
	switch attackType {
	case "command_injection":
		return "critical"
	case "sql_injection":
		return "high"
	case "path_traversal":
		return "high"
	case "xss":
		return "medium"
	case "brute_force":
		return "medium"
	default:
		return "low"
	}
}
