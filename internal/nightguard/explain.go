package nightguard

// ExplainNightDecision returns a morning explanation for a night mode decision.
func ExplainNightDecision(decision string) string {
	switch decision {
	case "minimal_notifications":
		return "Notifications were minimized overnight to avoid disturbance."
	case "extra_caution_siren":
		return "Extra caution was used before sounding the siren during quiet hours."
	case "no_override":
		return "No actions were taken without your confirmation during the night."
	default:
		return "Night mode was active. All decisions prioritized your rest and security."
	}
}
