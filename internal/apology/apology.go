// Package apology provides templates and logging for system responsibility in failure scenarios.
package apology

import (
	"log"
	"time"
)

// ApologyType enumerates types of system failures.
type ApologyType int

const (
	ApologyNone ApologyType = iota
	ApologyHAOutage
	ApologyHardwareDegraded
	ApologyDelayedResponse
)

// ApologyTemplate returns a clear, professional apology for the given type.
func ApologyTemplate(t ApologyType) string {
	switch t {
	case ApologyHAOutage:
		return "We apologize for the inconvenience. The system is currently experiencing a Home Assistant outage. We are working to restore service."
	case ApologyHardwareDegraded:
		return "We apologize for the degraded performance. Some hardware components are not functioning optimally. Our team is addressing the issue."
	case ApologyDelayedResponse:
		return "We apologize for the delayed response. The system is taking longer than expected to process your request. Thank you for your patience."
	default:
		return ""
	}
}

// LogApology records an apology event internally.
func LogApology(t ApologyType, details string) {
	log.Printf("[Apology] %s | Details: %s | Time: %s", ApologyTemplate(t), details, time.Now().Format(time.RFC3339))
}
