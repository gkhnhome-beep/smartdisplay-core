// Package alarm provides alarm state machine and screen state management.
// screen.go implements DESIGN Phase D3: Alarm screen state exposure and APIs.
package alarm

import (
	"fmt"
	"smartdisplay-core/internal/logger"
	"time"
)

// ScreenMode represents the alarm screen display mode
type ScreenMode string

const (
	ModeDisarmed  ScreenMode = "disarmed"
	ModeArming    ScreenMode = "arming"
	ModeArmed     ScreenMode = "armed"
	ModeTriggered ScreenMode = "triggered"
	ModeBlocked   ScreenMode = "blocked"
)

// BlockReason explains why alarm screen is blocked
type BlockReason string

const (
	BlockFirstBootActive     BlockReason = "first_boot_active"
	BlockGuestRequestPending BlockReason = "guest_request_pending"
	BlockFailsafeActive      BlockReason = "failsafe_active"
)

// CountdownInfo provides countdown state data
type CountdownInfo struct {
	TotalSeconds     int    `json:"total_seconds"`
	RemainingSeconds int    `json:"remaining_seconds"`
	Percentage       int    `json:"percentage"`
	StartedAt        string `json:"started_at"`       // RFC 3339
	WillCompleteAt   string `json:"will_complete_at"` // RFC 3339
}

// TriggerInfo provides alarm trigger details
type TriggerInfo struct {
	Severity                string `json:"severity"`       // critical, high, medium
	Priority                int    `json:"priority"`       // 1=highest
	TriggeredAt             string `json:"triggered_at"`   // RFC 3339
	TriggerReason           string `json:"trigger_reason"` // door_unlock, motion_detected, etc.
	TriggerLocation         string `json:"trigger_location,omitempty"`
	TimeTriggeredAgoSeconds int    `json:"time_triggered_ago_seconds"`
}

// ActionInfo describes an action button for the alarm screen
type ActionInfo struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Enabled      bool   `json:"enabled"`
	RequiresAuth bool   `json:"requires_auth"`
}

// BlockedInfo describes why alarm is blocked
type BlockedInfo struct {
	Reason           string `json:"reason"` // block reason enum
	RecoveryAction   string `json:"recovery_action,omitempty"`
	RedirectURL      string `json:"redirect_url,omitempty"`
	EstimatedSeconds int    `json:"estimated_seconds,omitempty"`
}

// GuestRequestInfo describes pending guest entry request
type GuestRequestInfo struct {
	GuestID            string `json:"guest_id"`
	RequestedAt        string `json:"requested_at"` // RFC 3339
	TimeWaitingSeconds int    `json:"time_waiting_seconds"`
	ExpiresAt          string `json:"expires_at,omitempty"` // RFC 3339
}

// FirstBootInfo describes first-boot blocking
type FirstBootInfo struct {
	WizardActive   bool   `json:"wizard_active"`
	CurrentStep    string `json:"current_step"`
	StepsRemaining int    `json:"steps_remaining"`
}

// FailsafeInfo describes failsafe blocking
type FailsafeInfo struct {
	FailsafeActive        bool   `json:"failsafe_active"`
	Reason                string `json:"reason"`                  // connection_lost, power_failure, sensor_malfunction
	StartedAt             string `json:"started_at"`              // RFC 3339
	EstimatedRecoveryTime int    `json:"estimated_recovery_time"` // seconds
}

