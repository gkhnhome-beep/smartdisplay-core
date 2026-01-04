package coherence

// ToneRegistry defines the expected tone for different message types.
type ToneRegistry struct {
	Expected map[string]string // messageType -> tone
}

// CheckToneConsistency verifies that outputs maintain consistent tone.
func CheckToneConsistency(outputs []OutputContext, registry ToneRegistry) (consistent bool, issues []string) {
	consistent = true
	issues = []string{}

	for _, out := range outputs {
		if expected, ok := registry.Expected[out.MessageType]; ok {
			if out.Tone != expected {
				consistent = false
				issues = append(issues, "tone_mismatch: "+out.MessageType+" should use "+expected+" but uses "+out.Tone)
			}
		}
	}

	return consistent, issues
}
