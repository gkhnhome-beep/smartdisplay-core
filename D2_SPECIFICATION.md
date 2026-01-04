# DESIGN PHASE D2 SPECIFICATION

**Phase:** DESIGN Phase D2  
**Goal:** Define first-impression behavior of Home / Idle screen after setup completion  
**Date:** January 4, 2026  
**Status:** IN PROGRESS - DEFINING

---

## Overview

The Home / Idle screen is the primary interface for SmartDisplay. D2 defines deterministic behavioral states, data contracts, and accessibility considerations—WITHOUT visual design, CSS, or animations.

**Applies to:**
- First startup after first-boot completion
- Every normal startup of the system
- Normal user interaction throughout the day

---

## Home Screen State Machine

### States Definition

```
State Transitions:

    SetupRedirect
         ↓ (first-boot incomplete)
    
    Idle ←→ Active ←→ Alert
     ↓
    (system idle → no interaction)
```

#### State 1: SetupRedirect
**When:** System started but `wizard_completed = false`  
**Behavior:** Immediately redirect to first-boot flow  
**Duration:** Until wizard completes  
**User Action:** Complete setup or system reset  
**API Response:** `{"state": "setup_redirect", "action": "go_to_firstboot"}`

#### State 2: Idle
**When:** System running, no recent interaction, no alerts  
**Behavior:** Display calm status summary, minimal UI  
**Duration:** Until user interaction or alert triggered  
**User Action:** Touch/tap wakes to Active, or critical event triggers Alert  
**Display:** 
- Alarm state indicator
- HA connectivity status
- Time/date (localized)
- Optional AI one-liner (non-intrusive)

#### State 3: Active
**When:** User interaction detected (touch, gesture, or API call)  
**Behavior:** Show primary actions, reveal controls, expand options  
**Duration:** 5-30 minutes (configurable, reset on interaction)  
**User Action:** Back to Idle on inactivity timeout  
**Display:**
- Full status summary
- Primary action buttons (based on role)
- Expanded information
- Guest/Admin controls (if authorized)

#### State 4: Alert
**When:** Alarm triggered, critical event, or system error  
**Behavior:** Override Idle content, show alarm details  
**Duration:** Until alarm state normalizes  
**User Action:** Acknowledge alarm or resolve event  
**Priority:** Overrides all other states  
**Display:**
- Alarm type and details
- Suggested actions
- Time triggered
- Override options (if authorized)

---

## Detailed State Behaviors

### SetupRedirect State

**Trigger Condition:**
```
if !runtimeConfig.WizardCompleted {
    return StateSetupRedirect
}
```

**API Response Structure:**
```json
{
  "ok": true,
  "data": {
    "state": "setup_redirect",
    "action": "go_to_firstboot",
    "message": "Setup required"
  }
}
```

**i18n Keys:**
```
home.state.setup_redirect = "Setup required"
home.action.go_to_firstboot = "Complete setup"
```

**Logging:**
```
INFO home: setup_redirect state (wizard not completed)
```

---

### Idle State

**Trigger Condition:**
```
if firstBootComplete && 
   noRecentInteraction && 
   !alertActive {
    return StateIdle
}
```

**Purpose:** Minimal, calming display showing system health at a glance

**Display Components:**

#### 1. Alarm Status
```
Field: alarm_state
Type: string ("DISARMED", "ARMED", "TRIGGERED")
Display: Current state with visual indicator
Localization: Use i18n for state names
Example: "Alarm: Disarmed"
```

**i18n Keys:**
```
home.alarm.disarmed = "Disarmed"
home.alarm.armed = "Armed"
home.alarm.triggered = "Triggered"
home.alarm.label = "Alarm"
```

#### 2. Home Assistant Connectivity
```
Field: ha_connected
Type: boolean
Display: "Connected" (true) or "Not connected" (false)
Styling: Non-intrusive indicator
Icon: Optional checkmark or status dot
```

**i18n Keys:**
```
home.ha.connected = "Connected"
home.ha.disconnected = "Not connected"
home.ha.label = "Home Assistant"
```

