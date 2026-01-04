# DESIGN PHASE D4 SPECIFICATION
## Guest Access First-Impression Flow & Behavioral Model

**Phase:** DESIGN Phase D4  
**Focus:** Guest access flow and interaction model  
**Date:** January 4, 2026  
**Status:** SPECIFICATION IN PROGRESS

---

## Overview

The Guest Access flow enables temporary, controlled access to SmartDisplay for visitors. This specification defines:
- Six distinct guest states (Idle → Requesting → Approved/Denied → Exit/Expired)
- Reassuring, non-judgmental messaging
- Safe disarm and alarm control during guest access
- Owner notifications (backend only, no HA calls)
- Accessibility and voice integration

**Key Principle:** Guest experience should feel premium, safe, and non-restrictive—while owners maintain full control.

---

## Guest Flow States

### State 1: GuestIdle ✅

**When:** Guest device connected, no request active  
**Primary State:** `"guest_idle"`  
**Default:** System initialized, guest just connected  
**Duration:** Until guest initiates request or time-based auto-idle

#### Primary Message
```
EN: "Welcome. Request access to enter?"
TR: "Hoş geldiniz. Girmek için erişim talep etmek ister misiniz?"

Tone: Warm, welcoming, non-judgmental
Context: Guest is ready to request
```

