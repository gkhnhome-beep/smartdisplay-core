package system

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"smartdisplay-core/internal/ai"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/firstboot"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/ha/alarmo"
	"smartdisplay-core/internal/haadapter"
	"smartdisplay-core/internal/hal"
	"smartdisplay-core/internal/hanotify"
	"smartdisplay-core/internal/home"
	"smartdisplay-core/internal/logbook"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/menu"
	"smartdisplay-core/internal/platform"
	"smartdisplay-core/internal/plugin"
	"smartdisplay-core/internal/settings"
	"strings"
	"sync"
	"time"
)

// FailsafeState tracks when system is in degraded mode
type FailsafeState struct {
	Active      bool
	Explanation string
}

// AlarmContext holds evaluated context for smart alarm scenarios
type AlarmContext struct {
	TimeOfDay      string // "night", "day", etc.
	GuestPresent   bool
	GuestState     string
	AlarmState     string
	LastAlarmEvent string
	LastTrigger    time.Time
	DeviceActive   bool
	DeviceStates   []string
}

// SelfCheckResult holds results of system self-check
type SelfCheckResult struct {
	HAConnected bool
	AlarmValid  bool
	AIRunning   bool
	Details     []string
	Hardware    []hal.DeviceHealth
}

// Coordinator manages system state and interactions between components
type Coordinator struct {
	// Core subsystems (D0-D7 design phases)
	Alarm        *alarm.StateMachine
	Guest        *guest.StateMachine
	Countdown    *countdown.Countdown
	FirstBoot    *firstboot.FirstBootManager // D0: First-boot flow manager
	Home         *home.HomeStateManager      // D2: Home screen state machine
	AlarmScreen  *alarm.ScreenStateManager   // D3: Alarm screen state exposure
	GuestScreen  *guest.ScreenStateManager   // D4: Guest access flow state machine
	GuestRequest *guest.Manager              // FAZ L2: Guest approval flow
	Menu         *menu.MenuManager           // D5: Menu structure and role-based visibility
	Logbook      *logbook.LogbookManager     // D6: History and logbook
	Settings     *settings.SettingsManager   // D7: Settings management

	// Home automation & hardware
	HA            *haadapter.Adapter
	Notifier      hanotify.Notifier
	HALRegistry   *hal.Registry
	Platform      platform.Platform
	AlarmoAdapter *alarmo.Adapter    // A2: Read-only Alarmo state
	AlarmoState   alarmo.AlarmoState // A2: Normalized alarm state (single source of truth)
	AlarmoMu      sync.RWMutex       // A2: Protect AlarmoState updates

	// AI & insights
	AI          *ai.InsightEngine
	lastInsight ai.Insight

	// Device & runtime state
	DeviceStates    []string
	Cfg             config.Config
	hardwareProfile hal.HardwareProfile

	// Internal managers
	pluginRegistry *plugin.Registry
	failsafe       FailsafeState
}

