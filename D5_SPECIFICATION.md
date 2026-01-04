# DESIGN PHASE D5 SPECIFICATION
## Menu Structure & Role-Based Perception Model

**Phase:** DESIGN Phase D5  
**Focus:** Menu structure, role visibility, and role-based perception  
**Date:** January 4, 2026  
**Status:** SPECIFICATION IN PROGRESS

---

## Overview

The menu system is the **backbone of user perception** in SmartDisplay. It defines:
- Which screens users can access
- Which functions are available to their role
- How the interface adapts to system state
- Clear, role-appropriate information hierarchy

**Key Principle:** Users should instantly understand what they can see and do—without confusion, without disabled buttons, without hidden options.

---

## Global Menu Sections

### Section 1: Home ✅

**Purpose:** Dashboard overview, system status at a glance  
**Visibility:** All roles see, different content per role  
**Actionability:** Read-only (info only, no state changes)

**Content Varies By Role:**
```
Admin: Full system summary (home screen state)
User: Personal status, household summary
Guest: Welcome message, access status only
```

**Key Features:**
- Calm, non-overwhelming summary
- Alarm state
- HA connectivity (if available)
- Time/date
- Optional AI insight or notification

**Always Visible:** Yes (even during first-boot)

#### Sub-Items
```
(no sub-items for flat structure)
Can include shortcuts to Alarm, Guest
```

---

### Section 2: Alarm ✅

**Purpose:** Control and monitor alarm state  
**Visibility:** Admin always, User conditional, Guest conditional  
**Actionability:** Admin full control, User read-only, Guest limited disarm

**Content Varies By Role:**
```
Admin: Full alarm control (arm, disarm, view history)
User: View alarm state, cannot arm/disarm
Guest: View alarm state, can disarm if approved (for entry)
```

**Key Features:**
- Current alarm mode (Disarmed, Arming, Armed, Triggered)
- Countdown display (if arming)
- Triggered details (if triggered)
- Alarm history (Admin only)

**Visibility Rules:**
- Always visible for Admin
- Visible for User (no arm/disarm buttons)
- Visible for Guest (only if guest_requesting OR guest_approved)
- Hidden for Guest in Idle mode (if setting disabled)

#### Sub-Items
```
- View Current State (all roles)
- History (Admin only)
- Arm/Disarm (Admin only)
- Acknowledge Alert (Admin/User only)
```

---

### Section 3: Guest ✅

**Purpose:** Manage guest access and view guest status  
**Visibility:** Admin always, User conditional, Guest shows self  
**Actionability:** Admin manages, User views, Guest self-service

**Content Varies By Role:**
```
Admin: Manage guests, approve/deny requests, view history
User: View pending guest requests (no approve/deny)
Guest: View own access status, request entry, exit
```

**Key Features:**
- Pending guest requests (Admin)
- Approve/deny buttons (Admin)
- Guest history (Admin)
- Own access status (Guest)
- Exit button (Guest if approved)

**Visibility Rules:**
- Always visible for Admin
- Visible for User if guest_idle or guest_requesting (info only)
- Visible for Guest always
- Hidden for User if no guests configured

#### Sub-Items
```
- Pending Requests (Admin/User view only)
- Guest History (Admin only)
- Manage Guest (Admin only)
- My Status (Guest only)
- Request Access (Guest in Idle state)
- Exit (Guest in Approved state)
```

---

### Section 4: Devices (Read-Only) ✅

**Purpose:** View connected device status and information  
**Visibility:** Admin always, User sometimes, Guest never  
**Actionability:** Read-only (no control)

**Content Varies By Role:**
```
Admin: Full device list, status, battery levels, last contact
User: Basic device status (no technical details)
Guest: Not visible
```

**Key Features:**
- Device name and type
- Current status (Online/Offline)
- Battery level (if applicable)
- Last contact time
- Optional: signal strength (for diagnostic purposes)

**Visibility Rules:**
- Always visible for Admin
- Visible for User (simplified view)
- Invisible for Guest (no device info)
- Hidden for Admin during failsafe (temporarily, shows status only)

#### Sub-Items
```
- Device List (Admin/User)
- Device Details (Admin only)
- Connectivity Status (Admin/User)
- Battery Status (Admin/User)
- Technical Info (Admin only)
```

---

### Section 5: History / Logbook ✅

**Purpose:** View system activity and event history  
**Visibility:** Admin always, User sometimes, Guest never  
**Actionability:** Read-only (info only, no state changes)

**Content Varies By Role:**
```
Admin: Full system history (all events, all users)
User: Personal activity and system events (filtered)
Guest: Not visible
```

**Key Features:**
- Event timestamp
- Event type (Alarm triggered, Guest approved, Disarmed by User X)
- Details (location, reason, outcome)
- Optional: search and filter

