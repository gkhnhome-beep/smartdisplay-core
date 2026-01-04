# DESIGN PHASE D2 COMPLETION REPORT

**Phase:** DESIGN Phase D2  
**Goal:** Define first-impression behavior of Home / Idle screen after setup completion  
**Date:** January 4, 2026  
**Status:** ✅ COMPLETE

---

## Implementation Summary

DESIGN Phase D2 successfully defines the behavioral state machine, data contracts, and accessibility considerations for the Home/Idle screen. Comprehensive specification covers all four states (SetupRedirect, Idle, Active, Alert), API contracts, localization, and logging—WITHOUT any visual design, CSS, or animations.

---

## Deliverables

### 1. Home Screen State Machine ✅

Four clearly defined states with deterministic transitions:

#### State 1: SetupRedirect
- **When:** System starts with `wizard_completed = false`
- **Behavior:** Redirect to first-boot flow
- **Duration:** Until wizard completes
- **API Response:** `{"state": "setup_redirect", "action": "go_to_firstboot"}`

#### State 2: Idle (Default)
- **When:** System running, no interaction, no alerts
- **Behavior:** Calm status summary display
- **Duration:** Until user interaction or alert
- **Display:** Alarm state, HA connectivity, time/date, optional AI insight
- **Purpose:** Minimal, reassuring interface

#### State 3: Active (User Engaged)
- **When:** User interaction detected
- **Behavior:** Expand controls, show action buttons
- **Duration:** 5 minutes inactivity timeout (configurable)
- **Display:** Full status, role-based actions, expanded info
- **Purpose:** Enable user input and control

#### State 4: Alert (Critical)
- **When:** Alarm triggered, critical event, or system error
- **Behavior:** Override all states, show alarm details
- **Duration:** Until acknowledged/resolved
- **Display:** Alarm type, details, suggested actions
- **Purpose:** Demand attention for critical events
- **Priority:** Always overrides other states

### 2. Idle State Display Components ✅

Six core components for calm, informative display:

```
1. Alarm Status (required)
   - Shows: DISARMED | ARMED | TRIGGERED
   - Localized: Via i18n keys
   - Example: "Alarm: Disarmed"

2. HA Connectivity (required)
   - Shows: Connected | Not connected
   - Status Indicator: Optional visual dot/checkmark
   - Localized: home.ha.connected / home.ha.disconnected

3. Time & Date (required)
   - Format: Localized (English: "3:45 PM • Saturday, January 4")
   - Updates: Per minute (or per second if visible)
   - Respects: Language and regional settings

4. AI One-Liner (optional, non-intrusive)
   - Max: 100 characters
   - Example: "All calm. System learned your morning routine."
   - No auto-cycling (respects reduced_motion)
   - From: coordinator.AI.GetCurrentInsight()

5. Guest Status (if active)
   - Shows: PENDING | APPROVED | DENIED | EXPIRED
   - Time Remaining: If approved
   - Localized: Via i18n keys

6. Countdown Status (if active)
   - Shows: "Disarm in {seconds}"
   - Updates: Per second
   - From: coordinator.Countdown state
```

### 3. Active State Enhancements ✅

When user interacts, display expands:

```
New Elements Added:
- Action buttons (role-based)
  * Admin: Arm, Disarm, Guest Requests, Anomalies, Settings
  * Guest: Request Entry, View Rules
  * Restricted: View Status only

- Expanded information
  * Full AI insights (not truncated)
  * Recent events (last 5)
  * Device status (if HA connected)
  * System health indicators

- Inactivity timeout
  * Default: 5 minutes
  * Configurable: Via runtime config
  * Resets: On any user interaction
```

### 4. Alert State Behavior ✅

Three priority levels for different event types:

#### Priority 1: Alarm Triggered
```
Display: "ALARM TRIGGERED: {reason}"
- Time triggered
- Location (if known)
- Suggested actions (Disarm, Acknowledge, Call Support)
- Voice hook: Speak("Alarm triggered: {reason}")
- No auto-dismiss
```

#### Priority 2: Critical Event
```
Examples: Failsafe activated, HA lost, Device error
Display: "Attention Required: {event_type}"
- Description
- Action items
- Logging: WARN level
```

#### Priority 3: System Warning
```
Examples: Battery low, Device offline, Sensor malfunction
Display: "System Warning: {description}"
- Description
- Dismiss or Learn More
- Logging: INFO level
```

### 5. State Transition Diagram ✅

Complete flow defined:

```
Startup
  ↓
wizard_completed?
  ├─ NO → SetupRedirect → (redirect to firstboot)
  └─ YES → Idle

Idle
  ├─ User Input → Active → (inactivity timeout) → Idle
  ├─ Critical Event → Alert
  └─ System Running → Idle (steady state)

Alert
  ├─ Acknowledged → Idle or Active
  ├─ Resolved → Idle
  └─ User Interaction → Alert (stays)
```

