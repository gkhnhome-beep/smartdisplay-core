// Package awaymode manages system behavior when the owner is away.
package awaymode

import "time"

// OwnerAway returns true if no local interaction and alarm has been armed for a long duration.
func OwnerAway(lastInteraction time.Time, alarmArmed bool, armedSince time.Time, awayThreshold time.Duration) bool {
	if !alarmArmed {
		return false
	}
	if time.Since(lastInteraction) > awayThreshold && time.Since(armedSince) > awayThreshold {
		return true
	}
	return false
}

// ReassuringTone returns a reassuring tone string.
func ReassuringTone() string {
	return "reassuring"
}

// ProactiveSummary returns a proactive summary message for the owner.
func ProactiveSummary(status string, events []string) string {
	msg := "All is well at home. " + status
	if len(events) > 0 {
		msg += " Recent events: " + events[0] // Simple, can be expanded
	}
	return msg
}
