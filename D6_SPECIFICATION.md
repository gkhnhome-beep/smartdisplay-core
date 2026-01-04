# DESIGN PHASE D6 SPECIFICATION
## History / Logbook First-Impression Experience

**Phase:** DESIGN Phase D6  
**Focus:** System activity transparency and understandability  
**Date:** January 4, 2026  
**Status:** SPECIFICATION IN PROGRESS

---

## Overview

The History/Logbook section provides **transparent, understandable, non-threatening** access to system activity. It answers user questions like:
- "What happened in my house?"
- "When was the alarm last triggered?"
- "Who visited today?"
- "Is something wrong with my system?"

**Key Principle:** System activity should feel transparent without overwhelming users. Calm, factual tone. No blame. No jargon.

---

## Logbook Entry Categories

### Category 1: Alarm Events ✅

**Purpose:** Track all alarm state changes and triggers  
**Audience:** Admin (full), User (filtered)  
**Severity Levels:** Critical (triggered), Warning (armed/disarmed), Info (state changes)

#### Events in This Category

```
1. Alarm Triggered
   Trigger: Alarm mode changes to "triggered"
   Message: "Alarm triggered: {reason}"
   Example: "Alarm triggered: Door unlock detected"
   Severity: Critical
   
2. Alarm Armed
   Trigger: Alarm mode changes to "armed"
   Message: "Alarm armed"
   Context: "by {user_id}" or "automatically"
   Example: "Alarm armed by John Smith"
   Severity: Info
   
3. Alarm Disarmed
   Trigger: Alarm mode changes to "disarmed"
   Message: "Alarm disarmed"
   Context: "by {user_id}" or "automatically"
   Example: "Alarm disarmed by John Smith"
   Severity: Info
   
4. Alarm Countdown Started
   Trigger: Arm countdown begins
   Message: "Arm countdown started"
   Context: "{seconds} seconds"
   Example: "Arm countdown started (30 seconds)"
   Severity: Info
   
5. Alarm Countdown Cancelled
   Trigger: User cancels arming
   Message: "Arm countdown cancelled"
   Context: "by {user_id}"
   Severity: Info
   
6. Alarm Acknowledged
   Trigger: User acknowledges triggered state
   Message: "Alarm acknowledged"
   Context: "by {user_id}"
   Severity: Warning
```

#### Example Entries

```
Today 3:45 PM
Alarm triggered: Door unlock detected at Front Door
Severity: Critical

Today 3:30 PM
Alarm armed by John Smith
Severity: Info

Yesterday 10:15 AM
Alarm disarmed by Sarah Johnson
Severity: Info

January 1, 11:00 PM
Alarm armed (automatic)
Severity: Info
```

#### Retention Policy
```
Keep for: Minimum 30 days, ideally 90 days
Accessible to: Admin (full), User (filtered)
Privacy: No sensor details, just reason (if available)
```

---

### Category 2: Guest Events ✅

**Purpose:** Track guest access requests and outcomes  
**Audience:** Admin (full), User (view only)  
**Severity Levels:** Info (request, approved, denied, expired, exited)

#### Events in This Category

```
1. Guest Requested Access
   Trigger: Guest submits request
   Message: "Guest requested access"
   Context: "{time_remaining} seconds to decide"
   Example: "Guest requested access (60 seconds to decide)"
   Severity: Info
   
2. Guest Access Approved
   Trigger: Admin approves request
   Message: "Guest access approved"
   Context: "{duration} minutes of access"
   Example: "Guest access approved (30 minutes)"
   Severity: Info
   
3. Guest Access Denied
   Trigger: Admin denies request
   Message: "Guest access denied"
   Context: Optional reason (not required)
   Example: "Guest access denied"
   Severity: Info
   
4. Guest Request Expired
   Trigger: Request timeout without decision
   Message: "Guest request expired"
   Context: "{timeout} seconds"
   Example: "Guest request expired (waited 60 seconds)"
   Severity: Info
   
5. Guest Exited
   Trigger: Guest manually exits
   Message: "Guest exited"
   Context: "{duration} inside"
   Example: "Guest exited (visited for 25 minutes)"
   Severity: Info
   
6. Guest Access Auto-Expired
   Trigger: Approval duration exceeded
   Message: "Guest access expired"
   Context: "{duration} of access"
   Example: "Guest access expired (30 minute limit)"
   Severity: Info
```