**Visibility Rules:**
- Always visible for Admin
- Visible for User (filtered to safe events)
- Invisible for Guest (no history access)
- Visible during failsafe (shows events from safe period)

#### Sub-Items
```
- System Events (Admin/User)
- Alarm Events (Admin/User)
- Guest Events (Admin/User)
- User Actions (Admin only)
- Search History (Admin only)
```

---

### Section 6: Settings ✅

**Purpose:** Configure system behavior and preferences  
**Visibility:** Admin only, User never, Guest never  
**Actionability:** Admin full control

**Content Varies By Role:**
```
Admin: All settings (alarm timing, notifications, device config, etc.)
User: Not visible
Guest: Not visible
```

**Key Features:**
- Alarm settings (arm delay, sensors, etc.)
- Notification preferences
- User management (add/remove users)
- Device management
- System configuration
- Accessibility preferences (shared)

**Visibility Rules:**
- Always visible for Admin
- Hidden for User
- Hidden for Guest
- Disabled during first-boot (no settings changes until setup complete)
- Disabled during failsafe (temporarily read-only)

#### Sub-Items
```
- Alarm Config (Admin only)
- Notifications (Admin only)
- User Management (Admin only)
- Device Management (Admin only)
- System Settings (Admin only)
- Accessibility (all roles can view/modify their own)
```

---

## User Roles

### Role 1: Admin ✅

**Who:** Property owner, system administrator  
**Default Count:** 1-2 per property  
**Control Level:** FULL

#### Visibility
```
Home: ✓ Full
Alarm: ✓ Full control (arm, disarm, acknowledge)
Guest: ✓ Full management (approve, deny, view history)
Devices: ✓ Full (all devices, all details)
History: ✓ Full (all events, all users)
Settings: ✓ Full (all configuration)
```

#### Permissions
```
- Arm/Disarm alarm
- Manage guest requests
- Manage users
- Configure devices
- Access all history
- Modify settings
- Acknowledge alerts
```

#### Menu Behavior
```
No hidden sections
All sections enabled by default
May be disabled due to first-boot, failsafe (temporary)
Voice notifications available
```

#### Typical Actions
```
- Review home screen
- Check alarm status
- Approve/deny guest requests
- View recent activity
- Adjust settings
- Invite new users
```

---

### Role 2: User ✅

**Who:** Trusted household member (family, roommate)  
**Default Count:** 0-3 per property  
**Control Level:** MEDIUM

#### Visibility
```
Home: ✓ Household view (no admin functions)
Alarm: ✓ View only (cannot arm/disarm)
Guest: ✓ View pending (cannot approve/deny)
Devices: ✓ Basic status (simplified)
History: ✓ Filtered events (no sensitive details)
Settings: ✗ Hidden (not visible at all)
```

#### Permissions
```
- View home status
- View alarm state
- Cannot arm/disarm
- View guest requests (info only)
- View device status (simplified)
- View personal activity
- Cannot configure anything
```

#### Menu Behavior
```
Settings section invisible (not disabled)
History limited to non-sensitive events
Devices show basic status only
Cannot see admin actions in history
```

#### Typical Actions
```
- Check if system is armed
- View home status
- See if guest is visiting
- Check recent personal activity
- View device battery status
```

---

### Role 3: Guest ✅

**Who:** Temporary visitor with temporary access  
**Default Count:** 0-10 per day  
**Control Level:** MINIMAL

#### Visibility
```
Home: ✓ Welcome/status (very limited)
Alarm: ~ Conditional (only if guest_requesting or guest_approved)
Guest: ✓ Self-service (view own status, request access, exit)
Devices: ✗ Hidden (not visible)
History: ✗ Hidden (not visible)
Settings: ✗ Hidden (not visible)
```

#### Permissions
```
- View welcome message
- Request access
- View access status
- Disarm alarm if approved (for entry)
- Exit when done
- View house rules (if available)
```

#### Menu Behavior
```
Minimal, purpose-driven UI
Only 2-3 menu items visible at a time
No admin or settings sections
No access to system information
Cannot see other users or guests
```

#### Typical Actions
```
- Request access
- Wait for owner approval
- Enter (if approved)
- Exit when leaving
- Call owner if needed (optional)
```

---

## Role-Based Visibility Matrix

### By Menu Section

```
Section     | Admin | User  | Guest | Notes
------------|-------|-------|-------|--------------------
Home        |   ✓   |   ✓   |   ✓   | Always visible, content differs
Alarm       |   ✓   |   ✓   |  ~*   | Guest only if active
Guest       |   ✓   |   ✓   |   ✓   | Different views per role
Devices     |   ✓   |   ✓   |   ✗   | Hidden for guest
History     |   ✓   |   ✓   |   ✗   | Hidden for guest
Settings    |   ✓   |   ✗   |   ✗   | Admin only

Legend:
✓ = Always visible and actionable
~ = Conditional (depends on state)
✗ = Never visible to this role
✓* = Visible only if guest is requesting or approved
```

