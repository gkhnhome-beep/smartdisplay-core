package settings

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// UserRole defines access control levels for settings
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

// SettingsSection groups related settings
type SettingsSection string

const (
	SectionGeneral  SettingsSection = "general"
	SectionSecurity SettingsSection = "security"
	SectionSystem   SettingsSection = "system"
	SectionAdvanced SettingsSection = "advanced"
)

// FieldType indicates the type of a settings field
type FieldType string

const (
	TypeString   FieldType = "string"
	TypeInteger  FieldType = "integer"
	TypeBoolean  FieldType = "boolean"
	TypeReadOnly FieldType = "readonly"
)

// ActionType indicates the type of a settings action
type ActionType string

const (
	ActionFieldChange   ActionType = "field_change"
	ActionBackupCreate  ActionType = "backup_create"
	ActionBackupRestore ActionType = "backup_restore"
	ActionRestart       ActionType = "restart_now"
	ActionFactoryReset  ActionType = "factory_reset"
)

// ConfirmationType indicates the strength of confirmation required
type ConfirmationType string

const (
	ConfirmTypeNone   ConfirmationType = "none"
	ConfirmTypeSimple ConfirmationType = "simple"
	ConfirmTypeStrong ConfirmationType = "strong"
	ConfirmTypeDouble ConfirmationType = "double"
)

// SettingsField represents a single setting
type SettingsField struct {
	ID             string          `json:"id"`
	Section        SettingsSection `json:"section"`
	Type           FieldType       `json:"type"`
	Value          interface{}     `json:"value"`
	DefaultValue   interface{}     `json:"default_value,omitempty"`
	Help           string          `json:"help"`
	Warning        string          `json:"warning,omitempty"`
	RequireConfirm bool            `json:"require_confirm"`
	Options        []string        `json:"options,omitempty"`
	MinValue       *int            `json:"min_value,omitempty"`
	MaxValue       *int            `json:"max_value,omitempty"`
}

// SettingsAction represents a dangerous action
type SettingsAction struct {
	ID               string           `json:"id"`
	Section          SettingsSection  `json:"section"`
	Name             string           `json:"name"`
	Help             string           `json:"help"`
	Danger           string           `json:"danger,omitempty"`
	ConfirmType      ConfirmationType `json:"confirm_type"`
	ConfirmText      string           `json:"confirm_text,omitempty"`
	CountdownSeconds int              `json:"countdown_seconds"`
	RequiresRestart  bool             `json:"requires_restart"`
}

// FieldChangeRequest represents a field change request
type FieldChangeRequest struct {
	Action        ActionType  `json:"action"`
	FieldID       string      `json:"field_id"`
	NewValue      interface{} `json:"new_value"`
	Confirm       bool        `json:"confirm"`
	ConfirmDialog bool        `json:"confirm_dialog,omitempty"`
}

// ActionRequest represents an action request (backup, restart, reset)
type ActionRequest struct {
	Action      ActionType       `json:"action"`
	BackupID    string           `json:"backup_id,omitempty"`
	ConfirmType ConfirmationType `json:"confirm_type,omitempty"`
	ConfirmText string           `json:"confirm_text,omitempty"`
	Confirm     bool             `json:"confirm"`
}

// FieldChangeResponse represents the response to a field change
type FieldChangeResponse struct {
	Action          ActionType  `json:"action"`
	FieldID         string      `json:"field_id"`
	Status          string      `json:"status"`
	OldValue        interface{} `json:"old_value,omitempty"`
	NewValue        interface{} `json:"new_value,omitempty"`
	Timestamp       time.Time   `json:"timestamp"`
	RequiresRestart bool        `json:"requires_restart"`
	Warning         string      `json:"warning,omitempty"`
	Message         string      `json:"message,omitempty"`
	LogEntry        string      `json:"log_entry,omitempty"`
}

