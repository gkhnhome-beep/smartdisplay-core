// Package postalarm manages graceful handling of post-alarm moments.
package postalarm

import (
	"time"
)

// PostAlarmState tracks the state after an alarm has been triggered.
type PostAlarmState struct {
	TriggeredAt   time.Time
	CooldownUntil time.Time
	Reason        string
	Acknowledged  bool
}

// NewPostAlarmState creates a new post-alarm state with a default cooldown of 5 minutes.
func NewPostAlarmState(reason string, cooldownDuration time.Duration) *PostAlarmState {
	now := time.Now()
	return &PostAlarmState{
		TriggeredAt:   now,
		CooldownUntil: now.Add(cooldownDuration),
		Reason:        reason,
		Acknowledged:  false,
	}
}

// IsInCooldown returns true if the cooldown period is still active.
func (p *PostAlarmState) IsInCooldown() bool {
	return time.Now().Before(p.CooldownUntil)
}

// Acknowledge marks the post-alarm state as acknowledged by the user.
func (p *PostAlarmState) Acknowledge() {
	p.Acknowledged = true
}

// TimeSinceTrigger returns the duration since the alarm was triggered.
func (p *PostAlarmState) TimeSinceTrigger() time.Duration {
	return time.Since(p.TriggeredAt)
}