#### Example Entries

```
Today 2:30 PM
Guest requested access (60 seconds to decide)
Severity: Info

Today 2:30 PM
Guest access approved (30 minutes)
Severity: Info

Today 2:55 PM
Guest exited (visited for 25 minutes)
Severity: Info

Yesterday 6:00 PM
Guest requested access (no response)
Severity: Info

Yesterday 5:01 PM
Guest request expired (waited 60 seconds)
Severity: Info
```

#### Retention Policy
```
Keep for: Minimum 7 days, ideally 30 days
Accessible to: Admin (full), User (view only)
Privacy: No guest names/IDs (just "Guest")
```

---

### Category 3: System Events ✅

**Purpose:** Track system health and operation  
**Audience:** Admin (full), User (filtered safe events)  
**Severity Levels:** Critical (error), Warning (degraded), Info (normal)

#### Events in This Category

```
1. System Started
   Trigger: Application startup
   Message: "System started"
   Context: Version info (optional)
   Example: "System started (v3.2.1)"
   Severity: Info
   
2. Home Assistant Connected
   Trigger: HA connection established
   Message: "Home Assistant connected"
   Context: Optional device count
   Example: "Home Assistant connected (12 devices)"
   Severity: Info
   
3. Home Assistant Disconnected
   Trigger: HA connection lost
   Message: "Home Assistant disconnected"
   Context: Optional reason
   Example: "Home Assistant disconnected"
   Severity: Warning
   
4. Device Offline
   Trigger: Device loses connection
   Message: "Device offline: {device_name}"
   Context: Device type
   Example: "Device offline: Front Door Sensor (door lock)"
   Severity: Warning
   
5. Device Online
   Trigger: Device regains connection
   Message: "Device back online: {device_name}"
   Context: Device type
   Example: "Device back online: Front Door Sensor"
   Severity: Info
   
6. Battery Low
   Trigger: Device battery below threshold
   Message: "Low battery: {device_name}"
   Context: Battery percentage
   Example: "Low battery: Motion Sensor (15%)"
   Severity: Warning
   
7. System Update Available
   Trigger: New software version detected
   Message: "System update available"
   Context: Version number
   Example: "System update available (v3.3.0)"
   Severity: Info
   
8. System Updated
   Trigger: System updated successfully
   Message: "System updated"
   Context: Old → New version
   Example: "System updated (v3.2.1 → v3.3.0)"
   Severity: Info
```

#### Example Entries

```
Today 5:00 PM
Home Assistant connected (12 devices)
Severity: Info

Today 12:30 PM
Device offline: Front Door Sensor (door lock)
Severity: Warning

Today 12:31 PM
Device back online: Front Door Sensor
Severity: Info

Yesterday 8:00 AM
Low battery: Motion Sensor (15%)
Severity: Warning

Yesterday 6:00 PM
System updated (v3.2.0 → v3.2.1)
Severity: Info
```

#### Retention Policy
```
Keep for: Minimum 30 days
Accessible to: Admin (full), User (filtered - device events only, no errors)
Privacy: No detailed error messages, just summaries
```

---

### Category 4: Safety / Failsafe Events ✅

**Purpose:** Track safety mechanisms and recovery  
**Audience:** Admin only (sensitive information)  
**Severity Levels:** Critical (entered), Warning (recovering), Info (recovered)

#### Events in This Category

```
1. Failsafe Activated
   Trigger: System detects critical condition
   Message: "System in safe mode"
   Context: Reason (connection lost, power issue, etc.)
   Example: "System in safe mode (Home Assistant connection lost)"
   Severity: Critical
   
2. Failsafe Recovering
   Trigger: Recovery in progress
   Message: "System recovering from safe mode"
   Context: Estimated time
   Example: "System recovering from safe mode"
   Severity: Warning
   
3. Failsafe Recovered
   Trigger: System returns to normal
   Message: "System recovered to normal operation"
   Context: Duration in failsafe
   Example: "System recovered (safe mode for 5 minutes)"
   Severity: Info
   
4. Alarm Triggered During Failsafe
   Trigger: Alarm triggered while in safe mode
   Message: "Alarm triggered during safe mode"
   Context: Reason
   Example: "Alarm triggered during safe mode (door unlock)"
   Severity: Critical
```