### By Actionability

```
Section     | Admin Action | User Action | Guest Action
------------|--------------|-------------|------------------
Home        | View summary | View summary| View welcome
Alarm       | Full control | View only   | Disarm if approved*
Guest       | Full manage  | View only   | Self-service
Devices     | Full details | Basic view  | No access
History     | Full view    | Filtered    | No access
Settings    | Full config  | No access   | No access

* Guest can disarm only if guest_approved state active
```

---

## Dynamic Visibility Rules

### System State: First-Boot Active (D0)

**When:** System initializing, setup wizard running  
**Duration:** Until setup complete  
**Impact on Menu:**

```
Home:      ✓ Visible (setup progress)
Alarm:     ✓ Visible (read-only, no control)
Guest:     ✗ Hidden (cannot access yet)
Devices:   ✗ Hidden (setup not complete)
History:   ✗ Hidden (system new)
Settings:  ✗ Hidden (cannot modify during setup)

Message: "System setup in progress. Access limited."

Admin can:
- View home (setup status)
- View alarm (read-only)
- Cannot manage guests or change settings
- Cannot configure anything
```

---

### System State: Failsafe Mode (From Alarm)

**When:** System in safe mode (connection lost, power issue, etc.)  
**Duration:** Until recovery  
**Impact on Menu:**

```
Home:      ✓ Visible (shows failsafe status)
Alarm:     ✓ Visible (state read-only, no control)
Guest:     ✗ Hidden (system recovering)
Devices:   ✓ Visible (status only, simplified)
History:   ✓ Visible (up to failsafe event)
Settings:  ✓ Visible (read-only, no changes allowed)

Message: "System in safe mode. Changes disabled."

Admin can:
- View home (recovery progress)
- View alarm (read-only)
- View devices (status only)
- View history (up to failure)
- View settings (read-only)
- Cannot make changes until recovered
```

---

### System State: Guest Requesting (D4)

**When:** Guest has submitted request, waiting for approval  
**Duration:** 60 seconds (or until approved/denied)  
**Impact on Menu:**

```
Admin Menu:
  Home:    ✓ Visible (shows pending guest)
  Alarm:   ✓ Visible (full control)
  Guest:   ✓ Visible (with pending request badge)
  Devices: ✓ Visible
  History: ✓ Visible
  Settings: ✓ Visible

User Menu:
  Home:    ✓ Visible (shows pending guest)
  Alarm:   ✓ Visible (view-only)
  Guest:   ✓ Visible (shows pending request)
  Devices: ✓ Visible
  History: ✓ Visible
  Settings: ✗ Hidden

Guest Menu:
  Home:    ✓ Visible (waiting message)
  Alarm:   ✓ Visible (read-only)
  Guest:   ✓ Visible (request status)
  Devices: ✗ Hidden
  History: ✗ Hidden
  Settings: ✗ Hidden

Visible Changes:
- "Guest" section shows badge: "Requesting"
- Admin/User see pending request in Guest section
- Guest sees countdown timer
```

---

### System State: Guest Approved (D4)

**When:** Guest has been approved, access active  
**Duration:** Until guest exits or approval expires  
**Impact on Menu:**

```
Admin Menu:
  Home:    ✓ Visible (shows guest inside)
  Alarm:   ✓ Visible (full control)
  Guest:   ✓ Visible (with active guest badge)
  Devices: ✓ Visible
  History: ✓ Visible
  Settings: ✓ Visible

User Menu:
  Home:    ✓ Visible (shows guest inside)
  Alarm:   ✓ Visible (view-only)
  Guest:   ✓ Visible (shows active guest)
  Devices: ✓ Visible
  History: ✓ Visible
  Settings: ✗ Hidden

Guest Menu:
  Home:    ✓ Visible (welcome/inside)
  Alarm:   ✓ Visible (can disarm)
  Guest:   ✓ Visible (own status, exit button)
  Devices: ✗ Hidden
  History: ✗ Hidden
  Settings: ✗ Hidden

Visible Changes:
- "Guest" section shows badge: "Approved"
- Guest sees "Exit" button prominently
- Admin/User see countdown to expiry
- Alarm section shows "Disarmed (by guest approval)"
```

---

## Menu Interaction Principles

### Invisible vs Disabled

**Invisible (Hidden):**
```
Section doesn't appear in menu at all
User cannot see it exists
Used for: Roles without permission, system states that block access

Example: Guest never sees Settings section
         (not even as disabled button)
```