// ActionResponse represents the response to an action
type ActionResponse struct {
	Action           ActionType `json:"action"`
	Status           string     `json:"status"`
	Timestamp        time.Time  `json:"timestamp"`
	BackupID         string     `json:"backup_id,omitempty"`
	Size             float64    `json:"size_mb,omitempty"`
	Location         string     `json:"location,omitempty"`
	DownloadURL      string     `json:"download_url,omitempty"`
	Message          string     `json:"message,omitempty"`
	LogEntry         string     `json:"log_entry,omitempty"`
	CountdownSeconds int        `json:"countdown_s,omitempty"`
	RequiresRestart  bool       `json:"requires_restart,omitempty"`
	SettingsApplied  int        `json:"settings_applied,omitempty"`
	Changes          []string   `json:"changes,omitempty"`
}

// SettingsResponse represents the complete settings state
type SettingsResponse struct {
	Timestamp     time.Time                   `json:"timestamp"`
	Sections      map[string]*SectionResponse `json:"sections"`
	Accessibility map[string]bool             `json:"accessibility"`
	Advanced      *SectionResponse            `json:"advanced"` // Separate key for progressive disclosure
}

// SectionResponse groups fields/actions for a section
type SectionResponse struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Fields      []SettingsField  `json:"fields,omitempty"`
	Actions     []SettingsAction `json:"actions,omitempty"`
	Collapsed   bool             `json:"collapsed,omitempty"` // Advanced section only
}

// SettingsManager manages all application settings
type SettingsManager struct {
	mu       sync.RWMutex
	userRole UserRole

	// Core settings storage
	generalSettings  map[string]interface{}
	securitySettings map[string]interface{}
	systemSettings   map[string]interface{}

	// Confirmation tracking for dangerous actions
	pendingConfirmations map[string]*ConfirmationState

	// Dependencies
	getHAStatus     func() (bool, error)
	getSystemHealth func() (uptime string, storage string, memory string, err error)
	getVersion      func() string
	onRestart       func() error
	onBackupCreate  func() (backupID string, sizeMB float64, location string, err error)
	onBackupRestore func(backupID string) (settingsApplied int, changes []string, err error)
	onFactoryReset  func() error
	onLogEntry      func(level string, message string)
}

// ConfirmationState tracks a pending confirmation
type ConfirmationState struct {
	ActionID    string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ConfirmType ConfirmationType
}

// NewSettingsManager creates a new SettingsManager with dependency injection
func NewSettingsManager(
	getHAStatus func() (bool, error),
	getSystemHealth func() (uptime string, storage string, memory string, err error),
	getVersion func() string,
	onRestart func() error,
	onBackupCreate func() (backupID string, sizeMB float64, location string, err error),
	onBackupRestore func(backupID string) (settingsApplied int, changes []string, err error),
	onFactoryReset func() error,
	onLogEntry func(level string, message string),
) *SettingsManager {
	return &SettingsManager{
		userRole: RoleUser,

		// Initialize default settings
		generalSettings: map[string]interface{}{
			"language":       "en",
			"timezone":       "UTC",
			"high_contrast":  false,
			"large_text":     false,
			"reduced_motion": false,
		},

		securitySettings: map[string]interface{}{
			"alarm_entry_delay_s":         30,
			"alarm_exit_delay_s":          30,
			"alarm_arm_delay_s":           30,
			"alarm_trigger_sound_enabled": true,
			"guest_max_active":            1,
			"guest_request_timeout_s":     300,
			"guest_max_requests_per_hour": 10,
			"force_ha_connection":         true,
		},

		systemSettings: map[string]interface{}{
			"ha_status": "unknown",
			"last_sync": time.Now(),
			"uptime":    "0h",
			"storage":   "N/A",
			"memory":    "N/A",
			"version":   "unknown",
		},

		pendingConfirmations: make(map[string]*ConfirmationState),

		// Inject dependencies
		getHAStatus:     getHAStatus,
		getSystemHealth: getSystemHealth,
		getVersion:      getVersion,
		onRestart:       onRestart,
		onBackupCreate:  onBackupCreate,
		onBackupRestore: onBackupRestore,
		onFactoryReset:  onFactoryReset,
		onLogEntry:      onLogEntry,
	}
}

// SetUserRole updates the user's role for access control
func (sm *SettingsManager) SetUserRole(role UserRole) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.userRole = role
}