// NewCoordinator creates a new Coordinator with all subsystems
func NewCoordinator(a *alarm.StateMachine, g *guest.StateMachine, c *countdown.Countdown, ha *haadapter.Adapter, n hanotify.Notifier, halReg *hal.Registry, plat platform.Platform, haBaseURL string, haToken string) *Coordinator {
	aiEngine := ai.NewInsightEngine()
	cfg := config.Config{} // Will be populated by main

	// Initialize Alarmo adapter (read-only access to HA alarm state)
	// A2: Fetch from same HA instance as main adapter
	var alarmoAdapter *alarmo.Adapter
	if haBaseURL != "" && haToken != "" {
		alarmoAdapter = alarmo.New(haBaseURL, haToken)
	}

	// Create home state manager with dependency injection
	homeMgr := home.NewHomeStateManager(
		func() bool { return false }, // FirstBoot placeholder
		func() string { return a.CurrentState() },
		func() bool { return ha != nil && ha.IsConnected() },
		func() string {
			if aiEngine != nil {
				return aiEngine.GetCurrentInsight().Detail
			}
			return ""
		},
		func() string { return g.CurrentState() },
		func() bool { return c != nil }, // placeholder
		func() int { return 0 },         // placeholder
	)

	// Create alarm screen state manager with dependency injection (D3)
	alarmScreenMgr := alarm.NewScreenStateManager(
		func() bool { return false }, // FirstBoot placeholder
		func() string { return a.CurrentState() },
		func() bool { return c != nil && c.IsActive() }, // Countdown active check
		func() int { return c.Remaining() },             // Countdown remaining seconds
		func() time.Time { return time.Now() },          // Countdown started time (placeholder)
		func() bool { return g.HasPendingRequest() },    // Guest request pending
		func() (string, time.Time, time.Time) { // Guest request info (placeholder)
			return "", time.Now(), time.Now()
		},
		func() bool { return false },           // Failsafe active (placeholder)
		func() string { return "" },            // Failsafe reason (placeholder)
		func() time.Time { return time.Now() }, // Failsafe started (placeholder)
		func() int { return 0 },                // Failsafe estimate (placeholder)
	)

	// Create guest screen state manager with dependency injection (D4)
	guestScreenMgr := guest.NewScreenStateManager(
		func() bool { return false }, // FirstBoot placeholder
		func() string { return a.CurrentState() },
		func() time.Time { return time.Now() },
	)

	// Create menu manager with dependency injection (D5)
	menuMgr := menu.NewMenuManager(
		func() bool { return false }, // FirstBoot placeholder
		func() bool { return false }, // Failsafe placeholder
		func() bool { return false }, // Guest active placeholder
		menu.RoleAdmin,               // Default to admin role
	)

	// Create logbook manager with retention policy (D6)
	logbookMgr := logbook.NewLogbookManager(30, 90) // 30 days normal, 90 days safety

	// Create settings manager with dependency injection (D7)
	settingsMgr := settings.NewSettingsManager(
		func() (bool, error) { // getHAStatus
			if ha != nil {
				return ha.IsConnected(), nil
			}
			return false, nil
		},
		func() (string, string, string, error) { // getSystemHealth
			uptime := getSystemUptime()
			storage := getStorageInfo()
			memory := getMemoryInfo()
			return uptime, storage, memory, nil
		},
		func() string { // getVersion
			return "1.0.0" // Placeholder, should come from version.Version
		},
		func() error { // onRestart
			logger.Info("Settings: System restart initiated")
			return nil
		},
		func() (string, float64, string, error) { // onBackupCreate
			logger.Info("Settings: Backup creation initiated")
			return "backup-001", 2.4, "/data/backups/backup-001.json", nil
		},
		func(backupID string) (int, []string, error) { // onBackupRestore
			logger.Info(fmt.Sprintf("Settings: Backup restore initiated for %s", backupID))
			return 15, []string{"language: en→tr", "guest_max_active: 1→2"}, nil
		},
		func() error { // onFactoryReset
			logger.Info("Settings: Factory reset initiated")
			return nil
		},
		func(level string, message string) { // onLogEntry
			if level == "WARN" || level == "ERROR" {
				logger.Error(fmt.Sprintf("Settings [%s]: %s", level, message))
			} else {
				logger.Info(fmt.Sprintf("Settings: %s", message))
			}
		},
	)

	coord := &Coordinator{
		Alarm:          a,
		Guest:          g,
		Countdown:      c,
		HA:             ha,
		Notifier:       n,
		AI:             aiEngine,
		FirstBoot:      firstboot.New(false),               // Placeholder, will be set from runtimeCfg at startup
		Home:           homeMgr,                            // D2: Home state manager
		AlarmScreen:    alarmScreenMgr,                     // D3: Alarm screen state manager
		GuestScreen:    guestScreenMgr,                     // D4: Guest screen state manager
		GuestRequest:   guest.NewManager(60 * time.Second), // FAZ L2: Guest approval flow
		Menu:           menuMgr,                            // D5: Menu structure and role-based visibility
		Logbook:        logbookMgr,                         // D6: History and logbook
		Settings:       settingsMgr,                        // D7: Settings management
		DeviceStates:   []string{"online"},
		Cfg:            cfg,
		HALRegistry:    halReg,
		Platform:       plat,
		AlarmoAdapter:  alarmoAdapter, // A2: Alarmo adapter
		pluginRegistry: plugin.NewRegistry(),
	}

	// FAZ L3: Wire guest approval callbacks
	coord.setupGuestApprovalCallbacks()

	// A2/A3: Initialize Alarmo state from first fetch
	if coord.AlarmoAdapter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		state, err := coord.AlarmoAdapter.FetchState(ctx, alarmo.AlarmoState{})
		cancel()
		if err != nil {
			logger.Error("alarmo initial fetch failed: " + err.Error())
			coord.failsafe.Active = true
			coord.failsafe.Explanation = "Alarmo unreachable at startup"
		} else {
			coord.AlarmoState = state
			logger.Info("alarmo state loaded")
		}
	}

	logger.Info("platform detected: " + plat.Name())
	if !cfg.AIEnabled {
		logger.Info("config: AI disabled")
	}
	if !cfg.GuestAccess {
		logger.Info("config: Guest access disabled")
	}
	coord.feedAI()
	return coord
}

// === FIRST-BOOT MODE (D0) ===

