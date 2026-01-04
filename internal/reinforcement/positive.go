package reinforcement

// PositiveLine returns a single positive line for a given success type.
func PositiveLine(t SuccessType) string {
	switch t {
	case SuccessWeekNoAlarmIssues:
		return "Great job! A full week of secure peace."
	case SuccessSmoothGuestVisit:
		return "Guest visit went smoothly—well done."
	case SuccessCleanExit:
		return "Exit was clean and secure. Nice attention to detail."
	default:
		return ""
	}
}