**Disabled (Visible but Inactive):**
```
Section appears but is grayed out
User can see it exists but cannot interact
Used for: Temporary state blocks, role limitations that might change

Example: User role sees Alarm section (can understand it exists)
         but cannot arm/disarm (buttons disabled)
         
         Admin sees Settings section during first-boot
         but cannot modify (disabled)
```

**This Specification Uses Invisible:**
```
If a role cannot perform actions in a section,
the section should be invisible (not in menu at all)
This prevents confusion and clutter
```

---

## Menu Structure Examples

### Admin User Menu

```
┌─ HOME ────────────────────────┐
│ • System status overview       │
│ • Quick links to Alarm, Guest  │
└────────────────────────────────┘

┌─ ALARM ───────────────────────┐
│ • Current state (Armed/Disarmed)│
│ • Arm/Disarm buttons           │
│ • View History                 │
│ • Triggered details (if active)│
└────────────────────────────────┘

┌─ GUEST ───────────────────────┐
│ • Pending requests (with badge)│
│ • Approve/Deny buttons         │
│ • Guest history                │
│ • Current guest status (if any)│
└────────────────────────────────┘

┌─ DEVICES ─────────────────────┐
│ • Device list (Online/Offline) │
│ • Battery levels               │
│ • Last contact times           │
│ • Technical details            │
└────────────────────────────────┘

┌─ HISTORY ─────────────────────┐
│ • All system events            │
│ • Filter and search            │
│ • Export (optional)            │
└────────────────────────────────┘

┌─ SETTINGS ────────────────────┐
│ • Alarm configuration          │
│ • User management              │
│ • Device setup                 │
│ • System preferences           │
│ • Accessibility settings       │
└────────────────────────────────┘
```

### User (Trusted Member) Menu

```
┌─ HOME ────────────────────────┐
│ • Household status             │
│ • Quick info                   │
└────────────────────────────────┘

┌─ ALARM ───────────────────────┐
│ • Current state (read-only)    │
│ • Recent events                │
│ (Cannot arm/disarm)            │
└────────────────────────────────┘

┌─ GUEST ───────────────────────┐
│ • Pending requests (view only) │
│ • Guest status (view only)     │
│ (Cannot approve/deny)          │
└────────────────────────────────┘

┌─ DEVICES ─────────────────────┐
│ • Device status (simplified)   │
│ • Battery levels               │
│ (No technical details)         │
└────────────────────────────────┘

┌─ HISTORY ─────────────────────┐
│ • Personal activity            │
│ • Recent system events         │
│ (No sensitive info)            │
└────────────────────────────────┘

[Settings section is completely hidden]
```

### Guest (Temporary Visitor) Menu

```
┌─ HOME ────────────────────────┐
│ • Welcome message              │
│ • Access status                │
│ • Instructions                 │
└────────────────────────────────┘

┌─ ALARM ───────────────────────┐
│ • Alarm state (if approved)    │
│ • Disarm button (if approved)  │
│ (Only visible if guest active) │
└────────────────────────────────┘

┌─ GUEST ───────────────────────┐
│ • My access status             │
│ • Request Access button (idle) │
│ • Exit button (approved)       │
│ • Countdown (if requesting)    │
└────────────────────────────────┘

[Devices section is completely hidden]
[History section is completely hidden]
[Settings section is completely hidden]
```

---

## API Contract

### GET /api/ui/menu

**Purpose:** Get menu structure and visibility for authenticated user  
**Auth:** Required (any authenticated user)  
**Response:** Visible sections with enabled actions