// InfoContext holds contextual information for the alarm screen response
type InfoContext struct {
	CanArm                  bool              `json:"can_arm,omitempty"`
	CanDisarm               bool              `json:"can_disarm,omitempty"`
	CanCancel               bool              `json:"can_cancel,omitempty"`
	AcknowledgmentRequired  bool              `json:"acknowledgment_required,omitempty"`
	ReasonBlocked           string            `json:"reason_blocked,omitempty"`
	NextAction              string            `json:"next_action,omitempty"`
	SensorsActive           int               `json:"sensors_active,omitempty"`
	LastCheck               string            `json:"last_check,omitempty"` // RFC 3339
	ProtectionStatus        string            `json:"protection_status,omitempty"`
	AutoArmAt               string            `json:"auto_arm_at,omitempty"`
	EscalationTimeRemaining int               `json:"escalation_time_remaining,omitempty"`
	RecoveryAction          string            `json:"recovery_action,omitempty"`
	EstimatedSeconds        int               `json:"estimated_seconds,omitempty"`
	GuestInfo               *GuestRequestInfo `json:"guest_info,omitempty"`
	FirstBootInfo           *FirstBootInfo    `json:"first_boot_info,omitempty"`
	FailsafeInfo            *FailsafeInfo     `json:"failsafe_info,omitempty"`
}

// ScreenState represents complete alarm screen state for API response
type ScreenState struct {
	Mode        ScreenMode     `json:"mode"`
	BlockReason string         `json:"block_reason,omitempty"` // Only if mode==blocked
	Message     string         `json:"message"`
	Context     string         `json:"context"`
	Timestamp   string         `json:"timestamp"`           // RFC 3339
	ArmedAt     string         `json:"armed_at,omitempty"`  // For armed state
	Countdown   *CountdownInfo `json:"countdown,omitempty"` // For arming state
	Alert       *TriggerInfo   `json:"alert,omitempty"`     // For triggered state
	Actions     []ActionInfo   `json:"actions"`
	Info        InfoContext    `json:"info"`
}

// SummaryState represents lightweight alarm summary for frequent polling
type SummaryState struct {
	Mode                      ScreenMode `json:"mode"`
	Message                   string     `json:"message"`
	Context                   string     `json:"context"`
	CountdownRemainingSeconds *int       `json:"countdown_remaining_seconds"`
	TriggeredAgoSeconds       *int       `json:"triggered_ago_seconds"`
	ActionsAvailable          int        `json:"actions_available"`
	Priority                  string     `json:"priority"` // normal, warning, critical
}

// ScreenStateManager maps alarm core state to screen presentation state
type ScreenStateManager struct {
	// Dependency injection functions
	firstBootActiveFn      func() bool                                                         // Check if first-boot is active
	alarmStateFn           func() string                                                       // Get alarm state (DISARMED, ARMED, TRIGGERED, etc.)
	countdownActiveFn      func() bool                                                         // Check if countdown is active
	countdownRemainFn      func() int                                                          // Get countdown remaining seconds
	countdownStartedAtFn   func() time.Time                                                    // Get when countdown started
	guestRequestPendingFn  func() bool                                                         // Check if guest request pending
	guestRequestInfoFn     func() (guestID string, requestedAt time.Time, expiresAt time.Time) // Get guest info
	failsafeActiveFn       func() bool                                                         // Check if failsafe is active
	failsafeReasonFn       func() string                                                       // Get failsafe reason
	failsafeStartedAtFn    func() time.Time                                                    // Get when failsafe started
	failsafeEstimateSecsFn func() int                                                          // Get estimated recovery time

	// State tracking
	triggeredAt       time.Time
	triggeredReason   string
	triggeredLocation string
	armedAt           time.Time

	// Configuration
	countdownTotal int // Total countdown seconds (default 30)

	// Cached state
	lastEvaluatedMode ScreenMode
	lastEvaluatedTime time.Time
}

// NewScreenStateManager creates a new screen state manager with dependency injection
func NewScreenStateManager(
	firstBootActiveFn func() bool,
	alarmStateFn func() string,
	countdownActiveFn func() bool,
	countdownRemainFn func() int,
	countdownStartedAtFn func() time.Time,
	guestRequestPendingFn func() bool,
	guestRequestInfoFn func() (string, time.Time, time.Time),
	failsafeActiveFn func() bool,
	failsafeReasonFn func() string,
	failsafeStartedAtFn func() time.Time,
	failsafeEstimateSecsFn func() int,
) *ScreenStateManager {
	return &ScreenStateManager{
		firstBootActiveFn:      firstBootActiveFn,
		alarmStateFn:           alarmStateFn,
		countdownActiveFn:      countdownActiveFn,
		countdownRemainFn:      countdownRemainFn,
		countdownStartedAtFn:   countdownStartedAtFn,
		guestRequestPendingFn:  guestRequestPendingFn,
		guestRequestInfoFn:     guestRequestInfoFn,
		failsafeActiveFn:       failsafeActiveFn,
		failsafeReasonFn:       failsafeReasonFn,
		failsafeStartedAtFn:    failsafeStartedAtFn,
		failsafeEstimateSecsFn: failsafeEstimateSecsFn,
		countdownTotal:         30, // Default 30 second countdown
		lastEvaluatedMode:      ModeDisarmed,
		lastEvaluatedTime:      time.Now(),
	}
}