#### Secondary Context
```
EN: "Your request will be sent to the property owner."
TR: "Talebiniz mülk sahibine gönderilecektir."

Purpose: Set expectations (not automatic, owner decides)
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- Request Access (→ GuestRequesting state)
- View House Rules (read-only)
- Call Owner (optional, TBD)

**Guest Cannot:**
- Disarm alarm
- Enter property
- See owner contact details (for safety)

#### Owner Notifications
```
None yet (no request active)
Display: "No pending guest requests" in owner view
```

#### Data Structure
```json
{
  "state": "guest_idle",
  "message": "Welcome. Request access to enter?",
  "context": "Your request will be sent to the property owner.",
  "timestamp": "2026-01-04T10:30:00Z",
  "guest_id": "guest_123",
  "actions": [
    {"id": "request", "label": "Request Access", "enabled": true},
    {"id": "rules", "label": "House Rules", "enabled": true}
  ],
  "info": {
    "can_request": true,
    "reason_blocked": null,
    "last_request": null
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display, no animation
- **large_text:** "Request Access" button prominent, large text
- **high_contrast:** Clear button outline, high contrast

#### Voice Variant (Optional FAZ 81)
```
Voice.SpeakInfo("Welcome. You can request access to enter.")
(Optional, on state entry)
```

#### i18n Keys
```
guest.state.idle = "Guest Idle"
guest.message.idle = "Welcome. Request access to enter?"
guest.context.idle.welcome = "Your request will be sent to the property owner."
guest.context.idle.info = "The property owner will review your request."
guest.action.request = "Request Access"
guest.action.rules = "House Rules"
guest.info.idle = "Request access to enter the property."
```

---

### State 2: GuestRequesting ✅

**When:** Guest initiates access request, countdown in progress  
**Primary State:** `"guest_requesting"`  
**Duration:** Default timeout 60 seconds (configurable)  
**Cannot Self-Cancel:** Guest must wait for owner decision or timeout  
**Auto-Transition:** At timeout → GuestExpired

#### Primary Message
```
EN: "Request sent to owner"
TR: "İstek sahibine gönderildi"

Tone: Calm, patient
Context: Waiting for owner response
```

#### Secondary Context
```
EN: "Waiting for approval. {remaining} seconds remaining."
TR: "Onay bekleniyor. {remaining} saniye kaldı."

OR (if long wait):

EN: "Owner has {remaining} seconds to review your request."
TR: "Sahibinin talebinizi incelemesi için {remaining} saniyesi var."

Purpose: Show countdown, manage expectations
```

#### Countdown Behavior

**Standard Display (all users):**
```
Remaining: 60 seconds → 59 → 58 ... → 1 → 0 (auto-expire)
Updates: Every second
Display: Clear number display
Example: "Waiting for approval. 45 seconds remaining."
```

**Reduced Motion Users (FAZ 80):**
```
Display: Static countdown (not animated)
Updates: Every second, but no animation
Format: Clear number display only
Example: "Waiting for approval. Please wait 45 seconds."
```

**Large Text Users (FAZ 80):**
```
Display: Large countdown number
Context: Single line, simplified
Actions: None (must wait)
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- None (must wait for owner decision)
- View countdown (information only)

**Guest Cannot:**
- Cancel request (must wait timeout)
- Proceed to entry
- Interrupt process

#### Owner Notifications
```
Owner View Shows:
- Guest requesting entry
- Pending request badge
- Time remaining (60s default)
- Approve or Deny buttons
- Guest contact info (if available)

Owner Can:
- Approve (→ GuestApproved state)
- Deny (→ GuestDenied state)
- Ignore (→ GuestExpired at timeout)
```

#### Data Structure
```json
{
  "state": "guest_requesting",
  "message": "Request sent to owner",
  "context": "Waiting for approval. {remaining} seconds remaining.",
  "countdown": {
    "total_seconds": 60,
    "remaining_seconds": 45,
    "percentage": 75,
    "started_at": "2026-01-04T10:30:00Z",
    "will_expire_at": "2026-01-04T10:31:00Z"
  },
  "timestamp": "2026-01-04T10:30:15Z",
  "guest_id": "guest_123",
  "request_id": "req_abc123",
  "actions": [],
  "owner_notification": {
    "title": "Guest requesting entry",
    "pending": true,
    "time_remaining_seconds": 45
  },
  "info": {
    "can_cancel": false,
    "auto_expire_at": "2026-01-04T10:31:00Z",
    "approval_pending": true
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static countdown, no animation
- **large_text:** Large countdown number, minimal text
- **high_contrast:** Clear countdown display, bold text

#### i18n Keys
```
guest.state.requesting = "Requesting Access"
guest.message.requesting = "Request sent to owner"
guest.context.requesting = "Waiting for approval. {remaining} seconds remaining."
guest.context.requesting.owner = "Owner has {remaining} seconds to review your request."
guest.countdown.seconds = "{remaining} seconds remaining"
guest.info.requesting = "Your request is pending owner approval."
guest.timeout.default = "Request expires in 60 seconds"
guest.notification.pending = "Guest requesting entry"
```

---

### State 3: GuestApproved ✅

**When:** Owner approves guest request  
**Primary State:** `"guest_approved"`  
**Duration:** Until guest exits or approval expires (configurable)  
**Default Approval Duration:** 30 minutes (configurable)  
**Alarm Impact:** Disarmed safely during approved period

#### Primary Message
```
EN: "Welcome! Access approved."
TR: "Hoş geldiniz! Erişim onaylandı."

Tone: Warm, welcoming
Context: Guest is now inside (or can enter)
```

#### Secondary Context
```
EN: "Your access is active until {time}."
TR: "Erişiminiz {time} kadar etkindir."

Purpose: Show expiration time, clarity
```

#### Approval Details
```json
{
  "approved_at": "2026-01-04T10:30:45Z",
  "expires_at": "2026-01-04T11:00:45Z",
  "duration_minutes": 30,
  "time_remaining_seconds": 1800
}
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- Exit Access (→ GuestExit state)
- View House Rules (read-only)
- Call Owner (optional)
- Disarm Alarm (while approved)

**Guest Cannot:**
- Extend approval
- Change access duration
- Access owner settings

#### Owner Notifications
```
Owner View Shows:
- Guest currently inside
- Approval active badge
- Time until auto-expiry
- Guest info (if available)
- Manual kick-out option (optional)

Owner Can:
- Revoke access (force exit)
- View guest activity
- Message guest (optional)
```

#### Auto-Expiration

**At Expiration Time:**
```
1. Check time_remaining_seconds = 0
2. Automatically transition to GuestExpired
3. Re-arm alarm to previous mode
4. Log: "INFO guest: approval expired for {guest_id}"
5. Notify owner: "Guest access expired"
```

#### Safe Disarm During Approval

```
When guest tries to disarm alarm during approval:
1. Check: guest_approved state is active
2. Check: approval not expired
3. Check: alarm is armed
4. Action: Disarm alarm safely (set mode to disarmed)
5. Log: "INFO guest: alarm disarmed by approved guest {guest_id}"
6. Notify owner: "Guest disarmed alarm"
7. Voice hook: Voice.SpeakInfo("Alarm disarmed by guest")
```

#### Data Structure
```json
{
  "state": "guest_approved",
  "message": "Welcome! Access approved.",
  "context": "Your access is active until {time}.",
  "approval": {
    "approved_at": "2026-01-04T10:30:45Z",
    "expires_at": "2026-01-04T11:00:45Z",
    "duration_minutes": 30,
    "time_remaining_seconds": 1800
  },
  "timestamp": "2026-01-04T10:30:45Z",
  "guest_id": "guest_123",
  "request_id": "req_abc123",
  "actions": [
    {"id": "exit", "label": "Exit", "enabled": true},
    {"id": "rules", "label": "House Rules", "enabled": true},
    {"id": "disarm", "label": "Disarm Alarm", "enabled": true}
  ],
  "owner_notification": {
    "title": "Guest inside (approved)",
    "active": true,
    "time_remaining_minutes": 30
  },
  "alarm_status": {
    "can_disarm": true,
    "currently_disarmed": true
  },
  "info": {
    "access_active": true,
    "will_expire_at": "2026-01-04T11:00:45Z",
    "approval_duration_minutes": 30
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static expiry time, no countdown animation
- **large_text:** Clear approval message, large "Exit" button
- **high_contrast:** Green indicator for "approved" state

#### Voice Variant (Optional FAZ 81)
```
Voice.SpeakInfo("Access approved. You may now enter.")
(On state entry)
```

#### i18n Keys
```
guest.state.approved = "Access Approved"
guest.message.approved = "Welcome! Access approved."
guest.context.approved = "Your access is active until {time}."
guest.context.approved.info = "Guest access is currently active."
guest.action.exit = "Exit"
guest.action.disarm = "Disarm Alarm"
guest.info.approved = "You have temporary access to the property."
guest.approval.expires_at = "Access expires at {time}"
guest.approval.duration = "{minutes} minutes of access"
guest.notification.inside = "Guest inside (approved)"
guest.alarm.disarmed_by_guest = "Alarm disarmed by approved guest"
```

---

### State 4: GuestDenied ✅

**When:** Owner explicitly denies guest request  
**Primary State:** `"guest_denied"`  
**Duration:** Permanent (for this session) or auto-timeout  
**No Re-request:** Until new session starts (optional)

#### Primary Message
```
EN: "Access denied"
TR: "Erişim reddedildi"

Tone: Professional, not harsh
Context: Entry not permitted (no judgment)
```

#### Secondary Context
```
EN: "Your request has been denied. Please contact the owner for assistance."
TR: "Talebiniz reddedilmiştir. Yardım için lütfen sahibi ile iletişime geçin."

Purpose: Provide next steps, maintain dignity
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- View House Rules (read-only)
- Call Owner (optional)
- Disconnect (exit guest mode)

**Guest Cannot:**
- Request access again (in same session)
- Disarm alarm
- Proceed to entry

#### Owner Notifications
```
Owner View Shows:
- Guest request denied
- Time of denial
- Optional reason (if owner provided)

Owner Can:
- View denial history
- Remove guest device (optional)
```

#### Data Structure
```json
{
  "state": "guest_denied",
  "message": "Access denied",
  "context": "Your request has been denied. Please contact the owner for assistance.",
  "denial": {
    "denied_at": "2026-01-04T10:30:55Z",
    "reason": null,
    "permanent": false,
    "can_retry": false
  },
  "timestamp": "2026-01-04T10:30:55Z",
  "guest_id": "guest_123",
  "request_id": "req_abc123",
  "actions": [
    {"id": "rules", "label": "House Rules", "enabled": true},
    {"id": "call", "label": "Call Owner", "enabled": true},
    {"id": "disconnect", "label": "Disconnect", "enabled": true}
  ],
  "owner_notification": {
    "title": "Guest request denied",
    "status": "denied"
  },
  "info": {
    "can_request_again": false,
    "reason": null
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display, no animation
- **large_text:** Clear denial message, large text
- **high_contrast:** Orange indicator for "denied" state (not threatening)

#### i18n Keys
```
guest.state.denied = "Access Denied"
guest.message.denied = "Access denied"
guest.context.denied = "Your request has been denied. Please contact the owner for assistance."
guest.context.denied.info = "The property owner has denied your request."
guest.action.call = "Call Owner"
guest.action.disconnect = "Disconnect"
guest.info.denied = "You do not have access to the property."
guest.notification.denied = "Guest request denied"
guest.denied.no_retry = "You cannot retry this request in this session."
```

---

### State 5: GuestExpired ✅

**When:** Request timeout reached without owner decision  
**Primary State:** `"guest_expired"`  
**Duration:** Permanent (for this request)  
**Auto-Trigger:** At 60s countdown expiration

#### Primary Message
```
EN: "Request expired"
TR: "İstek süresi doldu"

Tone: Matter-of-fact, not apologetic
Context: Owner didn't respond in time
```

#### Secondary Context
```
EN: "Your request was not answered within the time limit. You may try again later."
TR: "Talebiniz zaman sınırı içinde yanıtlanmadı. Daha sonra tekrar deneebilirsiniz."

Purpose: Explain what happened, suggest retry
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- Request Access Again (→ GuestRequesting state, new request)
- View House Rules (read-only)
- Call Owner (optional)
- Disconnect (exit guest mode)

**Guest Cannot:**
- Proceed to entry
- Disarm alarm
- Access anything

#### Owner Notifications
```
Owner View Shows:
- Guest request expired (no action taken)
- Time of expiration
- Number of expired requests from this guest

Owner Can:
- View expired request history
- Pre-approve future guests (optional)
```

#### Data Structure
```json
{
  "state": "guest_expired",
  "message": "Request expired",
  "context": "Your request was not answered within the time limit. You may try again later.",
  "expiration": {
    "expired_at": "2026-01-04T10:31:00Z",
    "reason": "timeout",
    "timeout_seconds": 60
  },
  "timestamp": "2026-01-04T10:31:00Z",
  "guest_id": "guest_123",
  "request_id": "req_abc123",
  "actions": [
    {"id": "request_again", "label": "Request Again", "enabled": true},
    {"id": "rules", "label": "House Rules", "enabled": true},
    {"id": "call", "label": "Call Owner", "enabled": true},
    {"id": "disconnect", "label": "Disconnect", "enabled": true}
  ],
  "owner_notification": {
    "title": "Guest request expired (no action)",
    "status": "expired"
  },
  "info": {
    "can_retry": true,
    "time_until_retry_allowed": 0
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display, no animation
- **large_text:** Clear message, "Request Again" button prominent
- **high_contrast:** Gray indicator for "expired" state

#### i18n Keys
```
guest.state.expired = "Request Expired"
guest.message.expired = "Request expired"
guest.context.expired = "Your request was not answered within the time limit. You may try again later."
guest.context.expired.info = "The property owner did not respond to your request."
guest.action.request_again = "Request Again"
guest.info.expired = "You may request access again."
guest.notification.expired = "Guest request expired (no action)"
guest.expired.retry = "You can submit a new request."
```

---

### State 6: GuestExit ✅

**When:** Approved guest manually exits before expiration  
**Primary State:** `"guest_exit"`  
**Duration:** Permanent (for this session)  
**Alarm Impact:** Re-arms to previous mode

#### Primary Message
```
EN: "You have exited the property"
TR: "Mülkü terk ettiniz"

Tone: Professional, farewell
Context: Guest access concluded
```

#### Secondary Context
```
EN: "Thank you for visiting. The alarm has been re-armed."
TR: "Ziyaretiniz için teşekkürler. Alarm yeniden silahlandı."

Purpose: Confirm exit, explain re-arm
```

#### Exit Sequence

```
1. Guest clicks "Exit" button (from Approved state)
2. Action: POST /api/ui/guest/exit
3. Coordinator receives exit signal
4. Sets state to guest_exit
5. Re-arms alarm to previous mode (before approval)
6. Logs: "INFO guest: guest exited manually (guest_id: {id})"
7. Notifies owner: "Guest has exited"
8. Returns exit confirmation to guest
```

#### Alarm Re-Arming

```
When Guest Exits:
- Check alarm mode before approval (saved state)
- If previously Armed: Re-arm to Armed
- If previously Disarmed: Keep Disarmed
- If previously Arming: Resume countdown
- Voice hook: Voice.SpeakInfo("Guest exited. Alarm re-armed.")
```

#### Allowed Actions (Guest-Facing)

**Guest Can:**
- Disconnect (exit guest mode, return to GuestIdle or disconnect)
- View House Rules (read-only)
- Call Owner (optional)

**Guest Cannot:**
- Re-enter (must request again)
- Access alarm
- Return to approved state

#### Owner Notifications
```
Owner View Shows:
- Guest has exited
- Time of exit
- Alarm re-armed status
- Duration of guest visit

Owner Can:
- View guest activity history
- Rate guest experience (optional)
```

#### Data Structure
```json
{
  "state": "guest_exit",
  "message": "You have exited the property",
  "context": "Thank you for visiting. The alarm has been re-armed.",
  "exit": {
    "exited_at": "2026-01-04T10:50:00Z",
    "duration_minutes": 20,
    "approved_until": "2026-01-04T11:00:45Z",
    "alarm_status_before": "armed",
    "alarm_status_now": "armed"
  },
  "timestamp": "2026-01-04T10:50:00Z",
  "guest_id": "guest_123",
  "request_id": "req_abc123",
  "actions": [
    {"id": "disconnect", "label": "Disconnect", "enabled": true},
    {"id": "rules", "label": "House Rules", "enabled": true}
  ],
  "owner_notification": {
    "title": "Guest has exited",
    "status": "exited",
    "visit_duration_minutes": 20,
    "alarm_status": "re-armed"
  },
  "info": {
    "visit_duration_seconds": 1200,
    "alarm_restored": true
  }
}
```

#### Accessibility Variants
- **reduced_motion:** Static display, no animation
- **large_text:** Clear exit message, large text
- **high_contrast:** Blue indicator for "exit" state

#### Voice Variant (Optional FAZ 81)
```
Voice.SpeakInfo("You have exited the property. Thank you for visiting.")
(On state entry)
```

#### i18n Keys
```
guest.state.exit = "Exited"
guest.message.exit = "You have exited the property"
guest.context.exit = "Thank you for visiting. The alarm has been re-armed."
guest.context.exit.info = "Guest access has ended and the alarm is re-armed."
guest.action.disconnect = "Disconnect"
guest.info.exit = "Your visit has concluded."
guest.exit.duration = "You were inside for {minutes} minutes"
guest.exit.alarm_rearmed = "The alarm has been re-armed to its previous state."
guest.notification.exited = "Guest has exited"
guest.thank_you = "Thank you for visiting!"
```

---

## Request Timeout & Countdown Behavior (Detailed)

### Timeout State Machine

The request timeout is a sub-state within "GuestRequesting" mode:

```
GuestRequesting Countdown Sequence:
60s: "Waiting for approval. 60 seconds remaining."
59s: "Waiting for approval. 59 seconds remaining."
...
2s: "Waiting for approval. 2 seconds remaining."
1s: "Waiting for approval. 1 second remaining."
0s: Auto-transition to GuestExpired state
```

### Countdown Data Contract

```json
{
  "countdown": {
    "total_seconds": 60,
    "remaining_seconds": 45,
    "percentage": 75,
    "started_at": "2026-01-04T10:30:00Z",
    "will_expire_at": "2026-01-04T10:31:00Z"
  }
}
```

### Countdown Display (No Animation)

**Standard Users:**
```
"Waiting for approval. 45 seconds remaining."
Updates: Every second via polling
Format: Clear number display
```

**Reduced Motion Users (FAZ 80):**
```
"Waiting for approval. Please wait."
Display: Static text, minimal updates
Updates: Once per second, no animation
Alternative: "45 seconds remaining"
```

### Owner Dashboard Countdown

**Owner View Shows:**
```
"Guest requesting entry"
"45 seconds remaining"
[Approve Button] [Deny Button]
```

**Owner Can Decide At Any Time:**
- Before countdown ends: Click Approve/Deny (immediate transition)
- At countdown end: Auto-transitions to GuestExpired (no action needed)

---

## Approval/Denial Behavior (Detailed)

### Approval Sequence

```
Owner Dashboard:
1. Sees "Guest requesting entry"
2. Clicks [Approve]
3. Server receives approval action
4. Sets guest state to guest_approved
5. Disarms alarm (for guest entry)
6. Logs: "INFO guest: request approved for {guest_id}"
7. Notifies owner: "Guest approved - access active"
8. Notifies guest: "Welcome! Access approved."
9. Sets expiration timer (30 min default)
```

**Guest Timeline:**
```
Before: GuestRequesting (waiting)
↓ [Approve clicked by owner]
↓
Now: GuestApproved (access active)
Display: "Welcome! Access approved."
Alarm: Disarmed
Countdown: Time until auto-expiry (30 min)
```

### Denial Sequence

```
Owner Dashboard:
1. Sees "Guest requesting entry"
2. Clicks [Deny]
3. Server receives denial action
4. Sets guest state to guest_denied
5. Logs: "INFO guest: request denied for {guest_id}"
6. Notifies owner: "Guest request denied"
7. Notifies guest: "Access denied"
8. Alarm: Remains in previous state (no change)
```

**Guest Timeline:**
```
Before: GuestRequesting (waiting)
↓ [Deny clicked by owner]
↓
Now: GuestDenied (not permitted)
Display: "Access denied"
Alarm: Unchanged (still armed if was armed)
Actions: Can try again, call owner, disconnect
```

### Voice Integration (Optional FAZ 81)

```
On Approval:
Voice.SpeakInfo("Guest access approved. Alarm disarmed.")

On Denial:
Voice.SpeakInfo("Guest request denied.")

On Expiration:
Voice.SpeakInfo("Request expired. Please try again.")
```

---

## Exit Behavior (Detailed)

### Exit Sequence

```
Guest View (Approved State):
1. Sees "Welcome! Access approved."
2. Sees "Exit" button
3. Clicks "Exit"
4. Action: POST /api/ui/guest/exit
5. Server receives exit action
6. Sets guest state to guest_exit
7. Re-arms alarm to previous mode
8. Logs: "INFO guest: guest exited manually"
9. Notifies owner: "Guest has exited"
10. Shows guest exit confirmation
```

**Guest Timeline:**
```
Before: GuestApproved (inside)
↓ [Guest clicks Exit]
↓
Now: GuestExit (departed)
Display: "You have exited the property. Thank you for visiting."
Alarm: Re-armed to previous state
Actions: Can disconnect, view rules, call owner
```

### Alarm Recovery on Exit

```
Important Rules:
- If alarm was Armed before approval → Re-arm to Armed
- If alarm was Disarmed before approval → Keep Disarmed
- If alarm was Arming before approval → Resume Arming from beginning
- If alarm was Triggered before approval → Return to Triggered
- If alarm was Blocked before approval → Return to Blocked (if condition still true)
```

**Example Scenarios:**

**Scenario 1: Owner armed, guest approved**
```
1. Alarm is Armed
2. Guest requests and is approved
3. Alarm disarms (for guest entry)
4. Guest enters and does things
5. Guest exits
6. Alarm returns to Armed state ✓
```

**Scenario 2: Owner disarmed (no protection needed), guest approved**
```
1. Alarm is Disarmed
2. Guest requests and is approved
3. Alarm stays Disarmed (or stays Disarmed)
4. Guest enters and does things
5. Guest exits
6. Alarm returns to Disarmed state ✓
```

**Scenario 3: Owner in middle of arming countdown, guest approved**
```
1. Alarm is Arming (15s remaining)
2. Guest requests and is approved
3. Alarm stops countdown, disarms
4. Guest enters and does things
5. Guest exits
6. Alarm restarts arming countdown from 30s ✓
```

---

## API Contracts

### GET /api/ui/guest/state

**Purpose:** Full guest access state with all details  
**Auth:** Required (guest or authenticated owner)  
**Role-Based:** Guest sees own request; owner sees all guests  
**Response:** Complete guest state object

**Example Response (GuestIdle):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_idle",
    "message": "Welcome. Request access to enter?",
    "context": "Your request will be sent to the property owner.",
    "timestamp": "2026-01-04T10:30:00Z",
    "guest_id": "guest_123",
    "actions": [
      {"id": "request", "label": "Request Access", "enabled": true},
      {"id": "rules", "label": "House Rules", "enabled": true}
    ],
    "info": {
      "can_request": true,
      "reason_blocked": null,
      "last_request": null
    }
  }
}
```

**Example Response (GuestRequesting):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_requesting",
    "message": "Request sent to owner",
    "context": "Waiting for approval. 45 seconds remaining.",
    "countdown": {
      "total_seconds": 60,
      "remaining_seconds": 45,
      "percentage": 75,
      "started_at": "2026-01-04T10:30:00Z",
      "will_expire_at": "2026-01-04T10:31:00Z"
    },
    "timestamp": "2026-01-04T10:30:15Z",
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions": [],
    "owner_notification": {
      "title": "Guest requesting entry",
      "pending": true,
      "time_remaining_seconds": 45
    },
    "info": {
      "can_cancel": false,
      "auto_expire_at": "2026-01-04T10:31:00Z",
      "approval_pending": true
    }
  }
}
```

**Example Response (GuestApproved):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_approved",
    "message": "Welcome! Access approved.",
    "context": "Your access is active until {time}.",
    "approval": {
      "approved_at": "2026-01-04T10:30:45Z",
      "expires_at": "2026-01-04T11:00:45Z",
      "duration_minutes": 30,
      "time_remaining_seconds": 1800
    },
    "timestamp": "2026-01-04T10:30:45Z",
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions": [
      {"id": "exit", "label": "Exit", "enabled": true},
      {"id": "rules", "label": "House Rules", "enabled": true},
      {"id": "disarm", "label": "Disarm Alarm", "enabled": true}
    ],
    "owner_notification": {
      "title": "Guest inside (approved)",
      "active": true,
      "time_remaining_minutes": 30
    },
    "alarm_status": {
      "can_disarm": true,
      "currently_disarmed": true
    },
    "info": {
      "access_active": true,
      "will_expire_at": "2026-01-04T11:00:45Z",
      "approval_duration_minutes": 30
    }
  }
}
```

**Example Response (GuestDenied):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_denied",
    "message": "Access denied",
    "context": "Your request has been denied. Please contact the owner for assistance.",
    "denial": {
      "denied_at": "2026-01-04T10:30:55Z",
      "reason": null,
      "permanent": false,
      "can_retry": false
    },
    "timestamp": "2026-01-04T10:30:55Z",
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions": [
      {"id": "rules", "label": "House Rules", "enabled": true},
      {"id": "call", "label": "Call Owner", "enabled": true},
      {"id": "disconnect", "label": "Disconnect", "enabled": true}
    ],
    "owner_notification": {
      "title": "Guest request denied",
      "status": "denied"
    },
    "info": {
      "can_request_again": false,
      "reason": null
    }
  }
}
```

**Example Response (GuestExpired):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_expired",
    "message": "Request expired",
    "context": "Your request was not answered within the time limit. You may try again later.",
    "expiration": {
      "expired_at": "2026-01-04T10:31:00Z",
      "reason": "timeout",
      "timeout_seconds": 60
    },
    "timestamp": "2026-01-04T10:31:00Z",
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions": [
      {"id": "request_again", "label": "Request Again", "enabled": true},
      {"id": "rules", "label": "House Rules", "enabled": true},
      {"id": "call", "label": "Call Owner", "enabled": true},
      {"id": "disconnect", "label": "Disconnect", "enabled": true}
    ],
    "owner_notification": {
      "title": "Guest request expired (no action)",
      "status": "expired"
    },
    "info": {
      "can_retry": true,
      "time_until_retry_allowed": 0
    }
  }
}
```

**Example Response (GuestExit):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_exit",
    "message": "You have exited the property",
    "context": "Thank you for visiting. The alarm has been re-armed.",
    "exit": {
      "exited_at": "2026-01-04T10:50:00Z",
      "duration_minutes": 20,
      "approved_until": "2026-01-04T11:00:45Z",
      "alarm_status_before": "armed",
      "alarm_status_now": "armed"
    },
    "timestamp": "2026-01-04T10:50:00Z",
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions": [
      {"id": "disconnect", "label": "Disconnect", "enabled": true},
      {"id": "rules", "label": "House Rules", "enabled": true}
    ],
    "owner_notification": {
      "title": "Guest has exited",
      "status": "exited",
      "visit_duration_minutes": 20,
      "alarm_status": "re-armed"
    },
    "info": {
      "visit_duration_seconds": 1200,
      "alarm_restored": true
    }
  }
}
```

### GET /api/ui/guest/summary

**Purpose:** Lightweight endpoint for frequent polling  
**Auth:** Required (guest or authenticated owner)  
**Response:** Key metrics only

**Response Structure:**
```json
{
  "ok": true,
  "data": {
    "state": "guest_idle|requesting|approved|denied|expired|exit",
    "message": "string (short)",
    "context": "string (short)",
    "countdown_remaining_seconds": 45,
    "approved_remaining_minutes": 30,
    "guest_id": "string",
    "request_id": "string",
    "actions_available": 2,
    "alarm_disarmed_by_guest": boolean
  }
}
```

**Example (Idle):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_idle",
    "message": "Welcome. Request access?",
    "context": "Send request to owner",
    "countdown_remaining_seconds": null,
    "approved_remaining_minutes": null,
    "guest_id": "guest_123",
    "request_id": null,
    "actions_available": 2,
    "alarm_disarmed_by_guest": false
  }
}
```

**Example (Requesting):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_requesting",
    "message": "Request sent to owner",
    "context": "Waiting for approval",
    "countdown_remaining_seconds": 45,
    "approved_remaining_minutes": null,
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions_available": 0,
    "alarm_disarmed_by_guest": false
  }
}
```