**Response Structure:**
```json
{
  "ok": true,
  "data": {
    "user_id": "user_123",
    "role": "admin|user|guest",
    "first_boot_active": false,
    "failsafe_active": false,
    "guest_active": false,
    "sections": [
      {
        "id": "home",
        "name": "Home",
        "description": "Dashboard overview",
        "visible": true,
        "actions": [
          {"id": "view_summary", "name": "View Summary", "enabled": true},
          {"id": "quick_alarm", "name": "Alarm Status", "enabled": true}
        ],
        "sub_sections": [],
        "reason_hidden": null
      },
      {
        "id": "alarm",
        "name": "Alarm",
        "description": "Control and monitor alarm",
        "visible": true,
        "actions": [
          {"id": "view_state", "name": "View State", "enabled": true},
          {"id": "arm", "name": "Arm", "enabled": true},
          {"id": "disarm", "name": "Disarm", "enabled": true},
          {"id": "view_history", "name": "History", "enabled": true}
        ],
        "sub_sections": [
          {"id": "current_state", "name": "Current State", "visible": true},
          {"id": "history", "name": "History", "visible": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "guest",
        "name": "Guest",
        "description": "Manage guest access",
        "visible": true,
        "actions": [
          {"id": "view_requests", "name": "View Requests", "enabled": true},
          {"id": "approve", "name": "Approve", "enabled": true},
          {"id": "deny", "name": "Deny", "enabled": true}
        ],
        "sub_sections": [
          {"id": "pending", "name": "Pending Requests", "visible": true},
          {"id": "history", "name": "Guest History", "visible": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "devices",
        "name": "Devices",
        "description": "View device status",
        "visible": true,
        "actions": [
          {"id": "view_list", "name": "View List", "enabled": true},
          {"id": "view_details", "name": "Details", "enabled": true}
        ],
        "sub_sections": [
          {"id": "status", "name": "Status", "visible": true},
          {"id": "battery", "name": "Battery Levels", "visible": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "history",
        "name": "History",
        "description": "System activity log",
        "visible": true,
        "actions": [
          {"id": "view_events", "name": "View Events", "enabled": true}
        ],
        "sub_sections": [
          {"id": "system_events", "name": "System Events", "visible": true},
          {"id": "alarm_events", "name": "Alarm Events", "visible": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "settings",
        "name": "Settings",
        "description": "System configuration",
        "visible": true,
        "actions": [
          {"id": "alarm_config", "name": "Alarm Settings", "enabled": true},
          {"id": "user_mgmt", "name": "User Management", "enabled": true},
          {"id": "device_mgmt", "name": "Device Management", "enabled": true}
        ],
        "sub_sections": [
          {"id": "alarm", "name": "Alarm Config", "visible": true},
          {"id": "users", "name": "Users", "visible": true},
          {"id": "devices", "name": "Devices", "visible": true}
        ],
        "reason_hidden": null
      }
    ]
  }
}
```

**Example Response (User Role):**
```json
{
  "ok": true,
  "data": {
    "user_id": "user_456",
    "role": "user",
    "first_boot_active": false,
    "failsafe_active": false,
    "guest_active": false,
    "sections": [
      {
        "id": "home",
        "name": "Home",
        "visible": true,
        "actions": [
          {"id": "view_summary", "name": "View Summary", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "alarm",
        "name": "Alarm",
        "visible": true,
        "actions": [
          {"id": "view_state", "name": "View State", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "guest",
        "name": "Guest",
        "visible": true,
        "actions": [
          {"id": "view_requests", "name": "View Requests", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "devices",
        "name": "Devices",
        "visible": true,
        "actions": [
          {"id": "view_list", "name": "View List", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "history",
        "name": "History",
        "visible": true,
        "actions": [
          {"id": "view_events", "name": "View Events", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "settings",
        "name": "Settings",
        "visible": false,
        "actions": [],
        "reason_hidden": "permission_insufficient"
      }
    ]
  }
}
```

**Example Response (Guest Role):**
```json
{
  "ok": true,
  "data": {
    "user_id": "guest_789",
    "role": "guest",
    "first_boot_active": false,
    "failsafe_active": false,
    "guest_active": true,
    "sections": [
      {
        "id": "home",
        "name": "Home",
        "visible": true,
        "actions": [
          {"id": "view_welcome", "name": "Welcome", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "alarm",
        "name": "Alarm",
        "visible": true,
        "actions": [
          {"id": "view_state", "name": "View State", "enabled": true},
          {"id": "disarm", "name": "Disarm", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "guest",
        "name": "Guest",
        "visible": true,
        "actions": [
          {"id": "view_status", "name": "My Status", "enabled": true},
          {"id": "exit", "name": "Exit", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "devices",
        "name": "Devices",
        "visible": false,
        "actions": [],
        "reason_hidden": "role_insufficient"
      },
      {
        "id": "history",
        "name": "History",
        "visible": false,
        "actions": [],
        "reason_hidden": "role_insufficient"
      },
      {
        "id": "settings",
        "name": "Settings",
        "visible": false,
        "actions": [],
        "reason_hidden": "role_insufficient"
      }
    ]
  }
}
```

**Example Response (First-Boot Active):**
```json
{
  "ok": true,
  "data": {
    "user_id": "admin_123",
    "role": "admin",
    "first_boot_active": true,
    "failsafe_active": false,
    "guest_active": false,
    "sections": [
      {
        "id": "home",
        "name": "Home",
        "visible": true,
        "actions": [
          {"id": "view_setup_status", "name": "Setup Status", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "alarm",
        "name": "Alarm",
        "visible": true,
        "actions": [
          {"id": "view_state", "name": "View State", "enabled": true}
        ],
        "reason_hidden": "first_boot_active"
      },
      {
        "id": "guest",
        "name": "Guest",
        "visible": false,
        "actions": [],
        "reason_hidden": "first_boot_active"
      },
      {
        "id": "devices",
        "name": "Devices",
        "visible": false,
        "actions": [],
        "reason_hidden": "first_boot_active"
      },
      {
        "id": "history",
        "name": "History",
        "visible": false,
        "actions": [],
        "reason_hidden": "first_boot_active"
      },
      {
        "id": "settings",
        "name": "Settings",
        "visible": false,
        "actions": [],
        "reason_hidden": "first_boot_active"
      }
    ]
  }
}
```