// EvaluateMode evaluates and returns the current screen mode
func (s *ScreenStateManager) EvaluateMode() ScreenMode {
	// Priority 1: Check blocked conditions first
	if s.firstBootActiveFn() {
		return ModeBlocked
	}
	if s.guestRequestPendingFn() {
		return ModeBlocked
	}
	if s.failsafeActiveFn() {
		return ModeBlocked
	}

	// Priority 2: Check alarm state
	alarmState := s.alarmStateFn()

	// Triggered takes precedence over other states
	if alarmState == "TRIGGERED" {
		mode := ModeTriggered
		if s.lastEvaluatedMode != mode {
			logger.Info(fmt.Sprintf("alarm: screen mode transition (%s → %s, trigger: alarm_triggered)",
				s.lastEvaluatedMode, mode))
			s.lastEvaluatedMode = mode
		}
		return mode
	}

	// Check countdown for arming state
	if s.countdownActiveFn() {
		mode := ModeArming
		if s.lastEvaluatedMode != mode {
			logger.Info(fmt.Sprintf("alarm: screen mode transition (%s → %s, trigger: countdown_active)",
				s.lastEvaluatedMode, mode))
			s.lastEvaluatedMode = mode
		}
		return mode
	}

	// Map alarm state to screen mode
	var mode ScreenMode
	switch alarmState {
	case "ARMED":
		mode = ModeArmed
	case "DISARMED":
		mode = ModeDisarmed
	default:
		// Unknown state, default to disarmed
		mode = ModeDisarmed
	}

	if s.lastEvaluatedMode != mode {
		logger.Info(fmt.Sprintf("alarm: screen mode transition (%s → %s, trigger: alarm_state:%s)",
			s.lastEvaluatedMode, mode, alarmState))
		s.lastEvaluatedMode = mode
	}
	return mode
}

// OnAlertTriggered records alert trigger info
func (s *ScreenStateManager) OnAlertTriggered(reason, location string) {
	s.triggeredAt = time.Now()
	s.triggeredReason = reason
	s.triggeredLocation = location
	logger.Info(fmt.Sprintf("alarm: triggered (reason: %s, location: %s)", reason, location))
}

// OnAlertResolved clears alert info
func (s *ScreenStateManager) OnAlertResolved() {
	s.triggeredAt = time.Time{}
	s.triggeredReason = ""
	s.triggeredLocation = ""
	logger.Info("alarm: alert resolved")
}

// OnArmed records when system was armed
func (s *ScreenStateManager) OnArmed() {
	s.armedAt = time.Now()
}

// GetScreenState returns full alarm screen state for API response
func (s *ScreenStateManager) GetScreenState() *ScreenState {
	mode := s.EvaluateMode()
	state := &ScreenState{
		Mode:      mode,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Actions:   []ActionInfo{},
		Info:      InfoContext{},
	}

	// Build response based on mode
	switch mode {
	case ModeDisarmed:
		s.buildDisarmedState(state)
	case ModeArming:
		s.buildArmingState(state)
	case ModeArmed:
		s.buildArmedState(state)
	case ModeTriggered:
		s.buildTriggeredState(state)
	case ModeBlocked:
		s.buildBlockedState(state)
	}

	return state
}

