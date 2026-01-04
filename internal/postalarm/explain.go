package postalarm

// ExplainWhatHappened returns a clear, non-blaming explanation of what triggered the alarm.
func ExplainWhatHappened(reason string) string {
	switch reason {
	case "door_open":
		return "The system detected a door opening. This is a normal trigger."
	case "motion_detected":
		return "The system detected motion. This is a normal trigger."
	case "window_open":
		return "The system detected a window opening. This is a normal trigger."
	case "tampering":
		return "The system detected unusual activity. Everything is secure now."
	default:
		return "The alarm was triggered as a precaution. Everything is under control."
	}
}

// ExplainNormalNow returns a reassurance message about the current state.
func ExplainNormalNow() string {
	return "The system is now stable. All sensors are monitored. No further action needed."
}

// PostAlarmMessage combines both explanations into a reassuring message.
func PostAlarmMessage(reason string) string {
	return ExplainWhatHappened(reason) + " " + ExplainNormalNow()
}

// ReducedNoiseGuidance returns guidance for reduced notifications during cooldown.
func ReducedNoiseGuidance() string {
	return "Notifications are minimized during this period. You can dismiss this at any time."
}