**Example (Approved):**
```json
{
  "ok": true,
  "data": {
    "state": "guest_approved",
    "message": "Welcome! Access approved.",
    "context": "Active until {time}",
    "countdown_remaining_seconds": null,
    "approved_remaining_minutes": 30,
    "guest_id": "guest_123",
    "request_id": "req_abc123",
    "actions_available": 3,
    "alarm_disarmed_by_guest": true
  }
}
```

---

## Localization Keys (i18n)

### State Names
```
guest.state.idle = "Guest Idle"
guest.state.requesting = "Requesting Access"
guest.state.approved = "Access Approved"
guest.state.denied = "Access Denied"
guest.state.expired = "Request Expired"
guest.state.exit = "Exited"
```

### Messages (Primary - by State)
```
guest.message.idle = "Welcome. Request access to enter?"
guest.message.requesting = "Request sent to owner"
guest.message.approved = "Welcome! Access approved."
guest.message.denied = "Access denied"
guest.message.expired = "Request expired"
guest.message.exit = "You have exited the property"
```

### Context (Secondary Messages - by State)
```
guest.context.idle = "Your request will be sent to the property owner."
guest.context.requesting = "Waiting for approval. {remaining} seconds remaining."
guest.context.approved = "Your access is active until {time}."
guest.context.denied = "Your request has been denied. Please contact the owner for assistance."
guest.context.expired = "Your request was not answered within the time limit. You may try again later."
guest.context.exit = "Thank you for visiting. The alarm has been re-armed."
```

