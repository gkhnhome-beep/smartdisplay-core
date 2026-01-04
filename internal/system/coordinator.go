// Failsafe mode state
type FailsafeState struct {
	Active      bool
	Explanation string
}

// Add to Coordinator struct:
//   failsafe FailsafeState
// Trust learning: call these methods when user actions are detected
func (c *Coordinator) UserApprovedAlarmQuickly() {
	if c.AI != nil {
		c.AI.TrackQuickApproval()
	}
}
func (c *Coordinator) UserCancelledAlarm() {
	if c.AI != nil {
		c.AI.TrackFrequentCancel()
	}
}
func (c *Coordinator) UserIgnoredWarning() {
	if c.AI != nil {
		c.AI.TrackIgnoredWarning()
	}
}
// IsQuietHours returns true if the current time is within configured quiet hours
func (c *Coordinator) IsQuietHours() bool {
	start := c.Cfg.QuietHoursStart
	end := c.Cfg.QuietHoursEnd
	if start == "" || end == "" {
		return false
	}
	now := time.Now()
	parse := func(s string) (int, int) {
		var h, m int
		fmt.Sscanf(s, "%d:%d", &h, &m)
		return h, m
	}
	sh, sm := parse(start)
	teh, tem := parse(end)
	startMin := sh*60 + sm
	endMin := teh*60 + tem
	nowMin := now.Hour()*60 + now.Minute()
	if startMin < endMin {
		return nowMin >= startMin && nowMin < endMin
	}
	// Overnight (e.g., 22:00-06:00)
	return nowMin >= startMin || nowMin < endMin
}
// ArrivalDetected is called when an arrival event is detected (RFID, RF remote, HA presence)
func (c *Coordinator) ArrivalDetected(source, identity string) {
	logger.Info("Arrival detected via: " + source + ", identity: " + identity)
	// TODO: Implement arrival actions (LED, AI, alarm, guest flow)
}
// CheckSmartAlarmScenarios evaluates and logs smart alarm scenarios
func (c *Coordinator) CheckSmartAlarmScenarios() {
		// False Alarm Cooldown: If alarm was triggered and reset within cooldown period, suppress re-trigger
		cooldown := 60 // seconds
		now := time.Now()
		if ctx.AlarmState == "ARMED" && !ctx.LastTrigger.IsZero() {
			if now.Sub(ctx.LastTrigger).Seconds() < float64(cooldown) {
				msg := "False Alarm Cooldown: Alarm was recently triggered and reset. Suppressing re-trigger for cooldown period."
				logger.Info(msg)
				if c.AI != nil {
					c.AI.Observe("ARMED", ctx.GuestState, ctx.DeviceStates...)
					c.lastInsight = c.AI.GetCurrentInsight()
				}
				// Optionally, record audit or domain event here
			}
		}
	ctx := c.EvaluateAlarmContext()

	// Silent Night Entry: ARMED, guest APPROVED, night
	if ctx.AlarmState == "ARMED" && ctx.GuestState == "APPROVED" && ctx.TimeOfDay == "night" {
		msg := "Silent Night Entry: Guest entry detected at night while alarm is armed. No auto-disarm."
		logger.Info(msg)
		if c.AI != nil {
			c.AI.Observe("ARMED", "APPROVED", ctx.DeviceStates...)
			c.lastInsight = c.AI.GetCurrentInsight()
		}
		// Optionally, record audit or domain event here
	}
}
import (
	"time"
)
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

	// Device activity (any device not "offline" or "error")
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

// AlarmLastEvent returns the last event for the alarm (if tracked)
func (c *Coordinator) AlarmLastEvent() string {
	if c.Alarm == nil {
		return ""
	}
	t, ok := c.Alarm.(interface{ LastEvent() string })
	if ok {
		return t.LastEvent()
	}
	return ""
}