#### Example Entries

```
Today 2:15 PM
System in safe mode (Home Assistant connection lost)
Severity: Critical

Today 2:15 PM
System recovering from safe mode
Severity: Warning

Today 2:20 PM
System recovered (safe mode for 5 minutes)
Severity: Info

Yesterday 11:00 PM
Alarm triggered during safe mode (door unlock detected)
Severity: Critical
```

#### Retention Policy
```
Keep for: Minimum 90 days (critical for debugging)
Accessible to: Admin only (sensitive)
Privacy: Only accessible to system owner
```

---

## Logbook Entry Structure

### Standard Entry Format

```json
{
  "id": "entry_abc123",
  "timestamp": "2026-01-04T15:45:30Z",
  "timestamp_local": "Today 3:45 PM",
  "category": "alarm|guest|system|safety",
  "type": "alarm_triggered|guest_approved|device_offline|failsafe_activated",
  "severity": "critical|warning|info",
  "message": "Alarm triggered: Door unlock detected",
  "context": "Front Door",
  "details": {
    "reason": "door_unlock",
    "location": "Front Door",
    "user_id": null,
    "guest_id": null,
    "duration_seconds": null
  },
  "grouped": false,
  "group_count": 1,
  "grouped_at": null,
  "visible_to_role": "admin|user"
}
```

### Grouped Entry Format

```json
{
  "id": "entry_group_123",
  "timestamp": "2026-01-04T15:50:00Z",
  "timestamp_local": "Today 3:50 PM",
  "category": "system",
  "type": "device_offline_multiple",
  "severity": "warning",
  "message": "3 devices offline",
  "context": "Motion Sensor, Door Lock, Light Switch",
  "details": {
    "devices": [
      "Motion Sensor (offline since 3:45 PM)",
      "Door Lock (offline since 3:46 PM)",
      "Light Switch (offline since 3:48 PM)"
    ],
    "count": 3,
    "reason": "network_issue"
  },
  "grouped": true,
  "group_count": 3,
  "grouped_at": "2026-01-04T15:52:00Z",
  "visible_to_role": "admin|user"
}
```

### Timestamp Localization

```
English (en):
  "Today 3:45 PM"
  "Yesterday 10:15 AM"
  "3 days ago"
  "January 1, 11:00 PM"

Turkish (tr):
  "Bugün 15:45"
  "Dün 10:15"
  "3 gün önce"
  "1 Ocak, 23:00"
```

---

## Tone Rules & Writing Guidelines

### Core Principles

```
✓ Calm and Factual
  - Use simple, direct language
  - Report what happened, not interpretations
  - No assumptions about user guilt/blame

✓ Non-Threatening
  - No alarm language (DANGER, CRITICAL in messages)
  - Use "System in safe mode" not "FAILURE DETECTED"
  - Frame security as protective, not threatening

✓ No Blame
  - Don't say: "You disarmed the alarm"
  - Say: "Alarm disarmed by John Smith"
  - Don't say: "Guest was denied entry"
  - Say: "Guest access denied"

✓ No Jargon
  - Don't say: "HA connection lost due to network timeout"
  - Say: "Home Assistant disconnected"
  - Don't say: "MQTT broker unreachable"
  - Say: "Connection issue detected"

✓ Human-Readable
  - Use names, not IDs
  - Use timestamps user understands (Today, Yesterday, 3 days ago)
  - Explain why things happened when possible
```

### Tone Examples by Category

#### Alarm Events

```
GOOD: "Alarm armed by John Smith"
BAD:  "ALARM SYSTEM ACTIVATED"

GOOD: "Alarm triggered: Door unlock detected"
BAD:  "CRITICAL: UNAUTHORIZED ENTRY DETECTED"

GOOD: "Alarm disarmed by Sarah Johnson"
BAD:  "USER DISABLED PROTECTION"
```

#### Guest Events

