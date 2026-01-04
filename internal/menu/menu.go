// Package menu provides menu structure and role-based visibility management.
// menu.go implements DESIGN Phase D5: Menu structure and role-based perception.
package menu

import (
	"fmt"
	"smartdisplay-core/internal/logger"
)

// UserRole represents the user's role in the system
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

// MenuSection represents a top-level menu section
type MenuSection struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Visible      bool         `json:"visible"`
	Actions      []MenuAction `json:"actions"`
	SubSections  []SubSection `json:"sub_sections,omitempty"`
	ReasonHidden string       `json:"reason_hidden,omitempty"`
}

// MenuAction represents an actionable item within a section
type MenuAction struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// SubSection represents a sub-item within a menu section
type SubSection struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
}

// MenuResponse represents the complete menu for the authenticated user
type MenuResponse struct {
	UserID          string        `json:"user_id"`
	Role            UserRole      `json:"role"`
	FirstBootActive bool          `json:"first_boot_active"`
	FailsafeActive  bool          `json:"failsafe_active"`
	GuestActive     bool          `json:"guest_active"`
	Sections        []MenuSection `json:"sections"`
}

// MenuManager manages menu structure and role-based visibility
type MenuManager struct {
	// Dependency injection functions
	firstBootActiveFn  func() bool // Check if first-boot active
	failsafeActiveFn   func() bool // Check if failsafe active
	guestActiveCheckFn func() bool // Check if guest requesting/approved

	// Configuration
	userRole UserRole

	// Tracking
	lastResolvedMenu *MenuResponse
}

// NewMenuManager creates a new menu manager
func NewMenuManager(
	firstBootActiveFn func() bool,
	failsafeActiveFn func() bool,
	guestActiveCheckFn func() bool,
	userRole UserRole,
) *MenuManager {
	return &MenuManager{
		firstBootActiveFn:  firstBootActiveFn,
		failsafeActiveFn:   failsafeActiveFn,
		guestActiveCheckFn: guestActiveCheckFn,
		userRole:           userRole,
		lastResolvedMenu:   nil,
	}
}

// SetUserRole sets the current user role (for role changes)
func (m *MenuManager) SetUserRole(role UserRole) {
	if m.userRole != role {
		logger.Info(fmt.Sprintf("menu: role changed (%s → %s)", m.userRole, role))
		m.userRole = role
		m.lastResolvedMenu = nil // Invalidate cached menu
	}
}

// ResolveMenu builds the menu for the current user role and system state
func (m *MenuManager) ResolveMenu(userID string) *MenuResponse {
	firstBootActive := m.firstBootActiveFn()
	failsafeActive := m.failsafeActiveFn()
	guestActive := m.guestActiveCheckFn()

	resp := &MenuResponse{
		UserID:          userID,
		Role:            m.userRole,
		FirstBootActive: firstBootActive,
		FailsafeActive:  failsafeActive,
		GuestActive:     guestActive,
		Sections:        []MenuSection{},
	}

	// Always add Home section
	resp.Sections = append(resp.Sections, m.buildHomeSection(firstBootActive))

	// Alarm section visibility rules
	resp.Sections = append(resp.Sections, m.buildAlarmSection(firstBootActive))

	// Guest section visibility rules
	resp.Sections = append(resp.Sections, m.buildGuestSection(firstBootActive))

	// Devices section visibility rules
	if !firstBootActive {
		resp.Sections = append(resp.Sections, m.buildDevicesSection())
	}

	// History section visibility rules
	if !firstBootActive {
		resp.Sections = append(resp.Sections, m.buildHistorySection())
	}

	// Settings section visibility rules
	if m.userRole == RoleAdmin && !firstBootActive {
		resp.Sections = append(resp.Sections, m.buildSettingsSection(failsafeActive))
	}

	m.lastResolvedMenu = resp

	logger.Info(fmt.Sprintf("menu: resolved for role=%s (sections=%d, first_boot=%v, failsafe=%v, guest_active=%v)",
		m.userRole, len(resp.Sections), firstBootActive, failsafeActive, guestActive))

	return resp
}