// AlarmLastTriggerTime returns the last time the alarm was triggered (if tracked)
func (c *Coordinator) AlarmLastTriggerTime() time.Time {
	if c.Alarm == nil {
		return time.Time{}
	}
	t, ok := c.Alarm.(interface{ LastTriggerTime() time.Time })
	if ok {
		return t.LastTriggerTime()
	}
	return time.Time{}
}
// StartHealthMonitor starts a goroutine to monitor HA connection health
func (c *Coordinator) StartHealthMonitor() {
   const maxDisconnects = 5
   const memCheckInterval = 6 // every 6 cycles (1 min)
   const memWarnThreshold = 1.5 // 50% growth triggers warning
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
				   disconnects = 0 // only log once per N
			   }
		   } else {
			   disconnects = 0
		   }
		   // Memory monitor (soft limit, not noisy)
		   memChecks++
		   if memChecks >= memCheckInterval {
			   memChecks = 0
			   var m runtime.MemStats
			   runtime.ReadMemStats(&m)
			   if lastMem > 0 && float64(m.Alloc) > float64(lastMem)*memWarnThreshold {
				   logger.Error("memory usage increased unexpectedly: " + itoa(int(m.Alloc/1024/1024)) + "MB (prev " + itoa(int(lastMem/1024/1024)) + "MB)")
			   }
			   lastMem = m.Alloc
		   }
	   }
   }()
}
// DegradedMode returns true if HA is offline or any hardware is missing
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
import (
	"smartdisplay-core/internal/hal"
)
// SetHardwareProfile sets and logs the active hardware profile.
func (c *Coordinator) SetHardwareProfile(profile hal.HardwareProfile) {
	c.hardwareProfile = profile
	logger.Info("hardware profile: " + string(profile))
}

// GetHardwareProfile returns the active hardware profile.
func (c *Coordinator) GetHardwareProfile() hal.HardwareProfile {
	return c.hardwareProfile
}
// ValidateHardwareProfile checks required HAL devices for the active profile.
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
	hardwareProfile hal.HardwareProfile
// BootHardwareValidation initializes all registered HAL devices and validates readiness at startup.
// Logs errors and continues in degraded mode if critical hardware is missing.
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

// GetHardwareReadiness returns a summary of hardware readiness for API exposure.
func (c *Coordinator) GetHardwareReadiness() []hal.DeviceHealth {
   return c.HALRegistry.DeviceHealthReport()
}
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
	   logger.Info("fan command: " + fanID + " set_level=" + itoa(level))
   }
}
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
   logger.Info("led state set: " + ledID + " color=" + itoa(int(color[0])) + "," + itoa(int(color[1])) + "," + itoa(int(color[2])) + " mode=" + mode)
}
// HandleRFEvent is legacy, use HandleRF433Edges for edge-based RF433
func (c *Coordinator) HandleRFEvent(code string) {
   if code != "" {
	   logger.Info("rf433 code: " + code)
   }
}

