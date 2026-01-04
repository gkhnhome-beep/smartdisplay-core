package principles

// FilterOutput validates an AI output and returns it if valid, or a safe fallback if not.
func FilterOutput(output string, principles ProductPrinciples) string {
	valid, violations := ValidateOutput(output, principles)
	if !valid {
		LogRejection(output, violations)
		return SafeFallback()
	}
	return output
}

// SafeFallback returns a universally safe message when output is rejected.
func SafeFallback() string {
	return "The system is operating normally. Please let me know how I can help."
}
