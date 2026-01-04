// Package guest provides guest access flow and screen state management.
// screen.go implements DESIGN Phase D4: Guest access flow and behavioral model.
package guest

import (
	"fmt"
	"smartdisplay-core/internal/logger"
	"time"
)

// GuestScreenState represents the guest screen display state
type GuestScreenState string

const (
	StateGuestIdle       GuestScreenState = "guest_idle"
	StateGuestRequesting GuestScreenState = "guest_requesting"
	StateGuestApproved   GuestScreenState = "guest_approved"
	StateGuestDenied     GuestScreenState = "guest_denied"
	StateGuestExpired    GuestScreenState = "guest_expired"
	StateGuestExit       GuestScreenState = "guest_exit"
)

// RequestCountdownInfo provides request timeout countdown data
type RequestCountdownInfo struct {
	TotalSeconds     int    `json:"total_seconds"`
	RemainingSeconds int    `json:"remaining_seconds"`
	Percentage       int    `json:"percentage"`
	StartedAt        string `json:"started_at"`     // RFC 3339
	WillExpireAt     string `json:"will_expire_at"` // RFC 3339
}

// ApprovalInfo provides approval details
type ApprovalInfo struct {
	ApprovedAt           string `json:"approved_at"` // RFC 3339
	ExpiresAt            string `json:"expires_at"`  // RFC 3339
	DurationMinutes      int    `json:"duration_minutes"`
	TimeRemainingSeconds int    `json:"time_remaining_seconds"`
}

// DenialInfo provides denial details
type DenialInfo struct {
	DeniedAt  string  `json:"denied_at"` // RFC 3339
	Reason    *string `json:"reason,omitempty"`
	Permanent bool    `json:"permanent"`
	CanRetry  bool    `json:"can_retry"`
}