### Countdown
```
guest.countdown.seconds = "{remaining} seconds remaining"
guest.countdown.second = "1 second remaining"
guest.countdown.zero = "0 seconds - Request expired"
guest.countdown.owner = "Owner has {remaining} seconds to review request"
```

### Actions
```
guest.action.request = "Request Access"
guest.action.rules = "House Rules"
guest.action.call = "Call Owner"
guest.action.exit = "Exit"
guest.action.disarm = "Disarm Alarm"
guest.action.disconnect = "Disconnect"
guest.action.request_again = "Request Again"
```

### Approval/Denial
```
guest.approval.approved = "Your request has been approved"
guest.approval.denied = "Your request has been denied"
guest.approval.expired = "Your request has expired"
guest.approval.owner_button_approve = "Approve"
guest.approval.owner_button_deny = "Deny"
```

### Info Messages
```
guest.info.idle = "Request access to enter the property."
guest.info.requesting = "Your request is pending owner approval."
guest.info.approved = "You have temporary access to the property."
guest.info.denied = "You do not have access to the property."
guest.info.expired = "You may request access again."
guest.info.exit = "Your visit has concluded."
guest.access_duration = "Your access is active for {minutes} minutes"
guest.time_remaining = "{minutes} minutes remaining"
guest.visit_duration = "You were inside for {minutes} minutes"
```