// HandleGuestAction handles guest state machine actions with first-boot blocking (D0)
func (c *Coordinator) HandleGuestAction(action string) {
	logger.Info("coordinator: handling guest action")

	// First-boot mode: Block all guest actions (D0)
	if c.FirstBoot != nil && c.FirstBoot.Active() {
		logger.Info("firstboot: guest action blocked during setup")
		return
	}

	c.Guest.Handle(action)
	c.feedAI()
	c.CheckSmartAlarmScenarios()

	if action == "EXIT" {
		c.LeavingHomeDetected("guest_exit")
	}
}

// HandleAlarmAction handles alarm state machine actions with first-boot blocking (D0)
func (c *Coordinator) HandleAlarmAction(action string) {
	logger.Info("coordinator: handling alarm action")

	// First-boot mode: Block all alarm actions (D0)
	if c.FirstBoot != nil && c.FirstBoot.Active() {
		logger.Info("firstboot: alarm action blocked during setup")
		return
	}

	ctx := c.EvaluateAlarmContext()
	isTrigger := action == "TRIGGER" || action == alarm.TRIGGER
	shouldSoundSiren := true
	aiExplanation := ""

	if isTrigger && c.IsQuietHours() {
		// Only allow siren if confirmed threat
		confirmedThreat := ctx.GuestState == "DENIED" || ctx.GuestState == "EXPIRED"
		for _, ds := range ctx.DeviceStates {
			if ds == "offline" || ds == "error" {
				confirmedThreat = true
				break
			}
		}
		if !confirmedThreat {
			shouldSoundSiren = false
			aiExplanation = "Siren suppressed during quiet hours: no confirmed threat. UI/notification preferred."
		} else {
			aiExplanation = "Siren allowed during quiet hours: confirmed threat detected."
		}
	}

	if isTrigger && !shouldSoundSiren {
		logger.Info("[QUIET HOURS] Siren suppressed. Reason: " + aiExplanation)
		if c.Notifier != nil {
			c.Notifier.Notify("AlarmTriggered", map[string]interface{}{
				"reason":      aiExplanation,
				"quiet_hours": true,
			})
		}
		if c.AI != nil {
			c.AI.Observe(ctx.AlarmState, ctx.GuestState, ctx.DeviceStates...)
			c.lastInsight = c.AI.GetCurrentInsight()
		}
		return
	}

	c.Alarm.Handle(action)
	c.feedAI()
	c.CheckSmartAlarmScenarios()

	if aiExplanation != "" && c.AI != nil {
		c.lastInsight = c.AI.GetCurrentInsight()
	}
}

// === ALARM CONTEXT & SCENARIOS ===

// EvaluateAlarmContext gathers context for smart alarm scenarios
func (c *Coordinator) EvaluateAlarmContext() AlarmContext {
	ctx := AlarmContext{}

	// Time of day (night: 22:00-06:00)
	hour := time.Now().Hour()
	if hour >= 22 || hour < 6 {
		ctx.TimeOfDay = "night"
	} else {
		ctx.TimeOfDay = "day"
	}

	// Guest presence/state
	if c.Guest != nil {
		ctx.GuestState = c.Guest.CurrentState()
		ctx.GuestPresent = ctx.GuestState == "APPROVED"
	}

	// Alarm state
	if c.Alarm != nil {
		ctx.AlarmState = c.Alarm.CurrentState()
		ctx.LastAlarmEvent = c.AlarmLastEvent()
		ctx.LastTrigger = c.AlarmLastTriggerTime()
	}

	// Device activity
	ctx.DeviceStates = c.DeviceStates
	ctx.DeviceActive = false
	for _, s := range c.DeviceStates {
		if s != "offline" && s != "error" {
			ctx.DeviceActive = true
			break
		}
	}

	return ctx
}

// CheckSmartAlarmScenarios evaluates smart alarm scenarios
func (c *Coordinator) CheckSmartAlarmScenarios() {
	ctx := c.EvaluateAlarmContext()

	// False Alarm Cooldown
	const cooldown = 60
	now := time.Now()
	if ctx.AlarmState == "ARMED" && !ctx.LastTrigger.IsZero() {
		if now.Sub(ctx.LastTrigger).Seconds() < float64(cooldown) {
			msg := "False Alarm Cooldown: Alarm was recently triggered and reset. Suppressing re-trigger for cooldown period."
			logger.Info(msg)
			if c.AI != nil {
				c.AI.Observe("ARMED", ctx.GuestState, ctx.DeviceStates...)
				c.lastInsight = c.AI.GetCurrentInsight()
			}
		}
	}

	// Silent Night Entry
	if ctx.AlarmState == "ARMED" && ctx.GuestState == "APPROVED" && ctx.TimeOfDay == "night" {
		msg := "Silent Night Entry: Guest entry detected at night while alarm is armed. No auto-disarm."
		logger.Info(msg)
		if c.AI != nil {
			c.AI.Observe("ARMED", "APPROVED", ctx.DeviceStates...)
			c.lastInsight = c.AI.GetCurrentInsight()
		}
	}
}

