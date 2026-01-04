package logbook

import (
	"fmt"
	"smartdisplay-core/internal/logger"
	"time"
)

// EntryCategory represents the category of a logbook entry
type EntryCategory string

const (
	CategoryAlarm  EntryCategory = "alarm"
	CategoryGuest  EntryCategory = "guest"
	CategorySystem EntryCategory = "system"
	CategorySafety EntryCategory = "safety"
)

// EntryType represents the specific type of event
type EntryType string

const (
	// Alarm events
	AlarmTriggered          EntryType = "alarm_triggered"
	AlarmArmed              EntryType = "alarm_armed"
	AlarmDisarmed           EntryType = "alarm_disarmed"
	AlarmCountdownStarted   EntryType = "alarm_countdown_started"
	AlarmCountdownCancelled EntryType = "alarm_countdown_cancelled"
	AlarmAcknowledged       EntryType = "alarm_acknowledged"

	// Guest events
	GuestRequested   EntryType = "guest_requested"
	GuestApproved    EntryType = "guest_approved"
	GuestDenied      EntryType = "guest_denied"
	GuestExpired     EntryType = "guest_expired"
	GuestExited      EntryType = "guest_exited"
	GuestAutoExpired EntryType = "guest_auto_expired"

	// System events
	SystemStarted   EntryType = "system_started"
	HAConnected     EntryType = "ha_connected"
	HADisconnected  EntryType = "ha_disconnected"
	DeviceOffline   EntryType = "device_offline"
	DeviceOnline    EntryType = "device_online"
	BatteryLow      EntryType = "battery_low"
	UpdateAvailable EntryType = "update_available"
	SystemUpdated   EntryType = "system_updated"

	// Safety/Failsafe events
	FailsafeActivated   EntryType = "failsafe_activated"
	FailsafeRecovering  EntryType = "failsafe_recovering"
	FailsafeRecovered   EntryType = "failsafe_recovered"
	AlarmDuringFailsafe EntryType = "alarm_during_failsafe"
)

// Severity represents the severity level of an entry
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// UserRole represents the user role (for filtering)
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

