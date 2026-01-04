// Package morning provides a single useful morning message (briefing) for the user.
package morning

import (
	"fmt"
	"time"
)

// MorningBriefing holds the content of the morning message.
type MorningBriefing struct {
	AlarmStatus  string   `json:"alarm_status"`
	NightEvents  []string `json:"night_events"`
	TodayContext string   `json:"today_context"`
	GeneratedAt  string   `json:"generated_at"`
}

// GenerateBriefing creates a morning briefing from system state.
func GenerateBriefing(alarmStatus string, nightEvents []string, todayContext string) MorningBriefing {
	return MorningBriefing{
		AlarmStatus:  alarmStatus,
		NightEvents:  nightEvents,
		TodayContext: todayContext,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}
}

// FormatBriefing returns a single useful message for the user.
func FormatBriefing(b MorningBriefing) string {
	msg := fmt.Sprintf("Good morning! Alarm: %s.", b.AlarmStatus)
	if len(b.NightEvents) > 0 {
		msg += " Night: " + b.NightEvents[0] // Simple, can be expanded
	}
	if b.TodayContext != "" {
		msg += " Today: " + b.TodayContext
	}
	return msg
}