// IsQuietHours returns true if current time is within configured quiet hours
// Note: Quiet hours not yet configurable, returns false for now
func (c *Coordinator) IsQuietHours() bool {
	// TODO: Add QuietHoursStart and QuietHoursEnd to Config when quiet hours feature is implemented
	return false
}

// AlarmLastEvent returns the last event for the alarm
// Note: StateMachine does not expose LastEvent() method currently
func (c *Coordinator) AlarmLastEvent() string {
	if c.Alarm == nil {
		return ""
	}
	// TODO: Add LastEvent() method to alarm.StateMachine if needed
	return ""
}

// AlarmLastTriggerTime returns the last time the alarm was triggered
// Note: StateMachine does not expose LastTriggerTime() method currently
func (c *Coordinator) AlarmLastTriggerTime() time.Time {
	if c.Alarm == nil {
		return time.Time{}
	}
	// TODO: Add LastTriggerTime() method to alarm.StateMachine if needed
	return time.Time{}
}

// === DEVICE & HARDWARE ===

// RegisterDevice registers a device in the HAL registry
func (c *Coordinator) RegisterDevice(device hal.Device) {
	c.HALRegistry.RegisterDevice(device)
	logger.Info("device registered: " + device.ID() + " type=" + device.Type())
}

// GetDevice retrieves a device by ID
func (c *Coordinator) GetDevice(id string) hal.Device {
	return c.HALRegistry.GetDevice(id)
}

// ListDevices lists all registered devices
func (c *Coordinator) ListDevices() []hal.Device {
	return c.HALRegistry.ListDevices()
}

// HardwareHealth returns hardware health report
func (c *Coordinator) HardwareHealth() []hal.DeviceHealth {
	return c.HALRegistry.DeviceHealthReport()
}

// SetHardwareProfile sets the active hardware profile
func (c *Coordinator) SetHardwareProfile(profile hal.HardwareProfile) {
	c.hardwareProfile = profile
	logger.Info("hardware profile: " + string(profile))
}

// GetHardwareProfile returns the active hardware profile
func (c *Coordinator) GetHardwareProfile() hal.HardwareProfile {
	return c.hardwareProfile
}

// ValidateHardwareProfile checks required HAL devices for the active profile
func (c *Coordinator) ValidateHardwareProfile() (missing []string) {
	spec, ok := hal.HardwareProfiles[c.hardwareProfile]
	if !ok {
		logger.Info("unknown hardware profile: " + string(c.hardwareProfile))
		return nil
	}
	devices := c.ListDevices()
	found := make(map[string]bool)
	for _, d := range devices {
		found[d.Type()] = true
	}
	for _, req := range spec.Required {
		if !found[req] {
			logger.Info("required hardware missing: " + req)
			missing = append(missing, req)
		}
	}
	return
}

// BootHardwareValidation initializes all registered HAL devices at startup
func (c *Coordinator) BootHardwareValidation() {
	devices := c.ListDevices()
	for _, dev := range devices {
		err := dev.Init()
		if err != nil {
			logger.Info("hardware init error: " + dev.Type() + " id=" + dev.ID() + " err=" + err.Error())
			audit.Record("hardware_fault", dev.Type()+":"+dev.ID()+":"+err.Error())
		} else if !dev.IsReady() {
			logger.Info("hardware not ready after init: " + dev.Type() + " id=" + dev.ID())
			audit.Record("hardware_fault", dev.Type()+":"+dev.ID()+":not ready after init")
		} else {
			logger.Info("hardware ready: " + dev.Type() + " id=" + dev.ID())
		}
	}
}

// === FAILSAFE MODE ===

// UpdateFailsafeState checks and updates failsafe mode status
func (c *Coordinator) UpdateFailsafeState() {
	haOffline := c.HA == nil || !c.HA.IsConnected()
	hardwareDegraded := false
	for _, dev := range c.HALRegistry.DeviceHealthReport() {
		if !dev.Ready {
			hardwareDegraded = true
			break
		}
	}
	if haOffline && hardwareDegraded {
		if !c.failsafe.Active {
			c.failsafe.Active = true
			c.failsafe.Explanation = "Failsafe Mode: Home Assistant is offline and hardware is degraded. Limited UI and manual alarm control only."
			logger.Error(c.failsafe.Explanation)
		}
	} else {
		if c.failsafe.Active {
			c.failsafe.Active = false
			c.failsafe.Explanation = "System recovered: Failsafe Mode exited."
			logger.Info(c.failsafe.Explanation)
		}
	}
}

// InFailsafeMode returns whether system is in failsafe mode
func (c *Coordinator) InFailsafeMode() bool {
	return c.failsafe.Active
}