// GetSummaryState returns lightweight summary for polling
func (s *ScreenStateManager) GetSummaryState() *SummaryState {
	mode := s.EvaluateMode()
	state := &SummaryState{
		Mode:             mode,
		ActionsAvailable: 0,
	}

	// Set priority based on mode
	switch mode {
	case ModeTriggered:
		state.Priority = "critical"
	case ModeArming:
		state.Priority = "warning"
	default:
		state.Priority = "normal"
	}

	// Build summary based on mode
	switch mode {
	case ModeDisarmed:
		state.Message = "Alarm: Disarmed"
		state.Context = "Ready to arm"
		state.ActionsAvailable = 3
	case ModeArming:
		remaining := s.countdownRemainFn()
		state.Message = fmt.Sprintf("Arming in %d seconds...", remaining)
		state.Context = "Keep clear"
		state.CountdownRemainingSeconds = &remaining
		state.ActionsAvailable = 1
	case ModeArmed:
		state.Message = "Alarm: Armed"
		state.Context = "System protecting"
		state.ActionsAvailable = 3
	case ModeTriggered:
		triggeredAgo := int(time.Since(s.triggeredAt).Seconds())
		state.Message = "ALARM TRIGGERED"
		state.Context = fmt.Sprintf("Breach: %s", s.triggeredReason)
		state.TriggeredAgoSeconds = &triggeredAgo
		state.ActionsAvailable = 3
	case ModeBlocked:
		state.Message = "Alarm: Unavailable"
		state.Context = "Action required"
		state.ActionsAvailable = 0
	}

	return state
}

// buildDisarmedState populates state for Disarmed mode
func (s *ScreenStateManager) buildDisarmedState(state *ScreenState) {
	state.Message = "Alarm: Disarmed"
	state.Context = "Ready to arm when you leave"

	state.Actions = []ActionInfo{
		{ID: "arm", Label: "Arm", Enabled: true, RequiresAuth: false},
		{ID: "history", Label: "History", Enabled: true, RequiresAuth: false},
		{ID: "settings", Label: "Settings", Enabled: true, RequiresAuth: true},
	}

	state.Info.CanArm = true
	state.Info.NextAction = "Arm when leaving"
}

// buildArmingState populates state for Arming mode
func (s *ScreenStateManager) buildArmingState(state *ScreenState) {
	remaining := s.countdownRemainFn()
	state.Message = fmt.Sprintf("Arming in %d seconds...", remaining)
	state.Context = "Keep system clear until armed"

	// Build countdown info
	startedAt := s.countdownStartedAtFn()
	willCompleteAt := startedAt.Add(time.Duration(s.countdownTotal) * time.Second)

	state.Countdown = &CountdownInfo{
		TotalSeconds:     s.countdownTotal,
		RemainingSeconds: remaining,
		Percentage:       (s.countdownTotal - remaining) * 100 / s.countdownTotal,
		StartedAt:        startedAt.UTC().Format(time.RFC3339),
		WillCompleteAt:   willCompleteAt.UTC().Format(time.RFC3339),
	}

	state.Actions = []ActionInfo{
		{ID: "cancel", Label: "Cancel", Enabled: true, RequiresAuth: false},
	}

	state.Info.CanCancel = true
	state.Info.AutoArmAt = willCompleteAt.UTC().Format(time.RFC3339)
}

// buildArmedState populates state for Armed mode
func (s *ScreenStateManager) buildArmedState(state *ScreenState) {
	state.Message = "Alarm: Armed"
	state.Context = "Sensors active. System protecting."
	state.ArmedAt = s.armedAt.UTC().Format(time.RFC3339)

	state.Actions = []ActionInfo{
		{ID: "disarm", Label: "Disarm", Enabled: true, RequiresAuth: true},
		{ID: "status", Label: "Status", Enabled: true, RequiresAuth: false},
		{ID: "settings", Label: "Settings", Enabled: true, RequiresAuth: true},
	}

	state.Info.CanDisarm = true
	state.Info.SensorsActive = 5 // Placeholder
	state.Info.LastCheck = time.Now().UTC().Format(time.RFC3339)
	state.Info.ProtectionStatus = "active"
}