// GetSettings returns the complete settings state for the current user
func (sm *SettingsManager) GetSettings() (*SettingsResponse, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Only admin can access settings
	if sm.userRole != RoleAdmin {
		return nil, errors.New("insufficient permissions")
	}

	// Update system information
	haStatus, _ := sm.getHAStatus()
	uptime, storage, memory, _ := sm.getSystemHealth()
	version := sm.getVersion()

	sm.systemSettings["ha_status"] = haStatus
	sm.systemSettings["last_sync"] = time.Now()
	if uptime != "" {
		sm.systemSettings["uptime"] = uptime
	}
	if storage != "" {
		sm.systemSettings["storage"] = storage
	}
	if memory != "" {
		sm.systemSettings["memory"] = memory
	}
	sm.systemSettings["version"] = version

	// Build sections
	sections := make(map[string]*SectionResponse)

	sections["general"] = sm.buildGeneralSection()
	sections["security"] = sm.buildSecuritySection()
	sections["system"] = sm.buildSystemSection()

	// Advanced section for progressive disclosure
	advancedSection := sm.buildAdvancedSection()

	return &SettingsResponse{
		Timestamp: time.Now(),
		Sections:  sections,
		Accessibility: map[string]bool{
			"reduced_motion": sm.generalSettings["reduced_motion"].(bool),
			"large_text":     sm.generalSettings["large_text"].(bool),
			"high_contrast":  sm.generalSettings["high_contrast"].(bool),
		},
		Advanced: advancedSection,
	}, nil
}

// ApplyFieldChange applies a field change with confirmation handling
func (sm *SettingsManager) ApplyFieldChange(req *FieldChangeRequest) (*FieldChangeResponse, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Only admin can change settings
	if sm.userRole != RoleAdmin {
		return nil, errors.New("insufficient permissions")
	}

	field, err := sm.findField(req.FieldID)
	if err != nil {
		return nil, err
	}

	// Check if confirmation is required but not provided
	if field.RequireConfirm && !req.Confirm {
		return nil, errors.New("confirmation required")
	}

	// Get old value
	var oldValue interface{}
	switch field.Section {
	case SectionGeneral:
		oldValue = sm.generalSettings[field.ID]
		sm.generalSettings[field.ID] = req.NewValue
	case SectionSecurity:
		oldValue = sm.securitySettings[field.ID]
		sm.securitySettings[field.ID] = req.NewValue
	default:
		return nil, errors.New("cannot modify system or advanced settings")
	}

	// Determine if restart is needed
	requiresRestart := sm.settingRequiresRestart(field.ID)

	// Log the change
	logMsg := fmt.Sprintf("%s changed from %v to %v", field.ID, oldValue, req.NewValue)
	if field.Warning != "" {
		sm.onLogEntry("INFO", logMsg+" ("+field.Warning+")")
	} else {
		sm.onLogEntry("INFO", logMsg)
	}

	return &FieldChangeResponse{
		Action:          ActionFieldChange,
		FieldID:         req.FieldID,
		Status:          "success",
		OldValue:        oldValue,
		NewValue:        req.NewValue,
		Timestamp:       time.Now(),
		RequiresRestart: requiresRestart,
		Warning:         field.Warning,
		LogEntry:        logMsg,
	}, nil
}

// ApplyAction handles dangerous actions (restart, backup, factory reset)
func (sm *SettingsManager) ApplyAction(req *ActionRequest) (*ActionResponse, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Only admin can execute actions
	if sm.userRole != RoleAdmin {
		return nil, errors.New("insufficient permissions")
	}

	// Validate confirmation
	if !req.Confirm {
		return nil, errors.New("confirmation required")
	}

	switch req.Action {
	case ActionRestart:
		return sm.handleRestart()
	case ActionBackupCreate:
		return sm.handleBackupCreate()
	case ActionBackupRestore:
		return sm.handleBackupRestore(req.BackupID)
	case ActionFactoryReset:
		// Validate factory reset confirmation (requires "RESET" text and countdown)
		if req.ConfirmType != ConfirmTypeDouble || req.ConfirmText != "RESET" {
			return nil, errors.New("factory reset requires confirmation text 'RESET'")
		}
		return sm.handleFactoryReset()
	default:
		return nil, errors.New("unknown action")
	}
}

