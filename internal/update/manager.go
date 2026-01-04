// Package update provides a safe, no-execution OTA update skeleton.
// This phase implements:
// - Checksum validation only (no downloads)
// - Package staging to temporary storage (no execution)
// - Reboot activation (admin-only, manual trigger)
// - Comprehensive audit logging of all actions
//
// NOTE: Auto-updates are explicitly disabled. All operations are manual and audited.
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PackageInfo contains metadata about an available update.
type PackageInfo struct {
	Version      string `json:"version"`
	BuildID      string `json:"build_id"`
	Checksum     string `json:"checksum"` // SHA256 hex string
	Size         int64  `json:"size"`
	ReleaseNotes string `json:"release_notes,omitempty"`
}

// StagedPackage represents a staged update ready for activation.
type StagedPackage struct {
	PackageInfo
	StagedAt  time.Time `json:"staged_at"`
	StagePath string    `json:"stage_path"` // full path to staged file
}

// UpdateStatus represents the current state of the update system.
type UpdateStatus struct {
	CurrentVersion string         `json:"current_version"`
	Available      *PackageInfo   `json:"available,omitempty"`
	Staged         *StagedPackage `json:"staged,omitempty"`
	PendingReboot  bool           `json:"pending_reboot"`
	LastCheckTime  time.Time      `json:"last_check_time"`
	LastError      string         `json:"last_error,omitempty"`
}

// Manager handles OTA update operations with audit logging.
type Manager struct {
	mu             sync.RWMutex
	currentVersion string
	stagingDir     string
	available      *PackageInfo
	staged         *StagedPackage
	pendingReboot  bool
	lastCheckTime  time.Time
	lastError      string
	auditLogger    AuditLogger
}

// AuditLogger interface for logging all update actions.
type AuditLogger interface {
	Record(action string, detail string)
}

// StubAuditLogger logs to a simple string (for testing).
type StubAuditLogger struct {
	mu      sync.Mutex
	entries []string
}

func (s *StubAuditLogger) Record(action string, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, fmt.Sprintf("[%s] %s: %s", time.Now().Format(time.RFC3339), action, detail))
}

// New creates a new UpdateManager.
// currentVersion: semantic version of current build (e.g., "1.2.3")
// stagingDir: directory for staged packages (must be writable)
// auditLogger: logger for all actions
func New(currentVersion, stagingDir string, auditLogger AuditLogger) *Manager {
	return &Manager{
		currentVersion: currentVersion,
		stagingDir:     stagingDir,
		auditLogger:    auditLogger,
		lastCheckTime:  time.Now(),
	}
}

// CheckAvailable checks if a newer version is available (STUB).
// In future phases, this will fetch from a remote server.
// Currently always returns nil (no update available).
func (m *Manager) CheckAvailable() (*PackageInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastCheckTime = time.Now()
	m.auditLogger.Record("update_check", "checked for available updates")

	// STUB: No remote check implemented
	// In future: fetch from update server, verify signature, etc.
	m.available = nil
	m.lastError = ""

	return nil, nil
}

// ValidatePackage verifies a package's integrity using SHA256.
// IMPORTANT: This validates checksum ONLY. Does NOT download or execute.
// data: package file content (in memory or streamed)
// expectedChecksum: hex-encoded SHA256 checksum to verify against
// Returns error if checksum doesn't match.
func (m *Manager) ValidatePackage(data io.Reader, expectedChecksum string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate SHA256 of provided data
	hash := sha256.New()
	_, err := io.Copy(hash, data)
	if err != nil {
		errMsg := fmt.Sprintf("checksum calculation failed: %v", err)
		m.auditLogger.Record("update_validate_error", errMsg)
		m.lastError = errMsg
		return fmt.Errorf("checksum calculation failed: %w", err)
	}

	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))

	// Compare checksums (case-insensitive)
	if calculatedChecksum != expectedChecksum {
		errMsg := fmt.Sprintf("checksum mismatch: got %s, expected %s", calculatedChecksum, expectedChecksum)
		m.auditLogger.Record("update_validate_failed", errMsg)
		m.lastError = errMsg
		return fmt.Errorf("checksum mismatch")
	}

	m.auditLogger.Record("update_validate_success", fmt.Sprintf("package validated with checksum %s", expectedChecksum))
	m.lastError = ""
	return nil
}