**Example Response (Failsafe Active):**
```json
{
  "ok": true,
  "data": {
    "user_id": "admin_123",
    "role": "admin",
    "first_boot_active": false,
    "failsafe_active": true,
    "guest_active": false,
    "sections": [
      {
        "id": "home",
        "name": "Home",
        "visible": true,
        "actions": [
          {"id": "view_failsafe_status", "name": "Recovery Status", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "alarm",
        "name": "Alarm",
        "visible": true,
        "actions": [
          {"id": "view_state", "name": "View State", "enabled": true}
        ],
        "reason_hidden": "failsafe_active"
      },
      {
        "id": "guest",
        "name": "Guest",
        "visible": false,
        "actions": [],
        "reason_hidden": "failsafe_active"
      },
      {
        "id": "devices",
        "name": "Devices",
        "visible": true,
        "actions": [
          {"id": "view_list", "name": "View List", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "history",
        "name": "History",
        "visible": true,
        "actions": [
          {"id": "view_events", "name": "View Events", "enabled": true}
        ],
        "reason_hidden": null
      },
      {
        "id": "settings",
        "name": "Settings",
        "visible": true,
        "actions": [
          {"id": "view_config", "name": "View Settings", "enabled": false}
        ],
        "reason_hidden": null
      }
    ]
  }
}
```

---

## Accessibility Integration

### For `reduced_motion` Users (FAZ 80)

**Menu Behavior:**
```
❌ NO: Auto-expanding menus, slide-in animations
✅ YES: Static menu structure, instant display
```

**Keyboard Navigation:**
```
✓ Tab through menu items (no focus traps)
✓ Enter to activate
✓ Arrow keys work
✓ Clear focus indicator
```

**Structure:**
```
Flat menu preferred (no nested menus)
OR if nesting required: explicit [+] expand button (not hover)
No hover-activated menus
```

---

### For `large_text` Users (FAZ 80)

**Menu Structure:**
```
Font: 18pt+ for menu items
Flat structure: no nesting
Simplified naming: "Alarm" not "Alarm System Management"
Large spacing between items
Large buttons/targets (minimum 44pt)
```

**Example Simplified Names:**
```
Standard: "Alarm System Control and Monitoring"
Large Text: "Alarm"

Standard: "Device Status and Connectivity"
Large Text: "Devices"

Standard: "Historical Activity and Event Logs"
Large Text: "History"
```

---

### For `high_contrast` Users (FAZ 80)

**Visual Separation:**
```
Clear borders around each section
Different background colors if needed
High contrast text (4.5:1 minimum)
Bold fonts for section names
```

**Color Scheme:**
```
Section Indicators:
- Home: Blue
- Alarm: Red/Orange (alert color)
- Guest: Green
- Devices: Gray/Neutral
- History: Purple
- Settings: Dark/Bold

All WCAG AA compliant (4.5:1+)
```

---

## Localization Keys (i18n)

### Section Names
```
menu.section.home = "Home"
menu.section.alarm = "Alarm"
menu.section.guest = "Guest"
menu.section.devices = "Devices"
menu.section.history = "History"
menu.section.settings = "Settings"
```

### Section Descriptions
```
menu.description.home = "Dashboard overview"
menu.description.alarm = "Control and monitor alarm"
menu.description.guest = "Manage guest access"
menu.description.devices = "View device status"
menu.description.history = "System activity log"
menu.description.settings = "System configuration"
```

### Action Labels
```
menu.action.view_summary = "View Summary"
menu.action.arm = "Arm"
menu.action.disarm = "Disarm"
menu.action.view_history = "View History"
menu.action.approve = "Approve"
menu.action.deny = "Deny"
menu.action.view_devices = "View Devices"
menu.action.configure = "Configure"
```

### Sub-Section Names
```
menu.subsection.current_state = "Current State"
menu.subsection.alarm_history = "Alarm History"
menu.subsection.pending_requests = "Pending Requests"
menu.subsection.guest_history = "Guest History"
menu.subsection.device_list = "Device List"
menu.subsection.battery_status = "Battery Status"
menu.subsection.system_events = "System Events"
menu.subsection.alarm_events = "Alarm Events"
menu.subsection.alarm_config = "Alarm Configuration"
menu.subsection.user_management = "User Management"
menu.subsection.device_management = "Device Management"
```

