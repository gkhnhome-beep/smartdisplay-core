package trust

// TrustScoreLevel is used to adjust tone, explanation depth, and alert urgency.
type TrustScoreLevel int

const (
	TrustLow TrustScoreLevel = iota
	TrustMedium
	TrustHigh
)

// GetTrustScoreLevel returns a qualitative level for tone/explanation/urgency.
func GetTrustScoreLevel() TrustScoreLevel {
	score := getTrustScore()
	switch {
	case score < 40:
		return TrustLow
	case score > 70:
		return TrustHigh
	default:
		return TrustMedium
	}
}
