package trust

// Tone returns the system's tone based on TrustScoreLevel.
func Tone() string {
	switch GetTrustScoreLevel() {
	case TrustLow:
		return "urgent"
	case TrustHigh:
		return "friendly"
	default:
		return "neutral"
	}
}

// ExplanationDepth returns how detailed explanations should be.
func ExplanationDepth() string {
	switch GetTrustScoreLevel() {
	case TrustLow:
		return "detailed"
	case TrustHigh:
		return "concise"
	default:
		return "normal"
	}
}

// AlertUrgency returns the urgency level for alerts.
func AlertUrgency() string {
	switch GetTrustScoreLevel() {
	case TrustLow:
		return "high"
	case TrustHigh:
		return "low"
	default:
		return "medium"
	}
}