// handleRestart handles system restart
func (sm *SettingsManager) handleRestart() (*ActionResponse, error) {
	sm.onLogEntry("WARN", "System restart initiated by admin from Settings")

	if err := sm.onRestart(); err != nil {
		sm.onLogEntry("ERROR", fmt.Sprintf("System restart failed: %v", err))
		return nil, err
	}

	return &ActionResponse{
		Action:           ActionRestart,
		Status:           "success",
		Timestamp:        time.Now(),
		Message:          "settings.system.restart.success",
		LogEntry:         "System restart initiated by admin from Settings",
		CountdownSeconds: 10,
		RequiresRestart:  true,
	}, nil
}

// handleBackupCreate handles backup creation
func (sm *SettingsManager) handleBackupCreate() (*ActionResponse, error) {
	backupID, sizeMB, location, err := sm.onBackupCreate()
	if err != nil {
		sm.onLogEntry("ERROR", fmt.Sprintf("Backup creation failed: %v", err))
		return nil, err
	}

	sm.onLogEntry("INFO", fmt.Sprintf("Backup created: %s (%.1f MB)", backupID, sizeMB))

	downloadURL := fmt.Sprintf("/api/backups/%s.json", backupID)

	return &ActionResponse{
		Action:      ActionBackupCreate,
		Status:      "success",
		Timestamp:   time.Now(),
		BackupID:    backupID,
		Size:        sizeMB,
		Location:    location,
		DownloadURL: downloadURL,
		Message:     "settings.advanced.backup_create.success",
		LogEntry:    fmt.Sprintf("Backup created: %s (%.1f MB)", backupID, sizeMB),
	}, nil
}

// handleBackupRestore handles backup restoration
func (sm *SettingsManager) handleBackupRestore(backupID string) (*ActionResponse, error) {
	if backupID == "" {
		return nil, errors.New("backup_id required")
	}

	settingsApplied, changes, err := sm.onBackupRestore(backupID)
	if err != nil {
		sm.onLogEntry("ERROR", fmt.Sprintf("Backup restore failed: %v", err))
		return nil, err
	}

	logMsg := fmt.Sprintf("Backup restored from backup (%.0f settings applied)", float64(settingsApplied))
	sm.onLogEntry("WARN", logMsg)

	return &ActionResponse{
		Action:           ActionBackupRestore,
		Status:           "success",
		Timestamp:        time.Now(),
		BackupID:         backupID,
		Message:          "settings.advanced.backup_restore.success",
		LogEntry:         logMsg,
		CountdownSeconds: 10,
		RequiresRestart:  true,
		SettingsApplied:  settingsApplied,
		Changes:          changes,
	}, nil
}

// handleFactoryReset handles factory reset
func (sm *SettingsManager) handleFactoryReset() (*ActionResponse, error) {
	sm.onLogEntry("WARN", "Factory reset initiated by admin (all settings will be erased on startup)")

	if err := sm.onFactoryReset(); err != nil {
		sm.onLogEntry("ERROR", fmt.Sprintf("Factory reset failed: %v", err))
		return nil, err
	}

	return &ActionResponse{
		Action:           ActionFactoryReset,
		Status:           "success",
		Timestamp:        time.Now(),
		Message:          "settings.advanced.factory_reset.success",
		LogEntry:         "Factory reset initiated by admin (all settings will be erased on startup)",
		CountdownSeconds: 10,
		RequiresRestart:  true,
	}, nil
}

// Helper methods for building settings sections

func (sm *SettingsManager) buildGeneralSection() *SectionResponse {
	return &SectionResponse{
		Title:       "General Settings",
		Description: "Basic preferences and accessibility",
		Fields: []SettingsField{
			{
				ID:             "language",
				Section:        SectionGeneral,
				Type:           TypeString,
				Value:          sm.generalSettings["language"],
				DefaultValue:   "en",
				Help:           "Choose how SmartDisplay speaks to you",
				RequireConfirm: false,
				Options:        []string{"en", "tr"},
			},
			{
				ID:             "timezone",
				Section:        SectionGeneral,
				Type:           TypeString,
				Value:          sm.generalSettings["timezone"],
				DefaultValue:   "UTC",
				Help:           "Set your local time zone (ISO 8601)",
				RequireConfirm: false,
			},
			{
				ID:             "high_contrast",
				Section:        SectionGeneral,
				Type:           TypeBoolean,
				Value:          sm.generalSettings["high_contrast"],
				DefaultValue:   false,
				Help:           "Clearer separation between sections",
				RequireConfirm: false,
			},
			{
				ID:             "large_text",
				Section:        SectionGeneral,
				Type:           TypeBoolean,
				Value:          sm.generalSettings["large_text"],
				DefaultValue:   false,
				Help:           "Larger text for easier reading",
				RequireConfirm: false,
			},
			{
				ID:             "reduced_motion",
				Section:        SectionGeneral,
				Type:           TypeBoolean,
				Value:          sm.generalSettings["reduced_motion"],
				DefaultValue:   false,
				Help:           "No animations or transitions",
				RequireConfirm: false,
			},
		},
	}
}

