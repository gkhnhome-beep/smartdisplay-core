package audit

import (
	"fmt"
	"strings"
	"time"
)

type TimelineEntry struct {
	Timestamp string `json:"timestamp"`
	Sentence  string `json:"sentence"`
}

// ToTimeline transforms audit entries into human-readable timeline entries.
func ToTimeline(entries []Entry) []TimelineEntry {
	var timeline []TimelineEntry
	for _, e := range entries {
		ts := e.Timestamp
		sentence := humanize(e)
		timeline = append(timeline, TimelineEntry{
			Timestamp: ts,
			Sentence:  sentence,
		})
	}
	return timeline
}

// humanize converts an audit Entry to a readable sentence.
func humanize(e Entry) string {
	switch strings.ToLower(e.Action) {
	case "perm_check":
		return fmt.Sprintf("Permission check: %s", e.Detail)
	case "smoke_test":
		return "Admin ran a system smoke test."
	case "hardware_event":
		return fmt.Sprintf("Hardware event: %s", e.Detail)
	case "domain_event":
		return fmt.Sprintf("Domain event: %s", e.Detail)
	case "alarm":
		return fmt.Sprintf("Alarm event: %s", e.Detail)
	case "guest":
		return fmt.Sprintf("Guest event: %s", e.Detail)
	case "login":
		return fmt.Sprintf("Login: %s", e.Detail)
	case "config":
		return fmt.Sprintf("Config change: %s", e.Detail)
	case "restore":
		return "Configuration was restored from backup."
	case "trust_learn":
		return fmt.Sprintf("User trust learning: %s", e.Detail)
	default:
		return fmt.Sprintf("%s: %s", e.Action, e.Detail)
	}
}

// For testing/demo: generate a readable timestamp
func prettyTime(ts string) string {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Format("Jan 2, 15:04")
	}
	return ts
}
