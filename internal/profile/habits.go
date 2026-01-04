package profile

import (
	"sync"
	"time"
)

// UserProfile stores deterministic user habit data.
type UserProfile struct {
	ArrivalTimes     []time.Time // Confirmed arrival times
	ExitTimes        []time.Time // Confirmed exit times
	AlarmArmCount    int         // Number of times alarm armed
	AlarmDisarmCount int         // Number of times alarm disarmed
	GuestUsageCount  int         // Number of confirmed guest usages
	mu               sync.Mutex  // For concurrent access
}

var profile = &UserProfile{}

// RecordArrival logs a confirmed arrival time.
func RecordArrival(t time.Time) {
	profile.mu.Lock()
	defer profile.mu.Unlock()
	profile.ArrivalTimes = append(profile.ArrivalTimes, t)
}

// RecordExit logs a confirmed exit time.
func RecordExit(t time.Time) {
	profile.mu.Lock()
	defer profile.mu.Unlock()
	profile.ExitTimes = append(profile.ExitTimes, t)
}

// RecordAlarmArm increments the alarm arm count.
func RecordAlarmArm() {
	profile.mu.Lock()
	defer profile.mu.Unlock()
	profile.AlarmArmCount++
}

// RecordAlarmDisarm increments the alarm disarm count.
func RecordAlarmDisarm() {
	profile.mu.Lock()
	defer profile.mu.Unlock()
	profile.AlarmDisarmCount++
}

// RecordGuestUsage increments the guest usage count.
func RecordGuestUsage() {
	profile.mu.Lock()
	defer profile.mu.Unlock()
	profile.GuestUsageCount++
}
