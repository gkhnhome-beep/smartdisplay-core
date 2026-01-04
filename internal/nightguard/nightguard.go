// Package nightguard manages special system behavior during sleep hours (Night Mode).
package nightguard

import "time"

// NightModeActive returns true if quiet hours are active and the alarm is armed.
func NightModeActive(quietHours bool, alarmArmed bool) bool {
	return quietHours && alarmArmed
}

// NightGuardConfig holds settings for night behavior.
type NightGuardConfig struct {
	MinimalNotifications    bool
	ExtraCautionBeforeSiren bool
}

// GetNightGuardConfig returns config for night mode.
func GetNightGuardConfig(nightMode bool) NightGuardConfig {
	if nightMode {
		return NightGuardConfig{
			MinimalNotifications:    true,
			ExtraCautionBeforeSiren: true,
		}
	}
	return NightGuardConfig{}
}

// MorningSummary returns a clear summary of night events for the user.
func MorningSummary(events []string) string {
	if len(events) == 0 {
		return "All was quiet last night."
	}
	return "Night summary: " + joinEvents(events)
}

func joinEvents(events []string) string {
	return "- " + time.Now().Format("15:04") + ": " + events[0] // Simple, can be expanded
}