// FailsafeExplanation returns the current failsafe explanation
func (c *Coordinator) FailsafeExplanation() string {
	return c.failsafe.Explanation
}

// DegradedMode returns true if HA is offline or hardware is missing
func (c *Coordinator) DegradedMode() bool {
	haOffline := c.HA == nil || !c.HA.IsConnected()
	hardwareMissing := false
	for _, dev := range c.HALRegistry.DeviceHealthReport() {
		if !dev.Ready {
			hardwareMissing = true
			break
		}
	}
	return haOffline || hardwareMissing
}

// StartHealthMonitor starts goroutine to monitor system health
func (c *Coordinator) StartHealthMonitor() {
	const maxDisconnects = 5
	const memCheckInterval = 6
	const memWarnThreshold = 1.5
	go func() {
		disconnects := 0
		memChecks := 0
		var lastMem uint64 = 0
		for {
			time.Sleep(10 * time.Second)
			// HA disconnect monitor
			if c.HA == nil || !c.HA.IsConnected() {
				disconnects++
				if disconnects >= maxDisconnects {
					logger.Error("HA disconnected for extended period (degraded mode)")
					disconnects = 0
				}
			} else {
				disconnects = 0
			}
			// Memory monitor
			memChecks++
			if memChecks >= memCheckInterval {
				memChecks = 0
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if lastMem > 0 && float64(m.Alloc) > float64(lastMem)*memWarnThreshold {
					logger.Error("memory usage increased unexpectedly")
				}
				lastMem = m.Alloc
			}
		}
	}()
}

// StartAlarmPolling starts a goroutine to poll Alarmo state every 2 seconds
// A2: Keep synchronized with HA Alarmo integration; stops when ctx is cancelled
func (c *Coordinator) StartAlarmPolling(ctx context.Context) {
	if c.AlarmoAdapter == nil {
		logger.Error("alarmo: adapter not initialized, polling disabled")
		return
	}

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		logger.Info("alarmo: polling started (2s interval)")

		for {
			select {
			case <-ctx.Done():
				logger.Info("alarmo: polling stopped")
				return
			case <-ticker.C:
				// Snapshot previous state to preserve countdown when HA omits attributes
				c.AlarmoMu.RLock()
				prevState := c.AlarmoState
				c.AlarmoMu.RUnlock()
				// Fetch latest state
				fetchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				newState, err := c.AlarmoAdapter.FetchState(fetchCtx, prevState)
				cancel()

				if err != nil {
					// Check if it's a 404 error (Alarmo not installed)
					if strings.Contains(err.Error(), "http 404") {
						// Alarmo not installed - log once and reduce noise
						c.AlarmoMu.Lock()
						if !c.failsafe.Active {
							c.failsafe.Active = true
							c.failsafe.Explanation = "Alarmo not installed in Home Assistant"
							logger.Info("alarmo: not found (404) - alarm features disabled")
						}
						c.AlarmoMu.Unlock()
						// Slow down polling when Alarmo is not installed (reduce log spam)
						time.Sleep(58 * time.Second) // Sleep to make it ~60s total with ticker
						continue
					}
					// Other errors - log and activate failsafe
					logger.Error("alarmo: fetch error: " + err.Error())
					c.AlarmoMu.Lock()
					if !c.failsafe.Active {
						c.failsafe.Active = true
						c.failsafe.Explanation = "Alarmo unreachable"
						logger.Error("alarmo: failsafe activated")
					}
					c.AlarmoMu.Unlock()
					continue
				}

				// Update state (thread-safe)
				c.AlarmoMu.Lock()
				if c.AlarmoState.Mode != newState.Mode || c.AlarmoState.ArmedMode != newState.ArmedMode {
					oldMode := c.AlarmoState.Mode
					oldArmed := c.AlarmoState.ArmedMode
					c.AlarmoState = newState
					logger.Info(fmt.Sprintf("alarmo state change: %s/%s -> %s/%s",
						oldMode, oldArmed, newState.Mode, newState.ArmedMode))
				} else {
					c.AlarmoState = newState // Always update for timestamp
				}

				// Clear failsafe on successful fetch
				if c.failsafe.Active {
					c.failsafe.Active = false
					logger.Info("alarmo: failsafe cleared")
				}
				c.AlarmoMu.Unlock()
			}
		}
	}()
}

