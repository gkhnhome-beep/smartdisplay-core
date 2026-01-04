package voice

import (
	"smartdisplay-core/internal/logger"
)

// Hook represents the voice feedback system.
// It provides a mechanism to log intended speech output without actually playing audio.
// Voice feedback is disabled by default and controlled via runtime config.
type Hook struct {
	enabled bool
}

// New creates a new voice feedback hook with the given enabled state.
func New(enabled bool) *Hook {
	return &Hook{
		enabled: enabled,
	}
}

// Enabled returns whether voice feedback is currently enabled.
func (h *Hook) Enabled() bool {
	return h.enabled
}

// SetEnabled updates the voice feedback enabled state.
func (h *Hook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// Speak logs the intended speech if voice is enabled.
// Text should be brief (single phrase or sentence).
// Priority indicates the urgency: "critical", "warning", or "info".
// No audio is played - this only logs the intent.
func (h *Hook) Speak(text string, priority string) {
	if !h.enabled {
		return
	}

	// Log the voice event
	logger.Info("voice: priority=" + priority + " text=" + text)
}

// SpeakCritical logs critical voice feedback (e.g., alarms, system failures).
// Shorthand for Speak(text, "critical").
func (h *Hook) SpeakCritical(text string) {
	h.Speak(text, "critical")
}

// SpeakWarning logs warning-level voice feedback (e.g., confirmations needed).
// Shorthand for Speak(text, "warning").
func (h *Hook) SpeakWarning(text string) {
	h.Speak(text, "warning")
}

// SpeakInfo logs informational voice feedback.
// Shorthand for Speak(text, "info").
func (h *Hook) SpeakInfo(text string) {
	h.Speak(text, "info")
}
