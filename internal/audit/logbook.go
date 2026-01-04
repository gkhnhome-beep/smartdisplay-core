package audit

import (
	"fmt"
	"strings"
	"time"

	"smartdisplay-core/internal/i18n"
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
		return fmt.Sprintf(i18n.T("audit.perm_check"), e.Detail)
	case "smoke_test":
		return i18n.T("audit.smoke_test")
	case "hardware_event":
		return fmt.Sprintf(i18n.T("audit.hardware_event"), e.Detail)
	case "domain_event":
		return fmt.Sprintf(i18n.T("audit.domain_event"), e.Detail)
	case "alarm":
		return fmt.Sprintf(i18n.T("audit.alarm"), e.Detail)
	case "guest":
		return fmt.Sprintf(i18n.T("audit.guest"), e.Detail)
	case "login":
		return fmt.Sprintf(i18n.T("audit.login"), e.Detail)
	case "config":
		return fmt.Sprintf(i18n.T("audit.config"), e.Detail)
	case "restore":
		return i18n.T("audit.restore")
	case "trust_learn":
		return fmt.Sprintf(i18n.T("audit.trust_learn"), e.Detail)
	default:
		return fmt.Sprintf(i18n.T("audit.default"), e.Action, e.Detail)
	}
}

// For testing/demo: generate a readable timestamp
func prettyTime(ts string) string {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Format("Jan 2, 15:04")
	}
	return ts
}
