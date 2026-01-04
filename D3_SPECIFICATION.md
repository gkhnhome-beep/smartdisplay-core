# DESIGN PHASE D3 SPECIFICATION
## Alarm Screen First-Impression Behavior & Data Model

**Phase:** DESIGN Phase D3  
**Focus:** Alarm screen (PRIMARY screen) first-impression behavior  
**Date:** January 4, 2026  
**Status:** SPECIFICATION IN PROGRESS

---

## Overview

The Alarm screen is the **primary user interface** of SmartDisplay. It displays:
- Current alarm state (Disarmed, Arming, Armed, Triggered, Blocked)
- Next actions (based on role and state)
- System context (why actions unavailable, what's happening)
- Instructions (calm, clear language)

This specification defines all five alarm modes, their messages, available actions, and data contracts—**without visuals, CSS, or animations**.

---

## Alarm Screen Modes

### Mode 1: Disarmed ✅

**When:** System initialized and ready, alarm not armed  
**Primary State:** `"disarmed"`  
**Default After:** Startup (after first-boot), user disarms, failsafe recovers  
**Duration:** Until user arms system or external trigger

#### Primary Message
```
EN: "Alarm: Disarmed"
TR: "Alarm: Silindi"

Tone: Calm, reassuring
Context: System is ready but not protecting
```

#### Secondary Context
```
EN: "Ready to arm when you leave"
TR: "Ayrılmadan önce silahlanmaya hazır"

Purpose: Suggest next action without being pushy
```

#### Allowed Actions (by Role)

**Admin/Owner:**
- Arm (→ Arming mode, start countdown)
- View History (past alarms)
- Settings (configure behavior)

**Guest:**
- None (display only)

**Restricted:**
- None (display only)

#### Data Structure
```json
{
  "mode": "disarmed",
  "message": "Alarm: Disarmed",
  "context": "Ready to arm when you leave",
  "timestamp": "2026-01-04T10:30:00Z",
  "actions": [
    {"id": "arm", "label": "Arm", "enabled": true, "requires_auth": false},
    {"id": "history", "label": "History", "enabled": true, "requires_auth": false},
    {"id": "settings", "label": "Settings", "enabled": true, "requires_auth": true}
  ],
  "info": {
    "can_arm": true,
    "reason_blocked": null,
    "next_action": "Arm when leaving"
  }
}
```

#### Accessibility Variants
- **reduced_motion:** No animation on entry, static display
- **large_text:** "Arm" button prominent, context in smaller text below
- **high_contrast:** Green indicator for "ready" state

#### Voice Variant (Optional FAZ 81)
```
Voice.SpeakInfo("Alarm is disarmed. Ready to arm.")
(Only if voice_enabled and this is a state change entry)
```

#### i18n Keys
```
alarm.mode.disarmed = "Alarm: Disarmed"
alarm.context.disarmed.ready = "Ready to arm when you leave"
alarm.action.arm = "Arm"
alarm.action.history = "History"
alarm.action.settings = "Settings"
alarm.info.disarmed = "System is not protecting. Arm to activate protection."
```

---

### Mode 2: Arming (Countdown) ✅

**When:** User initiates arm action, countdown in progress  
**Primary State:** `"arming"`  
**Duration:** Countdown seconds → Auto-transition to Armed mode  
**Default Countdown:** 30 seconds (configurable)  
**Cannot Reverse:** Unless explicitly allowed (TBD)

#### Primary Message
```
EN: "Arming in {seconds}..."
TR: "Silahlanıyor: {seconds}..."

Updates: Every second
Format: No panic language, clear count
```

#### Secondary Context
```
EN: "Keep system clear until armed"
TR: "Silahlanıncaya kadar sistemi açık tutun"

Purpose: Explain what's happening, action needed
```

#### Countdown Behavior

**Standard Display (all users):**
```
Remaining: 30 seconds → 29 → 28 ... → 1 → 0 (auto-arm)
Updates: Every second
Display: "{seconds} seconds remaining"
Format: Large, clear number
```

**Reduced Motion Users (FAZ 80):**
```
Display: Static countdown (not animated)
Updates: Still every second, but no animation
Format: Clear number display only
Alternative: "Arming. Please wait {seconds} seconds"
```

**Large Text Users (FAZ 80):**
```
Display: Very large countdown number
Context: Single line explanation
Actions: Disabled (cannot cancel)
```

#### Allowed Actions

**Admin/Owner:**
- Cancel (optional, TBD - may transition back to Disarmed)
- None required to proceed (auto-arm after countdown)

**Guest:**
- View Only (no actions allowed)

**Restricted:**
- View Only (no actions allowed)

#### Auto-Transitions
```
Arming (30s) → At 0s → Armed mode
OR
Arming → User Cancel (if enabled) → Disarmed
```

#### Data Structure
```json
{
  "mode": "arming",
  "message": "Arming in {seconds}...",
  "context": "Keep system clear until armed",
  "countdown": {
    "total_seconds": 30,
    "remaining_seconds": 15,
    "percentage": 50,
    "started_at": "2026-01-04T10:30:00Z",
    "will_complete_at": "2026-01-04T10:30:30Z"
  },
  "timestamp": "2026-01-04T10:30:15Z",
  "actions": [
    {"id": "cancel", "label": "Cancel", "enabled": true, "requires_auth": false}
  ],
  "info": {
    "can_cancel": true,
    "auto_arm_at": "2026-01-04T10:30:30Z",
    "time_format": "{remaining_seconds} seconds"
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display, no animation
- **large_text:** Large countdown number, minimal text
- **high_contrast:** Bold countdown display, clear border

#### i18n Keys
```
alarm.mode.arming = "Arming"
alarm.message.arming = "Arming in {seconds}..."
alarm.context.arming.clear = "Keep system clear until armed"
alarm.context.arming.waiting = "System is checking sensors..."
alarm.countdown.seconds = "{seconds} seconds remaining"
alarm.action.cancel = "Cancel"
alarm.info.arming = "System will arm automatically when countdown reaches zero."
```

---

### Mode 3: Armed ✅

**When:** Arm countdown complete, system actively protecting  
**Primary State:** `"armed"`  
**Default:** Steady state until trigger or user disarm  
**Visual Indicator:** System is ACTIVELY PROTECTING (important)

#### Primary Message
```
EN: "Alarm: Armed"
TR: "Alarm: Silahlanmış"

Tone: Confident, secure
Context: System is protecting
```

#### Secondary Context
```
EN: "Sensors active. System protecting."
TR: "Sensörler aktif. Sistem koruma altında."

Purpose: Reassure that protection is active
```

#### Allowed Actions (by Role)

**Admin/Owner:**
- Disarm (→ Disarmed mode immediately)
- View Status (sensor details)
- Settings

**Guest:**
- None (display only, awaiting entry approval)

**Restricted:**
- None (display only)

#### Data Structure
```json
{
  "mode": "armed",
  "message": "Alarm: Armed",
  "context": "Sensors active. System protecting.",
  "timestamp": "2026-01-04T10:30:30Z",
  "armed_at": "2026-01-04T10:30:30Z",
  "actions": [
    {"id": "disarm", "label": "Disarm", "enabled": true, "requires_auth": true},
    {"id": "status", "label": "Status", "enabled": true, "requires_auth": false},
    {"id": "settings", "label": "Settings", "enabled": true, "requires_auth": true}
  ],
  "info": {
    "sensors_active": 5,
    "last_check": "2026-01-04T10:30:30Z",
    "protection_status": "active",
    "can_disarm": true
  }
}
```

#### Accessibility Variants
- **reduced_motion:** No pulsing indicator, static display
- **large_text:** Prominent "Disarm" button, large mode name
- **high_contrast:** Green indicator for "armed" state, bold fonts

#### Voice Variant (Optional FAZ 81)
```
Voice.SpeakInfo("Alarm is armed. System is protecting.")
(Only if voice_enabled and this is a state change entry)
```

#### i18n Keys
```
alarm.mode.armed = "Alarm: Armed"
alarm.context.armed.active = "Sensors active. System protecting."
alarm.context.armed.secure = "System is secure and protecting."
alarm.action.disarm = "Disarm"
alarm.action.status = "Status"
alarm.action.settings = "Settings"
alarm.info.armed = "Disarm to stop protection and allow entry."
alarm.info.sensors = "{count} sensors active"
```

---

### Mode 4: Triggered ✅

**When:** Alarm detection activated (door unlock, motion, breach)  
**Primary State:** `"triggered"`  
**Duration:** Until acknowledged/resolved (no auto-reset)  
**ALERT PRIORITY:** HIGHEST - Overrides all other UI states  
**Voice Hook:** Critical event (FAZ 81 integration)

#### Primary Message
```
EN: "ALARM TRIGGERED"
TR: "ALARM AKTIVE"

Tone: Urgent but not panicked
Format: ALL CAPS (high priority)
Color: Red/Orange (high_contrast users)
```

#### Secondary Context
```
EN: "Breach detected: Door unlock"
TR: "Kesinti algılandı: Kapı açılması"

OR

EN: "Breach detected: Motion in secure area"
TR: "Kesinti algılandı: Güvenli alanda hareket"

Purpose: Explain what triggered the alarm
```

#### Alert Details
```json
{
  "trigger_reason": "door_unlock|motion_detected|sensor_breach|unknown",
  "trigger_location": "Front Door|Living Room|Back Yard",
  "triggered_at": "2026-01-04T10:45:23Z",
  "severity": "critical",
  "time_until_escalation": 300  // seconds (5 min example)
}
```

#### Allowed Actions (by Role)

**Admin/Owner:**
- Disarm (with auth)
- Acknowledge (with auth)
- Call Support

**Guest:**
- Cannot interact (locked out during alarm)

**Restricted:**
- Cannot interact (locked out during alarm)

#### Countdown to Escalation
```
Optional: System may escalate if not disarmed within time window
Display: "30 seconds until emergency contact" (if applicable)
Purpose: Warn about next action
No Panic: Clear language, not rushed tone
```

#### Data Structure
```json
{
  "mode": "triggered",
  "message": "ALARM TRIGGERED",
  "context": "Breach detected: Door unlock",
  "alert": {
    "severity": "critical",
    "priority": 1,
    "triggered_at": "2026-01-04T10:45:23Z",
    "trigger_reason": "door_unlock",
    "trigger_location": "Front Door",
    "time_triggered_ago_seconds": 5
  },
  "timestamp": "2026-01-04T10:45:28Z",
  "actions": [
    {"id": "disarm", "label": "Disarm", "enabled": true, "requires_auth": true},
    {"id": "acknowledge", "label": "Acknowledge", "enabled": true, "requires_auth": true},
    {"id": "call_support", "label": "Call Support", "enabled": true, "requires_auth": false}
  ],
  "info": {
    "can_disarm": true,
    "acknowledgment_required": true,
    "escalation_time_remaining": 300
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display (no flashing or animation)
- **large_text:** Very large "TRIGGERED" text, single button prominently displayed
- **high_contrast:** Red background, white/high-contrast text, clear button outlines

#### Voice Variant (FAZ 81 Integration)
```
Voice.SpeakCritical("Alarm triggered: Door unlock detected")

Timing: Immediately on state change
Priority: CRITICAL
No Playback: Just logging (standard library only)
```

#### i18n Keys
```
alarm.mode.triggered = "ALARM TRIGGERED"
alarm.context.triggered.door_unlock = "Breach detected: Door unlock"
alarm.context.triggered.motion = "Breach detected: Motion in secure area"
alarm.context.triggered.sensor = "Breach detected: Sensor triggered"
alarm.alert.severity.critical = "CRITICAL ALERT"
alarm.action.disarm = "Disarm"
alarm.action.acknowledge = "Acknowledge"
alarm.action.call_support = "Call Support"
alarm.info.triggered = "System detected a breach. Disarm or call support."
alarm.triggered.escalation = "Emergency contact in {seconds} seconds"
```

---

### Mode 5: Blocked ✅

**When:** Actions unavailable due to:
- Guest request pending (awaiting approval)
- First-boot active (setup not complete)
- Failsafe state (system recovering)

**Primary State:** `"blocked"`  
**Duration:** Until condition resolved  
**Purpose:** Explain why actions unavailable, not frustrate user

#### Sub-Types

##### 5a: Blocked by Guest Request
**Scenario:** Guest action pending approval (awaiting decision from admin)

```
Primary Message:
EN: "Alarm: Waiting for approval"
TR: "Alarm: Onay bekleniyor"

Context:
EN: "Guest entry request pending. Admin must approve or deny."
TR: "Konuk giriş talebine karar verilmesi bekleniyor."

Timeline: Shows how long guest has been waiting
```

**Data Structure:**
```json
{
  "mode": "blocked",
  "block_reason": "guest_request_pending",
  "message": "Alarm: Waiting for approval",
  "context": "Guest entry request pending. Admin must approve or deny.",
  "guest_info": {
    "guest_id": "guest_123",
    "requested_at": "2026-01-04T10:30:00Z",
    "time_waiting_seconds": 45,
    "expires_at": "2026-01-04T10:40:00Z"
  },
  "actions": [],
  "info": {
    "reason": "guest_request_pending",
    "recovery_action": "Admin approves or denies request"
  }
}
```

**i18n Keys:**
```
alarm.blocked.reason.guest_request = "Guest entry request pending"
alarm.blocked.guest.message = "Guest entry request pending. Admin must approve or deny."
alarm.blocked.guest.waiting = "Waiting for {time_remaining} seconds"
```

##### 5b: Blocked by First-Boot (Setup Active)
**Scenario:** System still in initial setup phase (from D0)

```
Primary Message:
EN: "Alarm: Setup in progress"
TR: "Alarm: Kurulum devam ediyor"

Context:
EN: "Complete initial setup to activate alarm controls."
TR: "Alarm kontrollerini etkinleştirmek için ilk kurulumu tamamlayın."

Action: Direct to setup completion
```

**Data Structure:**
```json
{
  "mode": "blocked",
  "block_reason": "first_boot_active",
  "message": "Alarm: Setup in progress",
  "context": "Complete initial setup to activate alarm controls.",
  "first_boot_info": {
    "wizard_active": true,
    "current_step": "alarm_role",
    "steps_remaining": 2
  },
  "actions": [
    {"id": "continue_setup", "label": "Continue Setup", "enabled": true, "requires_auth": false}
  ],
  "info": {
    "reason": "first_boot_active",
    "recovery_action": "Complete setup wizard",
    "redirect_url": "/api/setup/firstboot/status"
  }
}
```

**i18n Keys:**
```
alarm.blocked.reason.setup = "Setup in progress"
alarm.blocked.setup.message = "Complete initial setup to activate alarm controls."
alarm.blocked.setup.step = "Setup step {current}/{total}"
alarm.action.continue_setup = "Continue Setup"
```

##### 5c: Blocked by Failsafe State
**Scenario:** System in failsafe recovery mode (from FAZ 81 or system state)

```
Primary Message:
EN: "Alarm: System recovering"
TR: "Alarm: Sistem kurtarılıyor"

Context:
EN: "System in safe mode. Alarm controls unavailable during recovery."
TR: "Sistem güvenli modda. Kurtarma sırasında alarm kontrolleri kullanılamaz."

Timeline: Shows recovery progress if known
```

**Data Structure:**
```json
{
  "mode": "blocked",
  "block_reason": "failsafe_active",
  "message": "Alarm: System recovering",
  "context": "System in safe mode. Alarm controls unavailable during recovery.",
  "failsafe_info": {
    "failsafe_active": true,
    "reason": "connection_lost|power_failure|sensor_malfunction",
    "started_at": "2026-01-04T10:45:00Z",
    "estimated_recovery_time": 120
  },
  "actions": [],
  "info": {
    "reason": "failsafe_active",
    "recovery_action": "System will recover automatically",
    "estimated_time_seconds": 120
  }
}
```

**i18n Keys:**
```
alarm.blocked.reason.failsafe = "System recovering"
alarm.blocked.failsafe.message = "System in safe mode. Alarm controls unavailable during recovery."
alarm.blocked.failsafe.reason.connection = "Connection lost"
alarm.blocked.failsafe.reason.power = "Power failure"
alarm.blocked.failsafe.recovery_time = "Estimated recovery time: {seconds} seconds"
```

#### General Blocked Behavior

**All Blocked Scenarios:**
- No state-changing actions (no Arm/Disarm)
- May have recovery actions (Continue Setup, Acknowledge, etc.)
- Explain reason clearly
- Suggest next steps
- Timeout or auto-resolve when condition clears

**Accessibility for Blocked Mode:**
- **reduced_motion:** No loading animation, static message
- **large_text:** Clear explanation text, simple language
- **high_contrast:** Orange/yellow indicator, clear borders

---

## Countdown Behavior (Detailed)

### Countdown State Machine

The countdown is a sub-state within "Arming" mode:

```
Arming Mode Countdown Sequence:
30s: "Arming in 30 seconds..."
29s: "Arming in 29 seconds..."
...
2s: "Arming in 2 seconds..."
1s: "Arming in 1 second..."
0s: Auto-transition to Armed mode
```

### Countdown Data Contract

```json
{
  "countdown": {
    "total_seconds": 30,
    "remaining_seconds": 15,
    "percentage": 50,
    "started_at": "2026-01-04T10:30:00Z",
    "will_complete_at": "2026-01-04T10:30:30Z"
  }
}
```

### Countdown Display (No Animation)

**Standard Users:**
```
"Arming in 15 seconds..."
Updates: Every second via polling or server push
Format: Clear number display
```

**Reduced Motion Users (FAZ 80):**
```
"Arming. Please wait."
Display: Static text, minimal updates
Updates: Once per second, no animation
Alternative: "15 seconds remaining"
```

### Countdown Cancel (Optional)

If cancel is allowed:
```
User clicks Cancel during countdown
→ Cancel action sent to server
→ Countdown stops
→ Transition back to Disarmed mode
→ Log: "INFO alarm: arming cancelled by user"
```

---

## Triggered Behavior (Detailed)

### State Entry (First Moment)

When alarm is first triggered:

```
1. Coordinator detects trigger (from HA adapter or local sensor)
2. Sets alarm mode to "triggered"
3. Calls Voice.SpeakCritical("Alarm triggered: {reason}")
4. Logs: "WARN alarm: triggered at {timestamp}, reason: {reason}"
5. API responds with alert details
6. UI displays "ALARM TRIGGERED" prominently
```

### Alert Acknowledgment

```
User clicks Acknowledge button:
1. Send POST /api/ui/alarm/acknowledge
2. Update alarm.acknowledged = true
3. Optionally keep triggered state (no auto-dismiss)
4. Log: "INFO alarm: triggered alarm acknowledged by {user_id}"
```

### Escalation (Optional)

```
If configured with escalation_time:
- Show countdown: "Emergency contact in {seconds}"
- At 0: Optionally trigger external alert (out of scope)
- Give user time to disarm before escalation
```

### No Auto-Reset

```
Important:
- Alarm does NOT automatically reset after time passes
- Requires explicit Disarm action
- This prevents missing actual breaches
```

---

## Blocked Behavior (Detailed)

### Blocked Due to First-Boot (Most Common)

**Timeline:**
```
1. System startup
2. wizard_completed = false
3. Alarm screen checks coordinator.FirstBoot.Active()
4. Returns blocked mode with "Setup in progress"
5. User directed to /api/setup/firstboot/status
6. After setup complete, alarm screen returns normal state
```

**Message Tone:**
```
NOT: "You can't use this yet"
BUT: "Complete setup to activate alarm"

Positive, encouraging language
```

### Blocked Due to Guest Request

**Timeline:**
```
1. Guest initiates entry request
2. Coordinator creates guest request (awaiting approval)
3. Alarm screen checks coordinator.Guest.HasPendingRequest()
4. Returns blocked mode with guest info
5. Admin approves or denies
6. Alarm returns to normal state
```

**Information Shown:**
```
- Guest waiting (time_waiting)
- When request expires
- No action required from user (just blocked)
```

### Blocked Due to Failsafe

**Timeline:**
```
1. System detects critical condition (HA lost, etc.)
2. Coordinator activates failsafe
3. Alarm screen checks coordinator.Failsafe.Active()
4. Returns blocked mode with recovery info
5. System recovers, failsafe clears
6. Alarm returns to normal state
```

**Information Shown:**
```
- What happened (connection lost, power failure, etc.)
- Estimated recovery time (if known)
- No action required (system self-healing)
```

---

## API Contracts

### GET /api/ui/alarm/state

**Purpose:** Full alarm screen state with all details  
**Auth:** Required (any authenticated user)  
**Response:** Complete alarm state object

**Example Response (Disarmed):**
```json
{
  "ok": true,
  "data": {
    "mode": "disarmed",
    "message": "Alarm: Disarmed",
    "context": "Ready to arm when you leave",
    "timestamp": "2026-01-04T10:30:00Z",
    "actions": [
      {"id": "arm", "label": "Arm", "enabled": true, "requires_auth": false},
      {"id": "history", "label": "History", "enabled": true, "requires_auth": false},
      {"id": "settings", "label": "Settings", "enabled": true, "requires_auth": true}
    ],
    "info": {
      "can_arm": true,
      "reason_blocked": null,
      "next_action": "Arm when leaving"
    }
  }
}
```

**Example Response (Arming):**
```json
{
  "ok": true,
  "data": {
    "mode": "arming",
    "message": "Arming in 15 seconds...",
    "context": "Keep system clear until armed",
    "countdown": {
      "total_seconds": 30,
      "remaining_seconds": 15,
      "percentage": 50,
      "started_at": "2026-01-04T10:30:00Z",
      "will_complete_at": "2026-01-04T10:30:30Z"
    },
    "timestamp": "2026-01-04T10:30:15Z",
    "actions": [
      {"id": "cancel", "label": "Cancel", "enabled": true, "requires_auth": false}
    ],
    "info": {
      "can_cancel": true,
      "auto_arm_at": "2026-01-04T10:30:30Z"
    }
  }
}
```

**Example Response (Armed):**
```json
{
  "ok": true,
  "data": {
    "mode": "armed",
    "message": "Alarm: Armed",
    "context": "Sensors active. System protecting.",
    "timestamp": "2026-01-04T10:30:30Z",
    "armed_at": "2026-01-04T10:30:30Z",
    "actions": [
      {"id": "disarm", "label": "Disarm", "enabled": true, "requires_auth": true},
      {"id": "status", "label": "Status", "enabled": true, "requires_auth": false},
      {"id": "settings", "label": "Settings", "enabled": true, "requires_auth": true}
    ],
    "info": {
      "sensors_active": 5,
      "last_check": "2026-01-04T10:30:30Z",
      "protection_status": "active",
      "can_disarm": true
    }
  }
}
```

**Example Response (Triggered):**
```json
{
  "ok": true,
  "data": {
    "mode": "triggered",
    "message": "ALARM TRIGGERED",
    "context": "Breach detected: Door unlock",
    "alert": {
      "severity": "critical",
      "priority": 1,
      "triggered_at": "2026-01-04T10:45:23Z",
      "trigger_reason": "door_unlock",
      "trigger_location": "Front Door",
      "time_triggered_ago_seconds": 5
    },
    "timestamp": "2026-01-04T10:45:28Z",
    "actions": [
      {"id": "disarm", "label": "Disarm", "enabled": true, "requires_auth": true},
      {"id": "acknowledge", "label": "Acknowledge", "enabled": true, "requires_auth": true},
      {"id": "call_support", "label": "Call Support", "enabled": true, "requires_auth": false}
    ],
    "info": {
      "can_disarm": true,
      "acknowledgment_required": true,
      "escalation_time_remaining": 300
    }
  }
}
```

**Example Response (Blocked - First-Boot):**
```json
{
  "ok": true,
  "data": {
    "mode": "blocked",
    "block_reason": "first_boot_active",
    "message": "Alarm: Setup in progress",
    "context": "Complete initial setup to activate alarm controls.",
    "first_boot_info": {
      "wizard_active": true,
      "current_step": "alarm_role",
      "steps_remaining": 2
    },
    "actions": [
      {"id": "continue_setup", "label": "Continue Setup", "enabled": true, "requires_auth": false}
    ],
    "info": {
      "reason": "first_boot_active",
      "recovery_action": "Complete setup wizard",
      "redirect_url": "/api/setup/firstboot/status"
    }
  }
}
```

**Example Response (Blocked - Guest Request):**
```json
{
  "ok": true,
  "data": {
    "mode": "blocked",
    "block_reason": "guest_request_pending",
    "message": "Alarm: Waiting for approval",
    "context": "Guest entry request pending. Admin must approve or deny.",
    "guest_info": {
      "guest_id": "guest_123",
      "requested_at": "2026-01-04T10:30:00Z",
      "time_waiting_seconds": 45,
      "expires_at": "2026-01-04T10:40:00Z"
    },
    "actions": [],
    "info": {
      "reason": "guest_request_pending",
      "recovery_action": "Admin approves or denies request"
    }
  }
}
```

**Example Response (Blocked - Failsafe):**
```json
{
  "ok": true,
  "data": {
    "mode": "blocked",
    "block_reason": "failsafe_active",
    "message": "Alarm: System recovering",
    "context": "System in safe mode. Alarm controls unavailable during recovery.",
    "failsafe_info": {
      "failsafe_active": true,
      "reason": "connection_lost",
      "started_at": "2026-01-04T10:45:00Z",
      "estimated_recovery_time": 120
    },
    "actions": [],
    "info": {
      "reason": "failsafe_active",
      "recovery_action": "System will recover automatically",
      "estimated_time_seconds": 120
    }
  }
}
```

### GET /api/ui/alarm/summary

**Purpose:** Lightweight endpoint for frequent polling (minimal data)  
**Auth:** Required (any authenticated user)  
**Response:** Key metrics only, fast polling

**Response Structure:**
```json
{
  "ok": true,
  "data": {
    "mode": "armed|disarmed|arming|triggered|blocked",
    "message": "string (short)",
    "context": "string (short)",
    "countdown_remaining_seconds": 15,
    "triggered_ago_seconds": null,
    "actions_available": 3,
    "priority": "normal|warning|critical"
  }
}
```

**Example (Disarmed):**
```json
{
  "ok": true,
  "data": {
    "mode": "disarmed",
    "message": "Alarm: Disarmed",
    "context": "Ready to arm",
    "countdown_remaining_seconds": null,
    "triggered_ago_seconds": null,
    "actions_available": 3,
    "priority": "normal"
  }
}
```

**Example (Arming):**
```json
{
  "ok": true,
  "data": {
    "mode": "arming",
    "message": "Arming in 15 seconds...",
    "context": "Keep clear",
    "countdown_remaining_seconds": 15,
    "triggered_ago_seconds": null,
    "actions_available": 1,
    "priority": "warning"
  }
}
```

**Example (Triggered):**
```json
{
  "ok": true,
  "data": {
    "mode": "triggered",
    "message": "ALARM TRIGGERED",
    "context": "Door unlock",
    "countdown_remaining_seconds": null,
    "triggered_ago_seconds": 5,
    "actions_available": 3,
    "priority": "critical"
  }
}
```

---

## Localization Keys (i18n)

### Mode Names
```
alarm.mode.disarmed = "Alarm: Disarmed"
alarm.mode.armed = "Alarm: Armed"
alarm.mode.arming = "Alarm: Arming"
alarm.mode.triggered = "ALARM TRIGGERED"
alarm.mode.blocked = "Alarm: Unavailable"
```

### Messages (Primary)
```
alarm.message.disarmed = "Alarm: Disarmed"
alarm.message.armed = "Alarm: Armed"
alarm.message.arming = "Arming in {seconds}..."
alarm.message.triggered = "ALARM TRIGGERED"
alarm.message.triggered.critical = "CRITICAL ALERT"
alarm.message.blocked.setup = "Alarm: Setup in progress"
alarm.message.blocked.guest = "Alarm: Waiting for approval"
alarm.message.blocked.failsafe = "Alarm: System recovering"
```

### Context (Secondary Messages)
```
alarm.context.disarmed = "Ready to arm when you leave"
alarm.context.disarmed.ready = "Ready to arm when you leave"
alarm.context.disarmed.info = "System is not protecting. Arm to activate protection."

alarm.context.armed = "Sensors active. System protecting."
alarm.context.armed.active = "Sensors active. System protecting."
alarm.context.armed.secure = "System is secure and protecting."
alarm.context.armed.info = "Disarm to stop protection and allow entry."

alarm.context.arming = "Keep system clear until armed"
alarm.context.arming.clear = "Keep system clear until armed"
alarm.context.arming.waiting = "System is checking sensors..."
alarm.context.arming.info = "System will arm automatically when countdown reaches zero."

alarm.context.triggered = "Breach detected: {reason}"
alarm.context.triggered.door = "Breach detected: Door unlock"
alarm.context.triggered.motion = "Breach detected: Motion in secure area"
alarm.context.triggered.sensor = "Breach detected: Sensor triggered"
alarm.context.triggered.info = "System detected a breach. Disarm or call support."

alarm.context.blocked.setup = "Complete initial setup to activate alarm controls."
alarm.context.blocked.guest = "Guest entry request pending. Admin must approve or deny."
alarm.context.blocked.failsafe = "System in safe mode. Alarm controls unavailable during recovery."
```

### Countdown
```
alarm.countdown.seconds = "{seconds} seconds remaining"
alarm.countdown.second = "1 second remaining"
alarm.countdown.zero = "0 seconds - Arming now"
alarm.countdown.warning = "System will arm automatically"
```

### Actions
```
alarm.action.arm = "Arm"
alarm.action.disarm = "Disarm"
alarm.action.cancel = "Cancel"
alarm.action.acknowledge = "Acknowledge"
alarm.action.call_support = "Call Support"
alarm.action.history = "History"
alarm.action.status = "Status"
alarm.action.settings = "Settings"
alarm.action.continue_setup = "Continue Setup"
```

### Info Messages
```
alarm.info.disarmed = "System is not protecting. Arm to activate protection."
alarm.info.armed = "Disarm to stop protection and allow entry."
alarm.info.arming = "System will arm automatically when countdown reaches zero."
alarm.info.triggered = "System detected a breach. Disarm or call support."
alarm.info.blocked = "Alarm unavailable until condition resolves."
alarm.info.sensors = "{count} sensors active"
alarm.info.guest_waiting = "Guest waiting for {time_remaining} seconds"
alarm.info.escalation = "Emergency contact in {seconds} seconds"
```

### Blocked-Specific Keys
```
alarm.blocked.reason.setup = "Setup in progress"
alarm.blocked.reason.guest_request = "Guest entry request pending"
alarm.blocked.reason.failsafe = "System recovering"

alarm.blocked.setup.message = "Complete initial setup to activate alarm controls."
alarm.blocked.setup.step = "Setup step {current}/{total}"

alarm.blocked.guest.message = "Guest entry request pending. Admin must approve or deny."
alarm.blocked.guest.waiting = "Waiting for {time_remaining} seconds"

alarm.blocked.failsafe.message = "System in safe mode. Alarm controls unavailable during recovery."
alarm.blocked.failsafe.reason.connection = "Connection lost"
alarm.blocked.failsafe.reason.power = "Power failure"
alarm.blocked.failsafe.recovery_time = "Estimated recovery time: {seconds} seconds"
```

**Total i18n Keys:** ~80+ organized by component

---

## Accessibility Integration

### For `reduced_motion` Users (FAZ 80)

**Countdown Behavior:**
```
❌ NO: Animated counting animation
✅ YES: Static number display
Display: "Arming. 15 seconds remaining" (no animation)
Updates: Still every second, just no visual animation
```

**Idle Animation:**
```
❌ NO: Pulsing indicator, animated borders
✅ YES: Static display with clear indicator
Example: Simple text or emoji (no animation)
```

**Transition Animation:**
```
❌ NO: Sliding or fading transitions between states
✅ YES: Instant display change
```

### For `large_text` Users (FAZ 80)

**Font Sizes:**
```
Mode Name: 32pt (very large)
Message: 24pt (large)
Context: 16pt (readable)
Actions: 18pt+ (easy to tap/click)
```

**Simplified Text:**
```
Standard: "Alarm: Armed. Sensors active. System protecting."
Large Text: "Armed. Sensors on."

OR even simpler in some modes:
Standard: "Arming in 15 seconds. Keep system clear until armed."
Large Text: "Arming. Wait."
```

**Spacing:**
```
More whitespace between elements
Clear separations between sections
No text wrapping issues
```

### For `high_contrast` Users (FAZ 80)

**Color Scheme:**
```
Disarmed: Green indicator (ready)
Arming: Yellow/Orange indicator (in progress)
Armed: Green indicator (protecting)
Triggered: Red indicator (critical)
Blocked: Gray indicator (unavailable)

WCAG AA Compliant: 4.5:1 minimum contrast
```

**Bold/Clear Text:**
```
Use bold fonts for mode names
Clear borders around action buttons
High contrast between background and text
No subtle colors or gradients
```

---

## Logging Strategy

### INFO Level (Normal Operations)

```
INFO alarm: mode changed from {old_mode} to {new_mode}
Example: INFO alarm: mode changed from disarmed to arming

INFO alarm: countdown started (duration: 30 seconds)
INFO alarm: countdown ended, mode transition to armed
INFO alarm: action executed by {user_id} (action: {action_id}, status: success)
Example: INFO alarm: action executed by admin_user (action: arm, status: success)

INFO alarm: mode: {current_mode}
Example: INFO alarm: mode: armed
```

### WARN Level (Alerts & Blocked States)

```
WARN alarm: triggered at {timestamp} (reason: {reason}, location: {location})
Example: WARN alarm: triggered at 2026-01-04T10:45:23Z (reason: door_unlock, location: Front Door)

WARN alarm: action blocked (action: {action_id}, user_id: {user_id}, reason: {reason})
Example: WARN alarm: action blocked (action: disarm, user_id: guest_123, reason: insufficient_auth)

WARN alarm: state blocked (reason: {block_reason})
Example: WARN alarm: state blocked (reason: failsafe_active)
```

### What NOT to Log

```
❌ Alarm codes or secret disarm sequences
❌ User passwords or authentication tokens
❌ Full HA configuration details
❌ Every API polling request (too noisy)
❌ Internal counter values (percentage, remaining_ms)
❌ Sensor details or raw HA data
```

### Log Examples

**Startup (Normal):**
```
INFO alarm: initialized in disarmed mode
INFO alarm: mode: disarmed
```

**Arm Sequence:**
```
INFO alarm: action executed by admin_user (action: arm, status: success)
INFO alarm: mode changed from disarmed to arming
INFO alarm: countdown started (duration: 30 seconds)
INFO alarm: countdown ended, mode transition to armed
INFO alarm: mode: armed
```

**Trigger Event:**
```
WARN alarm: triggered at 2026-01-04T10:45:23Z (reason: door_unlock, location: Front Door)
INFO alarm: mode changed from armed to triggered
INFO voice: spoke critical "Alarm triggered: Door unlock detected"
WARN alarm: action blocked (action: disarm, user_id: guest_123, reason: insufficient_auth)
```

**Blocked State (First-Boot):**
```
INFO alarm: mode: blocked
INFO alarm: state blocked (reason: first_boot_active)
WARN alarm: action blocked (action: arm, user_id: admin, reason: first_boot_active)
```

---

## State Transition Diagram

```
                    [Startup]
                        |
                        v
                  [Disarmed] ←──────────────────┐
                   |     ↑                       |
                   |     |                       |
              [Arm] |     | [Disarm/Cancel]      |
                   |     |                       |
                   v     └──────────────────── [Failsafe Recovery]
                [Arming]
                   |
              [Timer=0]
                   |
                   v
                [Armed] ────────┐
                   |            |
              [Trigger]    [Disarm]
                   |            |
                   v            v
              [Triggered]  [Disarmed]


BLOCKED Substates:
  [Disarmed/Armed/Triggered] ←→ [Blocked-FirstBoot]
  [Disarmed/Armed/Triggered] ←→ [Blocked-GuestRequest]
  [Disarmed/Armed/Triggered] ←→ [Blocked-Failsafe]

Transitions:
  Disarmed --[Arm]--> Arming
  Arming --[Timer=0]--> Armed
  Arming --[Cancel]--> Disarmed
  Armed --[Trigger]--> Triggered
  Armed --[Disarm]--> Disarmed
  Triggered --[Disarm]--> Disarmed
  
  Any --> Blocked-FirstBoot (if first-boot active)
  Any --> Blocked-GuestRequest (if guest pending)
  Any --> Blocked-Failsafe (if failsafe active)
  
  Blocked --> Previous (when condition resolves)
```

---

## Key Design Decisions

### 1. Five Distinct Modes
**Decision:** Separate Disarmed, Arming, Armed, Triggered, Blocked  
**Rationale:** Each has different UX needs; no overlap  
**Alternative Rejected:** Combine Arming into Armed (too ambiguous)

### 2. Countdown as Sub-State
**Decision:** Countdown is part of Arming, not separate mode  
**Rationale:** Clear progression toward Armed state  
**Alternative Rejected:** Countdown as separate state (confusing)

### 3. Triggered Is Highest Priority
**Decision:** Triggered overrides all other states, no suppression  
**Rationale:** Safety—alarm cannot be hidden  
**Alternative Rejected:** Can dismiss triggered (unsafe)

### 4. Blocked Is Separate Category
**Decision:** First-boot/guest/failsafe are "blocked" modes, not state variants  
**Rationale:** Clear separation of why actions unavailable  
**Alternative Rejected:** Merge with other states (confusing)

### 5. No Disarm Logic
**Decision:** D3 defines data/UI only, not actual disarm mechanism  
**Rationale:** Scope compliance; actual disarm is separate (future phase)  
**Note:** D3 shows "Disarm" button but doesn't implement it

### 6. Calm Language for Triggered
**Decision:** Use "Breach detected" not "INTRUDER DETECTED"  
**Rationale:** Professional, calm tone; respects Product Principles  
**Alternative Rejected:** Panic language (violates Product Principles)

### 7. Separate Summary Endpoint
**Decision:** GET /api/ui/alarm/summary distinct from state  
**Rationale:** UI can poll frequently without full payload  
**Cache:** Acceptable 1-2 sec staleness

### 8. Voice Hook Integration
**Decision:** D3 defines voice contract, FAZ 81 provides execution  
**Rationale:** Scope compliance; backend-driven only  
**No Playback:** Voice logging only (standard library)

---

## Design Principles Applied

**From SmartDisplay Product Principles:**

✅ **Calm:** No panic language, clear instructions  
✅ **Predictable:** Deterministic state machine, no surprises  
✅ **Respectful:** Blocked states explain why, not frustrate  
✅ **Protective:** Triggered state highest priority, can't suppress  
✅ **Accessible:** Variants for reduced_motion, large_text, high_contrast (FAZ 80)  
✅ **Localized:** 80+ i18n keys for English and Turkish (FAZ 79)  
✅ **Voice-Ready:** Optional voice integration defined (FAZ 81)  
✅ **Setup-Aware:** First-boot blocking enforced (D0)  

---

## Integration with Previous Phases

### FAZ 79 (Localization)
```
All text through alarm.* i18n key namespace
80+ keys defined for English source
Turkish localization strategy documented
```

### FAZ 80 (Accessibility)
```
reduced_motion: Static display, minimal updates
large_text: Simplified summaries, larger fonts
high_contrast: Clear colors, WCAG AA contrast
```

### FAZ 81 (Voice Feedback)
```
Triggered state calls Voice.SpeakCritical()
Disarmed/Armed states call Voice.SpeakInfo()
Optional voice variants for all modes
```

### D0 (First-Boot Flow)
```
Blocked mode if wizard_completed = false
Redirect to setup if first-boot active
```

### D1 (First-Boot Copy)
```
Uses system message patterns from D1
Follows tone guidelines (Calm, Respectful, etc.)
```

### D2 (Home Screen)
```
Alert state in D2 references alarm triggered state
Home screen shows alarm summary
```

---

## Testing Checklist

✅ All 5 modes defined (Disarmed, Arming, Armed, Triggered, Blocked)  
✅ Blocked sub-types documented (First-Boot, Guest, Failsafe)  
✅ Primary messages defined for each mode  
✅ Secondary context provided  
✅ Allowed actions specified by role  
✅ Countdown behavior fully specified  
✅ Triggered behavior with escalation path  
✅ Blocked behavior with recovery actions  
✅ API endpoints documented with full contracts  
✅ Example JSON responses for all modes  
✅ 80+ i18n keys organized  
✅ Accessibility variants for all 3 preferences  
✅ Voice integration path clear  
✅ Logging strategy defined (INFO/WARN levels)  
✅ State transition diagram complete  
✅ Design decisions documented  
✅ Product Principles validated  
✅ Integration with FAZ 79, 80, 81, D0, D1, D2 confirmed  

---

## Next Implementation Steps

### Phase 1: Coordinator Integration
- Add alarm mode tracking to Coordinator
- Implement state machine logic
- Add countdown timer management

### Phase 2: API Implementation
- Create GET /api/ui/alarm/state endpoint
- Create GET /api/ui/alarm/summary endpoint
- Implement role-based action filtering

### Phase 3: i18n Integration
- Add alarm.* keys to i18n system
- Populate English source from specification
- Create Turkish translations

### Phase 4: Voice Integration (Optional)
- Wire Coordinator.Voice.Speak() calls to state changes
- Test voice variants from D1

### Phase 5: UI Implementation
- Call GET /api/ui/alarm/state on load
- Render based on mode and accessibility preferences
- Poll GET /api/ui/alarm/summary for updates

---

## Future Enhancements (Out of D3 Scope)

1. **Escalation Sequence** - Automatic notifications after triggered (Phase D4?)
2. **Alarm History** - Past alarm events and reasons (Phase D4?)
3. **Sensor Details** - Show which sensor triggered (Phase D5?)
4. **Guest Notifications** - Notify guest of approval/denial (Phase D5?)
5. **HA Integration** - Detailed HA sensor status (Phase D6?)

---

## Summary

DESIGN Phase D3 successfully defines:

- ✅ **5 alarm screen modes** - Disarmed, Arming, Armed, Triggered, Blocked
- ✅ **Mode-specific messages** - Primary + secondary context for each
- ✅ **Role-based actions** - Different buttons for Admin/Guest/Restricted
- ✅ **Countdown behavior** - Clear timing, reduced_motion variants
- ✅ **Triggered behavior** - Highest priority, optional escalation
- ✅ **Blocked behavior** - 3 sub-types with recovery actions
- ✅ **API contracts** - Full endpoints with example responses
- ✅ **Localization** - 80+ i18n keys organized
- ✅ **Accessibility** - Variants for all 3 FAZ 80 preferences
- ✅ **Voice integration** - Optional voice feedback path
- ✅ **Logging strategy** - Appropriate levels without noise
- ✅ **State machine** - Deterministic transitions documented

The specification is **ready for implementation** where the API endpoints are created, state machine is implemented in Coordinator, and i18n keys are added to the localization system.

---

**Status:** ✅ SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