// buildHomeSection creates the Home menu section
func (m *MenuManager) buildHomeSection(firstBootActive bool) MenuSection {
	actions := []MenuAction{
		{ID: "view_summary", Name: "View Summary", Enabled: true},
		{ID: "quick_alarm", Name: "Alarm Status", Enabled: true},
	}

	if m.userRole == RoleAdmin && firstBootActive {
		actions = append(actions, MenuAction{ID: "setup_progress", Name: "Setup Progress", Enabled: true})
	}

	return MenuSection{
		ID:          "home",
		Name:        "Home",
		Description: "Dashboard overview",
		Visible:     true,
		Actions:     actions,
	}
}

// buildAlarmSection creates the Alarm menu section
func (m *MenuManager) buildAlarmSection(firstBootActive bool) MenuSection {
	section := MenuSection{
		ID:          "alarm",
		Name:        "Alarm",
		Description: "Control and monitor alarm",
		Visible:     true,
		Actions:     []MenuAction{},
	}

	switch m.userRole {
	case RoleAdmin:
		section.Actions = []MenuAction{
			{ID: "view_state", Name: "View State", Enabled: true},
			{ID: "arm", Name: "Arm", Enabled: !firstBootActive},
			{ID: "disarm", Name: "Disarm", Enabled: !firstBootActive},
			{ID: "view_history", Name: "History", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "current_state", Name: "Current State", Visible: true},
			{ID: "history", Name: "History", Visible: true},
		}

	case RoleUser:
		section.Actions = []MenuAction{
			{ID: "view_state", Name: "View State", Enabled: true},
			{ID: "view_history", Name: "History", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "current_state", Name: "Current State", Visible: true},
			{ID: "history", Name: "History", Visible: true},
		}

	case RoleGuest:
		// Guest sees Alarm only if requesting or approved
		// For now, always include in section resolution
		section.Actions = []MenuAction{
			{ID: "view_state", Name: "View State", Enabled: true},
		}
	}

	return section
}

// buildGuestSection creates the Guest menu section
func (m *MenuManager) buildGuestSection(firstBootActive bool) MenuSection {
	section := MenuSection{
		ID:          "guest",
		Name:        "Guest",
		Description: "Manage guest access",
		Visible:     !firstBootActive,
		Actions:     []MenuAction{},
	}

	if !section.Visible {
		section.ReasonHidden = "first_boot_active"
		return section
	}

	switch m.userRole {
	case RoleAdmin:
		section.Actions = []MenuAction{
			{ID: "view_requests", Name: "View Requests", Enabled: true},
			{ID: "approve", Name: "Approve", Enabled: true},
			{ID: "deny", Name: "Deny", Enabled: true},
			{ID: "view_history", Name: "History", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "pending", Name: "Pending Requests", Visible: true},
			{ID: "history", Name: "Guest History", Visible: true},
		}

	case RoleUser:
		section.Actions = []MenuAction{
			{ID: "view_requests", Name: "View Requests", Enabled: true},
			{ID: "view_history", Name: "History", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "pending", Name: "Pending Requests", Visible: true},
			{ID: "history", Name: "Guest History", Visible: true},
		}

	case RoleGuest:
		section.Actions = []MenuAction{
			{ID: "view_status", Name: "My Status", Enabled: true},
			{ID: "request_access", Name: "Request Access", Enabled: true},
			{ID: "exit", Name: "Exit", Enabled: true},
		}
	}

	return section
}