// ExpirationInfo provides expiration details
type ExpirationInfo struct {
	ExpiredAt      string `json:"expired_at"` // RFC 3339
	Reason         string `json:"reason"`     // "timeout"
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// ExitInfo provides exit details
type ExitInfo struct {
	ExitedAt          string `json:"exited_at"` // RFC 3339
	DurationMinutes   int    `json:"duration_minutes"`
	ApprovedUntil     string `json:"approved_until"` // RFC 3339
	AlarmStatusBefore string `json:"alarm_status_before"`
	AlarmStatusNow    string `json:"alarm_status_now"`
}

// GuestActionInfo describes an action available to guest
type GuestActionInfo struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

// GuestOwnerNotification describes notification to owner
type GuestOwnerNotification struct {
	Title                string `json:"title"`
	Pending              bool   `json:"pending,omitempty"`
	Status               string `json:"status,omitempty"`
	TimeRemainingSeconds int    `json:"time_remaining_seconds,omitempty"`
	Active               bool   `json:"active,omitempty"`
	TimeRemainingMinutes int    `json:"time_remaining_minutes,omitempty"`
	VisitDurationMinutes int    `json:"visit_duration_minutes,omitempty"`
	AlarmStatus          string `json:"alarm_status,omitempty"`
}

// GuestInfoContext holds contextual information for guest screen response
type GuestInfoContext struct {
	CanRequest            bool    `json:"can_request,omitempty"`
	ReasonBlocked         string  `json:"reason_blocked,omitempty"`
	LastRequest           *string `json:"last_request,omitempty"`
	CanCancel             bool    `json:"can_cancel,omitempty"`
	ApprovalPending       bool    `json:"approval_pending,omitempty"`
	AccessActive          bool    `json:"access_active,omitempty"`
	WillExpireAt          string  `json:"will_expire_at,omitempty"`
	ApprovalDurationMin   int     `json:"approval_duration_minutes,omitempty"`
	CanRequestAgain       bool    `json:"can_request_again,omitempty"`
	CanRetry              bool    `json:"can_retry,omitempty"`
	TimeUntilRetryAllowed int     `json:"time_until_retry_allowed,omitempty"`
	VisitDurationSeconds  int     `json:"visit_duration_seconds,omitempty"`
	AlarmRestored         bool    `json:"alarm_restored,omitempty"`
}

// ScreenStateResponse represents complete guest screen state for API response
type ScreenStateResponse struct {
	State             GuestScreenState        `json:"state"`
	Message           string                  `json:"message"`
	Context           string                  `json:"context"`
	Timestamp         string                  `json:"timestamp"` // RFC 3339
	GuestID           string                  `json:"guest_id,omitempty"`
	RequestID         string                  `json:"request_id,omitempty"`
	Countdown         *RequestCountdownInfo   `json:"countdown,omitempty"`
	Approval          *ApprovalInfo           `json:"approval,omitempty"`
	Denial            *DenialInfo             `json:"denial,omitempty"`
	Expiration        *ExpirationInfo         `json:"expiration,omitempty"`
	Exit              *ExitInfo               `json:"exit,omitempty"`
	Actions           []GuestActionInfo       `json:"actions"`
	OwnerNotification *GuestOwnerNotification `json:"owner_notification,omitempty"`
	Info              GuestInfoContext        `json:"info"`
}

// SummaryResponse represents lightweight guest summary for polling
type SummaryResponse struct {
	State              GuestScreenState `json:"state"`
	Message            string           `json:"message"`
	Context            string           `json:"context"`
	CountdownRemaining *int             `json:"countdown_remaining_seconds"`
	ApprovalRemaining  *int             `json:"approval_remaining_seconds"`
	ActionsAvailable   int              `json:"actions_available"`
	Priority           string           `json:"priority"`
}

// ScreenStateManager manages guest screen state and transitions
type ScreenStateManager struct {
	// Dependency injection functions
	firstBootActiveFn   func() bool           // Check if first-boot is active
	alarmStateFn        func() string         // Get alarm state
	systemTimeFn        func() time.Time      // Get current time
	firstBootBlockingFn func() (bool, string) // Check first-boot blocking

	// State tracking
	currentState             GuestScreenState
	guestID                  string
	requestID                string
	requestStartedAt         time.Time
	approvedAt               time.Time
	approvalExpiresAt        time.Time
	deniedAt                 time.Time
	expiredAt                time.Time
	exitedAt                 time.Time
	alarmStateBeforeApproval string

	// Configuration
	requestTimeoutSeconds   int // Default 60 seconds
	approvalDurationMinutes int // Default 30 minutes

	// Tracking
	lastEvaluatedState GuestScreenState
}

// NewScreenStateManager creates a new guest screen state manager
func NewScreenStateManager(
	firstBootActiveFn func() bool,
	alarmStateFn func() string,
	systemTimeFn func() time.Time,
) *ScreenStateManager {
	return &ScreenStateManager{
		firstBootActiveFn:       firstBootActiveFn,
		alarmStateFn:            alarmStateFn,
		systemTimeFn:            systemTimeFn,
		currentState:            StateGuestIdle,
		requestTimeoutSeconds:   60,
		approvalDurationMinutes: 30,
		lastEvaluatedState:      StateGuestIdle,
	}
}

// EvaluateState evaluates and returns the current guest screen state
func (s *ScreenStateManager) EvaluateState() GuestScreenState {
	// Check blocking conditions
	if s.firstBootActiveFn() {
		return StateGuestIdle // Blocked, but stay in Idle with blocker message
	}

	// Evaluate current state and check timeouts
	now := s.systemTimeFn()

	switch s.currentState {
	case StateGuestRequesting:
		// Check if request timed out
		if now.After(s.requestStartedAt.Add(time.Duration(s.requestTimeoutSeconds) * time.Second)) {
			s.currentState = StateGuestExpired
			logger.Info(fmt.Sprintf("guest: request expired (guest_id: %s, request_id: %s)", s.guestID, s.requestID))
			s.lastEvaluatedState = StateGuestExpired
			return StateGuestExpired
		}

	case StateGuestApproved:
		// Check if approval expired
		if now.After(s.approvalExpiresAt) {
			s.currentState = StateGuestExpired
			logger.Info(fmt.Sprintf("guest: approval expired (guest_id: %s)", s.guestID))
			s.lastEvaluatedState = StateGuestExpired
			return StateGuestExpired
		}
	}

	if s.lastEvaluatedState != s.currentState {
		logger.Info(fmt.Sprintf("guest: state transition (%s → %s)", s.lastEvaluatedState, s.currentState))
		s.lastEvaluatedState = s.currentState
	}

	return s.currentState
}

// OnRequestInitiated starts a new guest access request
func (s *ScreenStateManager) OnRequestInitiated(guestID string) {
	s.guestID = guestID
	s.requestID = fmt.Sprintf("req_%d", s.systemTimeFn().Unix())
	s.requestStartedAt = s.systemTimeFn()
	s.currentState = StateGuestRequesting
	logger.Info(fmt.Sprintf("guest: request initiated (guest_id: %s, request_id: %s)", guestID, s.requestID))
}

// OnApproval approves guest access
func (s *ScreenStateManager) OnApproval(alarmState string) {
	s.alarmStateBeforeApproval = alarmState
	s.currentState = StateGuestApproved
	s.approvedAt = s.systemTimeFn()
	s.approvalExpiresAt = s.approvedAt.Add(time.Duration(s.approvalDurationMinutes) * time.Minute)
	logger.Info(fmt.Sprintf("guest: request approved (guest_id: %s, expires_at: %s)", s.guestID, s.approvalExpiresAt.Format(time.RFC3339)))
}

// OnDenial denies guest access
func (s *ScreenStateManager) OnDenial() {
	s.currentState = StateGuestDenied
	s.deniedAt = s.systemTimeFn()
	logger.Info(fmt.Sprintf("guest: request denied (guest_id: %s)", s.guestID))
}

// OnExit handles guest exiting
func (s *ScreenStateManager) OnExit() {
	s.currentState = StateGuestExit
	s.exitedAt = s.systemTimeFn()
	logger.Info(fmt.Sprintf("guest: guest exited manually (guest_id: %s)", s.guestID))
}

// ReturnToIdle resets to idle state (for new session or retry)
func (s *ScreenStateManager) ReturnToIdle() {
	if s.currentState != StateGuestIdle {
		logger.Info(fmt.Sprintf("guest: returning to idle from %s", s.currentState))
	}
	s.currentState = StateGuestIdle
	s.guestID = ""
	s.requestID = ""
	s.requestStartedAt = time.Time{}
}

// IsFirstBootBlocking returns whether first-boot is blocking access
func (s *ScreenStateManager) IsFirstBootBlocking() bool {
	return s.firstBootActiveFn()
}

// GetScreenState returns full guest screen state for API response
func (s *ScreenStateManager) GetScreenState() *ScreenStateResponse {
	state := s.EvaluateState()
	resp := &ScreenStateResponse{
		State:     state,
		Timestamp: s.systemTimeFn().UTC().Format(time.RFC3339),
		GuestID:   s.guestID,
		RequestID: s.requestID,
		Actions:   []GuestActionInfo{},
		Info:      GuestInfoContext{},
	}

	// Build response based on state
	switch state {
	case StateGuestIdle:
		s.buildIdleResponse(resp)
	case StateGuestRequesting:
		s.buildRequestingResponse(resp)
	case StateGuestApproved:
		s.buildApprovedResponse(resp)
	case StateGuestDenied:
		s.buildDeniedResponse(resp)
	case StateGuestExpired:
		s.buildExpiredResponse(resp)
	case StateGuestExit:
		s.buildExitResponse(resp)
	}

	return resp
}

// GetSummaryState returns lightweight summary for polling
func (s *ScreenStateManager) GetSummaryState() *SummaryResponse {
	state := s.EvaluateState()
	summary := &SummaryResponse{
		State:            state,
		ActionsAvailable: 0,
		Priority:         "normal",
	}

	now := s.systemTimeFn()

	switch state {
	case StateGuestIdle:
		summary.Message = "Welcome. Request access to enter?"
		summary.Context = "Your request will be sent to the property owner."
		summary.ActionsAvailable = 2
		summary.Priority = "normal"

	case StateGuestRequesting:
		remaining := int(s.requestStartedAt.Add(time.Duration(s.requestTimeoutSeconds) * time.Second).Sub(now).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		summary.Message = "Request sent to owner"
		summary.Context = fmt.Sprintf("Waiting for approval. %d seconds remaining.", remaining)
		summary.CountdownRemaining = &remaining
		summary.ActionsAvailable = 0
		summary.Priority = "warning"

	case StateGuestApproved:
		remaining := int(s.approvalExpiresAt.Sub(now).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		summary.Message = "Welcome! Access approved."
		summary.Context = fmt.Sprintf("Access active for %d more seconds.", remaining)
		summary.ApprovalRemaining = &remaining
		summary.ActionsAvailable = 3
		summary.Priority = "normal"

	case StateGuestDenied:
		summary.Message = "Access denied"
		summary.Context = "Your request has been denied. Please contact the owner for assistance."
		summary.ActionsAvailable = 3
		summary.Priority = "warning"

	case StateGuestExpired:
		summary.Message = "Request expired"
		summary.Context = "Your request was not answered within the time limit. You may try again later."
		summary.ActionsAvailable = 4
		summary.Priority = "normal"

	case StateGuestExit:
		summary.Message = "You have exited the property"
		summary.Context = "Thank you for visiting. The alarm has been re-armed."
		summary.ActionsAvailable = 2
		summary.Priority = "normal"
	}

	return summary
}

// buildIdleResponse populates response for GuestIdle state
func (s *ScreenStateManager) buildIdleResponse(resp *ScreenStateResponse) {
	if s.IsFirstBootBlocking() {
		resp.Message = "Alarm: Setup in progress"
		resp.Context = "Complete initial setup to activate guest access."
		resp.Info.ReasonBlocked = "first_boot_active"
		resp.Actions = []GuestActionInfo{}
		return
	}

	resp.Message = "Welcome. Request access to enter?"
	resp.Context = "Your request will be sent to the property owner."

	resp.Actions = []GuestActionInfo{
		{ID: "request", Label: "Request Access", Enabled: true},
		{ID: "rules", Label: "House Rules", Enabled: true},
	}

	resp.Info.CanRequest = true
	resp.Info.ReasonBlocked = ""
	resp.Info.LastRequest = nil
}

// buildRequestingResponse populates response for GuestRequesting state
func (s *ScreenStateManager) buildRequestingResponse(resp *ScreenStateResponse) {
	now := s.systemTimeFn()
	remaining := int(s.requestStartedAt.Add(time.Duration(s.requestTimeoutSeconds) * time.Second).Sub(now).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	resp.Message = "Request sent to owner"
	resp.Context = fmt.Sprintf("Waiting for approval. %d seconds remaining.", remaining)

	resp.Countdown = &RequestCountdownInfo{
		TotalSeconds:     s.requestTimeoutSeconds,
		RemainingSeconds: remaining,
		Percentage:       (s.requestTimeoutSeconds - remaining) * 100 / s.requestTimeoutSeconds,
		StartedAt:        s.requestStartedAt.UTC().Format(time.RFC3339),
		WillExpireAt:     s.requestStartedAt.Add(time.Duration(s.requestTimeoutSeconds) * time.Second).UTC().Format(time.RFC3339),
	}

	resp.Actions = []GuestActionInfo{}

	resp.OwnerNotification = &GuestOwnerNotification{
		Title:                "Guest requesting entry",
		Pending:              true,
		TimeRemainingSeconds: remaining,
	}

	resp.Info.CanCancel = false
	resp.Info.ApprovalPending = true
}

// buildApprovedResponse populates response for GuestApproved state
func (s *ScreenStateManager) buildApprovedResponse(resp *ScreenStateResponse) {
	now := s.systemTimeFn()
	remaining := int(s.approvalExpiresAt.Sub(now).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	resp.Message = "Welcome! Access approved."
	resp.Context = fmt.Sprintf("Your access is active until %s.", s.approvalExpiresAt.Format(time.Kitchen))

	resp.Approval = &ApprovalInfo{
		ApprovedAt:           s.approvedAt.UTC().Format(time.RFC3339),
		ExpiresAt:            s.approvalExpiresAt.UTC().Format(time.RFC3339),
		DurationMinutes:      s.approvalDurationMinutes,
		TimeRemainingSeconds: remaining,
	}

	resp.Actions = []GuestActionInfo{
		{ID: "exit", Label: "Exit", Enabled: true},
		{ID: "rules", Label: "House Rules", Enabled: true},
		{ID: "disarm", Label: "Disarm Alarm", Enabled: true},
	}

	resp.OwnerNotification = &GuestOwnerNotification{
		Title:                "Guest inside (approved)",
		Active:               true,
		TimeRemainingMinutes: remaining / 60,
	}

	resp.Info.AccessActive = true
	resp.Info.WillExpireAt = s.approvalExpiresAt.UTC().Format(time.RFC3339)
	resp.Info.ApprovalDurationMin = s.approvalDurationMinutes
}

// buildDeniedResponse populates response for GuestDenied state
func (s *ScreenStateManager) buildDeniedResponse(resp *ScreenStateResponse) {
	resp.Message = "Access denied"
	resp.Context = "Your request has been denied. Please contact the owner for assistance."

	resp.Denial = &DenialInfo{
		DeniedAt:  s.deniedAt.UTC().Format(time.RFC3339),
		Reason:    nil,
		Permanent: false,
		CanRetry:  false,
	}

	resp.Actions = []GuestActionInfo{
		{ID: "rules", Label: "House Rules", Enabled: true},
		{ID: "call", Label: "Call Owner", Enabled: true},
		{ID: "disconnect", Label: "Disconnect", Enabled: true},
	}

	resp.OwnerNotification = &GuestOwnerNotification{
		Title:  "Guest request denied",
		Status: "denied",
	}

	resp.Info.CanRequestAgain = false
}

// buildExpiredResponse populates response for GuestExpired state
func (s *ScreenStateManager) buildExpiredResponse(resp *ScreenStateResponse) {
	resp.Message = "Request expired"
	resp.Context = "Your request was not answered within the time limit. You may try again later."

	resp.Expiration = &ExpirationInfo{
		ExpiredAt:      s.expiredAt.UTC().Format(time.RFC3339),
		Reason:         "timeout",
		TimeoutSeconds: s.requestTimeoutSeconds,
	}

	resp.Actions = []GuestActionInfo{
		{ID: "request_again", Label: "Request Again", Enabled: true},
		{ID: "rules", Label: "House Rules", Enabled: true},
		{ID: "call", Label: "Call Owner", Enabled: true},
		{ID: "disconnect", Label: "Disconnect", Enabled: true},
	}

	resp.OwnerNotification = &GuestOwnerNotification{
		Title:  "Guest request expired (no action)",
		Status: "expired",
	}

	resp.Info.CanRetry = true
	resp.Info.TimeUntilRetryAllowed = 0
}

// buildExitResponse populates response for GuestExit state
func (s *ScreenStateManager) buildExitResponse(resp *ScreenStateResponse) {
	visitDuration := int(s.exitedAt.Sub(s.approvedAt).Seconds())

	resp.Message = "You have exited the property"
	resp.Context = "Thank you for visiting. The alarm has been re-armed."

	resp.Exit = &ExitInfo{
		ExitedAt:          s.exitedAt.UTC().Format(time.RFC3339),
		DurationMinutes:   visitDuration / 60,
		ApprovedUntil:     s.approvalExpiresAt.UTC().Format(time.RFC3339),
		AlarmStatusBefore: s.alarmStateBeforeApproval,
		AlarmStatusNow:    s.alarmStateFn(), // Current status (re-armed)
	}

	resp.Actions = []GuestActionInfo{
		{ID: "disconnect", Label: "Disconnect", Enabled: true},
		{ID: "rules", Label: "House Rules", Enabled: true},
	}

	resp.OwnerNotification = &GuestOwnerNotification{
		Title:                "Guest has exited",
		Status:               "exited",
		VisitDurationMinutes: visitDuration / 60,
		AlarmStatus:          "re-armed",
	}

	resp.Info.VisitDurationSeconds = visitDuration
	resp.Info.AlarmRestored = true
}

// SetRequestTimeoutSeconds sets the request timeout duration
func (s *ScreenStateManager) SetRequestTimeoutSeconds(seconds int) {
	s.requestTimeoutSeconds = seconds
}

// SetApprovalDurationMinutes sets the approval duration
func (s *ScreenStateManager) SetApprovalDurationMinutes(minutes int) {
	s.approvalDurationMinutes = minutes
}

// GetCurrentState returns the current guest screen state
func (s *ScreenStateManager) GetCurrentState() GuestScreenState {
	return s.EvaluateState()
}

// HasPendingRequest returns whether a request is currently pending
func (s *ScreenStateManager) HasPendingRequest() bool {
	state := s.EvaluateState()
	return state == StateGuestRequesting
}

// IsApproved returns whether guest access is currently approved
func (s *ScreenStateManager) IsApproved() bool {
	return s.EvaluateState() == StateGuestApproved
}