// HandleRF433Edges forwards raw edge patterns from RF433 GPIO device
func (c *Coordinator) HandleRF433Edges(id string, edges []interface{}) {
   if len(edges) == 0 {
	   return
   }
   logger.Info("rf433 edges: " + id + " count=" + itoa(len(edges)))
   for _, e := range edges {
	   logger.Info("rf433 edge: " + id + " event=" + fmt.Sprintf("%v", e))
	   audit.Record("hardware_event", "rf433 edge: "+id+" event="+fmt.Sprintf("%v", e))
   }
   audit.Record("domain_event", "remote_signal: "+id+" edges count="+itoa(len(edges)))
}
func (c *Coordinator) HandleRFIDEvent(cardID string) {
   if cardID != "" {
	   logger.Info("rfid scanned: " + cardID)
       // For now, treat any RFID scan as exit if cardID matches exit pattern (expand as needed)
       if cardID == "EXIT" {
           c.LeavingHomeDetected("rfid_exit")
       }
   }
}
type SelfCheckResult struct {
	HAConnected   bool
	AlarmValid    bool
	AIRunning     bool
	Details       []string
	Hardware      []hal.DeviceHealth
}

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
   hardwareFault := false
   for _, dev := range hardware {
	   if dev.Error != "" {
		   logger.Info("hardware error: " + dev.Type + " id=" + dev.ID + " err=" + dev.Error)
		   audit.Record("hardware_fault", dev.Type+":"+dev.ID+":"+dev.Error)
		   details = append(details, "hardware error: "+dev.Type+" id="+dev.ID+" err="+dev.Error)
		   hardwareFault = true
	   } else if !dev.Ready {
		   logger.Info("hardware not ready: " + dev.Type + " id=" + dev.ID)
		   audit.Record("hardware_fault", dev.Type+":"+dev.ID+":not ready")
		   details = append(details, "hardware not ready: "+dev.Type+" id="+dev.ID)
		   hardwareFault = true
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
func (c *Coordinator) HardwareHealth() []hal.DeviceHealth {
	return c.HALRegistry.DeviceHealthReport()
}
package system

import (
	"fmt"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/haadapter"
	"smartdisplay-core/internal/hanotify"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/ai"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/hal"
	"smartdisplay-core/internal/platform"
	"smartdisplay-core/internal/plugin"
)

	Alarm        *alarm.StateMachine
	Guest        *guest.StateMachine
	Countdown    *countdown.Countdown
	HA           *haadapter.Adapter
	Notifier     hanotify.Notifier
	AI           *ai.InsightEngine
	lastInsight  ai.Insight
	DeviceStates []string
	Cfg          config.Config
	HALRegistry  *hal.Registry
	Platform     platform.Platform
	failsafe     FailsafeState
	pluginRegistry *plugin.Registry
// Call this periodically or after health checks
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

func (c *Coordinator) InFailsafeMode() bool {
	return c.failsafe.Active
}

func (c *Coordinator) FailsafeExplanation() string {
	return c.failsafe.Explanation
}

// RegisterPlugin registers an internal plugin with the plugin registry.
// Plugins should be registered before StartPlugins is called.
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

// StartPlugins starts all registered plugins.
// Returns error map with plugin IDs that failed to start.
func (c *Coordinator) StartPlugins() map[string]error {
	if c.pluginRegistry == nil {
		return map[string]error{"coordinator": fmt.Errorf("plugin registry not initialized")}
	}
	errs := c.pluginRegistry.StartAll()
	if len(errs) > 0 {
		logger.Error("some plugins failed to start:")
		for id, err := range errs {
			logger.Error("  " + id + ": " + err.Error())
		}
	} else {
		logger.Info("all plugins started successfully")
	}
	return errs
}

// StopPlugins stops all running plugins.
// Returns error map with plugin IDs that failed to stop.
func (c *Coordinator) StopPlugins() map[string]error {
	if c.pluginRegistry == nil {
		return map[string]error{"coordinator": fmt.Errorf("plugin registry not initialized")}
	}
	errs := c.pluginRegistry.StopAll()
	if len(errs) > 0 {
		logger.Error("some plugins failed to stop:")
		for id, err := range errs {
			logger.Error("  " + id + ": " + err.Error())
		}
	} else {
		logger.Info("all plugins stopped successfully")
	}
	return errs
}

// GetPluginStatus returns status for a specific plugin.
func (c *Coordinator) GetPluginStatus(id string) (plugin.Status, error) {
	if c.pluginRegistry == nil {
		return plugin.Status{}, fmt.Errorf("plugin registry not initialized")
	}
	return c.pluginRegistry.GetStatus(id)
}

// GetAllPluginStatus returns status for all registered plugins.
func (c *Coordinator) GetAllPluginStatus() map[string]plugin.Status {
	if c.pluginRegistry == nil {
		return make(map[string]plugin.Status)
	}
	return c.pluginRegistry.GetAllStatus()
}

// ListPlugins returns a list of all registered plugin IDs.
func (c *Coordinator) ListPlugins() []string {
	if c.pluginRegistry == nil {
		return []string{}
	}
	return c.pluginRegistry.List()
}
}
}

