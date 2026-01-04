package criticalmoment

// ShouldPromptConfirmation returns true only if the context is truly risky and not recently confirmed.
// Use a cache or timestamp to avoid confirmation spam (implementation stub).
func ShouldPromptConfirmation(moment MomentType, context map[string]interface{}, lastConfirmed map[MomentType]int64, now int64) bool {
	decision := ExplainCriticalMoment(moment, context)
	if !decision.Risky || !decision.RequireConfirmation {
		return false
	}
	const minInterval = 300 // seconds (5 minutes) between prompts for same moment
	if last, ok := lastConfirmed[moment]; ok && now-last < minInterval {
		return false
	}
	return true
}
