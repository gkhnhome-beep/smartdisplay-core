// Package security provides redaction utilities for secrets.
package security

// Redact replaces any secret/token with a fixed string for logs or output.
func Redact(_ string) string {
	return "***"
}