### Owner Notifications
```
guest.notification.idle = "No pending guest requests"
guest.notification.requesting = "Guest requesting entry"
guest.notification.approved = "Guest inside (approved)"
guest.notification.denied = "Guest request denied"
guest.notification.expired = "Guest request expired (no action)"
guest.notification.exited = "Guest has exited"
guest.notification.time_waiting = "Waiting for {time} seconds"
guest.notification.time_approved = "Access expires in {time} minutes"
```

### Exit Messages
```
guest.exit.message = "You have exited the property"
guest.exit.thank_you = "Thank you for visiting!"
guest.exit.alarm_rearmed = "The alarm has been re-armed to its previous state."
guest.exit.duration = "You were inside for {minutes} minutes"
```

### Call to Action
```
guest.request_help = "For assistance, call the property owner."
guest.retry_later = "You can submit a new request later."
guest.contact_owner = "Contact the owner for more information."
```

**Total i18n Keys:** ~70+ organized by component and context

---

## Accessibility Integration

### For `reduced_motion` Users (FAZ 80)

**Countdown Behavior:**
```
❌ NO: Animated counting, visual pulse
✅ YES: Static number display
Display: "Waiting for approval. 45 seconds remaining" (no animation)
Updates: Every second, but no visual animation
```

**State Changes:**
```
❌ NO: Sliding transitions, fade-in effects
✅ YES: Instant display change
```