// RequestAlarmAction sends a controlled arm/disarm request to Alarmo
// A4: Write operations - does NOT modify local state
// Valid actions: arm_home, arm_away, arm_night, disarm
// Returns error if request fails or validation fails
// Caller must wait for polling to reflect changes
func (c *Coordinator) RequestAlarmAction(ctx context.Context, action string) error {
	if c.AlarmoAdapter == nil {
		logger.Error("alarmo: adapter not initialized")
		return fmt.Errorf("alarmo adapter not initialized")
	}

	// Read current state for validation
	c.AlarmoMu.RLock()
	currentState := c.AlarmoState
	alarmoReachable := !c.failsafe.Active
	c.AlarmoMu.RUnlock()

	// Validation: reject if Alarmo is unreachable
	if !alarmoReachable {
		logger.Error(fmt.Sprintf("alarmo action rejected: alarmo unreachable (action=%s)", action))
		return fmt.Errorf("alarmo unreachable")
	}

	// Validation: reject arm/disarm if triggered
	if currentState.Triggered || currentState.Mode == "triggered" {
		logger.Error(fmt.Sprintf("alarmo action rejected: system triggered (action=%s)", action))
		return fmt.Errorf("action blocked: system triggered")
	}

	// Log action request (INFO level, action name only)
	logger.Info(fmt.Sprintf("alarmo action requested: %s", action))
	audit.Record("alarmo_action", action)

	// Send request to Alarmo (non-blocking, does not modify state)
	err := c.AlarmoAdapter.RequestAction(ctx, action)
	if err != nil {
		logger.Error(fmt.Sprintf("alarmo action failed: %s (error=%s)", action, err.Error()))
		audit.Record("alarmo_action_failed", action)
		return fmt.Errorf("alarmo action failed: %w", err)
	}

	logger.Info(fmt.Sprintf("alarmo action sent: %s (waiting for state change)", action))
	return nil
}

// === SELF-CHECK & DIAGNOSTICS ===

// SelfCheck runs a diagnostic check on all subsystems
func (c *Coordinator) SelfCheck() SelfCheckResult {
	var details []string

	haOk := c.HA != nil && c.HA.IsConnected()
	if haOk {
		logger.Info("self-check: HA adapter connected")
		details = append(details, "HA adapter connected")
	} else {
		logger.Info("self-check: HA adapter NOT connected")
		details = append(details, "HA adapter NOT connected")
	}

	alarmOk := c.Alarm != nil && c.Alarm.CurrentState() != ""
	if alarmOk {
		logger.Info("self-check: Alarm state valid")
		details = append(details, "Alarm state valid")
	} else {
		logger.Info("self-check: Alarm state INVALID")
		details = append(details, "Alarm state INVALID")
	}

	aiOk := c.AI != nil
	if aiOk {
		logger.Info("self-check: AI engine running")
		details = append(details, "AI engine running")
	} else {
		logger.Info("self-check: AI engine NOT running")
		details = append(details, "AI engine NOT running")
	}

	hardware := c.HALRegistry.DeviceHealthReport()
	for _, dev := range hardware {
		if dev.Error != "" {
			logger.Info("hardware error: " + dev.Type + " id=" + dev.ID + " err=" + dev.Error)
			audit.Record("hardware_fault", dev.Type+":"+dev.ID+":"+dev.Error)
			details = append(details, "hardware error: "+dev.Type+" id="+dev.ID)
		} else if !dev.Ready {
			logger.Info("hardware not ready: " + dev.Type + " id=" + dev.ID)
			audit.Record("hardware_fault", dev.Type+":"+dev.ID+":not ready")
			details = append(details, "hardware not ready: "+dev.Type+" id="+dev.ID)
		}
	}

	return SelfCheckResult{
		HAConnected: haOk,
		AlarmValid:  alarmOk,
		AIRunning:   aiOk,
		Details:     details,
		Hardware:    hardware,
	}
}

// === HA & EVENTS ===

// HandleHAEvent handles events from Home Assistant adapter
func (c *Coordinator) HandleHAEvent(event haadapter.Event) {
	logger.Info("coordinator: handling HA event")
	c.HA.HandleEvent(event)
	c.feedAI()
}

// ArrivalDetected is called when arrival is detected
func (c *Coordinator) ArrivalDetected(source, identity string) {
	logger.Info("Arrival detected via: " + source + ", identity: " + identity)
}

// LeavingHomeDetected is called when leaving home is detected
func (c *Coordinator) LeavingHomeDetected(source string) {
	logger.Info("Leaving Home detected via: " + source)
}

// === HARDWARE CONTROL ===

// EnableFanOnStartup enables a fan at startup
func (c *Coordinator) EnableFanOnStartup(fanID string) {
	fan := c.GetDevice(fanID)
	if fan == nil {
		return
	}
	out, ok := fan.(interface{ Write(any) error })
	if !ok {
		return
	}
	out.Write(map[string]any{"cmd": "on"})
	logger.Info("fan command: " + fanID + " on")
}