func (sm *SettingsManager) buildSecuritySection() *SectionResponse {
	return &SectionResponse{
		Title:       "Security",
		Description: "Alarm and access control",
		Fields: []SettingsField{
			{
				ID:             "alarm_entry_delay_s",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["alarm_entry_delay_s"],
				DefaultValue:   30,
				Help:           "Entry delay in seconds before alarm triggers when a protected door/window opens",
				RequireConfirm: false,
				MinValue:       intPtr(0),
				MaxValue:       intPtr(600),
			},
			{
				ID:             "alarm_exit_delay_s",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["alarm_exit_delay_s"],
				DefaultValue:   30,
				Help:           "Exit delay in seconds to allow leaving before the alarm arms",
				RequireConfirm: false,
				MinValue:       intPtr(0),
				MaxValue:       intPtr(600),
			},
			{
				ID:             "alarm_arm_delay_s",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["alarm_arm_delay_s"],
				DefaultValue:   30,
				Help:           "Extra time to cancel alarm after voice confirmation",
				RequireConfirm: false,
				MinValue:       intPtr(10),
				MaxValue:       intPtr(300),
			},
			{
				ID:             "alarm_trigger_sound_enabled",
				Section:        SectionSecurity,
				Type:           TypeBoolean,
				Value:          sm.securitySettings["alarm_trigger_sound_enabled"],
				DefaultValue:   true,
				Help:           "Play audio alert when alarm triggers",
				Warning:        "Disabling sound means alarm will trigger silently. Enable sound to restore audible alerts.",
				RequireConfirm: true,
			},
			{
				ID:             "guest_max_active",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["guest_max_active"],
				DefaultValue:   1,
				Help:           "Maximum number of guests with access at the same time",
				Warning:        "Allowing more than 2 concurrent guests increases access risk",
				RequireConfirm: true,
				MinValue:       intPtr(1),
				MaxValue:       intPtr(10),
			},
			{
				ID:             "guest_request_timeout_s",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["guest_request_timeout_s"],
				DefaultValue:   300,
				Help:           "Seconds to wait for approval before expiring guest request",
				RequireConfirm: false,
				MinValue:       intPtr(60),
				MaxValue:       intPtr(3600),
			},
			{
				ID:             "guest_max_requests_per_hour",
				Section:        SectionSecurity,
				Type:           TypeInteger,
				Value:          sm.securitySettings["guest_max_requests_per_hour"],
				DefaultValue:   10,
				Help:           "Prevent spam by limiting guest requests",
				RequireConfirm: false,
				MinValue:       intPtr(1),
				MaxValue:       intPtr(100),
			},
			{
				ID:             "force_ha_connection",
				Section:        SectionSecurity,
				Type:           TypeBoolean,
				Value:          sm.securitySettings["force_ha_connection"],
				DefaultValue:   true,
				Help:           "Disable to allow offline operation if Home Assistant fails",
				Warning:        "SmartDisplay will stop working if Home Assistant becomes unavailable and this is enabled",
				RequireConfirm: true,
			},
		},
	}
}