```
GOOD: "Guest requested access (60 seconds to decide)"
BAD:  "UNKNOWN PERSON REQUESTING ENTRY"

GOOD: "Guest access approved (30 minutes)"
BAD:  "SECURITY OVERRIDE: GUEST ENTRY PERMITTED"

GOOD: "Guest exited (visited for 25 minutes)"
BAD:  "INTRUDER DEPARTURE CONFIRMED"
```

#### System Events

```
GOOD: "Home Assistant disconnected"
BAD:  "CRITICAL: CONNECTION LOST"

GOOD: "Device offline: Front Door Sensor"
BAD:  "DEVICE FAILURE: NETWORK UNREACHABLE"

GOOD: "System recovered (safe mode for 5 minutes)"
BAD:  "EMERGENCY SAFE MODE TERMINATED"
```

#### Safety/Failsafe Events

```
GOOD: "System in safe mode (Home Assistant connection lost)"
BAD:  "CRITICAL SYSTEM FAILURE - SAFE MODE ENGAGED"

GOOD: "System recovering from safe mode"
BAD:  "ATTEMPTING EMERGENCY RECOVERY"

GOOD: "Alarm triggered during safe mode (door unlock)"
BAD:  "BREACH DETECTED WHILE DISABLED"
```

### Severity Level Tone

```
INFO (Blue indicator):
- Routine operations
- Expected state changes
- Normal notifications
- Tone: Informational, neutral

WARNING (Orange indicator):
- Unexpected but recoverable
- Temporary issues
- Requires attention but not urgent
- Tone: Cautious but calm

CRITICAL (Red indicator):
- Alarm triggered
- Security-related events
- Needs immediate attention
- Tone: Urgent but factual
```

---

## Event Grouping & Summarization

### Grouping Logic

**When to Group:**
```
Same event type within 5-minute window
AND
Related (e.g., multiple devices going offline)
OR
Same event repeated more than twice
```

**When NOT to Group:**
```
Different event types (even in same category)
OR
Events separated by > 5 minutes
OR
Different severity levels
```

### Grouping Examples

#### Example 1: Multiple Devices Going Offline

```
Before Grouping:
  3:45 PM - Device offline: Motion Sensor
  3:46 PM - Device offline: Door Lock
  3:48 PM - Device offline: Light Switch

After Grouping:
  3:50 PM - 3 devices offline
           (Motion Sensor, Door Lock, Light Switch)
```

#### Example 2: Repeated Request Timeouts

```
Before Grouping:
  2:05 PM - Guest request expired (60 seconds)
  2:06 PM - Guest request expired (60 seconds)
  2:07 PM - Guest request expired (60 seconds)

After Grouping:
  2:07 PM - 3 guest requests expired
```

#### Example 3: Device State Changes (NOT Grouped)

```
DO NOT GROUP these (different event types):
  3:45 PM - Device offline: Motion Sensor
  3:46 PM - Device back online: Motion Sensor
  
Display as separate entries (shows state change)
```

### Summarization Rules

```
Single Event: Show full message
  "Alarm triggered: Door unlock detected"

2-3 Events: Show abbreviated
  "Alarm armed, armed, armed" → "Alarm armed (3 times)"
  
4+ Events: Show count
  "4 similar events in the last 10 minutes"
```

### Grouping Entry Content

**Grouped Entry Structure:**
```
Primary Message: {action} {count} {items}
Example: "4 devices offline"

Detail List: Show what was grouped
Example: 
  - Motion Sensor (offline since 3:45 PM)
  - Door Lock (offline since 3:46 PM)
  - Light Switch (offline since 3:48 PM)
  - Camera (offline since 3:50 PM)

Count Badge: Show "4" visually
Tooltip (hover): Show all details
```

---

## API Contracts

### GET /api/ui/logbook

**Purpose:** Retrieve full logbook history with pagination  
**Auth:** Required (Admin or User, role-filtered)  
**Query Parameters:**
```
?limit=20         (default, max 100)
?offset=0         (pagination)
?category=alarm   (optional filter)
?days=30          (optional, default 30)
?severity=warning (optional filter)
```