// FanCommand sends a fan control command
func (c *Coordinator) FanCommand(fanID string, cmd string, level int) {
	fan := c.GetDevice(fanID)
	if fan == nil {
		return
	}
	out, ok := fan.(interface{ Write(any) error })
	if !ok {
		return
	}
	switch cmd {
	case "on":
		out.Write(map[string]any{"cmd": "on"})
		logger.Info("fan command: " + fanID + " on")
	case "off":
		out.Write(map[string]any{"cmd": "off"})
		logger.Info("fan command: " + fanID + " off")
	case "set_level":
		out.Write(map[string]any{"cmd": "set_level", "level": level})
		logger.Info("fan command: " + fanID + " set_level")
	}
}

// SetLEDState sets LED state based on alarm state
func (c *Coordinator) SetLEDState(ledID string, alarmState string) {
	led := c.GetDevice(ledID)
	if led == nil {
		return
	}
	out, ok := led.(interface{ Write(any) error })
	if !ok {
		return
	}
	var color [3]uint8
	var mode string
	switch alarmState {
	case "DISARMED":
		color = [3]uint8{0, 255, 0}
		mode = "solid"
	case "ARMED":
		color = [3]uint8{255, 255, 0}
		mode = "solid"
	case "TRIGGERED":
		color = [3]uint8{255, 0, 0}
		mode = "blink"
	default:
		color = [3]uint8{0, 0, 255}
		mode = "pulse"
	}
	out.Write(map[string]any{"cmd": "set_color", "r": color[0], "g": color[1], "b": color[2]})
	out.Write(map[string]any{"cmd": "set_mode", "mode": mode})
	logger.Info("led state set: " + ledID + " mode=" + mode)
}

// === RF & RFID ===

// HandleRFEvent handles RF433 events (legacy)
func (c *Coordinator) HandleRFEvent(code string) {
	if code != "" {
		logger.Info("rf433 code: " + code)
	}
}

// HandleRF433Edges forwards raw edge patterns from RF433
func (c *Coordinator) HandleRF433Edges(id string, edges []interface{}) {
	if len(edges) == 0 {
		return
	}
	logger.Info("rf433 edges: " + id)
	audit.Record("domain_event", "remote_signal: "+id)
}

// HandleRFIDEvent handles RFID card scans
func (c *Coordinator) HandleRFIDEvent(cardID string) {
	if cardID != "" {
		logger.Info("rfid scanned: " + cardID)
		if cardID == "EXIT" {
			c.LeavingHomeDetected("rfid_exit")
		}
	}
}

// === AI & INSIGHTS ===

// feedAI feeds current state to AI engine
func (c *Coordinator) feedAI() {
	if !c.Cfg.AIEnabled {
		logger.Info("config: AI disabled, skipping insight generation")
		return
	}
	alarmState := c.Alarm.CurrentState()
	guestState := c.Guest.CurrentState()
	c.AI.Observe(alarmState, guestState, c.DeviceStates...)
	c.lastInsight = c.AI.GetCurrentInsight()
	logger.Info("ai insight: " + c.lastInsight.Detail)
}

// GetCurrentInsight returns the current AI insight
func (c *Coordinator) GetCurrentInsight() ai.Insight {
	return c.lastInsight
}

// ExplainInsight returns AI explanation of current insight
func (c *Coordinator) ExplainInsight() string {
	return c.AI.ExplainInsight()
}

// SetDeviceStates updates device states and feeds to AI
func (c *Coordinator) SetDeviceStates(states []string) {
	c.DeviceStates = states
	c.feedAI()
}

// === PLUGINS ===

// RegisterPlugin registers a plugin with the system
func (c *Coordinator) RegisterPlugin(p plugin.Plugin) error {
	if c.pluginRegistry == nil {
		return fmt.Errorf("plugin registry not initialized")
	}
	if err := c.pluginRegistry.Register(p); err != nil {
		logger.Error("failed to register plugin " + p.ID() + ": " + err.Error())
		return err
	}
	logger.Info("plugin registered: " + p.ID())
	return nil
}

// StartPlugins starts all registered plugins
func (c *Coordinator) StartPlugins() map[string]error {
	if c.pluginRegistry == nil {
		return map[string]error{"coordinator": fmt.Errorf("plugin registry not initialized")}
	}
	errs := c.pluginRegistry.StartAll()
	if len(errs) > 0 {
		logger.Error("some plugins failed to start")
	} else {
		logger.Info("all plugins started successfully")
	}
	return errs
}

// StopPlugins stops all running plugins
func (c *Coordinator) StopPlugins() map[string]error {
	if c.pluginRegistry == nil {
		return map[string]error{"coordinator": fmt.Errorf("plugin registry not initialized")}
	}
	errs := c.pluginRegistry.StopAll()
	if len(errs) > 0 {
		logger.Error("some plugins failed to stop")
	} else {
		logger.Info("all plugins stopped successfully")
	}
	return errs
}

