package criticalmoment

// MomentDecision holds the result of a critical moment check.
type MomentDecision struct {
	Moment              MomentType // The type of critical moment detected
	RequireConfirmation bool       // Should extra confirmation be required?
	Explanation         string     // Human-friendly explanation for the user
	Risky               bool       // Is this context truly risky?
}

// ExplainCriticalMoment returns the decision and explanation for a given moment and context.
func ExplainCriticalMoment(moment MomentType, context map[string]interface{}) MomentDecision {
	switch moment {
	case LeavingWithAlarmDisarmed:
		if isLeaving(context) && !isAlarmArmed(context) {
			return MomentDecision{
				Moment:              moment,
				RequireConfirmation: true,
				Explanation:         "You are leaving while the alarm is disarmed. This increases the risk of unauthorized entry.",
				Risky:               true,
			}
		}
	case ArmingWhileGuestPresent:
		if isGuestPresent(context) && isArming(context) {
			return MomentDecision{
				Moment:              moment,
				RequireConfirmation: true,
				Explanation:         "A guest is present while you are arming the system. This may inconvenience or lock out your guest.",
				Risky:               true,
			}
		}
	case DisarmingDuringQuietHours:
		if isDisarming(context) && isQuietHours(context) {
			return MomentDecision{
				Moment:              moment,
				RequireConfirmation: true,
				Explanation:         "Disarming during quiet hours may be unexpected and could indicate a security risk.",
				Risky:               true,
			}
		}
	}
	return MomentDecision{Moment: MomentNone, RequireConfirmation: false, Explanation: "", Risky: false}
}

// The following helpers are stubs to be implemented with real context logic.
func isLeaving(context map[string]interface{}) bool      { return context["leaving"] == true }
func isAlarmArmed(context map[string]interface{}) bool   { return context["alarmArmed"] == true }
func isGuestPresent(context map[string]interface{}) bool { return context["guestPresent"] == true }
func isArming(context map[string]interface{}) bool       { return context["arming"] == true }
func isDisarming(context map[string]interface{}) bool    { return context["disarming"] == true }
func isQuietHours(context map[string]interface{}) bool   { return context["quietHours"] == true }