// EntryDetail holds optional context details for an entry
type EntryDetail struct {
	Reason           string                 `json:"reason,omitempty"`
	Location         string                 `json:"location,omitempty"`
	UserID           string                 `json:"user_id,omitempty"`
	GuestID          string                 `json:"guest_id,omitempty"`
	DurationSeconds  int                    `json:"duration_seconds,omitempty"`
	DurationMinutes  int                    `json:"duration_minutes,omitempty"`
	Count            int                    `json:"count,omitempty"`
	DeviceName       string                 `json:"device_name,omitempty"`
	DeviceType       string                 `json:"device_type,omitempty"`
	DeviceList       []string               `json:"devices,omitempty"`
	Version          string                 `json:"version,omitempty"`
	OldVersion       string                 `json:"old_version,omitempty"`
	NewVersion       string                 `json:"new_version,omitempty"`
	Percentage       int                    `json:"percentage,omitempty"`
	RequestTimeouts  []string               `json:"request_timeouts,omitempty"`
	EstimatedSeconds int                    `json:"estimated_seconds,omitempty"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
}

// Entry represents a single logbook entry
type Entry struct {
	ID             string        `json:"id"`
	Timestamp      time.Time     `json:"timestamp"`
	TimestampLocal string        `json:"timestamp_local"`
	Category       EntryCategory `json:"category"`
	Type           EntryType     `json:"type"`
	Severity       Severity      `json:"severity"`
	Message        string        `json:"message"`
	Context        string        `json:"context,omitempty"`
	Details        EntryDetail   `json:"details"`
	Grouped        bool          `json:"grouped"`
	GroupCount     int           `json:"group_count"`
	GroupedAt      *time.Time    `json:"grouped_at,omitempty"`
	VisibleToRole  UserRole      `json:"visible_to_role"`
}

// PaginationInfo holds pagination metadata
type PaginationInfo struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// MetadataInfo holds metadata about the logbook
type MetadataInfo struct {
	DateRangeStart    *time.Time     `json:"date_range_start,omitempty"`
	DateRangeEnd      *time.Time     `json:"date_range_end,omitempty"`
	CategoryCounts    map[string]int `json:"category_counts,omitempty"`
	Last24Hours       *SeverityCount `json:"last_24_hours,omitempty"`
	HasCriticalEvents bool           `json:"has_critical_events,omitempty"`
}

// SeverityCount holds counts by severity
type SeverityCount struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
}

// LogbookResponse represents the response for /api/ui/logbook
type LogbookResponse struct {
	UserID     string         `json:"user_id"`
	Role       UserRole       `json:"role"`
	Entries    []Entry        `json:"entries"`
	Pagination PaginationInfo `json:"pagination"`
	Metadata   MetadataInfo   `json:"metadata"`
}

// LogbookSummaryResponse represents the response for /api/ui/logbook/summary
type LogbookSummaryResponse struct {
	RecentEntries []SummaryEntry `json:"recent_entries"`
	Metadata      MetadataInfo   `json:"metadata"`
}

// SummaryEntry represents a simplified entry for summary view
type SummaryEntry struct {
	TimestampLocal string        `json:"timestamp_local"`
	Message        string        `json:"message"`
	Severity       Severity      `json:"severity"`
	Category       EntryCategory `json:"category"`
}

// LogbookManager manages logbook entries
type LogbookManager struct {
	entries             []Entry
	retentionDays       int // for normal entries
	retentionSafetyDays int // for safety events
	entryIDCounter      int64
	lastGroupCheck      time.Time
	userRole            UserRole
}

// NewLogbookManager creates a new LogbookManager
func NewLogbookManager(retentionDays, retentionSafetyDays int) *LogbookManager {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	if retentionSafetyDays <= 0 {
		retentionSafetyDays = 90
	}
	mgr := &LogbookManager{
		entries:             make([]Entry, 0),
		retentionDays:       retentionDays,
		retentionSafetyDays: retentionSafetyDays,
		entryIDCounter:      1000,
		lastGroupCheck:      time.Now(),
		userRole:            RoleAdmin,
	}
	logger.Info(fmt.Sprintf("logbook: initialized (retention: %d days, retention_safety: %d days)",
		retentionDays, retentionSafetyDays))
	return mgr
}

// SetUserRole sets the current user role (for role-based filtering)
func (m *LogbookManager) SetUserRole(role UserRole) {
	m.userRole = role
	logger.Info(fmt.Sprintf("logbook: user role set to %s", role))
}

// AddEntry adds a new entry to the logbook
func (m *LogbookManager) AddEntry(category EntryCategory, entryType EntryType, severity Severity,
	message string, context string, details EntryDetail, visibleToRole UserRole) {
	now := time.Now()
	m.entryIDCounter++
	entry := Entry{
		ID:             fmt.Sprintf("entry_%d", m.entryIDCounter),
		Timestamp:      now,
		TimestampLocal: formatTimestamp(now),
		Category:       category,
		Type:           entryType,
		Severity:       severity,
		Message:        message,
		Context:        context,
		Details:        details,
		Grouped:        false,
		GroupCount:     1,
		GroupedAt:      nil,
		VisibleToRole:  visibleToRole,
	}

	// Add to entries (prepend for newest first)
	m.entries = append([]Entry{entry}, m.entries...)

	// Clean up old entries
	m.cleanupOldEntries()

	logger.Info(fmt.Sprintf("logbook: entry created (category: %s, severity: %s)", category, severity))
}

// GetEntries returns logbook entries with optional filtering
func (m *LogbookManager) GetEntries(userRole UserRole, limit, offset int, categoryFilter string) LogbookResponse {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Filter by role visibility
	filtered := m.filterByRole(userRole)

	// Filter by category if specified
	if categoryFilter != "" {
		var categoryFiltered []Entry
		for _, e := range filtered {
			if string(e.Category) == categoryFilter {
				categoryFiltered = append(categoryFiltered, e)
			}
		}
		filtered = categoryFiltered
	}

	// Calculate pagination
	total := len(filtered)
	hasMore := (offset + limit) < total
	if offset > total {
		offset = total
	}

	var entries []Entry
	end := offset + limit
	if end > total {
		end = total
	}
	if offset < total {
		entries = filtered[offset:end]
	}

	// Build category counts
	categoryCounts := make(map[string]int)
	for _, e := range filtered {
		key := string(e.Category)
		categoryCounts[key]++
	}

	// Calculate severity counts for last 24 hours
	last24h := m.getSeverityCounts(24 * time.Hour)
	hasCritical := last24h.Critical > 0

	now := time.Now()
	retentionStart := now.AddDate(0, 0, -m.retentionDays)

	response := LogbookResponse{
		Entries: entries,
		Pagination: PaginationInfo{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasMore: hasMore,
		},
		Metadata: MetadataInfo{
			DateRangeStart:    &retentionStart,
			DateRangeEnd:      &now,
			CategoryCounts:    categoryCounts,
			Last24Hours:       &last24h,
			HasCriticalEvents: hasCritical,
		},
	}

	logger.Info(fmt.Sprintf("logbook: entries retrieved by role %s (count: %d, filtered)",
		userRole, len(entries)))

	return response
}

// GetSummary returns recent entries for dashboard view
func (m *LogbookManager) GetSummary(userRole UserRole, limit int) LogbookSummaryResponse {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	// Filter by role visibility
	filtered := m.filterByRole(userRole)

	// Get most recent entries
	var recent []Entry
	for i := 0; i < len(filtered) && i < limit; i++ {
		recent = append(recent, filtered[i])
	}

	// Convert to summary entries
	var summaryEntries []SummaryEntry
	for _, e := range recent {
		summaryEntries = append(summaryEntries, SummaryEntry{
			TimestampLocal: e.TimestampLocal,
			Message:        e.Message,
			Severity:       e.Severity,
			Category:       e.Category,
		})
	}

	// Build metadata
	last24h := m.getSeverityCounts(24 * time.Hour)
	hasCritical := last24h.Critical > 0

	response := LogbookSummaryResponse{
		RecentEntries: summaryEntries,
		Metadata: MetadataInfo{
			Last24Hours:       &last24h,
			HasCriticalEvents: hasCritical,
		},
	}

	return response
}

// filterByRole returns entries visible to the given role
func (m *LogbookManager) filterByRole(role UserRole) []Entry {
	var filtered []Entry
	for _, e := range m.entries {
		if canView(role, e.VisibleToRole, e.Category) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// canView determines if a role can view an entry
func canView(userRole UserRole, entryVisibility UserRole, category EntryCategory) bool {
	switch userRole {
	case RoleAdmin:
		return true // Admin sees everything
	case RoleUser:
		// User sees non-safety events
		if category == CategorySafety {
			return false
		}
		// User sees entries marked for user or admin
		return entryVisibility == RoleUser || entryVisibility == RoleAdmin
	case RoleGuest:
		return false // Guest sees nothing
	}
	return false
}

// cleanupOldEntries removes entries older than retention period
func (m *LogbookManager) cleanupOldEntries() {
	now := time.Now()
	var retained []Entry

	for _, e := range m.entries {
		retention := m.retentionDays
		if e.Category == CategorySafety {
			retention = m.retentionSafetyDays
		}

		if now.Sub(e.Timestamp) < time.Duration(retention*24)*time.Hour {
			retained = append(retained, e)
		}
	}

	if len(retained) < len(m.entries) {
		removed := len(m.entries) - len(retained)
		logger.Info(fmt.Sprintf("logbook: old entries archived (count: %d, days: %d)", removed, m.retentionDays))
		m.entries = retained
	}
}

// getSeverityCounts returns severity counts within a time window
func (m *LogbookManager) getSeverityCounts(window time.Duration) SeverityCount {
	cutoff := time.Now().Add(-window)
	count := SeverityCount{}

	for _, e := range m.entries {
		if e.Timestamp.After(cutoff) {
			count.Total++
			switch e.Severity {
			case SeverityCritical:
				count.Critical++
			case SeverityWarning:
				count.Warning++
			case SeverityInfo:
				count.Info++
			}
		}
	}

	return count
}

// formatTimestamp formats a timestamp for user display
func formatTimestamp(t time.Time) string {
	now := time.Now()
	days := int(now.Sub(t).Hours() / 24)

	switch {
	case days == 0:
		return fmt.Sprintf("Today %s", t.Format("3:04 PM"))
	case days == 1:
		return fmt.Sprintf("Yesterday %s", t.Format("3:04 PM"))
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("January 2, 3:04 PM")
	}
}

// ValidateEntry checks if an entry is valid
func ValidateEntry(e Entry) bool {
	if e.ID == "" || e.Message == "" {
		return false
	}
	if e.Category == "" || e.Type == "" || e.Severity == "" {
		return false
	}
	return true
}