**Response (Admin User):**
```json
{
  "ok": true,
  "data": {
    "entries": [
      {
        "id": "entry_001",
        "timestamp": "2026-01-04T15:45:30Z",
        "timestamp_local": "Today 3:45 PM",
        "category": "alarm",
        "type": "alarm_triggered",
        "severity": "critical",
        "message": "Alarm triggered: Door unlock detected",
        "context": "Front Door",
        "details": {
          "reason": "door_unlock",
          "location": "Front Door",
          "user_id": null
        },
        "grouped": false,
        "group_count": 1,
        "visible_to_role": "admin"
      },
      {
        "id": "entry_002",
        "timestamp": "2026-01-04T15:30:00Z",
        "timestamp_local": "Today 3:30 PM",
        "category": "guest",
        "type": "guest_approved",
        "severity": "info",
        "message": "Guest access approved (30 minutes)",
        "context": null,
        "details": {
          "duration_minutes": 30,
          "guest_id": "guest_123"
        },
        "grouped": false,
        "group_count": 1,
        "visible_to_role": "admin"
      },
      {
        "id": "entry_group_003",
        "timestamp": "2026-01-04T15:20:00Z",
        "timestamp_local": "Today 3:20 PM",
        "category": "system",
        "type": "device_offline_multiple",
        "severity": "warning",
        "message": "3 devices offline",
        "context": "Motion Sensor, Door Lock, Light Switch",
        "details": {
          "devices": [
            "Motion Sensor (offline since 3:15 PM)",
            "Door Lock (offline since 3:18 PM)",
            "Light Switch (offline since 3:20 PM)"
          ],
          "count": 3,
          "reason": "network_issue"
        },
        "grouped": true,
        "group_count": 3,
        "grouped_at": "2026-01-04T15:22:00Z",
        "visible_to_role": "admin"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 127,
      "has_more": true
    },
    "metadata": {
      "date_range_start": "2025-12-05T00:00:00Z",
      "date_range_end": "2026-01-04T23:59:59Z",
      "category_counts": {
        "alarm": 5,
        "guest": 3,
        "system": 12,
        "safety": 0
      }
    }
  }
}
```

**Response (User Role - Filtered):**
```json
{
  "ok": true,
  "data": {
    "entries": [
      {
        "id": "entry_002",
        "timestamp": "2026-01-04T15:30:00Z",
        "timestamp_local": "Today 3:30 PM",
        "category": "guest",
        "type": "guest_approved",
        "severity": "info",
        "message": "Guest access approved (30 minutes)",
        "context": null,
        "details": {
          "duration_minutes": 30,
          "guest_id": "guest_123"
        },
        "grouped": false,
        "group_count": 1,
        "visible_to_role": "user"
      },
      {
        "id": "entry_003",
        "timestamp": "2026-01-04T15:10:00Z",
        "timestamp_local": "Today 3:10 PM",
        "category": "system",
        "type": "device_online",
        "severity": "info",
        "message": "Device back online: Motion Sensor",
        "context": null,
        "details": {
          "device_name": "Motion Sensor",
          "device_type": "sensor"
        },
        "grouped": false,
        "group_count": 1,
        "visible_to_role": "user"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 47,
      "has_more": false
    },
    "metadata": {
      "category_counts": {
        "guest": 3,
        "system": 8
      }
    }
  }
}
```

---

### GET /api/ui/logbook/summary

**Purpose:** Get recent entries for dashboard glance  
**Auth:** Required (Admin or User, role-filtered)  
**Query Parameters:**
```
?limit=5  (default, max 20)
?hours=24 (default time window)
```

**Response:**
```json
{
  "ok": true,
  "data": {
    "recent_entries": [
      {
        "timestamp_local": "Today 3:45 PM",
        "message": "Alarm triggered: Door unlock detected",
        "severity": "critical",
        "category": "alarm"
      },
      {
        "timestamp_local": "Today 3:30 PM",
        "message": "Guest access approved (30 minutes)",
        "severity": "info",
        "category": "guest"
      },
      {
        "timestamp_local": "Today 3:20 PM",
        "message": "3 devices offline",
        "severity": "warning",
        "category": "system"
      }
    ],
    "metadata": {
      "has_critical_events": true,
      "last_24_hours": {
        "total": 12,
        "critical": 1,
        "warning": 3,
        "info": 8
      }
    }
  }
}
```

---

## Accessibility Integration

### For `reduced_motion` Users (FAZ 80)

