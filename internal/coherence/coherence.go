// Package coherence ensures the system speaks with ONE consistent voice.
package coherence

import (
	"log"
	"strings"
)

// OutputContext holds metadata about an AI output.
type OutputContext struct {
	MessageType string   // e.g., "alarm_advice", "morning_briefing", "guest_welcome"
	Tone        string   // e.g., "calm", "urgent", "welcoming"
	Terms       []string // Key terms used
	Advice      string   // Any advice given
}

// CoherenceCheck validates that multiple outputs are tonally and terminologically consistent.
type CoherenceCheck struct {
	RecentOutputs []OutputContext
	ToneRegistry  map[string]string // messageType -> expected tone
	TermRegistry  map[string]string // term variations -> canonical term
}

// NewCoherenceCheck creates a new coherence checker.
func NewCoherenceCheck() *CoherenceCheck {
	return &CoherenceCheck{
		RecentOutputs: []OutputContext{},
		ToneRegistry: map[string]string{
			"alarm_advice":     "calm",
			"morning_briefing": "friendly",
			"guest_welcome":    "welcoming",
			"night_guard":      "protective",
		},
		TermRegistry: map[string]string{
			"disarm":   "disarm",
			"turn off": "disarm",
			"arm":      "arm",
			"activate": "arm",
		},
	}
}

// CheckOutput adds an output to the coherence check and validates it.
func (c *CoherenceCheck) CheckOutput(ctx OutputContext) (coherent bool, warnings []string) {
	coherent = true
	warnings = []string{}

	// Check tone consistency
	if expectedTone, ok := c.ToneRegistry[ctx.MessageType]; ok {
		if ctx.Tone != expectedTone {
			coherent = false
			warnings = append(warnings, "tone_mismatch: expected "+expectedTone+" but got "+ctx.Tone)
		}
	}

	// Check for conflicting advice
	for _, prev := range c.RecentOutputs {
		if conflictDetected(prev.Advice, ctx.Advice) {
			coherent = false
			warnings = append(warnings, "conflicting_advice: previous output contradicts this one")
		}
	}

	// Keep recent outputs (max 5)
	c.RecentOutputs = append(c.RecentOutputs, ctx)
	if len(c.RecentOutputs) > 5 {
		c.RecentOutputs = c.RecentOutputs[1:]
	}

	return coherent, warnings
}

// conflictDetected checks if two pieces of advice contradict each other.
func conflictDetected(prev, curr string) bool {
	if prev == "" || curr == "" {
		return false
	}
	// Simple heuristic: detect opposite verbs
	if strings.Contains(prev, "arm") && strings.Contains(curr, "disarm") {
		return true
	}
	if strings.Contains(prev, "disarm") && strings.Contains(curr, "arm") {
		return true
	}
	return false
}

// LogCoherenceWarning logs a coherence issue.
func LogCoherenceWarning(messageType string, warnings []string) {
	log.Printf("[Coherence] Warning for %s: %v", messageType, warnings)
}
