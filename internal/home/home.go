// Package home manages the home/idle screen state machine and behavioral logic.
// Implements DESIGN Phase D2: Home screen state machine with 4 states and 2 APIs.
package home

import (
	"fmt"
	"smartdisplay-core/internal/logger"
	"time"
)

// HomeState represents the current state of the home screen
type HomeState string

const (
	StateSetupRedirect HomeState = "setup_redirect"
	StateIdle          HomeState = "idle"
	StateActive        HomeState = "active"
	StateAlert         HomeState = "alert"
)

// ActionButton represents an action available in Active state
type ActionButton struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Icon     string `json:"icon,omitempty"`
	Requires string `json:"requires,omitempty"` // auth level required
}

// AlertInfo holds alert details for Alert state
type AlertInfo struct {
	Priority    string         `json:"priority"` // critical, high, medium
	Type        string         `json:"type"`     // alarm_triggered, system_error, etc.
	Message     string         `json:"message"`
	TriggeredAt string         `json:"triggered_at"`
	Reason      string         `json:"reason,omitempty"`
	Location    string         `json:"location,omitempty"`
	Actions     []ActionButton `json:"actions"`
}

// Summary holds basic home screen information (for both Idle and full state responses)
type Summary struct {
	AlarmState         string `json:"alarm_state"` // DISARMED, ARMED, TRIGGERED
	HAConnected        bool   `json:"ha_connected"`
	CurrentTime        string `json:"current_time"`          // RFC 3339 format
	AIInsight          string `json:"ai_insight"`            // Optional one-liner
	GuestState         string `json:"guest_state,omitempty"` // null or state string
	CountdownActive    bool   `json:"countdown_active"`
	CountdownRemaining int    `json:"countdown_remaining"` // seconds
}

// ExpandedInfo holds additional information for Active state
type ExpandedInfo struct {
	RecentEvents  []map[string]interface{} `json:"recent_events,omitempty"`
	DeviceStatus  map[string]interface{}   `json:"device_status,omitempty"`
	SystemHealth  map[string]interface{}   `json:"system_health,omitempty"`
	FullAIInsight string                   `json:"full_ai_insight,omitempty"`
}

// HomeStateResponse is the full response structure for /api/ui/home/state
type HomeStateResponse struct {
	State        HomeState                 `json:"state"`
	SystemReady  bool                      `json:"system_ready"`
	Summary      Summary                   `json:"summary"`
	Alert        *AlertInfo                `json:"alert,omitempty"`
	Actions      map[string][]ActionButton `json:"actions,omitempty"`
	ExpandedInfo *ExpandedInfo             `json:"expanded_info,omitempty"`
	Message      string                    `json:"message"`
}

// SummaryResponse is the lightweight response for /api/ui/home/summary
type SummaryResponse struct {
	AlarmState         string `json:"alarm_state"`
	HAConnected        bool   `json:"ha_connected"`
	CurrentTime        string `json:"current_time"`
	AIInsight          string `json:"ai_insight"`
	GuestState         string `json:"guest_state,omitempty"`
	CountdownActive    bool   `json:"countdown_active"`
	CountdownRemaining int    `json:"countdown_remaining"`
	HasPendingAlerts   bool   `json:"has_pending_alerts"`
}

// HomeStateManager manages home screen state machine
type HomeStateManager struct {
	// State tracking
	currentState         HomeState
	activeStateEnteredAt time.Time // for timeout tracking
	activeStateTimeout   time.Duration
	lastInteractionTime  time.Time
	inactivityThreshold  time.Duration

	// Alert state
	currentAlert *AlertInfo
	alertEntered time.Time

	// Dependencies (injected)
	firstBootActive    func() bool         // Check if first-boot is active
	alarmState         func() string       // Get alarm state
	haConnected        func() bool         // Check HA connection
	aiInsight          func() string       // Get AI insight
	guestState         func() string       // Get guest state
	countdownActive    func() bool         // Check if countdown active
	countdownRemaining func() int          // Get countdown remaining seconds
	expandedInfo       func() ExpandedInfo // Get expanded info for Active state
	userRole           func() string       // Get current user role (admin/user/guest)
}