**Approval Display:**
```
❌ NO: Pulsing approval indicator
✅ YES: Static "Approved" indicator
```

### For `large_text` Users (FAZ 80)

**Font Sizes:**
```
State Message: 28pt (very large)
Context: 18pt (readable)
Countdown: 32pt (very prominent)
Actions: 20pt+ (easy to tap)
```

**Simplified Text:**
```
Standard: "Waiting for approval. {remaining} seconds remaining."
Large Text: "Waiting. {remaining}s"

Standard: "Your access is active until {time}."
Large Text: "Access: {time}"
```

**Spacing:**
```
More whitespace between elements
Clear separations
Large buttons, easy to tap
```

### For `high_contrast` Users (FAZ 80)

**Color Scheme:**
```
Idle: Blue indicator (ready to request)
Requesting: Yellow/Orange indicator (waiting)
Approved: Green indicator (access granted)
Denied: Red indicator (not permitted)
Expired: Gray indicator (request expired)
Exit: Purple/Blue indicator (concluded)

WCAG AA Compliant: 4.5:1 minimum contrast
```

**Text & Borders:**
```
Use bold fonts for state names
Clear borders around action buttons
High contrast between background and text
Distinct colors for each state
```

---

## Logging and Audit Strategy

### INFO Level (Normal Operations)