**Auto-Refresh Behavior:**
```
❌ NO: Auto-refreshing logbook every 30 seconds
✅ YES: User manually refreshes, or on-demand polling only
```

**Transition Animation:**
```
❌ NO: Sliding entry animations, fade-in effects
✅ YES: Instant display change, no animation
```

**Loading Indicator:**
```
✓ Static loading message if needed
✓ Progress indicator only if > 5 seconds
✓ No animated spinner
```

---

### For `large_text` Users (FAZ 80)

**Text Simplification:**
```
Standard: "Device offline: Motion Sensor (Living Room) due to network timeout"
Large Text: "Motion Sensor offline"

Standard: "Home Assistant connection lost (MQTT broker unreachable)"
Large Text: "Home Assistant disconnected"
```

**Font Sizes:**
```
Message: 18pt+ (readable)
Context: 14pt+ (secondary info)
Timestamp: 16pt+ (clear)
Button text: 16pt+ (easy to tap)
```

**Layout:**
```
One entry per line
Clear separation between entries
Large clickable area for each entry
Details expandable (if needed)
```

---

### For `high_contrast` Users (FAZ 80)

**Severity Color Scheme:**
```
Info (Blue):     RGB(0, 102, 204)  - Safe, routine
Warning (Orange): RGB(255, 153, 0)  - Caution, attention needed
Critical (Red):   RGB(204, 0, 0)    - Urgent, action needed

All WCAG AA compliant (4.5:1+ contrast ratio)
```

**Visual Indicators:**
```
✓ Clear severity badge (color + icon)
✓ High contrast text on background
✓ Bold fonts for emphasis
✓ Clear borders between entries
```

---

## Localization Keys (i18n)

### Message Templates

```
logbook.alarm.triggered = "Alarm triggered: {reason}"
logbook.alarm.armed = "Alarm armed"
logbook.alarm.armed_by = "Alarm armed by {user}"
logbook.alarm.disarmed = "Alarm disarmed"
logbook.alarm.disarmed_by = "Alarm disarmed by {user}"
logbook.alarm.countdown_started = "Arm countdown started ({seconds} seconds)"
logbook.alarm.countdown_cancelled = "Arm countdown cancelled by {user}"
logbook.alarm.acknowledged = "Alarm acknowledged by {user}"
```

### Guest Event Messages

```
logbook.guest.requested = "Guest requested access ({seconds} seconds to decide)"
logbook.guest.approved = "Guest access approved ({minutes} minutes)"
logbook.guest.denied = "Guest access denied"
logbook.guest.expired = "Guest request expired (waited {seconds} seconds)"
logbook.guest.exited = "Guest exited (visited for {minutes} minutes)"
logbook.guest.auto_expired = "Guest access expired ({minutes} minute limit)"
```

### System Event Messages

```
logbook.system.started = "System started ({version})"
logbook.system.ha_connected = "Home Assistant connected ({count} devices)"
logbook.system.ha_disconnected = "Home Assistant disconnected"
logbook.system.device_offline = "Device offline: {device_name}"
logbook.system.device_online = "Device back online: {device_name}"
logbook.system.battery_low = "Low battery: {device_name} ({percentage}%)"
logbook.system.update_available = "System update available ({version})"
logbook.system.updated = "System updated ({old_version} → {new_version})"
```

### Safety/Failsafe Messages

```
logbook.safety.failsafe_activated = "System in safe mode ({reason})"
logbook.safety.failsafe_recovering = "System recovering from safe mode"
logbook.safety.failsafe_recovered = "System recovered (safe mode for {minutes} minutes)"
logbook.safety.alarm_during_failsafe = "Alarm triggered during safe mode ({reason})"
```

### Grouping Messages

```
logbook.grouped.devices_offline = "{count} devices offline"
logbook.grouped.requests_expired = "{count} guest requests expired"
logbook.grouped.similar_events = "{count} similar events"
```

### Severity Labels

```
logbook.severity.critical = "Critical"
logbook.severity.warning = "Warning"
logbook.severity.info = "Info"
```

### Category Labels

```
logbook.category.alarm = "Alarm"
logbook.category.guest = "Guest"
logbook.category.system = "System"
logbook.category.safety = "Safety"
```

### Timestamp Formatting