// NewHomeStateManager creates a new home state manager
func NewHomeStateManager(
	firstBootActiveFn func() bool,
	alarmStateFn func() string,
	haConnectedFn func() bool,
	aiInsightFn func() string,
	guestStateFn func() string,
	countdownActiveFn func() bool,
	countdownRemainingFn func() int,
) *HomeStateManager {
	mgr := &HomeStateManager{
		currentState:        StateIdle,
		activeStateTimeout:  5 * time.Minute, // Default 5 minute timeout
		inactivityThreshold: 5 * time.Minute,
		lastInteractionTime: time.Now(),
		firstBootActive:     firstBootActiveFn,
		alarmState:          alarmStateFn,
		haConnected:         haConnectedFn,
		aiInsight:           aiInsightFn,
		guestState:          guestStateFn,
		countdownActive:     countdownActiveFn,
		countdownRemaining:  countdownRemainingFn,
	}
	logger.Info("home: state manager initialized (state: idle)")
	return mgr
}

// SetExpandedInfoProvider sets the function to get expanded info
func (m *HomeStateManager) SetExpandedInfoProvider(fn func() ExpandedInfo) {
	m.expandedInfo = fn
}

// SetUserRoleProvider sets the function to get current user role
func (m *HomeStateManager) SetUserRoleProvider(fn func() string) {
	m.userRole = fn
}

// SetActiveStateTimeout configures the Active state timeout duration
func (m *HomeStateManager) SetActiveStateTimeout(d time.Duration) {
	m.activeStateTimeout = d
	logger.Info("home: active state timeout set to " + fmt.Sprintf("%v", d))
}

// OnUserInteraction is called when user interacts with the system
func (m *HomeStateManager) OnUserInteraction() {
	m.lastInteractionTime = time.Now()

	// Transition to Active unless already in Alert
	if m.currentState == StateSetupRedirect {
		return // Don't transition during setup
	}
	if m.currentState == StateAlert {
		return // Don't transition out of Alert
	}
	if m.currentState != StateActive {
		oldState := m.currentState
		m.currentState = StateActive
		m.activeStateEnteredAt = time.Now()
		logger.Info("home: state transition (" + string(oldState) + " → " + string(StateActive) + ", trigger: user_interaction)")
	}
}

// OnAlertTriggered transitions to Alert state with given alert info
func (m *HomeStateManager) OnAlertTriggered(alert *AlertInfo) {
	if m.currentAlert == nil || m.currentAlert.Type != alert.Type {
		oldState := m.currentState
		m.currentState = StateAlert
		m.currentAlert = alert
		m.alertEntered = time.Now()
		logger.Info("home: state transition (" + string(oldState) + " → " + string(StateAlert) + ", trigger: " + alert.Type + ")")
	}
}

// OnAlertResolved transitions out of Alert state back to Idle
func (m *HomeStateManager) OnAlertResolved(alertType string) {
	if m.currentState == StateAlert && m.currentAlert != nil && m.currentAlert.Type == alertType {
		logger.Info("home: alert dismissed (type: " + alertType + ")")
		m.currentState = StateIdle
		m.currentAlert = nil
		m.lastInteractionTime = time.Now()
	}
}

// EvaluateState evaluates and returns the current state (checks first-boot, timeout, etc.)
func (m *HomeStateManager) EvaluateState() HomeState {
	// Priority 1: Setup redirect (wizard not completed)
	if m.firstBootActive != nil && m.firstBootActive() {
		if m.currentState != StateSetupRedirect {
			m.currentState = StateSetupRedirect
			logger.Info("home: setup_redirect state (wizard not completed)")
		}
		return StateSetupRedirect
	}

	// Priority 2: Alert (overrides everything else)
	if m.currentState == StateAlert {
		return StateAlert
	}

	// Priority 3: Check for Active state timeout
	if m.currentState == StateActive && m.activeStateEnteredAt.Add(m.activeStateTimeout).Before(time.Now()) {
		m.currentState = StateIdle
		logger.Info("home: state transition (active → idle, trigger: inactivity_timeout)")
		return StateIdle
	}

	// Return current state
	return m.currentState
}

// GetCurrentState returns the current evaluated home state
func (m *HomeStateManager) GetCurrentState() HomeState {
	return m.EvaluateState()
}

// BuildSummary builds the Summary object from current dependencies
func (m *HomeStateManager) BuildSummary() Summary {
	guestState := ""
	if m.guestState != nil {
		gs := m.guestState()
		if gs != "" && gs != "IDLE" {
			guestState = gs
		}
	}

	return Summary{
		AlarmState:         m.alarmState(),
		HAConnected:        m.haConnected(),
		CurrentTime:        time.Now().UTC().Format(time.RFC3339),
		AIInsight:          m.aiInsight(),
		GuestState:         guestState,
		CountdownActive:    m.countdownActive(),
		CountdownRemaining: m.countdownRemaining(),
	}
}