// buildTriggeredState populates state for Triggered mode
func (s *ScreenStateManager) buildTriggeredState(state *ScreenState) {
	state.Message = "ALARM TRIGGERED"
	state.Context = fmt.Sprintf("Breach detected: %s", s.triggeredReason)
	state.BlockReason = "" // Not a blocked mode

	state.Alert = &TriggerInfo{
		Severity:                "critical",
		Priority:                1,
		TriggeredAt:             s.triggeredAt.UTC().Format(time.RFC3339),
		TriggerReason:           s.triggeredReason,
		TriggerLocation:         s.triggeredLocation,
		TimeTriggeredAgoSeconds: int(time.Since(s.triggeredAt).Seconds()),
	}

	state.Actions = []ActionInfo{
		{ID: "disarm", Label: "Disarm", Enabled: true, RequiresAuth: true},
		{ID: "acknowledge", Label: "Acknowledge", Enabled: true, RequiresAuth: true},
		{ID: "call_support", Label: "Call Support", Enabled: true, RequiresAuth: false},
	}

	state.Info.CanDisarm = true
	state.Info.AcknowledgmentRequired = true
	state.Info.EscalationTimeRemaining = 300 // Placeholder
}

// buildBlockedState populates state for Blocked mode
func (s *ScreenStateManager) buildBlockedState(state *ScreenState) {
	state.Mode = ModeBlocked
	state.Actions = []ActionInfo{}
	state.Info.ReasonBlocked = ""

	// Determine which blocking condition and populate accordingly
	if s.firstBootActiveFn() {
		state.BlockReason = string(BlockFirstBootActive)
		state.Message = "Alarm: Setup in progress"
		state.Context = "Complete initial setup to activate alarm controls."
		state.Info.ReasonBlocked = "first_boot_active"
		state.Info.RecoveryAction = "Complete setup wizard"
		state.Actions = []ActionInfo{
			{ID: "continue_setup", Label: "Continue Setup", Enabled: true, RequiresAuth: false},
		}
		state.Info.FirstBootInfo = &FirstBootInfo{
			WizardActive:   true,
			CurrentStep:    "alarm_role",
			StepsRemaining: 2,
		}
	} else if s.guestRequestPendingFn() {
		state.BlockReason = string(BlockGuestRequestPending)
		state.Message = "Alarm: Waiting for approval"
		state.Context = "Guest entry request pending. Admin must approve or deny."
		state.Info.ReasonBlocked = "guest_request_pending"
		state.Info.RecoveryAction = "Admin approves or denies request"

		guestID, requestedAt, expiresAt := s.guestRequestInfoFn()
		timeWaiting := int(time.Since(requestedAt).Seconds())
		state.Info.GuestInfo = &GuestRequestInfo{
			GuestID:            guestID,
			RequestedAt:        requestedAt.UTC().Format(time.RFC3339),
			TimeWaitingSeconds: timeWaiting,
			ExpiresAt:          expiresAt.UTC().Format(time.RFC3339),
		}
	} else if s.failsafeActiveFn() {
		state.BlockReason = string(BlockFailsafeActive)
		state.Message = "Alarm: System recovering"
		state.Context = "System in safe mode. Alarm controls unavailable during recovery."
		state.Info.ReasonBlocked = "failsafe_active"
		state.Info.RecoveryAction = "System will recover automatically"

		estimateSecs := s.failsafeEstimateSecsFn()
		state.Info.FailsafeInfo = &FailsafeInfo{
			FailsafeActive:        true,
			Reason:                s.failsafeReasonFn(),
			StartedAt:             s.failsafeStartedAtFn().UTC().Format(time.RFC3339),
			EstimatedRecoveryTime: estimateSecs,
		}
		state.Info.EstimatedSeconds = estimateSecs
	}
}

// SetCountdownTotal sets the countdown duration (in seconds)
func (s *ScreenStateManager) SetCountdownTotal(seconds int) {
	s.countdownTotal = seconds
}