```
INFO guest: state initialized (guest_id: {id})
Example: INFO guest: state initialized (guest_id: guest_123)

INFO guest: request submitted (guest_id: {id}, request_id: {req_id})
Example: INFO guest: request submitted (guest_id: guest_123, request_id: req_abc123)

INFO guest: request approved (guest_id: {id}, approved_for_minutes: {minutes})
Example: INFO guest: request approved (guest_id: guest_123, approved_for_minutes: 30)

INFO guest: request denied (guest_id: {id})
Example: INFO guest: request denied (guest_id: guest_123)

INFO guest: request expired (guest_id: {id}, timeout_seconds: {seconds})
Example: INFO guest: request expired (guest_id: guest_123, timeout_seconds: 60)

INFO guest: guest exited (guest_id: {id}, visit_duration_seconds: {seconds})
Example: INFO guest: guest exited (guest_id: guest_123, visit_duration_seconds: 1200)

INFO guest: alarm disarmed by approved guest (guest_id: {id})
Example: INFO guest: alarm disarmed by approved guest (guest_id: guest_123)

INFO guest: approval auto-expired (guest_id: {id})
Example: INFO guest: approval auto-expired (guest_id: guest_123)
```

### WARN Level (Important Events)

```
WARN guest: request timeout reached (guest_id: {id})
Example: WARN guest: request timeout reached (guest_id: guest_123)

WARN guest: access revoked by owner (guest_id: {id})
Example: WARN guest: access revoked by owner (guest_id: guest_123)

WARN guest: attempted disarm outside approved state (guest_id: {id})
Example: WARN guest: attempted disarm outside approved state (guest_id: guest_123)
```

### What NOT to Log

```
❌ Guest names or personal info
❌ Guest addresses or contact details
❌ Alarm codes or secret sequences
❌ HA tokens or configuration
❌ Full API request bodies
❌ Every countdown tick (too noisy)
❌ Summary polling calls (too frequent)
```

### Audit Trail

```
Each guest request should have:
- guest_id (unique identifier)
- request_id (unique per request)
- timestamp (when action occurred)
- action (request, approve, deny, expire, exit)
- outcome (success/blocked/timeout)

This allows owner to:
- View guest history
- Audit access patterns
- Identify frequent requesters
- Track visit durations
```

### Log Examples

**Normal Request → Approval → Exit:**
```
INFO guest: state initialized (guest_id: guest_123)
INFO guest: request submitted (guest_id: guest_123, request_id: req_abc123)
INFO guest: request approved (guest_id: guest_123, approved_for_minutes: 30)
INFO guest: alarm disarmed by approved guest (guest_id: guest_123)
INFO guest: guest exited (guest_id: guest_123, visit_duration_seconds: 1200)
```

**Request → Timeout:**
```
INFO guest: state initialized (guest_id: guest_456)
INFO guest: request submitted (guest_id: guest_456, request_id: req_def456)
WARN guest: request timeout reached (guest_id: guest_456)
INFO guest: request expired (guest_id: guest_456, timeout_seconds: 60)
```

**Request → Denial:**
```
INFO guest: state initialized (guest_id: guest_789)
INFO guest: request submitted (guest_id: guest_789, request_id: req_ghi789)
INFO guest: request denied (guest_id: guest_789)
```

---

## Design Principles Applied

**From SmartDisplay Product Principles:**

✅ **Calm:** No anxiety-inducing language, clear status
✅ **Predictable:** Deterministic state machine, no surprises
✅ **Respectful:** Non-judgmental denial messages, maintains dignity
✅ **Protective:** Owner maintains control, can approve/deny
✅ **Accessible:** Variants for reduced_motion, large_text, high_contrast
✅ **Localized:** 70+ i18n keys for English and Turkish
✅ **Voice-Ready:** Optional voice integration for all state changes
✅ **Safe:** Alarm restoration on exit, no data loss

---

## Integration with Previous Phases

### FAZ 79 (Localization)
```
All text through guest.* i18n key namespace
70+ keys for English, Turkish equivalents
Date/time formatting locale-aware
```

### FAZ 80 (Accessibility)
```
reduced_motion: Static countdown, no animation
large_text: Simplified messages, larger fonts
high_contrast: Clear state colors (Blue/Yellow/Green/Red/Gray)
```

### FAZ 81 (Voice Feedback)
```
On request submit: Voice.SpeakInfo("Request sent to owner")
On approval: Voice.SpeakInfo("Access approved. Alarm disarmed.")
On denial: Voice.SpeakInfo("Request denied.")
On exit: Voice.SpeakInfo("Thank you for visiting.")
```

### D0 (First-Boot Flow)
```
Guest requests blocked if first-boot active
Blocked message: "System setup in progress. Try again later."
```

### D3 (Alarm Screen)
```
Approved guest can disarm alarm
Alarm mode changes to disarmed during guest approval
Exit restores alarm to previous mode
```

### D2 (Home Screen)
```
Home screen shows guest approval status
Can navigate to guest flow from home
```

---

## State Transition Diagram