#### 3. Time / Date
```
Field: current_time
Type: ISO 8601 formatted datetime
Display: Localized time and date format
Format: Respects user's language and region
Example (EN): "3:45 PM • Saturday, January 4"
Example (TR): "15:45 • Cumartesi, 4 Ocak"
```

**i18n Keys:**
```
home.time.label = "Time"
home.date.format.en = "dddd, MMMM D"  (Monday, January 1)
home.date.format.tr = "dddd, D MMMM" (Pazartesi, 1 Ocak)
```

#### 4. AI One-Liner (Optional, Non-Intrusive)
```
Field: ai_insight
Type: string (max 100 chars)
Display: Brief AI observation about home or patterns
Frequency: Not auto-cycling (respect reduced_motion)
Example: "All calm. System learned your morning routine."
```

**i18n Keys:**
```
home.ai.insight.loading = "AI monitoring..."
home.ai.insight.none = ""  (empty if no current insight)
```

**Accessibility Notes:**
- For `reduced_motion`: Do not auto-cycle insights, show one static
- For `large_text`: Use simpler phrasing, shorter insights
- For `high_contrast`: Ensure text has sufficient contrast

#### 5. Guest Status (if guest mode active)
```
Field: guest_state
Type: string ("PENDING", "APPROVED", "DENIED", "EXPIRED")
Display: Guest status with time remaining (if approved)
Localization: Use i18n for state names
```

**i18n Keys:**
```
home.guest.pending = "Guest pending"
home.guest.approved = "Guest approved"
home.guest.denied = "Guest denied"
home.guest.expired = "Guest access expired"
```

#### 6. Countdown Status (if active)
```
Field: countdown_active
Type: boolean
Field: countdown_remaining
Type: integer (seconds)
Display: "Disarm in {seconds}" if countdown active
Purpose: Show time remaining to arm/disarm
```

**i18n Keys:**
```
home.countdown.disarm_in = "Disarm in {seconds}"
home.countdown.seconds = "{count}s"
```

**API Response Structure (Idle State):**
```json
{
  "ok": true,
  "data": {
    "state": "idle",
    "system_ready": true,
    "summary": {
      "alarm_state": "DISARMED",
      "ha_connected": true,
      "current_time": "2026-01-04T15:45:00Z",
      "ai_insight": "All calm. System learned your morning routine.",
      "guest_state": null,
      "countdown_active": false,
      "countdown_remaining": 0
    },
    "message": "All systems calm"
  }
}
```

**Logging:**
```
INFO home: idle state (system ready, no interaction)
```

---

### Active State

**Trigger Condition:**
```
if idle && 
   (userInteractionDetected || apiCallReceived) {
    return StateActive
    stateTimeout = now + 5min  (default configurable)
}
```

**Purpose:** Show expanded controls and options when user is interacting

**Display Components (extends Idle):**

#### Additional Fields
```
Fields added to Idle display:
- action_buttons: []  (based on role)
- expanded_info: {}   (detailed summaries)
- guest_controls: {}  (if guest authorized)
- admin_controls: {}  (if admin authorized)
```

#### Action Buttons (Role-Based)

**For Admin/Owner:**
```
Actions:
- Arm alarm
- Disarm alarm
- Check guest requests
- View anomalies
- Settings

i18n Keys:
home.action.arm = "Arm"
home.action.disarm = "Disarm"
home.action.guest_requests = "Guest Requests"
home.action.anomalies = "Anomalies"
home.action.settings = "Settings"
```

**For Guest:**
```
Actions:
- Request entry
- View rules
- (Limited controls only)

i18n Keys:
home.action.request_entry = "Request Entry"
home.action.view_rules = "Entry Rules"
```

**For Restricted User:**
```
Actions:
- View status
- (No controls)

i18n Keys:
home.action.view_status = "View Status"
```

#### Expanded Information
```
When Active state:
- Full AI insight (not truncated)
- Device status (if HA connected)
- Recent events (last 5)
- System health indicators

Localization: All via i18n keys
Accessibility: Simplified summaries for large_text
```