// BuildActions builds action buttons based on user role and current state
func (m *HomeStateManager) BuildActions() map[string][]ActionButton {
	role := ""
	if m.userRole != nil {
		role = m.userRole()
	}

	actions := make(map[string][]ActionButton)

	// Primary actions (always available in Active state)
	primary := []ActionButton{
		{
			ID:      "arm",
			Label:   "Arm",
			Enabled: m.alarmState() == "DISARMED",
		},
		{
			ID:      "disarm",
			Label:   "Disarm",
			Enabled: m.alarmState() == "ARMED",
		},
	}

	// Secondary actions (based on role)
	secondary := []ActionButton{}
	switch role {
	case "admin":
		secondary = []ActionButton{
			{
				ID:      "guests",
				Label:   "Guest Requests",
				Enabled: true,
			},
			{
				ID:      "anomalies",
				Label:   "Anomalies",
				Enabled: true,
			},
			{
				ID:      "settings",
				Label:   "Settings",
				Enabled: true,
			},
		}
	case "user":
		secondary = []ActionButton{
			{
				ID:      "anomalies",
				Label:   "Anomalies",
				Enabled: true,
			},
		}
	case "guest":
		secondary = []ActionButton{
			{
				ID:      "request_entry",
				Label:   "Request Entry",
				Enabled: true,
			},
		}
	}

	actions["primary"] = primary
	actions["secondary"] = secondary

	return actions
}

// BuildExpandedInfo builds expanded information for Active state
func (m *HomeStateManager) BuildExpandedInfo() *ExpandedInfo {
	if m.expandedInfo == nil {
		return nil
	}
	info := m.expandedInfo()
	return &info
}

// GetStateResponse returns the full state response for the API
func (m *HomeStateManager) GetStateResponse() HomeStateResponse {
	state := m.EvaluateState()
	summary := m.BuildSummary()
	systemReady := state != StateSetupRedirect && state != StateAlert

	resp := HomeStateResponse{
		State:       state,
		SystemReady: systemReady,
		Summary:     summary,
		Message:     m.getStateMessage(state),
	}

	// Add alert info if in Alert state
	if state == StateAlert && m.currentAlert != nil {
		resp.Alert = m.currentAlert
	}

	// Add actions if in Active state
	if state == StateActive {
		resp.Actions = m.BuildActions()
		resp.ExpandedInfo = m.BuildExpandedInfo()
	}

	return resp
}

// GetSummaryResponse returns the lightweight summary response
func (m *HomeStateManager) GetSummaryResponse() SummaryResponse {
	summary := m.BuildSummary()
	hasPendingAlerts := m.currentState == StateAlert || (m.currentAlert != nil)

	return SummaryResponse{
		AlarmState:         summary.AlarmState,
		HAConnected:        summary.HAConnected,
		CurrentTime:        summary.CurrentTime,
		AIInsight:          summary.AIInsight,
		GuestState:         summary.GuestState,
		CountdownActive:    summary.CountdownActive,
		CountdownRemaining: summary.CountdownRemaining,
		HasPendingAlerts:   hasPendingAlerts,
	}
}

// getStateMessage returns a human-readable message for the current state
func (m *HomeStateManager) getStateMessage(state HomeState) string {
	switch state {
	case StateSetupRedirect:
		return "Setup required"
	case StateIdle:
		return "All systems calm"
	case StateActive:
		return "Ready for input"
	case StateAlert:
		if m.currentAlert != nil {
			return "Critical alert"
		}
		return "Alert"
	default:
		return "Unknown state"
	}
}

// IsInActiveState returns true if currently in Active state
func (m *HomeStateManager) IsInActiveState() bool {
	return m.currentState == StateActive
}

// IsInAlertState returns true if currently in Alert state
func (m *HomeStateManager) IsInAlertState() bool {
	return m.currentState == StateAlert
}

// IsSetupRequired returns true if first-boot not completed
func (m *HomeStateManager) IsSetupRequired() bool {
	return m.firstBootActive != nil && m.firstBootActive()
}

// GetCurrentAlert returns the current alert info, or nil
func (m *HomeStateManager) GetCurrentAlert() *AlertInfo {
	return m.currentAlert
}

// UpdateAlert updates the alert information (used by subsystems)
func (m *HomeStateManager) UpdateAlert(alert *AlertInfo) {
	m.OnAlertTriggered(alert)
}

// ClearAlert clears the current alert (used by subsystems)
func (m *HomeStateManager) ClearAlert() {
	if m.currentAlert != nil {
		m.OnAlertResolved(m.currentAlert.Type)
	}
}