### 6. API Contracts ✅

#### GET /api/ui/home/state
Returns complete home screen state:
```json
{
  "state": "idle|active|alert|setup_redirect",
  "system_ready": boolean,
  "summary": {...},
  "alert": {...} (if state==alert),
  "actions": {...} (if state==active),
  "expanded_info": {...} (if state==active),
  "message": "string"
}
```

#### GET /api/ui/home/summary
Lightweight endpoint for frequent polling:
```json
{
  "alarm_state": "DISARMED|ARMED|TRIGGERED",
  "ha_connected": boolean,
  "current_time": "2026-01-04T15:45:00Z",
  "ai_insight": "string",
  "guest_state": null|string,
  "countdown_active": boolean,
  "countdown_remaining": integer,
  "has_pending_alerts": boolean
}
```

**Rationale:**
- Full state endpoint for UI initialization
- Summary endpoint for fast updates
- No caching except summary (1-2 sec acceptable)
- Auth-aware (different data for guest/user/admin)

### 7. Accessibility Integration ✅

#### For `reduced_motion` Users
```
Changes:
- No auto-cycling AI insights (show one static insight)
- No animated transitions between states
- Static, direct information display
- Simpler visual indicators

i18n Key:
home.ai.insight.single = (use if reduced_motion)
```

#### For `large_text` Users
```
Changes:
- Simplified summaries (fewer words)
- Fewer fields per screen
- Larger action buttons
- More spacing between elements

Display:
- Use .short variant of all text
- Increase font sizes
- Simplify alerts
- Focus on primary info

Example:
STANDARD: "SmartDisplay is monitoring. HA connected. All calm."
SIMPLIFIED: "All calm"
```

#### For `high_contrast` Users
```
Changes:
- WCAG AA contrast ratio (4.5:1 minimum)
- Bold, clear fonts
- Clear visual hierarchy
- Distinct alarm colors (red for triggered)

Colors:
- HA Connected: Green
- HA Disconnected: Gray
- Alert: Red/Orange
```

### 8. Localization Keys ✅

**100+ i18n keys defined:**

Organized by component:

```
home.alarm.* (alarm states)
home.ha.* (home assistant)
home.time.* (time/date formatting)
home.ai.* (AI insights)
home.guest.* (guest status)
home.countdown.* (countdown)
home.action.* (action buttons)
home.state.* (state names)
home.alert.* (alert messages)
home.accessibility.* (a11y options)
```

**Language Support:**
- English (en): Primary source
- Turkish (tr): Localized equivalents
- Date formatting: Locale-aware

### 9. Voice Integration Path ✅

Optional voice feedback for alerts:

```
On Alert state:
1. Get alert message from coordinator
2. Call coordinator.Voice.SpeakCritical(message)
3. Log voice event
4. Optional playback (if FAZ 81 enabled)

Example:
Voice.SpeakCritical("Alarm triggered: door unlock detected")
```

**No Audio Playback:** D2 defines contract only, FAZ 81 handles actual playback.

### 10. Logging Strategy ✅

**INFO Level (Normal Operations):**
```
INFO home: idle state (system ready, no interaction)
INFO home: active state (user interaction detected)
INFO home: state transition (idle → active, trigger: user_interaction)
INFO home: action executed (user_id: admin, action: arm, status: success)
```

**WARN Level (Alerts & Errors):**
```
WARN home: alert state (alarm triggered - door unlock detected)
WARN home: setup redirect (wizard not completed)
WARN home: action blocked (user_id: guest, action: disarm, reason: unauthorized)
```

**What NOT Logged:**
```
❌ Full user data
❌ HA tokens or credentials
❌ Every polling request
❌ Every state field
```

---

## Key Design Decisions

### 1. Idle Is Default, Not Active
**Decision:** System starts in Idle state (calm, minimal display)  
**Rationale:** Respects Product Principle of Calm; doesn't overwhelm on startup  
**Alternative Rejected:** Auto-active state (too busy for a display)

### 2. Alert Always Overrides
**Decision:** Alert state cannot be suppressed by user interaction  
**Rationale:** Critical events must get attention; can't accidentally dismiss  
**Example:** Alarm trigger overrides all UI interactions

### 3. Active Has Timeout
**Decision:** Active state reverts to Idle after inactivity (default 5 min)  
**Rationale:** Prevents leaving UI in expanded state; respects energy usage  
**Configurable:** Runtime config allows adjustment

### 4. SetupRedirect Is Separate State
**Decision:** First-boot redirect is its own state, not Idle variant  
**Rationale:** Clear separation of setup vs. normal operation  
**Clarity:** UI knows immediately if system is in setup phase