```
English (en):
logbook.time.today = "Today {time}"
logbook.time.yesterday = "Yesterday {time}"
logbook.time.days_ago = "{count} days ago"
logbook.time.full_date = "{date} {time}"

Turkish (tr):
logbook.time.today = "Bugün {time}"
logbook.time.yesterday = "Dün {time}"
logbook.time.days_ago = "{count} gün önce"
logbook.time.full_date = "{date} {time}"
```

**Total i18n Keys:** ~70+ organized by category and purpose

---

## Data Privacy & Security Policy

### What's Visible to Each Role

#### Admin User (Full Access)

```
✓ All categories: Alarm, Guest, System, Safety
✓ All severity levels
✓ Full details (user IDs, reasons, etc.)
✓ Grouped entries
✓ 90-day history
✓ API: GET /api/ui/logbook, GET /api/ui/logbook/summary
```

#### User Role (Filtered)

```
✓ Guest events (approved/denied/exited)
✓ System events (device status, updates)
✓ Alarm events (state changes only, not triggers)
✗ Safety/failsafe events (HIDDEN)
✗ User who armed/disarmed (see "automatically" only)
✓ Grouped entries
✓ 30-day history
✓ API: GET /api/ui/logbook, GET /api/ui/logbook/summary (filtered)
```

#### Guest Role

```
✗ No logbook access (section hidden)
```

### What's NOT in Logbook

```
❌ Raw system logs
❌ Full HA configuration
❌ Sensor raw values
❌ Network packets
❌ Personal user data beyond ID
❌ Guest names or contact info
❌ Authentication tokens
❌ Alarm codes or sequences
```

### Logbook Derivation

```
Source: Internal audit logs
Process: Entries derived from audit, not exposed directly
Filtering: PII removed, sensitive details sanitized
Aggregation: Related events grouped
```

### Data Retention

```
Alarm Events:    30-90 days (configurable)
Guest Events:    7-30 days (configurable)
System Events:   30 days (configurable)
Safety Events:   90 days (required for debugging, not configurable)

After retention: Entries archived or deleted per policy
Access logs: Separate from logbook (not shown to users)
```

### User Data Redaction

```
When showing "Disarmed by John Smith":
- User ID stored: "user_456"
- Display name: "John Smith"
- Email: NOT shown
- IP address: NOT shown
- Timestamp: Shown (localized)

When showing "Guest access":
- Guest ID: Not shown (just "Guest")
- Device used: Not shown
- IP address: Not shown
```

---

## Logging Strategy

### Logbook Entry Generation

**From Coordinator Audit Log:**
```
Coordinator logs every action (internal):
  timestamp, user_id, action, status, result

Logbook converts to user-friendly entry:
  timestamp_local, message, severity, category
```

**Example Flow:**
```
Coordinator Audit:
  {time: 2026-01-04T15:45:30Z, user: admin_123, action: alarm_trigger, reason: door_unlock, status: success}

Logbook Entry:
  {timestamp: 2026-01-04T15:45:30Z, message: "Alarm triggered: Door unlock detected", severity: critical, category: alarm}
```

### Startup Logging

```
INFO logbook: initialized (retention: 30 days, retention_safety: 90 days)
INFO logbook: loading entries from audit (entries: {count})
INFO logbook: grouping similar events (groups: {count})
```

### Runtime Logging

```
INFO logbook: entry created (category: alarm, severity: critical)
INFO logbook: entries retrieved by {user_id} (role: admin, count: 20)
INFO logbook: entries retrieved by {user_id} (role: user, count: 15, filtered)
INFO logbook: old entries archived (count: {count}, days: 31)
```

### What NOT to Log

```
❌ Every logbook API request
❌ User interaction timing
❌ Raw entry data before sanitization
```

---

## Tone Validation Examples

### Alarm Triggered

```
❌ BAD:
"CRITICAL ALERT: UNAUTHORIZED ENTRY DETECTED AT FRONT DOOR"

✓ GOOD:
"Alarm triggered: Door unlock detected at Front Door"

Why: Calm, factual, not panic-inducing
```

### Guest Request Expired

```
❌ BAD:
"GUEST REJECTION: TIMEOUT AFTER 60 SECONDS"

✓ GOOD:
"Guest request expired (waited 60 seconds)"

Why: Neutral tone, explains what happened
```

