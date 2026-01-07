// Package settings provides secure credential handling for Home Assistant integration.
// hahealth.go: HA runtime health monitoring and failure detection (FAZ S6)
package settings

import (
	"fmt"
	"smartdisplay-core/internal/logger"
	"time"
)

// FailureThreshold is the number of consecutive failures before marking HA as unreachable.
// FAZ S6: Conservative threshold to avoid flapping.
const FailureThreshold = 3

// RuntimeHealthMonitor tracks HA runtime availability and failure state.
// FAZ S6: Separate from ha_connected (test-based), focuses on operational health.
type RuntimeHealthMonitor struct {
	consecutiveFailures int
	isUnreachable       bool
	lastSeenAt          *time.Time
}

// NewRuntimeHealthMonitor creates a new health monitor.
// FAZ S6: Call once during startup, use same instance throughout lifetime.
func NewRuntimeHealthMonitor() *RuntimeHealthMonitor {
	return &RuntimeHealthMonitor{
		consecutiveFailures: 0,
		isUnreachable:       false,
		lastSeenAt:          nil,
	}
}

// RecordSuccess marks a successful HA read and resets failure counter.
// FAZ S6: Call after every successful HA API read (including reads from adapters).
func (m *RuntimeHealthMonitor) RecordSuccess() {
	wasUnreachable := m.isUnreachable

	now := time.Now()
	m.lastSeenAt = &now
	m.consecutiveFailures = 0
	m.isUnreachable = false

	// Log transition if recovered from unreachable state
	if wasUnreachable {
		logger.Info("ha runtime recovered")
	}
}

// RecordFailure increments failure counter and marks HA unreachable if threshold exceeded.
// FAZ S6: Call on any HA API failure (timeout, network error, 5xx, auth failure).
// Returns true if HA just transitioned to unreachable state (for logging).
func (m *RuntimeHealthMonitor) RecordFailure() bool {
	m.consecutiveFailures++

	wasReachable := !m.isUnreachable
	if m.consecutiveFailures >= FailureThreshold {
		m.isUnreachable = true
	}

	// Log transition if just became unreachable
	if wasReachable && m.isUnreachable {
		logger.Error("ha runtime unreachable after " + fmt.Sprintf("%d", m.consecutiveFailures) + " consecutive failures")
		return true
	}

	return false
}

// IsUnreachable returns the current unreachable state.
// FAZ S6: Safe to call frequently, no side effects.
func (m *RuntimeHealthMonitor) IsUnreachable() bool {
	return m.isUnreachable
}

// GetLastSeenAt returns the timestamp of the last successful HA read.
// FAZ S6: Returns nil if no successful read yet.
func (m *RuntimeHealthMonitor) GetLastSeenAt() *time.Time {
	if m.lastSeenAt == nil {
		return nil
	}
	copy := *m.lastSeenAt
	return &copy
}

// GetConsecutiveFailures returns the current failure counter.
// FAZ S6: Exposed for testing only, not used in normal operation.
func (m *RuntimeHealthMonitor) GetConsecutiveFailures() int {
	return m.consecutiveFailures
}

// UpdateRuntimeConfig updates RuntimeConfig with current health state.
// FAZ S6: Call periodically (e.g., with status endpoint) to persist state.
func (m *RuntimeHealthMonitor) UpdateRuntimeConfig(cfg interface{}) {
	// Type assert carefully since we can't import config package (circular import)
	// Caller must handle this or we accept interface{}
	// For now, this is a helper - actual update happens in handlers
}

// Global health monitor instance (created at server startup).
// FAZ S6: Shared across all HA operations.
var globalHealthMonitor *RuntimeHealthMonitor

// InitGlobalHealthMonitor initializes the global health monitor.
// FAZ S6: Called once from api.Server during initialization.
func InitGlobalHealthMonitor() {
	globalHealthMonitor = NewRuntimeHealthMonitor()
	logger.Info("ha health monitor initialized")
}

// GetGlobalHealthMonitor returns the global health monitor instance.
// FAZ S6: Safe to call after InitGlobalHealthMonitor.
func GetGlobalHealthMonitor() *RuntimeHealthMonitor {
	if globalHealthMonitor == nil {
		globalHealthMonitor = NewRuntimeHealthMonitor()
	}
	return globalHealthMonitor
}