#### Inactivity Timeout
```
Default: 5 minutes
Configurable: Via runtime config
On Timeout: Return to Idle state
Reset: On any user interaction
```

**i18n Keys:**
```
home.inactive_timeout = "Returning to idle view"
home.active_for = "Active for {minutes}m"
```

**API Response Structure (Active State):**
```json
{
  "ok": true,
  "data": {
    "state": "active",
    "system_ready": true,
    "summary": {
      "alarm_state": "DISARMED",
      "ha_connected": true,
      "current_time": "2026-01-04T15:45:00Z",
      "ai_insight": "All calm. System learned your morning routine. No anomalies detected.",
      "guest_state": null,
      "countdown_active": false,
      "countdown_remaining": 0
    },
    "actions": {
      "primary": [
        {"id": "arm", "label": "Arm", "enabled": true},
        {"id": "disarm", "label": "Disarm", "enabled": true}
      ],
      "secondary": [
        {"id": "guests", "label": "Guest Requests", "enabled": false},
        {"id": "anomalies", "label": "Anomalies", "enabled": true}
      ]
    },
    "expanded_info": {
      "recent_events": [...],
      "device_status": {...},
      "system_health": {...}
    },
    "message": "Ready for input"
  }
}
```

**Logging:**
```
INFO home: active state (user interaction detected)
```

---

### Alert State

**Trigger Condition:**
```
if alarmTriggered || 
   criticalEventDetected || 
   systemErrorOccurred {
    return StateAlert
    autoRevert = false  (stays until acknowledged or resolved)
}
```

**Purpose:** Override all other states, demand attention for critical events

**Priority Levels:**

#### Level 1: Alarm Triggered
```
Trigger: alarm.state == "TRIGGERED"
Display: 
  - "ALARM: {reason}"
  - Time triggered
  - Location (if known)
  - Suggested actions

i18n Keys:
home.alert.alarm_triggered = "ALARM TRIGGERED"
home.alert.triggered_at = "Triggered at {time}"
home.alert.triggered_reason = "Reason: {reason}"
home.alert.action_disarm = "Disarm"
home.alert.action_call_support = "Call Support"

Voice Hook (FAZ 81):
Speak("Alarm triggered: {reason}", priority="critical")
```

#### Level 2: Critical Event
```
Examples:
- System failsafe activated
- HA connection lost (critical state)
- Device error
- Security anomaly

Display:
  - Error/warning icon
  - Description
  - Action items

i18n Keys:
home.alert.critical_event = "Attention Required"
home.alert.event_type = "{event_type}"
```

#### Level 3: System Warning
```
Examples:
- Battery low
- Device offline
- Sensor malfunction

Display:
  - Warning indicator
  - Description
  - "Dismiss" or "Learn More"

i18n Keys:
home.alert.system_warning = "System Warning"
```

**Voice Hook Integration (Optional):**
```
On Alert state enter:
- Get alert message
- Call coord.Voice.SpeakCritical(message)
- Log voice event
```

**API Response Structure (Alert State):**
```json
{
  "ok": true,
  "data": {
    "state": "alert",
    "system_ready": false,
    "alert": {
      "priority": "critical",
      "type": "alarm_triggered",
      "message": "ALARM TRIGGERED: Door unlock detected",
      "triggered_at": "2026-01-04T15:45:32Z",
      "reason": "Door unlock detected",
      "location": "Front Door",
      "actions": [
        {"id": "disarm", "label": "Disarm", "requires_auth": true},
        {"id": "acknowledge", "label": "Acknowledge", "requires_auth": false},
        {"id": "call_support", "label": "Call Support", "requires_auth": false}
      ]
    },
    "summary": {
      "alarm_state": "TRIGGERED",
      "ha_connected": true,
      "current_time": "2026-01-04T15:45:32Z"
    },
    "message": "Critical alert"
  }
}
```

**Logging:**
```
WARN home: alert state (alarm triggered - door unlock detected)
INFO home: alert action (user_id: admin, action: disarm)
```