### 5. Summary Endpoint Is Separate
**Decision:** GET /api/ui/home/summary is distinct from state  
**Rationale:** UI can poll frequently for updates without full state payload  
**Cache:** Acceptable 1-2 sec staleness for performance

### 6. No Auto-Cycling (Respects reduced_motion)
**Decision:** AI insights shown as single static item, not rotating  
**Rationale:** respects FAZ 80 accessibility preference  
**Alternative:** Would require .reduced_motion variant logic

### 7. Role-Based Actions
**Decision:** Action buttons change based on user role  
**Rationale:** Guest sees different options than Admin  
**Auth:** API returns different actions based on authenticated role

---

## Data Structure Examples

### Complete Idle State Response
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

### Complete Active State Response
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
      "ai_insight": "All calm. No anomalies detected.",
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
      "recent_events": [
        {"time": "14:30", "event": "Disarm by admin"}
      ],
      "system_health": {"status": "optimal"}
    },
    "message": "Ready for input"
  }
}
```

### Complete Alert State Response
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

---

## Integration Points

### D2 ↔ D0 (First-Boot)
```
Connection: SetupRedirect state references first-boot
Contract: Checks wizard_completed flag
Behavior: If false, redirect to /api/setup/firstboot/status
```

### D2 ↔ D1 (Copy & Messaging)
```
Connection: Uses i18n keys from D1
Contract: All text through home.* key namespace
Behavior: Displays localized text from D1 specification
```

### D2 ↔ FAZ 80 (Accessibility)
```
Connection: Respects reduced_motion, large_text, high_contrast
Contract: Gets preferences from runtimeConfig
Behavior: Adapts display based on user preferences
```

### D2 ↔ FAZ 81 (Voice Feedback)
```
Connection: Optional voice integration for alerts
Contract: Calls coordinator.Voice.SpeakCritical()
Behavior: Voice hooks triggered on alert state entry
```

---

## Testing Checklist

✅ All 4 states defined and documented  
✅ State transitions clearly specified  
✅ API endpoints documented with full contracts  
✅ Data types and structures defined  
✅ Accessibility variants specified  
✅ Localization keys organized  
✅ Voice integration path clear  
✅ Logging strategy defined  
✅ Role-based access control noted  
✅ Example JSON responses provided  
✅ No visual design included  
✅ Deterministic behavior guaranteed  
✅ Timeout behavior specified  
✅ Alert priority levels defined  

---

## Next Implementation Steps

### Phase 1: API Implementation
- Add /api/ui/home/state endpoint
- Add /api/ui/home/summary endpoint
- Implement state machine logic in coordinator

### Phase 2: i18n Integration
- Add home.* keys to i18n system
- Populate English source text from D1
- Add Turkish translations

### Phase 3: State Logic
- Implement state transitions in coordinator
- Add inactivity timeout tracking
- Implement alert priority handling

### Phase 4: UI Integration
- Call GET /api/ui/home/state on startup
- Render based on state and accessibility preferences
- Poll GET /api/ui/home/summary for updates

### Phase 5: Voice Integration (Optional)
- Wire Voice.Speak() calls to alert entry
- Test voice variants from D1

---

## Future Enhancements (Out of D2 Scope)

1. **Device Widgets** - Show device status (Phase D3?)
2. **Guest Queue** - Dedicated guest request section (Phase D3?)
3. **Custom Summary** - User-configurable home screen (Phase D4?)
4. **Shortcut Actions** - Quick access favorites (Phase D4?)
5. **Weather Integration** - Local weather display (Phase D5?)

---

## Production Readiness

✅ Specification complete and comprehensive  
✅ API contracts documented  
✅ State machine deterministic  
✅ Accessibility fully considered  
✅ Localization strategy clear  
✅ No visual design (scope compliance)  
✅ No animations (scope compliance)  
✅ No CSS (scope compliance)  
✅ Backend-driven only  

---

## Summary

DESIGN Phase D2 successfully defines:

- ✅ **4-state Home screen model** - SetupRedirect, Idle, Active, Alert
- ✅ **Idle display components** - Calm, informative summary view
- ✅ **Active enhancements** - Expanded controls, role-based actions
- ✅ **Alert behavior** - Priority-based critical event handling
- ✅ **State transitions** - Deterministic flow between states
- ✅ **API contracts** - Full endpoints with example responses
- ✅ **Accessibility** - Variants for reduced_motion, large_text, high_contrast
- ✅ **Localization** - 100+ i18n keys organized and documented
- ✅ **Voice integration** - Optional voice feedback path defined
- ✅ **Logging strategy** - Appropriate log levels and examples

The specification is **ready for implementation** where API endpoints are created, state machine is implemented in coordinator, and i18n keys are added to the localization system.

---

**Status:** ✅ SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
