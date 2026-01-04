// Package principles defines and validates the product's core identity and behavior consistency.
package principles

import (
	"log"
	"strings"
)

// ProductPrinciples are the core values of the system.
type ProductPrinciples struct {
	Calm        bool // Tone must be soothing, not alarmist
	Predictable bool // Behavior must be consistent and understandable
	Respectful  bool // Never condescending, always user-centric
	Protective  bool // Security and safety first, but transparent
}

// DefaultPrinciples returns the standard product principles.
func DefaultPrinciples() ProductPrinciples {
	return ProductPrinciples{
		Calm:        true,
		Predictable: true,
		Respectful:  true,
		Protective:  true,
	}
}

// ValidateOutput checks if an AI output aligns with product principles.
func ValidateOutput(output string, principles ProductPrinciples) (valid bool, violations []string) {
	valid = true
	violations = []string{}

	if principles.Calm {
		if containsAlarmist(output) {
			valid = false
			violations = append(violations, "violates_calm: output is unnecessarily alarming")
		}
	}

	if principles.Predictable {
		if containsUnexpectedBehavior(output) {
			valid = false
			violations = append(violations, "violates_predictable: behavior may be unpredictable")
		}
	}

	if principles.Respectful {
		if containsCondescending(output) {
			valid = false
			violations = append(violations, "violates_respectful: output is condescending or demeaning")
		}
	}

	if principles.Protective {
		if containsSecurityGap(output) {
			valid = false
			violations = append(violations, "violates_protective: security concern detected")
		}
	}

	return valid, violations
}

// containsAlarmist checks for unnecessarily alarming language.
func containsAlarmist(output string) bool {
	alarming := []string{"DANGER", "DISASTER", "CRITICAL!!!", "EMERGENCY NOW"}
	lower := strings.ToUpper(output)
	for _, word := range alarming {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// containsUnexpectedBehavior checks for inconsistent or surprising statements.
func containsUnexpectedBehavior(output string) bool {
	// Stub: could check for behavioral contradictions
	return false
}

// containsCondescending checks for disrespectful language.
func containsCondescending(output string) bool {
	condescending := []string{"obviously", "just", "clearly", "obviously you"}
	lower := strings.ToLower(output)
	for _, word := range condescending {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// containsSecurityGap checks for security-related concerns.
func containsSecurityGap(output string) bool {
	// Stub: could check for security bypass suggestions
	return false
}

// LogRejection logs a rejected output with its violations.
func LogRejection(output string, violations []string) {
	log.Printf("[Principles] Rejected output: %q | Violations: %v", output, violations)
}
