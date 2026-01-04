// Package criticalmoment defines and detects moments requiring extra user confirmation to prevent costly mistakes.
package criticalmoment

// MomentType enumerates types of critical moments.
type MomentType int

const (
	MomentNone MomentType = iota
	LeavingWithAlarmDisarmed
	ArmingWhileGuestPresent
	DisarmingDuringQuietHours
)

func (m MomentType) String() string {
	switch m {
	case LeavingWithAlarmDisarmed:
		return "Leaving with alarm disarmed"
	case ArmingWhileGuestPresent:
		return "Arming while guest present"
	case DisarmingDuringQuietHours:
		return "Disarming during quiet hours"
	default:
		return "None"
	}
}