---

## Home Screen State Transitions

```
Startup
  ↓
Check: wizard_completed?
  ├─ NO → SetupRedirect
  │       (redirect to /api/setup/firstboot/status)
  │
  └─ YES → Idle
           (system running, no alerts)
           
Idle
  ├─ [User Interaction] → Active
  │                        (expand controls, show actions)
  │
  ├─ [Inactivity Timeout] → Idle
  │                         (after 5min, reset to idle)
  │
  └─ [Critical Event] → Alert
                        (override, show alarm/error)

Active
  ├─ [Inactivity Timeout] → Idle
  │                         (after configurable timeout)
  │
  └─ [Critical Event] → Alert
                        (override, show alarm/error)

Alert
  ├─ [Acknowledged] → Idle or Active
  │                   (depends on context)
  │
  ├─ [Resolved] → Idle
  │               (event cleared, return to calm)
  │
  └─ [User Interaction] → Alert (stays)
                         (don't dismiss alert on interaction)

SetupRedirect
  └─ [Wizard Complete] → Idle
                        (first-boot done, system ready)
```

---

## API Endpoints

### GET /api/ui/home/state

**Purpose:** Get current home screen state  
**Auth:** Optional (different data for guest/user/admin)  
**Rate Limit:** No limit (UI needs frequent polls)  
**Cache:** None (always fresh)  

**Response Structure:**
```json
{
  "ok": true,
  "data": {
    "state": "idle|active|alert|setup_redirect",
    "system_ready": boolean,
    "summary": {...},
    "alert": {...} (if state==alert),
    "actions": {...} (if state==active),
    "expanded_info": {...} (if state==active),
    "message": "Human-readable message"
  }
}
```

**Error Responses:**
```
400: Invalid request
401: Unauthorized (for restricted views)
500: System error
```

---

### GET /api/ui/home/summary

**Purpose:** Get just the summary data (lightweight for polling)  
**Auth:** Optional (different data for guest/user/admin)  
**Rate Limit:** No limit  
**Cache:** 1-2 seconds (acceptable staleness)  

**Response Structure:**
```json
{
  "ok": true,
  "data": {
    "alarm_state": "DISARMED|ARMED|TRIGGERED",
    "ha_connected": boolean,
    "current_time": "2026-01-04T15:45:00Z",
    "ai_insight": "string",
    "guest_state": null|string,
    "countdown_active": boolean,
    "countdown_remaining": integer,
    "has_pending_alerts": boolean
  }
}
```

---

## Accessibility Considerations

### For `reduced_motion` Users

**Behavior Changes:**
```
- No auto-cycling AI insights
- No animated transitions between states
- Static display of information
- Reduced flashing or blinking elements
```

**Display Changes:**
```
- Show single insight (not rotating)
- Simpler state transitions (no fade/slide)
- Clear, static layouts
- Direct information display
```

**i18n Keys:**
```
home.ai.insight.single = (show one insight, not cycling)
home.accessibility.motion_reduced = true
```

### For `large_text` Users

**Behavior Changes:**
```
- Simplified summaries (fewer words)
- Fewer fields on one screen
- Larger action buttons
- More spacing between elements
```

**Display Changes:**
```
- Use short_variant of all text
- Increase font sizes
- Simplify alerts (fewer details)
- Focus on primary information
```

**Summary Simplification:**
```
STANDARD:
"SmartDisplay is actively monitoring. Home Assistant connected. All calm."

SIMPLIFIED (large_text):
"All calm"
```

**i18n Keys:**
```
home.summary.full = (detailed summary)
home.summary.short = (simplified summary for large_text)
home.summary.for_accessibility = (use short variant if large_text==true)
```

### For `high_contrast` Users

**Display Changes:**
```
- Ensure all text meets WCAG AA contrast ratio (4.5:1)
- Use bold, clear fonts
- Avoid decorative text
- Clear visual hierarchy
```

**Color Usage:**
```
- Alarm state: Clear, distinct colors
- HA status: Green (connected) / Gray (disconnected)
- Alerts: Red / Orange with text confirmation
```