func (sm *SettingsManager) buildSystemSection() *SectionResponse {
	return &SectionResponse{
		Title:       "System",
		Description: "Health and status information",
		Fields: []SettingsField{
			{
				ID:      "ha_status",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["ha_status"],
				Help:    "Current connection status to Home Assistant",
			},
			{
				ID:      "last_sync",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["last_sync"],
				Help:    "When SmartDisplay last synchronized with Home Assistant",
			},
			{
				ID:      "uptime",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["uptime"],
				Help:    "How long SmartDisplay has been running",
			},
			{
				ID:      "storage",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["storage"],
				Help:    "Available disk space for backups and logs",
			},
			{
				ID:      "memory",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["memory"],
				Help:    "Available RAM for system operations",
			},
			{
				ID:      "version",
				Section: SectionSystem,
				Type:    TypeReadOnly,
				Value:   sm.systemSettings["version"],
				Help:    "Current SmartDisplay software version",
			},
		},
		Actions: []SettingsAction{
			{
				ID:               "restart_now",
				Section:          SectionSystem,
				Name:             "Restart System",
				Help:             "Restart SmartDisplay. Any active requests will be canceled.",
				ConfirmType:      ConfirmTypeSimple,
				ConfirmText:      "Restart SmartDisplay?",
				CountdownSeconds: 10,
				RequiresRestart:  true,
			},
		},
	}
}

func (sm *SettingsManager) buildAdvancedSection() *SectionResponse {
	return &SectionResponse{
		Title:       "Advanced",
		Description: "Backup, restore, and system reset",
		Collapsed:   true,
		Actions: []SettingsAction{
			{
				ID:               "backup_create",
				Section:          SectionAdvanced,
				Name:             "Create Backup",
				Help:             "Create an encrypted backup of all settings",
				ConfirmType:      ConfirmTypeSimple,
				ConfirmText:      "Create backup?",
				CountdownSeconds: 0,
				RequiresRestart:  false,
			},
			{
				ID:               "backup_restore",
				Section:          SectionAdvanced,
				Name:             "Restore Backup",
				Help:             "Restore all settings from a previous backup",
				ConfirmType:      ConfirmTypeStrong,
				ConfirmText:      "Restore Backup?",
				CountdownSeconds: 10,
				RequiresRestart:  true,
			},
			{
				ID:               "factory_reset",
				Section:          SectionAdvanced,
				Name:             "Factory Reset",
				Help:             "Erase all settings and return to first-boot setup",
				Danger:           "This will erase ALL custom settings. This action cannot be undone.",
				ConfirmType:      ConfirmTypeDouble,
				ConfirmText:      "Type 'RESET' to confirm factory reset",
				CountdownSeconds: 10,
				RequiresRestart:  true,
			},
		},
	}
}

// findField locates a field by ID across all sections
func (sm *SettingsManager) findField(fieldID string) (*SettingsField, error) {
	generalSection := sm.buildGeneralSection()
	for _, field := range generalSection.Fields {
		if field.ID == fieldID {
			return &field, nil
		}
	}

	securitySection := sm.buildSecuritySection()
	for _, field := range securitySection.Fields {
		if field.ID == fieldID {
			return &field, nil
		}
	}

	return nil, fmt.Errorf("field not found: %s", fieldID)
}

// settingRequiresRestart determines if a setting change requires restart
func (sm *SettingsManager) settingRequiresRestart(fieldID string) bool {
	restartFields := map[string]bool{
		"language":            true,
		"timezone":            true,
		"high_contrast":       true,
		"large_text":          true,
		"force_ha_connection": true,
	}
	return restartFields[fieldID]
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// ValidateSettings validates the current settings state
func (sm *SettingsManager) ValidateSettings() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Validate General settings
	if language, ok := sm.generalSettings["language"].(string); !ok || (language != "en" && language != "tr") {
		return errors.New("invalid language setting")
	}

	// Validate Security settings
	if entryDelay, ok := sm.securitySettings["alarm_entry_delay_s"].(int); !ok || entryDelay < 0 || entryDelay > 600 {
		return errors.New("invalid alarm_entry_delay_s")
	}

	if exitDelay, ok := sm.securitySettings["alarm_exit_delay_s"].(int); !ok || exitDelay < 0 || exitDelay > 600 {
		return errors.New("invalid alarm_exit_delay_s")
	}

	if armDelay, ok := sm.securitySettings["alarm_arm_delay_s"].(int); !ok || armDelay < 10 || armDelay > 300 {
		return errors.New("invalid alarm_arm_delay_s")
	}

	if guestMax, ok := sm.securitySettings["guest_max_active"].(int); !ok || guestMax < 1 || guestMax > 10 {
		return errors.New("invalid guest_max_active")
	}

	return nil
}