// GetPluginStatus returns status for a specific plugin
func (c *Coordinator) GetPluginStatus(id string) (plugin.Status, error) {
	if c.pluginRegistry == nil {
		return plugin.Status{}, fmt.Errorf("plugin registry not initialized")
	}
	return c.pluginRegistry.GetStatus(id)
}

// GetAllPluginStatus returns status for all plugins
func (c *Coordinator) GetAllPluginStatus() map[string]plugin.Status {
	if c.pluginRegistry == nil {
		return make(map[string]plugin.Status)
	}
	return c.pluginRegistry.GetAllStatus()
}

// ListPlugins returns list of all registered plugin IDs
func (c *Coordinator) ListPlugins() []string {
	if c.pluginRegistry == nil {
		return []string{}
	}
	return c.pluginRegistry.List()
}

// === TRUST & AI LEARNING ===

// UserApprovedAlarmQuickly tracks quick alarm approval for trust learning
func (c *Coordinator) UserApprovedAlarmQuickly() {
	if c.AI != nil {
		c.AI.TrackQuickApproval()
	}
}

// UserCancelledAlarm tracks alarm cancellation for trust learning
func (c *Coordinator) UserCancelledAlarm() {
	if c.AI != nil {
		c.AI.TrackFrequentCancel()
	}
}

// UserIgnoredWarning tracks ignored warnings for trust learning
func (c *Coordinator) UserIgnoredWarning() {
	if c.AI != nil {
		c.AI.TrackIgnoredWarning()
	}
}

// === SETTINGS HELPERS (D7) ===

// getSystemUptime returns human-readable system uptime
func getSystemUptime() string {
	uptime := time.Now().Unix()
	hours := uptime / 3600
	if hours == 0 {
		minutes := (uptime % 3600) / 60
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh", hours)
}

// getStorageInfo returns available storage as human-readable string
func getStorageInfo() string {
	// Placeholder: should query filesystem for available space
	return "12.5 GB available"
}

// getMemoryInfo returns available memory as human-readable string
func getMemoryInfo() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	available := m.Alloc / 1024 / 1024
	return fmt.Sprintf("%d MB", available)
}

// === GUEST APPROVAL CALLBACKS (FAZ L3) ===

// setupGuestApprovalCallbacks wires guest request callbacks to HA and alarm systems
func (c *Coordinator) setupGuestApprovalCallbacks() {
	if c.GuestRequest == nil {
		logger.Error("guest request manager not initialized")
		return
	}

	// Approved callback: Disarm alarm and send HA notification
	c.GuestRequest.SetApprovedCallback(func(req *guest.GuestRequest) error {
		logger.Info("guest approval callback triggered: request_id=" + req.ID)

		// Step 1: Disarm alarm via Alarmo
		if c.AlarmoAdapter != nil {
			// Request Alarmo disarm action
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := c.AlarmoAdapter.RequestAction(ctx, "disarm")
			cancel()

			if err != nil {
				logger.Error("guest approval: alarmo disarm failed: " + err.Error())
				// Don't fail the approval - HA might be offline
			} else {
				logger.Info("guest approval: alarmo disarm requested successfully")
			}
		}

		// Step 2: Send HA feedback notification
		if c.HA != nil {
			payload := map[string]interface{}{
				"title":   "Guest Access Approved",
				"message": "Guest access granted and alarm disarmed",
			}
			if err := c.sendHANotification(req.TargetUser, payload); err != nil {
				logger.Error("guest approval: failed to send HA notification: " + err.Error())
			}
		}

		return nil
	})

	// Rejected callback: Send HA notification
	c.GuestRequest.SetRejectedCallback(func(req *guest.GuestRequest) error {
		logger.Info("guest rejection callback triggered: request_id=" + req.ID)

		// Send HA feedback notification
		if c.HA != nil {
			payload := map[string]interface{}{
				"title":   "Guest Access Denied",
				"message": "Guest access request was rejected",
			}
			if err := c.sendHANotification(req.TargetUser, payload); err != nil {
				logger.Error("guest rejection: failed to send HA notification: " + err.Error())
			}
		}

		return nil
	})

	logger.Info("guest approval callbacks wired successfully")
}

// sendHANotification sends a mobile notification to a specific HA user
func (c *Coordinator) sendHANotification(targetUser string, payload map[string]interface{}) error {
	if c.HA == nil {
		return errors.New("HA adapter not available")
	}

	// Use HA CallService to send notification
	// Service: notify.mobile_app_<device>
	// For simplicity, we'll call notify.<user> and let HA route it
	serviceName := targetUser // e.g., "mobile_app_user1" or just "user1"

	return c.HA.CallService("notify", serviceName, payload)
}