### Role-Specific Messages
```
menu.role.admin = "Administrator"
menu.role.user = "Household Member"
menu.role.guest = "Guest"

menu.message.admin = "Full system control"
menu.message.user = "View and basic control"
menu.message.guest = "Limited temporary access"
```

### Hidden Section Messages (Backend Logging Only)
```
menu.hidden.first_boot_active = "Setup in progress"
menu.hidden.failsafe_active = "System recovering"
menu.hidden.role_insufficient = "Permission required"
menu.hidden.guest_inactive = "Guest access not active"
```

### Voice Variants
```
menu.voice.home = "Go to Home"
menu.voice.alarm = "Go to Alarm"
menu.voice.guest = "Go to Guest"
menu.voice.devices = "Go to Devices"
menu.voice.history = "Go to History"
menu.voice.settings = "Go to Settings"

menu.voice.permission_denied = "You don't have permission to access this section."
```

**Total i18n Keys:** ~50+ organized by section and purpose

---

## Logging Strategy

### Startup

**On Application Start:**
```
INFO menu: initializing menu for {user_id} (role: {role})
Example: INFO menu: initializing menu for admin_123 (role: admin)

INFO menu: menu structure resolved (visible: {count} sections, hidden: {count} sections)
Example: INFO menu: menu structure resolved (visible: 6 sections, hidden: 0 sections)

INFO menu: role-based visibility applied (role: {role})
Example: INFO menu: role-based visibility applied (role: user)

INFO menu: first_boot check: {status}
Example: INFO menu: first_boot check: false (setup complete)

INFO menu: failsafe check: {status}
Example: INFO menu: failsafe check: false (system nominal)
```

### Runtime

**On First Request (or Cache Invalid):**
```
INFO menu: menu retrieved for {user_id} (role: {role}, sections: {visible_count})
Example: INFO menu: menu retrieved for user_456 (role: user, sections: 5)
```

**On State Change:**
```
INFO menu: visibility updated (first_boot: {status})
Example: INFO menu: visibility updated (first_boot: true, sections: 2 visible)

INFO menu: visibility updated (failsafe: {status})
Example: INFO menu: visibility updated (failsafe: true, sections: 4 visible)

INFO menu: visibility updated (guest_state: {status})
Example: INFO menu: visibility updated (guest_state: approved, guest visible)
```

### What NOT to Log

```
❌ Per-click menu access (too noisy)
❌ Summary polling requests (too frequent)
❌ Every action enable/disable (too detailed)
❌ User interaction timing
```

### Log Examples

**Admin User at Startup:**
```
INFO menu: initializing menu for admin_123 (role: admin)
INFO menu: menu structure resolved (visible: 6 sections, hidden: 0 sections)
INFO menu: role-based visibility applied (role: admin)
INFO menu: first_boot check: false (setup complete)
INFO menu: failsafe check: false (system nominal)
```

**User Role at Startup:**
```
INFO menu: initializing menu for user_456 (role: user)
INFO menu: menu structure resolved (visible: 5 sections, hidden: 1 sections)
INFO menu: role-based visibility applied (role: user)
INFO menu: hidden section: settings (reason: role_insufficient)
```

**First-Boot Activated:**
```
INFO menu: visibility updated (first_boot: true)
INFO menu: menu structure resolved (visible: 2 sections, hidden: 4 sections)
INFO menu: hidden section: guest (reason: first_boot_active)
INFO menu: hidden section: devices (reason: first_boot_active)
INFO menu: hidden section: history (reason: first_boot_active)
INFO menu: hidden section: settings (reason: first_boot_active)
```

**Guest Approved:**
```
INFO menu: visibility updated (guest_state: approved)
INFO menu: menu structure resolved (visible: 3 sections, hidden: 3 sections)
INFO menu: alarm section visibility: enabled (guest_approved)
```

---

## Design Principles Applied

**From SmartDisplay Product Principles:**

✅ **Calm:** Clear menu structure, no confusion  
✅ **Predictable:** Role-based visibility is consistent  
✅ **Respectful:** Invisible > Disabled (no clutter)  
✅ **Protective:** Sensitive options hidden for lower roles  
✅ **Accessible:** Variants for all 3 preferences  
✅ **Localized:** 50+ i18n keys for all text  
✅ **Voice-Ready:** Menu readable by voice (optional)  

---

## Integration with Previous Phases

### FAZ 79 (Localization)
```
All menu text through menu.* key namespace
50+ keys for English and Turkish
Dynamic text generation from keys
```

### FAZ 80 (Accessibility)
```
reduced_motion: Static menu, no animations
large_text: Flat structure, simplified names, 18pt+ fonts
high_contrast: Color-coded sections, clear borders, 4.5:1 contrast
```

### FAZ 81 (Voice Feedback)
```
Menu structure readable by voice
Optional voice guidance: "Go to Alarm" (Voice.SpeakInfo)
Menu navigation can be voice-driven
```