---

## Localization Integration

### Language Support
```
English (en): Primary
Turkish (tr): Secondary
```

### Date/Time Formatting
```
English: 3:45 PM • Saturday, January 4
Turkish: 15:45 • Cumartesi, 4 Ocak
```

**i18n Keys for Formatting:**
```
home.datetime.format.en = "h:mm A • dddd, MMMM D"
home.datetime.format.tr = "HH:mm • dddd, D MMMM"
```

### State Names Localization
```
English: "Disarmed", "Armed", "Triggered"
Turkish: "Silah Disarmed", "Silah Armed", "Silah Triggered"
```

---

## Logging Strategy

### Log Levels

**INFO:**
```
- State transitions (Idle → Active, etc.)
- User actions (arm, disarm, acknowledge)
- System operations (inactivity timeout)

Example:
INFO home: idle state (system ready, no interaction)
INFO home: active state (user interaction detected)
INFO home: state transition (idle → alert)
```

**WARN:**
```
- Alert states
- Critical events
- System errors

Example:
WARN home: alert state (alarm triggered - door unlock)
WARN home: setup redirect (wizard not completed)
```

**DEBUG (not in production):**
```
- API call details
- State data dumps
- Timing information
```

### What NOT to Log
```
❌ Full user data
❌ HA token or credentials
❌ Excessive polling details
❌ Every state field value
```

### Log Examples

**Transition Logging:**
```
INFO home: state transition (idle → active, trigger: user_interaction)
INFO home: state transition (active → idle, trigger: inactivity_timeout)
INFO home: state transition (idle → alert, trigger: alarm_triggered)
```

**Action Logging:**
```
INFO home: action executed (user_id: admin, action: arm, status: success)
WARN home: action blocked (user_id: guest, action: disarm, reason: unauthorized)
```

**Event Logging:**
```
WARN home: alert dismissed (alert_type: alarm, user_id: admin)
INFO home: system ready (wizard_completed: true, ha_connected: true)
```

---

## State Data Types & Contracts

### Alarm State Type
```
type: string enum
values: "DISARMED" | "ARMED" | "TRIGGERED"
required: true
source: coordinator.Alarm.CurrentState()
```

### HA Connection Type
```
type: boolean
true: Connected and operational
false: Not connected or error
required: true
source: coordinator.HA.IsConnected()
```

### Current Time Type
```
type: RFC 3339 datetime string
format: "2026-01-04T15:45:00Z"
localization: Handled by UI layer
source: time.Now().UTC().Format(time.RFC3339)
```

### AI Insight Type
```
type: string
max_length: 100 characters
required: false (can be empty)
empty_meaning: No current insight
source: coordinator.AI.GetCurrentInsight().Detail
```

### Guest State Type
```
type: string enum or null
values: "PENDING" | "APPROVED" | "DENIED" | "EXPIRED" | null
null_meaning: No active guest
source: coordinator.Guest.CurrentState()
```

### Countdown Type
```
active: boolean
remaining: integer (seconds)
meaning: Time to disarm
source: coordinator.Countdown state
```

---

## Future Considerations

### Possible Extensions (Later Phases)
1. **Device Widget** - Show primary device status
2. **Guest Queue** - Display pending guest requests
3. **Weather Integration** - Show local weather
4. **Custom Summary** - User-configurable home screen
5. **Shortcut Actions** - Quick access to common tasks
6. **Recent Events** - Expanded event list
7. **Energy Usage** - Power consumption stats
8. **Daily Summary** - AI insights from yesterday

### Not in Scope for D2
- Visual design/CSS
- Animation timing
- Widget implementation
- Custom user layouts

---

## Validation Checklist

✅ All 4 states clearly defined  
✅ State transitions documented  
✅ API contracts specified  
✅ Data types defined  
✅ Accessibility considerations included  
✅ Localization keys identified  
✅ Voice integration path documented  
✅ Logging strategy defined  
✅ Deterministic behavior guaranteed  
✅ No visual design included  

---

**Status:** SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
