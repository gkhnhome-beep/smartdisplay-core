// Package scorecard computes a simple system quality score for UI display.
package scorecard

// StatusScore holds the system's quality scores.
type StatusScore struct {
	Security  string `json:"security"` // e.g. "Good", "Warning", "Critical"
	Stability string `json:"stability"`
	Awareness string `json:"awareness"`
}

// ComputeStatusScore returns a StatusScore based on system state (stub: replace with real checks).
func ComputeStatusScore(securityOK, stable, aware bool) StatusScore {
	return StatusScore{
		Security:  labelFor(securityOK),
		Stability: labelFor(stable),
		Awareness: labelFor(aware),
	}
}

func labelFor(ok bool) string {
	if ok {
		return "Good"
	}
	return "Warning"
}