### D0 (First-Boot Flow)
```
Menu hidden during setup (shows only Home and Alarm read-only)
Settings unavailable until setup complete
```

### D2 (Home Screen)
```
Home section shows dashboard
Links to other sections visible
```

### D3 (Alarm Screen)
```
Alarm section shows alarm control
History link in alarm section
```

### D4 (Guest Access)
```
Guest section shows guest management
Visible for all roles but content differs
Guest state affects menu visibility (requesting/approved)
```

---

## Design Decisions

### 1. Invisible > Disabled
**Decision:** Hidden sections instead of disabled buttons  
**Rationale:** Cleaner UI, no clutter, respects user role  
**Alternative Rejected:** Show all, disable some (confusing)

### 2. Three Distinct Roles
**Decision:** Admin, User, Guest (not per-action permissions)  
**Rationale:** Simple to understand, easy to manage  
**Alternative Rejected:** Complex per-action permissions (too hard to understand)

### 3. Dynamic Menu Based on System State
**Decision:** Menu adapts to first-boot, failsafe, guest state  
**Rationale:** Reflects what user can actually do  
**Alternative Rejected:** Static menu (doesn't reflect reality)

### 4. Separate API for Menu
**Decision:** GET /api/ui/menu separate from other endpoints  
**Rationale:** UI can check permissions before navigating  
**Alternative Rejected:** Inline menu in each response (scattered)

### 5. No Per-Click Logging
**Decision:** Log only startup/state change, not every access  
**Rationale:** Reduces noise, focuses on important events  
**Alternative Rejected:** Log every click (too verbose)

### 6. Flat Menu Structure (Preferred)
**Decision:** No sub-menus for large_text users  
**Rationale:** Simpler to navigate, especially with accessibility  
**Alternative Rejected:** Nested structure (hard to navigate)

---

## Testing Checklist

✅ 6 menu sections defined (Home, Alarm, Guest, Devices, History, Settings)  
✅ 3 user roles defined (Admin, User, Guest)  
✅ Visibility matrix complete (role × section × visibility)  
✅ Dynamic rules for all system states (first-boot, failsafe, guest)  
✅ API endpoint documented with full examples  
✅ 50+ i18n keys organized  
✅ Accessibility variants for all 3 preferences  
✅ Voice integration path defined  
✅ Logging strategy (INFO level, no PII, smart events)  
✅ Menu structure examples for each role  
✅ Design decisions documented  
✅ Product Principles validated  
✅ Integration with all previous phases confirmed  

---

## Next Implementation Steps

### Phase 1: Menu Manager
- Create MenuManager struct in coordinator
- Implement role-based visibility logic
- Add system state checks (first-boot, failsafe, guest)

### Phase 2: API Implementation
- Create GET /api/ui/menu endpoint
- Return visible sections with enabled actions
- Include reason_hidden for logging (backend only)

### Phase 3: i18n Integration
- Add menu.* keys to i18n system
- Populate English text
- Create Turkish translations

### Phase 4: Navigation Integration
- UI checks GET /api/ui/menu before navigation
- Hides unavailable sections
- Disables unavailable actions

### Phase 5: Voice Integration (Optional)
- Wire Voice.SpeakInfo() for menu announcements
- Test menu navigation by voice

---

## Future Enhancements (Out of D5 Scope)

1. **Breadcrumb Navigation** - Show where user is in menu (Phase D6?)
2. **Menu Search** - Find menu items by name (Phase D6?)
3. **Favorites/Shortcuts** - Quick access to frequent items (Phase D6?)
4. **Menu Customization** - User can reorder sections (Phase D7?)
5. **Persistent Menu State** - Remember expanded/collapsed (Phase D7?)
6. **Context Help** - Inline help for each section (Phase D7?)

---

## Summary

DESIGN Phase D5 successfully defines:

- ✅ **6 global menu sections** - Home, Alarm, Guest, Devices, History, Settings
- ✅ **3 user roles** - Admin (full), User (medium), Guest (minimal)
- ✅ **Visibility matrix** - Clear role-based permissions
- ✅ **Dynamic rules** - Menu adapts to system state (first-boot, failsafe, guest)
- ✅ **API contract** - GET /api/ui/menu with full examples
- ✅ **Accessibility** - Variants for all 3 FAZ 80 preferences
- ✅ **Localization** - 50+ i18n keys organized
- ✅ **Voice integration** - Menu readable and navigable by voice
- ✅ **Logging** - Startup and state change events, no per-click noise
- ✅ **Design principles** - Invisible > Disabled, role-based, clear perception

The specification is **ready for implementation** where MenuManager is created, API endpoint is built, and i18n keys are added.

---

**Status:** ✅ SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
