// Package narrative maintains a short narrative memory for continuity (max 7 days).
package narrative

import (
	"time"
)

// Event represents a notable event for narrative memory.
type Event struct {
	Timestamp time.Time
	Type      string
	Detail    string
}

// Memory holds recent events (max 7 days).
type Memory struct {
	Events []Event
}

// AddEvent adds a new event to memory, pruning old ones.
func (m *Memory) AddEvent(eventType, detail string) {
	e := Event{Timestamp: time.Now(), Type: eventType, Detail: detail}
	m.Events = append(m.Events, e)
	m.prune()
}

// GetRecent returns events within the last 7 days.
func (m *Memory) GetRecent() []Event {
	m.prune()
	return m.Events
}

// GetYesterday returns the most recent event from yesterday, if any.
func (m *Memory) GetYesterday() *Event {
	cutoff := time.Now().Add(-24 * time.Hour)
	for i := len(m.Events) - 1; i >= 0; i-- {
		if m.Events[i].Timestamp.Before(cutoff) {
			return &m.Events[i]
		}
	}
	return nil
}

// GetLastOfType returns the most recent event of a given type.
func (m *Memory) GetLastOfType(eventType string) *Event {
	for i := len(m.Events) - 1; i >= 0; i-- {
		if m.Events[i].Type == eventType {
			return &m.Events[i]
		}
	}
	return nil
}

// prune removes events older than 7 days.
func (m *Memory) prune() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var pruned []Event
	for _, e := range m.Events {
		if e.Timestamp.After(cutoff) {
			pruned = append(pruned, e)
		}
	}
	m.Events = pruned
}