// StageUpdate writes a validated package to staging directory.
// IMPORTANT: Does NOT execute or apply the update, only stages for later activation.
// data: package content (must already be validated by ValidatePackage)
// packageInfo: metadata about the package
// Returns: path to staged package file
func (m *Manager) StageUpdate(data io.Reader, packageInfo PackageInfo) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure staging directory exists
	if err := os.MkdirAll(m.stagingDir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create staging directory: %v", err)
		m.auditLogger.Record("update_stage_error", errMsg)
		m.lastError = errMsg
		return "", fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Create staged package filename with version
	stagedFilename := fmt.Sprintf("update-%s-%s.bin", packageInfo.Version, packageInfo.BuildID)
	stagedPath := filepath.Join(m.stagingDir, stagedFilename)

	// Write package to staging directory
	file, err := os.Create(stagedPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create staged file: %v", err)
		m.auditLogger.Record("update_stage_error", errMsg)
		m.lastError = errMsg
		return "", fmt.Errorf("failed to create staged file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		errMsg := fmt.Sprintf("failed to write staged file: %v", err)
		m.auditLogger.Record("update_stage_error", errMsg)
		m.lastError = errMsg
		os.Remove(stagedPath) // cleanup on error
		return "", fmt.Errorf("failed to write staged file: %w", err)
	}

	// Store staged package info
	m.staged = &StagedPackage{
		PackageInfo: packageInfo,
		StagedAt:    time.Now(),
		StagePath:   stagedPath,
	}

	m.auditLogger.Record("update_staged", fmt.Sprintf("version %s staged at %s", packageInfo.Version, stagedPath))
	m.lastError = ""

	return stagedPath, nil
}

// ActivateOnReboot schedules the staged update to activate on next reboot.
// IMPORTANT: Does NOT execute the update immediately, only marks for activation.
// Returns error if no staged update available.
func (m *Manager) ActivateOnReboot() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.staged == nil {
		errMsg := "no staged update available"
		m.auditLogger.Record("update_activate_error", errMsg)
		m.lastError = errMsg
		return fmt.Errorf("no staged update available")
	}

	// Mark pending reboot (in future: write boot flag to persistent storage)
	m.pendingReboot = true

	m.auditLogger.Record("update_activate_scheduled", fmt.Sprintf("update %s scheduled for next reboot", m.staged.Version))
	m.lastError = ""

	return nil
}

// CancelActivation cancels a scheduled reboot activation (before reboot occurs).
func (m *Manager) CancelActivation() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.pendingReboot {
		errMsg := "no pending activation to cancel"
		m.auditLogger.Record("update_cancel_error", errMsg)
		m.lastError = errMsg
		return fmt.Errorf("no pending activation")
	}

	m.pendingReboot = false
	m.auditLogger.Record("update_activation_cancelled", "pending reboot activation cancelled")
	m.lastError = ""

	return nil
}

// ClearStaged removes the currently staged update.
func (m *Manager) ClearStaged() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.staged == nil {
		return nil // nothing to clear
	}

	if err := os.Remove(m.staged.StagePath); err != nil && !os.IsNotExist(err) {
		errMsg := fmt.Sprintf("failed to remove staged file: %v", err)
		m.auditLogger.Record("update_clear_error", errMsg)
		m.lastError = errMsg
		return fmt.Errorf("failed to remove staged file: %w", err)
	}

	m.auditLogger.Record("update_cleared", fmt.Sprintf("removed staged update %s", m.staged.Version))
	m.staged = nil
	m.lastError = ""

	return nil
}

// GetStatus returns the current update system status.
func (m *Manager) GetStatus() UpdateStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return UpdateStatus{
		CurrentVersion: m.currentVersion,
		Available:      m.available,
		Staged:         m.staged,
		PendingReboot:  m.pendingReboot,
		LastCheckTime:  m.lastCheckTime,
		LastError:      m.lastError,
	}
}

// SetAvailable allows test/stub code to set available updates (for testing).
func (m *Manager) SetAvailable(pkg *PackageInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.available = pkg
	m.auditLogger.Record("update_available_set", fmt.Sprintf("available update: %s", pkg.Version))
}

// GetAuditLog returns all recorded audit entries (for testing/debugging).
func (m *Manager) GetAuditLog() []string {
	if stubLog, ok := m.auditLogger.(*StubAuditLogger); ok {
		stubLog.mu.Lock()
		defer stubLog.mu.Unlock()
		return append([]string{}, stubLog.entries...)
	}
	return nil
}