```
                    [Connection]
                        |
                        v
                  [GuestIdle] ←──────────────┐
                   |     ↑                    |
                   |     |                    |
            [Request] |   | [Disconnect]      |
                   |     |                    |
                   v     └─────────────── [GuestExpired]
            [GuestRequesting]          (request again)
                |          ↑
          [Approve] or   [Deny]
         [Timeout]        |
            |     \_______|
            |             |
            v             v
      [GuestApproved]  [GuestDenied]
            |                |
         [Exit]          [Disconnect]
            |                |
            v                v
      [GuestExit]     [GuestIdle] or [Disconnect]


Detailed Transitions:
  GuestIdle --[Request]--> GuestRequesting
  
  GuestRequesting --[Owner Approve]--> GuestApproved
  GuestRequesting --[Owner Deny]--> GuestDenied
  GuestRequesting --[Timeout (60s)]--> GuestExpired
  
  GuestApproved --[Guest Exit]--> GuestExit
  GuestApproved --[Auto-Expire (30min)]--> GuestExpired
  GuestApproved --[Owner Revoke]--> GuestExit
  
  GuestDenied --[Guest Disconnect]--> [End Session]
  GuestExpired --[Guest Request Again]--> GuestRequesting
  GuestExit --[Guest Disconnect]--> [End Session]
```

---

## Key Design Decisions

### 1. Six Distinct States
**Decision:** Separate Idle, Requesting, Approved, Denied, Expired, Exit  
**Rationale:** Each has unique UX needs; clear progression  
**Alternative Rejected:** Combine states (ambiguous)

### 2. Request Timeout
**Decision:** Request auto-expires after 60s if owner doesn't respond  
**Rationale:** Prevents indefinite waiting; respects guest time  
**Alternative Rejected:** Infinite timeout (frustrating for guest)

### 3. Non-Judgmental Denial
**Decision:** Use "Access denied" not "You are not welcome"  
**Rationale:** Respects guest dignity; maintains relationship  
**Alternative Rejected:** Harsh messaging (violates principles)

### 4. Auto-Disarm on Approval
**Decision:** Alarm automatically disarms when guest is approved  
**Rationale:** Enables guest entry; removes friction  
**Alternative Rejected:** Guest manually disarms (confusing)

### 5. Alarm Restoration on Exit
**Decision:** Alarm returns to pre-approval state when guest exits  
**Rationale:** Protects property after guest departure; deterministic  
**Alternative Rejected:** Keep disarmed (safety risk)

### 6. No Approval Extension
**Decision:** Guests cannot extend approval, only request new access  
**Rationale:** Owner maintains control; prevents unauthorized extension  
**Alternative Rejected:** Guest-initiated extension (safety risk)

### 7. Separate Summary Endpoint
**Decision:** GET /api/ui/guest/summary distinct from state  
**Rationale:** UI can poll frequently without full payload  
**Cache:** Acceptable 1-2 sec staleness

### 8. Voice Hook Integration
**Decision:** D4 defines voice contract, FAZ 81 provides execution  
**Rationale:** Scope compliance; backend-driven only  
**No Playback:** Voice logging only

---

## Testing Checklist

✅ All 6 states defined (Idle, Requesting, Approved, Denied, Expired, Exit)  
✅ Primary messages for each state  
✅ Secondary context for each state  
✅ Allowed actions by role (guest vs owner)  
✅ Request timeout behavior specified (60s default)  
✅ Countdown display (standard and reduced_motion)  
✅ Approval sequence defined  
✅ Denial sequence defined  
✅ Expiration sequence defined  
✅ Exit sequence with alarm restoration  
✅ API endpoints documented  
✅ Example JSON responses for all states  
✅ 70+ i18n keys organized  
✅ Accessibility variants for all 3 preferences  
✅ Voice integration path clear  
✅ Logging strategy (INFO/WARN levels, no PII)  
✅ State transition diagram complete  
✅ Design decisions documented  
✅ Product Principles validated  
✅ Integration with FAZ 79-81, D0, D2, D3 confirmed  

---

## Next Implementation Steps

### Phase 1: Data Model
- Define GuestRequest struct in coordinator
- Add guest state tracking to Coordinator
- Implement state machine logic

### Phase 2: API Implementation
- Create GET /api/ui/guest/state endpoint
- Create GET /api/ui/guest/summary endpoint
- Add approval/denial/exit action handlers

### Phase 3: Alarm Integration
- Wire safe disarm on approval
- Implement alarm restoration on exit
- Handle timeout-based expiration

### Phase 4: i18n Integration
- Add guest.* keys to i18n system
- Populate English source text
- Create Turkish translations

### Phase 5: Voice Integration (Optional)
- Wire Voice.Speak() calls to state changes
- Test voice variants

### Phase 6: Owner Dashboard (Future)
- Create owner view for pending requests
- Implement approve/deny buttons
- Show guest history and activity

---

## Future Enhancements (Out of D4 Scope)

1. **Guest Schedule** - Pre-approved guests with time windows (Phase D5?)
2. **Repeated Guests** - Remember frequent visitors (Phase D5?)
3. **Guest Notifications** - Notify guest of approval/denial (Phase D5?)
4. **Access Reasons** - Why is guest visiting? (Phase D6?)
5. **Guest Rating** - Owner rates guest experience (Phase D6?)
6. **PIN Backup** - Alternate entry method if system fails (Phase D7?)
7. **Photo Verification** - Visual guest verification (Phase D7?)
8. **Remote Approval** - Owner approves from anywhere (Phase D8?)

---

## Summary

DESIGN Phase D4 successfully defines:

- ✅ **6 guest flow states** - Idle, Requesting, Approved, Denied, Expired, Exit
- ✅ **State-specific messages** - Primary + secondary context
- ✅ **Request timeout** - 60s default, configurable
- ✅ **Approval behavior** - Safe disarm, alarm restoration
- ✅ **Denial behavior** - Non-judgmental, respects dignity
- ✅ **Exit behavior** - Alarm returns to pre-approval state
- ✅ **API contracts** - Full endpoints with examples
- ✅ **Localization** - 70+ i18n keys organized
- ✅ **Accessibility** - Variants for all 3 FAZ 80 preferences
- ✅ **Voice integration** - Optional voice feedback path
- ✅ **Logging** - Audit trail, no PII, INFO/WARN levels
- ✅ **State machine** - Deterministic transitions

The specification is **ready for implementation** where the API endpoints are created, state machine is implemented in Coordinator, and i18n keys are added.

---

**Status:** ✅ SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
