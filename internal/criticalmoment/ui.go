package criticalmoment

// UI helpers for exposing critical moment decisions and explanations.

type UIDecision struct {
	ShowConfirmation bool
	Explanation      string
	Moment           string
}

// GetUIDecision prepares UI data for a given moment and context.
func GetUIDecision(moment MomentType, context map[string]interface{}, lastConfirmed map[MomentType]int64, now int64) UIDecision {
	decision := ExplainCriticalMoment(moment, context)
	show := ShouldPromptConfirmation(moment, context, lastConfirmed, now)
	return UIDecision{
		ShowConfirmation: show,
		Explanation:      decision.Explanation,
		Moment:           moment.String(),
	}
}