func NewCoordinator(a *alarm.StateMachine, g *guest.StateMachine, c *countdown.Countdown, ha *haadapter.Adapter, n hanotify.Notifier, halReg *hal.Registry, plat platform.Platform) *Coordinator {
   aiEngine := ai.NewInsightEngine()
   coord := &Coordinator{
	   Alarm:           a,
	   Guest:           g,
	   Countdown:       c,
	   HA:              ha,
	   Notifier:        n,
	   AI:              aiEngine,
	   DeviceStates:    []string{"online"},
	   Cfg:             cfg,
	   HALRegistry:     halReg,
	   Platform:        plat,
	   pluginRegistry:  plugin.NewRegistry(),
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
func (c *Coordinator) RegisterDevice(device hal.Device) {
	c.HALRegistry.RegisterDevice(device)
	logger.Info("device registered: " + device.ID() + " type=" + device.Type())
}

func (c *Coordinator) GetDevice(id string) hal.Device {
	return c.HALRegistry.GetDevice(id)
}

func (c *Coordinator) ListDevices() []hal.Device {
	return c.HALRegistry.ListDevices()
}

func (c *Coordinator) HandleHAEvent(event haadapter.Event) {
	logger.Info("coordinator: handling HA event")
	c.HA.HandleEvent(event)
	c.feedAI()
}

func (c *Coordinator) HandleGuestAction(action string) {
	logger.Info("coordinator: handling guest action")
	c.Guest.Handle(action)
	c.feedAI()
    c.CheckSmartAlarmScenarios()
	if action == "EXIT" {
		c.LeavingHomeDetected("guest_exit")
	}
}

func (c *Coordinator) HandleAlarmAction(action string) {
	logger.Info("coordinator: handling alarm action")
	ctx := c.EvaluateAlarmContext()
	isTrigger := action == "TRIGGER" || action == alarm.TRIGGER
	shouldSoundSiren := true
	aiExplanation := ""
	if isTrigger && c.IsQuietHours() {
		// Only allow siren if confirmed threat (e.g., guest denied/expired, device anomaly)
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
		// Suppress siren, but still log and notify
		logger.Info("[QUIET HOURS] Siren suppressed. Reason: " + aiExplanation)
		if c.Notifier != nil {
			c.Notifier.Notify("Alarm triggered (quiet hours, no siren)", aiExplanation)
		}
		if c.AI != nil {
			c.AI.history = append(c.AI.history, ai.Insight{
				Type: ai.Explanation,
				Detail: aiExplanation,
				Severity: "low",
				Confidence: 1.0,
			})
			c.AI.SetSirenExplanation(aiExplanation)
			c.lastInsight = c.AI.GetCurrentInsight()
		}
		// Do not trigger siren in alarm state machine
		c.AI.Observe(ctx.AlarmState, ctx.GuestState, ctx.DeviceStates...)
		return
	}

	c.Alarm.Handle(action)
	c.feedAI()
	c.CheckSmartAlarmScenarios()
	if aiExplanation != "" && c.AI != nil {
		c.AI.history = append(c.AI.history, ai.Insight{
			Type: ai.Explanation,
			Detail: aiExplanation,
			Severity: "medium",
			Confidence: 1.0,
		})
		c.AI.SetSirenExplanation(aiExplanation)
		c.lastInsight = c.AI.GetCurrentInsight()
	}
}

func (c *Coordinator) feedAI() {
	if !c.Cfg.AIEnabled {
		logger.Info("config: AI disabled, skipping insight generation")
		return
	}
	alarmState := c.Alarm.CurrentState()
	guestState := c.Guest.CurrentState()
	c.AI.Observe(alarmState, guestState, c.DeviceStates...)
	c.lastInsight = c.AI.GetCurrentInsight()
	logger.Info("ai insight: " + string(c.lastInsight.Type) + " - " + c.lastInsight.Detail)
}
func (c *Coordinator) SetDeviceStates(states []string) {
	c.DeviceStates = states
			"smartdisplay-core/internal/audit"
	c.feedAI()
}

func (c *Coordinator) GetCurrentInsight() ai.Insight {
	return c.lastInsight
}

func (c *Coordinator) LeavingHomeDetected(source string) {
	logger.Info("Leaving Home detected via: " + source)
	// TODO: Implement security checks and AI summary
}

func (c *Coordinator) ExplainInsight() string {
	return c.AI.ExplainInsight()
}

			   audit.Record("hardware_event", "rf433 code: "+code)
			   logger.Info("domain event: remote_signal: " + code)
			   audit.Record("domain_event", "remote_signal: "+code)