// buildDevicesSection creates the Devices menu section
func (m *MenuManager) buildDevicesSection() MenuSection {
	section := MenuSection{
		ID:          "devices",
		Name:        "Devices",
		Description: "View device status",
		Visible:     m.userRole == RoleAdmin || m.userRole == RoleUser,
		Actions:     []MenuAction{},
	}

	if !section.Visible {
		section.ReasonHidden = "permission_insufficient"
		return section
	}

	switch m.userRole {
	case RoleAdmin:
		section.Actions = []MenuAction{
			{ID: "view_list", Name: "View List", Enabled: true},
			{ID: "view_details", Name: "Details", Enabled: true},
			{ID: "manage", Name: "Manage", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "status", Name: "Status", Visible: true},
			{ID: "battery", Name: "Battery Levels", Visible: true},
			{ID: "technical", Name: "Technical Info", Visible: true},
		}

	case RoleUser:
		section.Actions = []MenuAction{
			{ID: "view_list", Name: "View List", Enabled: true},
			{ID: "view_battery", Name: "Battery Status", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "status", Name: "Status", Visible: true},
			{ID: "battery", Name: "Battery Levels", Visible: true},
		}
	}

	return section
}

// buildHistorySection creates the History menu section
func (m *MenuManager) buildHistorySection() MenuSection {
	section := MenuSection{
		ID:          "history",
		Name:        "History",
		Description: "System activity log",
		Visible:     m.userRole == RoleAdmin || m.userRole == RoleUser,
		Actions:     []MenuAction{},
	}

	if !section.Visible {
		section.ReasonHidden = "permission_insufficient"
		return section
	}

	switch m.userRole {
	case RoleAdmin:
		section.Actions = []MenuAction{
			{ID: "view_events", Name: "View Events", Enabled: true},
			{ID: "search", Name: "Search", Enabled: true},
			{ID: "export", Name: "Export", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "system_events", Name: "System Events", Visible: true},
			{ID: "alarm_events", Name: "Alarm Events", Visible: true},
			{ID: "user_actions", Name: "User Actions", Visible: true},
		}

	case RoleUser:
		section.Actions = []MenuAction{
			{ID: "view_events", Name: "View Events", Enabled: true},
		}
		section.SubSections = []SubSection{
			{ID: "system_events", Name: "System Events", Visible: true},
			{ID: "alarm_events", Name: "Alarm Events", Visible: true},
		}
	}

	return section
}

// buildSettingsSection creates the Settings menu section
func (m *MenuManager) buildSettingsSection(failsafeActive bool) MenuSection {
	section := MenuSection{
		ID:          "settings",
		Name:        "Settings",
		Description: "System configuration",
		Visible:     m.userRole == RoleAdmin,
		Actions:     []MenuAction{},
	}

	if !section.Visible {
		section.ReasonHidden = "permission_insufficient"
		return section
	}

	// Admin sees settings, but actions may be disabled during failsafe
	section.Actions = []MenuAction{
		{ID: "alarm_config", Name: "Alarm Settings", Enabled: !failsafeActive},
		{ID: "user_mgmt", Name: "User Management", Enabled: !failsafeActive},
		{ID: "device_mgmt", Name: "Device Management", Enabled: !failsafeActive},
		{ID: "accessibility", Name: "Accessibility", Enabled: true},
	}

	section.SubSections = []SubSection{
		{ID: "alarm", Name: "Alarm Config", Visible: true},
		{ID: "users", Name: "Users", Visible: true},
		{ID: "devices", Name: "Devices", Visible: true},
		{ID: "accessibility", Name: "Accessibility", Visible: true},
	}

	return section
}

// GetVisibleSectionCount returns the number of visible sections
func (m *MenuManager) GetVisibleSectionCount(resp *MenuResponse) int {
	count := 0
	for _, section := range resp.Sections {
		if section.Visible {
			count++
		}
	}
	return count
}

// ValidateMenuState checks if menu is valid (no hidden sections with actions)
func (m *MenuManager) ValidateMenuState(resp *MenuResponse) bool {
	for _, section := range resp.Sections {
		if !section.Visible && len(section.Actions) > 0 {
			logger.Info(fmt.Sprintf("menu: validation error - hidden section '%s' has actions", section.ID))
			return false
		}
	}
	return true
}
