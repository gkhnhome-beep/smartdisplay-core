// Package reinforcement provides subtle positive feedback for successful user-system interactions.
package reinforcement

import (
	"time"
)

// SuccessType enumerates types of success moments.
type SuccessType int

const (
	SuccessNone SuccessType = iota
	SuccessWeekNoAlarmIssues
	SuccessSmoothGuestVisit
	SuccessCleanExit
)

// SuccessMoment holds info about a detected success.
type SuccessMoment struct {
	Type      SuccessType
	Timestamp time.Time
}

// DetectSuccessMoments checks for recent success moments (stub: implement with real checks).
func DetectSuccessMoments() []SuccessMoment {
	// TODO: Integrate with alarm, guest, and exit systems.
	return nil
}