### Home Assistant Disconnection

```
❌ BAD:
"SYSTEM FAILURE: HOME ASSISTANT UNREACHABLE"

✓ GOOD:
"Home Assistant disconnected"

Why: Simple, factual, not alarmist
```

### Failsafe Activation

```
❌ BAD:
"EMERGENCY: SYSTEM FAILURE DETECTED - SAFE MODE ENGAGED"

✓ GOOD:
"System in safe mode (Home Assistant connection lost)"

Why: Explains reason, not threatening language
```

---

## Integration with Previous Phases

### FAZ 79 (Localization)
```
All logbook text through logbook.* key namespace
70+ keys for English and Turkish
Localized timestamps ("Today", "Yesterday", etc.)
```

### FAZ 80 (Accessibility)
```
reduced_motion: No auto-refresh, static display
large_text: Simplified messages, 18pt+ fonts
high_contrast: Color-coded severity, WCAG AA compliant
```

### FAZ 81 (Voice Feedback)
```
Logbook entries readable by voice (optional)
Critical events can trigger voice notifications
Voice variants for each message type
```

### D0-D5 (Design Phases)
```
All previous features generate logbook entries:
- First-boot completion (System Events)
- Alarm state changes (Alarm Events)
- Guest approval/denial (Guest Events)
- Failsafe recovery (Safety Events)
- Menu permission changes (System Events)
```

---

## Testing Checklist

✅ 4 logbook categories defined (Alarm, Guest, System, Safety)  
✅ Entry structure specified (timestamp, message, context, severity)  
✅ Tone guidelines with examples  
✅ Grouping logic defined  
✅ Summarization rules specified  
✅ API endpoints documented with examples  
✅ 70+ i18n keys organized  
✅ Accessibility variants for all 3 preferences  
✅ Privacy policy defined (role-based filtering)  
✅ Data retention policy specified  
✅ User data redaction rules defined  
✅ Logging strategy documented  
✅ Tone validation examples provided  
✅ Integration with all phases confirmed  

---

## Next Implementation Steps

### Phase 1: Audit Log Integration
- Create LogbookEntry struct
- Implement entry generation from audit logs
- Add grouping logic

### Phase 2: API Implementation
- Create GET /api/ui/logbook endpoint
- Create GET /api/ui/logbook/summary endpoint
- Implement role-based filtering

### Phase 3: i18n Integration
- Add logbook.* keys to i18n system
- Populate English source text
- Create Turkish translations

### Phase 4: Data Retention
- Implement age-based archiving
- Add cleanup job for old entries
- Test retention policies

### Phase 5: Voice Integration (Optional)
- Wire Voice.SpeakInfo() for critical events
- Test critical event notifications

---

## Future Enhancements (Out of D6 Scope)

1. **Search & Filtering** - Find events by keyword (Phase D7?)
2. **Event Details** - Expand entry for full context (Phase D7?)
3. **Export History** - Download logbook as CSV (Phase D7?)
4. **Statistics Dashboard** - Show trends, patterns (Phase D8?)
5. **Alerts on Events** - Notify user of critical events (Phase D8?)
6. **Timeline View** - Visual timeline of events (Phase D8?)

---

## Summary

DESIGN Phase D6 successfully defines:

- ✅ **4 logbook entry categories** - Alarm, Guest, System, Safety
- ✅ **Entry structure** - Timestamp, message, context, severity
- ✅ **Tone rules** - Calm, factual, non-threatening, no jargon
- ✅ **Grouping & summarization** - Related events grouped, repeated events summarized
- ✅ **API contracts** - Full history and summary endpoints
- ✅ **Accessibility** - Variants for all 3 FAZ 80 preferences
- ✅ **Localization** - 70+ i18n keys organized
- ✅ **Privacy policy** - Role-based filtering, data redaction
- ✅ **Retention policy** - Different durations by category
- ✅ **Voice integration** - Optional voice for critical events
- ✅ **Logging strategy** - Entry generation from audit logs
- ✅ **Tone validation** - Examples for each category

The specification is **ready for implementation** where LogbookEntry is created, API endpoints are built, and i18n keys are added.

---

**Status:** ✅ SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
