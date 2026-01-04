package profile

import (
	"fmt"
	"time"
)

// ProfileSummary is a human-readable summary of user habits.
type ProfileSummary struct {
	TypicalArrival   string
	TypicalExit      string
	AlarmArmCount    int
	AlarmDisarmCount int
	GuestUsageCount  int
}

// GetUserProfileSummary returns a summary of the user's habits.
func GetUserProfileSummary() ProfileSummary {
	profile.mu.Lock()
	defer profile.mu.Unlock()

	return ProfileSummary{
		TypicalArrival:   summarizeTimes(profile.ArrivalTimes),
		TypicalExit:      summarizeTimes(profile.ExitTimes),
		AlarmArmCount:    profile.AlarmArmCount,
		AlarmDisarmCount: profile.AlarmDisarmCount,
		GuestUsageCount:  profile.GuestUsageCount,
	}
}

// summarizeTimes returns a string summary (e.g., average time) for a slice of times.
func summarizeTimes(times []time.Time) string {
	if len(times) == 0 {
		return "N/A"
	}
	var totalMinutes int
	for _, t := range times {
		totalMinutes += t.Hour()*60 + t.Minute()
	}
	avgMinutes := totalMinutes / len(times)
	hour := avgMinutes / 60
	minute := avgMinutes % 60
	return fmt.Sprintf("%02d:%02d", hour, minute)
}
