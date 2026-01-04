# D7_SPECIFICATION.md - Settings First-Impression Experience

**Phase:** D7 (Settings UX)  
**Date:** January 4, 2026  
**Scope:** Admin-only settings interface with progressive disclosure, clear warnings, and safe defaults  
**Status:** Specification (Backend-Only)

---

## 1. GOAL & PHILOSOPHY

### 1.1 Core Goal
Define a Settings experience where advanced power does not feel risky. Admin users should feel confident modifying system behavior without fear of irreversible mistakes or hidden consequences.

### 1.2 Design Principles

**Progressive Disclosure**
- Basic settings (language, accessibility) immediately visible
- Advanced settings (backup, factory reset) clearly marked and confirmed
- No "gotcha" surprises; consequences explained before action

**Clear Warnings**
- Only warn when consequences are real and significant
- Avoid warning fatigue; no "warning cry-wolf"
- Warnings explain impact, not just "Are you sure?"

**Calm Authority**
- Settings UX should feel like a trusted control panel, not a minefield
- Simple, flat UI structure (especially for large_text accessibility)
- Clear separation between safe and dangerous actions

---

## 2. SETTINGS SECTIONS

### 2.1 General Settings
**Purpose:** Customize basic behavior and presentation.  
**Access:** Admin (full read/write)  
**Safe to change:** Yes. No confirmation required.

#### Fields:
| Field | Type | Default | Notes |
|-------|------|---------|-------|
| language | enum | "en" | "en", "tr" |
| timezone | string | "UTC" | ISO 8601 format |
| high_contrast | bool | false | From FAZ 80 |
| large_text | bool | false | From FAZ 80 |
| reduced_motion | bool | false | From FAZ 80 |

#### Tone Examples:
- "Language: Choose how SmartDisplay speaks to you"
- "Text Size: Large text for easier reading"
- "High Contrast: Clearer separation between sections"

#### Copy Rule:
All General settings use descriptive labels with brief explanations. No warnings.

---

### 2.2 Security Settings
**Purpose:** Control alarm behavior and access restrictions.  
**Access:** Admin (full read/write)  
**Safe to change:** Yes. Some require confirmation.

#### Fields:
| Field | Type | Default | Notes |
|-------|------|---------|-------|
| alarm_arm_delay_s | int | 30 | Seconds before arming completes |
| alarm_trigger_sound_enabled | bool | true | Play alarm sound on trigger |
| guest_max_active | int | 1 | Max concurrent guest access |
| guest_request_timeout_s | int | 60 | Guest approval request timeout |
| guest_max_requests_per_hour | int | 5 | Rate limit on requests |
| force_ha_connection | bool | false | Block offline operation if true |

#### Confirmation Required:
- Disabling `alarm_trigger_sound_enabled` (silent alarms = risky)
- Setting `guest_max_active` to > 2 (security risk)
- Enabling `force_ha_connection` (breaks offline operation)

#### Tone Examples:
- "Arm Delay: Extra time to cancel alarm after voice confirmation"
- "Guest Requests Per Hour: Prevent access spam"
- "Require Home Assistant: Disable to keep system offline if HA fails"

#### Copy Rule:
Security settings explain impact before showing the toggle. For dangerous changes, add explicit confirmation.

---

### 2.3 System Settings
**Purpose:** Monitor and manage system health.  
**Access:** Admin (read-only for most; write for restart)  
**Safe to change:** Mostly. Restart is dangerous.

#### Fields:
| Field | Type | Writable | Notes |
|-------|------|----------|-------|
| ha_connection_status | enum | No | "connected", "disconnected", "error" |
| ha_last_sync_utc | timestamp | No | Last successful sync time |
| system_uptime_s | int | No | Seconds since boot |
| storage_available_mb | int | No | Free disk space |
| memory_available_mb | int | No | Free RAM |
| cpu_temp_c | float | No | Current CPU temperature (if available) |
| version | string | No | SmartDisplay version |

#### Writable Actions:
- `restart_now` (POST) - Requires confirmation with countdown
- `shutdown_graceful` (POST) - For kiosk mode systems

#### Tone Examples:
- "Home Assistant Status: Connected and synced 2 minutes ago"
- "System Health: All systems normal. No issues detected."
- "Restart: System will restart in 5 seconds. Save any work."

#### Copy Rule:
Health info is informational and calm. Restart action explicitly shows countdown and impact.

---

### 2.4 Advanced Settings
**Purpose:** Backup, restore, and factory reset operations.  
**Access:** Admin only  
**Safe to change:** No. All actions require explicit confirmation and consequence explanation.

#### Actions:
| Action | Impact | Confirmation | Recovery |
|--------|--------|--------------|----------|
| `backup_create` | Creates encrypted backup file | Simple (none required, but explain location) | Download backup |
| `backup_restore` | **Restores ALL settings from backup** | **STRONG** (show timestamp, list what will change) | Can restore from different backup |
| `factory_reset` | **Wipes ALL settings, returns to first-boot** | **VERY STRONG** (countdown, confirm twice) | None. Requires first-boot again. |

#### Backup/Restore Process:
```
1. User clicks "Create Backup"
   → System generates encrypted file: smartdisplay-backup-YYYYMMDD-HHMMSS.json
   → Filename and size shown
   → File ready for download/transfer
   → Log: INFO "Backup created: filename"

2. User clicks "Restore from Backup"
   → Show: Backup date, settings count, alarm arm status
   → Display: "This will restore X settings and overwrite current configuration"
   → Require confirmation: "Restore from this backup?"
   → On confirm: Apply settings, restart system
   → Log: WARN "Backup restored from {filename}, {X} settings applied"
```

#### Factory Reset Process:
```
1. User clicks "Factory Reset"
   → Show RED warning box:
      "This will erase ALL custom settings and return to first-boot."
      "SmartDisplay will restart and show the setup wizard."
      "This action cannot be undone."
   
2. User must:
   → Click "Yes, factory reset"
   → Confirm countdown: "System resets in 10 seconds... 9... 8..."
   → (Optional: Type "RESET" to confirm - prevents accidental confirmation)
   
3. On confirmation:
   → Delete all settings
   → Clear guest state, alarm history
   → Restart system to first-boot
   → Log: WARN "Factory reset initiated by admin"
```

#### Tone Examples:
- Backup: "Backup created: smartdisplay-backup-2026-01-04-15-30-45.json (2.4 MB). Download or transfer to safe location."
- Restore: "Restore from 2026-01-04 backup (created 2 days ago)? This will apply X settings and alarm arm status will change to: Armed."
- Factory Reset: "This erases ALL custom settings and returns system to first-boot. SmartDisplay will restart automatically."

#### Copy Rule:
Dangerous actions MUST explain consequences in clear language. Use "will erase", "cannot be undone", "will restart" (not "may" or "might").

---

### 2.5 Summary Table

| Section | Writable | Confirmation | Reversible |
|---------|----------|--------------|-----------|
| General | Yes | No | Yes (just toggle back) |
| Security | Yes | Some (4 fields) | Yes |
| System | Mostly No | Restart: Yes | Restart: Yes (reboot) |
| Advanced | Yes (3 actions) | All 3 required | Backup/Restore: Yes. Factory: No. |

---

## 3. PROGRESSIVE DISCLOSURE RULES

### 3.1 Visibility Model

```
GET /api/ui/settings response structure:

{
  "sections": [
    {
      "id": "general",
      "order": 1,
      "visibility": "always",
      "fields": [ ... ]
    },
    {
      "id": "security",
      "order": 2,
      "visibility": "always",
      "fields": [ ... ]
    },
    {
      "id": "system",
      "order": 3,
      "visibility": "always",
      "fields": [ ... ]
    },
    {
      "id": "advanced",
      "order": 4,
      "visibility": "always",
      "collapsed": true,  // Start collapsed for reduced UI complexity
      "fields": [ ... ]
    }
  ]
}
```

### 3.2 Collapsed vs Expanded

**Advanced section starts COLLAPSED** because:
- Reduces visual complexity
- Directs attention to safe settings first
- User must actively choose to see dangerous actions
- Prevents accidental clicks on restore/reset

**On Click "Advanced":**
- Expand to show 3 actions (Backup, Restore, Factory Reset)
- Each action has dedicated button with icon (optional)
- Red visual indicator for Factory Reset

### 3.3 No Irreversible Without Confirmation

Every non-reversible action (Backup Restore, Factory Reset) requires:
1. **Explanation** - "What will happen?"
2. **Confirmation** - "Are you sure?"
3. **Safety Delay** - Countdown timer (especially for Factory Reset)
4. **Secondary Confirmation** - Type word or click again (for Factory Reset)

---

## 4. DANGEROUS ACTIONS SPECIFICATION

### 4.1 Backup Restore Confirmation

```
POST /api/ui/settings/action
{
  "action": "backup_restore",
  "backup_id": "smartdisplay-backup-2026-01-04-15-30-45"
}

Confirmation Dialog:
┌─────────────────────────────────────────┐
│ Restore Backup?                         │
├─────────────────────────────────────────┤
│ Backup: 2026-01-04 (2 days ago)         │
│ Settings: 15 items                      │
│                                         │
│ This will restore:                      │
│ • Alarm settings (Armed → Disarmed)     │
│ • Language (English → Turkish)          │
│ • Guest limits (1 → 2)                  │
│ • All other settings from this backup   │
│                                         │
│ Your current settings will be lost.     │
│                                         │
│  [ Cancel ]  [ Restore ]                │
└─────────────────────────────────────────┘

Response (on Restore):
{
  "action": "backup_restore",
  "status": "success",
  "settings_applied": 15,
  "timestamp": "2026-01-04T15:45:32Z",
  "restart_required": true,
  "countdown_s": 10
}
```

**Log Entry (WARN):**
```
"Backup restored: smartdisplay-backup-2026-01-04-15-30-45 (15 settings applied)"
```

---

### 4.2 System Restart Confirmation

```
POST /api/ui/settings/action
{
  "action": "restart_now"
}

Confirmation Dialog:
┌─────────────────────────────────────────┐
│ Restart System?                         │
├─────────────────────────────────────────┤
│ SmartDisplay will restart in 5 seconds. │
│                                         │
│ • Any open requests will be canceled    │
│ • Guests will be logged out             │
│ • Alarm will remain Armed if currently  │
│ • System will be unavailable for ~30s   │
│                                         │
│ Restart in: 5... 4... 3...             │
│                                         │
│  [ Cancel Restart ]                     │
└─────────────────────────────────────────┘

Response (on Confirm):
{
  "action": "restart_now",
  "status": "confirmed",
  "countdown_s": 5,
  "timestamp": "2026-01-04T15:45:32Z"
}
```

**Log Entry (WARN):**
```
"System restart initiated by admin from Settings"
```

---

### 4.3 Factory Reset Confirmation (STRONGEST)

```
POST /api/ui/settings/action
{
  "action": "factory_reset",
  "confirm_type": "double"  // Requires double confirmation
}

Confirmation Dialog (Step 1):
┌─────────────────────────────────────────┐
│ ⚠ FACTORY RESET                         │
├─────────────────────────────────────────┤
│ This will ERASE ALL settings.            │
│                                         │
│ SmartDisplay will:                      │
│ • Reset to first-boot setup wizard      │
│ • Clear ALL custom settings             │
│ • Clear guest access & history          │
│ • Clear alarm configuration             │
│ • Restart automatically                 │
│                                         │
│ This CANNOT BE UNDONE.                  │
│                                         │
│ Type "RESET" to continue:               │
│ [ ______________ ]                      │
│                                         │
│  [ Cancel ]                             │
└─────────────────────────────────────────┘

After typing "RESET" (Step 2 - Countdown):
┌─────────────────────────────────────────┐
│ ⚠ CONFIRM FACTORY RESET                 │
├─────────────────────────────────────────┤
│ System will reset in: 10... 9... 8...   │
│                                         │
│  [ Cancel Reset ]  [ Confirm ]          │
└─────────────────────────────────────────┘

Response (on Confirm):
{
  "action": "factory_reset",
  "status": "confirmed",
  "countdown_s": 10,
  "timestamp": "2026-01-04T15:45:32Z",
  "restart_required": true
}
```

**Log Entry (WARN):**
```
"Factory reset initiated by admin (all settings will be erased on next startup)"
```

---

## 5. ACCESSIBILITY INTEGRATION

### 5.1 Reduced Motion (reduced_motion = true)

**Rule:** No collapsing/expanding animations. No transitions.

```go
Settings Response:
{
  "sections": [
    {
      "id": "advanced",
      "collapsed": true,
      "animation_enabled": false,  // Always false if user has reduced_motion
      "transition_duration_ms": 0
    }
  ]
}
```

**Behavior:**
- Advanced section expands instantly (no slide/fade animation)
- No delay between user click and section visibility
- Section changes appear immediately

---

### 5.2 Large Text (large_text = true)

**Rule:** Flat list structure. No nested collapse/expand.

```
Standard View (large_text = false):
┌─ General
├─ Security
├─ System
└─ Advanced ▼
    ├─ Backup
    ├─ Restore
    └─ Factory Reset

Large Text View (large_text = true):
General
Security
System
Advanced
  Backup
  Restore
  Factory Reset
```

**Changes:**
- No collapse/expand chevrons (just show all sections)
- Linear flow: user scrolls down for advanced
- Font size increased for all labels
- Padding increased for touch targets
- Button height ≥ 48px (touch-friendly)

---

### 5.3 High Contrast (high_contrast = true)

**Rule:** Clear visual separation between sections and action types.

```
General Settings
────────────────────────────
[Section background light gray]
Language:    [dropdown] ┃
Timezone:    [text]     ┃
Text Size:   [toggle]   ┃
────────────────────────────

Security Settings
────────────────────────────
[Section background light gray]
Alarm Delay: [slider]   ┃
Guest Limit: [input]    ┃
────────────────────────────

Dangerous Actions
────────────────────────────
[Section background RED/DARK]
🔴 Factory Reset    ┃ [Red button]
🟡 Restore Backup   ┃ [Orange button]
🟢 Create Backup    ┃ [Green button]
────────────────────────────
```

**Visual Indicators:**
- System (read-only) section: Light blue background
- Security section: Light gray background
- Advanced section (collapsed) header: Dark with right arrow "Advanced ▶"
- Advanced section (expanded) actions: Red/orange/green colored buttons
- Confirmation dialogs: High contrast borders, clear action buttons

---

## 6. API CONTRACTS

### 6.1 GET /api/ui/settings

**Purpose:** Fetch current settings and configuration state.  
**Method:** GET  
**Auth:** Admin only  
**Response:** Complete settings state with accessibility variants applied

```json
{
  "timestamp": "2026-01-04T15:45:32Z",
  "role": "admin",
  "accessibility": {
    "high_contrast": false,
    "large_text": false,
    "reduced_motion": false
  },
  "sections": [
    {
      "id": "general",
      "label": "settings.general.title",
      "description": "settings.general.description",
      "order": 1,
      "visibility": "always",
      "collapsed": false,
      "fields": [
        {
          "id": "language",
          "label": "settings.general.language",
          "type": "select",
          "value": "en",
          "options": [
            { "value": "en", "label": "English" },
            { "value": "tr", "label": "Türkçe" }
          ],
          "confirm_required": false,
          "help": "settings.general.language.help"
        },
        {
          "id": "timezone",
          "label": "settings.general.timezone",
          "type": "text",
          "value": "UTC",
          "confirm_required": false,
          "help": "settings.general.timezone.help"
        },
        {
          "id": "high_contrast",
          "label": "settings.general.high_contrast",
          "type": "toggle",
          "value": false,
          "confirm_required": false,
          "help": "settings.general.high_contrast.help"
        },
        {
          "id": "large_text",
          "label": "settings.general.large_text",
          "type": "toggle",
          "value": false,
          "confirm_required": false,
          "help": "settings.general.large_text.help"
        },
        {
          "id": "reduced_motion",
          "label": "settings.general.reduced_motion",
          "type": "toggle",
          "value": false,
          "confirm_required": false,
          "help": "settings.general.reduced_motion.help"
        }
      ]
    },
    {
      "id": "security",
      "label": "settings.security.title",
      "description": "settings.security.description",
      "order": 2,
      "visibility": "always",
      "collapsed": false,
      "fields": [
        {
          "id": "alarm_arm_delay_s",
          "label": "settings.security.alarm_arm_delay",
          "type": "number",
          "value": 30,
          "min": 5,
          "max": 300,
          "unit": "seconds",
          "confirm_required": false,
          "help": "settings.security.alarm_arm_delay.help"
        },
        {
          "id": "alarm_trigger_sound_enabled",
          "label": "settings.security.alarm_sound",
          "type": "toggle",
          "value": true,
          "confirm_required": true,
          "warning": "settings.security.alarm_sound.warning",
          "help": "settings.security.alarm_sound.help"
        },
        {
          "id": "guest_max_active",
          "label": "settings.security.guest_max_active",
          "type": "number",
          "value": 1,
          "min": 1,
          "max": 5,
          "confirm_required": false,
          "confirm_on_values": [3, 4, 5],
          "warning_on_values": [3, 4, 5],
          "help": "settings.security.guest_max_active.help"
        },
        {
          "id": "guest_request_timeout_s",
          "label": "settings.security.guest_request_timeout",
          "type": "number",
          "value": 60,
          "min": 30,
          "max": 600,
          "unit": "seconds",
          "confirm_required": false,
          "help": "settings.security.guest_request_timeout.help"
        },
        {
          "id": "guest_max_requests_per_hour",
          "label": "settings.security.guest_rate_limit",
          "type": "number",
          "value": 5,
          "min": 1,
          "max": 50,
          "confirm_required": false,
          "help": "settings.security.guest_rate_limit.help"
        },
        {
          "id": "force_ha_connection",
          "label": "settings.security.force_ha",
          "type": "toggle",
          "value": false,
          "confirm_required": true,
          "warning": "settings.security.force_ha.warning",
          "help": "settings.security.force_ha.help"
        }
      ]
    },
    {
      "id": "system",
      "label": "settings.system.title",
      "description": "settings.system.description",
      "order": 3,
      "visibility": "always",
      "collapsed": false,
      "fields": [
        {
          "id": "ha_connection_status",
          "label": "settings.system.ha_status",
          "type": "read_only",
          "value": "connected",
          "value_label": "settings.system.ha_status.connected",
          "help": "settings.system.ha_status.help"
        },
        {
          "id": "ha_last_sync_utc",
          "label": "settings.system.last_sync",
          "type": "read_only",
          "value": "2026-01-04T15:44:00Z",
          "display_format": "2 minutes ago",
          "help": "settings.system.last_sync.help"
        },
        {
          "id": "system_uptime_s",
          "label": "settings.system.uptime",
          "type": "read_only",
          "value": 345600,
          "display_format": "4 days, 0 hours",
          "help": "settings.system.uptime.help"
        },
        {
          "id": "storage_available_mb",
          "label": "settings.system.storage",
          "type": "read_only",
          "value": 1024,
          "display_format": "1024 MB (512 MB used, 20%)",
          "help": "settings.system.storage.help"
        },
        {
          "id": "memory_available_mb",
          "label": "settings.system.memory",
          "type": "read_only",
          "value": 512,
          "display_format": "512 MB available (256 MB used, 33%)",
          "health_status": "normal",
          "help": "settings.system.memory.help"
        },
        {
          "id": "version",
          "label": "settings.system.version",
          "type": "read_only",
          "value": "3.1.0",
          "help": "settings.system.version.help"
        }
      ],
      "actions": [
        {
          "id": "restart_now",
          "label": "settings.system.restart",
          "type": "action",
          "button_style": "warning",
          "confirm_required": true,
          "help": "settings.system.restart.help"
        }
      ]
    },
    {
      "id": "advanced",
      "label": "settings.advanced.title",
      "description": "settings.advanced.description",
      "order": 4,
      "visibility": "always",
      "collapsed": true,
      "animation_enabled": true,
      "actions": [
        {
          "id": "backup_create",
          "label": "settings.advanced.backup_create",
          "type": "action",
          "button_style": "normal",
          "confirm_required": false,
          "help": "settings.advanced.backup_create.help"
        },
        {
          "id": "backup_restore",
          "label": "settings.advanced.backup_restore",
          "type": "action",
          "button_style": "warning",
          "confirm_required": true,
          "confirm_strength": "strong",
          "help": "settings.advanced.backup_restore.help",
          "available_backups": [
            {
              "id": "smartdisplay-backup-2026-01-04-15-30-45",
              "timestamp": "2026-01-04T15:30:45Z",
              "display_name": "2026-01-04 (2 days ago)",
              "size_mb": 2.4,
              "settings_count": 15
            },
            {
              "id": "smartdisplay-backup-2026-01-02-10-15-20",
              "timestamp": "2026-01-02T10:15:20Z",
              "display_name": "2026-01-02 (4 days ago)",
              "size_mb": 2.3,
              "settings_count": 14
            }
          ]
        },
        {
          "id": "factory_reset",
          "label": "settings.advanced.factory_reset",
          "type": "action",
          "button_style": "danger",
          "confirm_required": true,
          "confirm_strength": "very_strong",
          "confirm_type": "double",
          "help": "settings.advanced.factory_reset.help",
          "danger_notice": "settings.advanced.factory_reset.danger"
        }
      ]
    }
  ]
}
```

---

### 6.2 POST /api/ui/settings/action

**Purpose:** Apply a settings change or perform an action.  
**Method:** POST  
**Auth:** Admin only  
**Request Body:** Action with confirmation context

#### Example 1: Change Language

```json
{
  "action": "field_change",
  "field_id": "language",
  "new_value": "tr",
  "confirm": true
}

Response:
{
  "action": "field_change",
  "field_id": "language",
  "status": "success",
  "old_value": "en",
  "new_value": "tr",
  "timestamp": "2026-01-04T15:45:32Z",
  "requires_restart": false,
  "log_entry": "Language changed from English to Turkish"
}
```

#### Example 2: Disable Alarm Sound (Requires Confirmation)

```json
{
  "action": "field_change",
  "field_id": "alarm_trigger_sound_enabled",
  "new_value": false,
  "confirm": true,
  "confirm_dialog": true  // User saw and clicked confirmation
}

Response:
{
  "action": "field_change",
  "field_id": "alarm_trigger_sound_enabled",
  "status": "success",
  "old_value": true,
  "new_value": false,
  "timestamp": "2026-01-04T15:45:32Z",
  "requires_restart": false,
  "warning": "Alarm will trigger silently. Enable sound to restore audible alerts.",
  "log_entry": "Alarm sound disabled (alarm will trigger silently)"
}
```

#### Example 3: Create Backup

```json
{
  "action": "backup_create"
}

Response:
{
  "action": "backup_create",
  "status": "success",
  "backup_id": "smartdisplay-backup-2026-01-04-15-45-32",
  "timestamp": "2026-01-04T15:45:32Z",
  "size_mb": 2.4,
  "location": "/data/backups/smartdisplay-backup-2026-01-04-15-45-32.json",
  "download_url": "/api/backups/smartdisplay-backup-2026-01-04-15-45-32.json",
  "message": "settings.advanced.backup_create.success",
  "log_entry": "Backup created: smartdisplay-backup-2026-01-04-15-45-32 (2.4 MB)"
}
```

#### Example 4: Restore Backup

```json
{
  "action": "backup_restore",
  "backup_id": "smartdisplay-backup-2026-01-04-15-30-45",
  "confirm": true
}

Response:
{
  "action": "backup_restore",
  "status": "success",
  "backup_id": "smartdisplay-backup-2026-01-04-15-30-45",
  "timestamp": "2026-01-04T15:45:32Z",
  "settings_applied": 15,
  "changes": [
    "language: en → tr",
    "guest_max_active: 1 → 2",
    "alarm_arm_delay_s: 30 → 45"
  ],
  "requires_restart": true,
  "countdown_s": 10,
  "message": "settings.advanced.backup_restore.success",
  "log_entry": "Backup restored from 2026-01-04 (15 settings applied)"
}
```

#### Example 5: Factory Reset

```json
{
  "action": "factory_reset",
  "confirm_type": "double",
  "confirm_text": "RESET",
  "confirm": true
}

Response:
{
  "action": "factory_reset",
  "status": "success",
  "timestamp": "2026-01-04T15:45:32Z",
  "countdown_s": 10,
  "requires_restart": true,
  "message": "settings.advanced.factory_reset.success",
  "log_entry": "Factory reset initiated by admin (all settings will be erased on startup)"
}
```

---

## 7. LOCALIZATION (i18n)

### 7.1 i18n Key Namespace: settings.*

All settings-related strings follow the `settings.{section}.{field}.{part}` pattern.

### 7.2 General Section Keys (8 keys)

```yaml
settings.general.title: "General Settings"
settings.general.description: "Basic preferences and accessibility"

settings.general.language: "Language"
settings.general.language.help: "Choose how SmartDisplay speaks to you"

settings.general.timezone: "Time Zone"
settings.general.timezone.help: "Set your local time zone (ISO 8601)"

settings.general.high_contrast: "High Contrast"
settings.general.high_contrast.help: "Clearer separation between sections"

settings.general.large_text: "Large Text"
settings.general.large_text.help: "Larger text for easier reading"

settings.general.reduced_motion: "Reduced Motion"
settings.general.reduced_motion.help: "No animations or transitions"
```

### 7.3 Security Section Keys (16 keys)

```yaml
settings.security.title: "Security"
settings.security.description: "Alarm and access control"

settings.security.alarm_arm_delay: "Arm Delay"
settings.security.alarm_arm_delay.help: "Extra time to cancel alarm after voice confirmation"

settings.security.alarm_sound: "Alarm Sound"
settings.security.alarm_sound.help: "Play audio alert when alarm triggers"
settings.security.alarm_sound.warning: "Disabling sound means alarm will trigger silently. Enable sound again to restore audible alerts."

settings.security.guest_max_active: "Max Active Guests"
settings.security.guest_max_active.help: "Maximum number of guests with access at the same time"
settings.security.guest_max_active.warning: "Allowing more than 2 concurrent guests increases access risk"

settings.security.guest_request_timeout: "Guest Request Timeout"
settings.security.guest_request_timeout.help: "Seconds to wait for approval before expiring guest request"

settings.security.guest_rate_limit: "Guest Requests Per Hour"
settings.security.guest_rate_limit.help: "Prevent spam by limiting guest requests"

settings.security.force_ha: "Require Home Assistant"
settings.security.force_ha.help: "Disable to allow offline operation if Home Assistant fails"
settings.security.force_ha.warning: "SmartDisplay will stop working if Home Assistant becomes unavailable and this is enabled"
```

### 7.4 System Section Keys (14 keys)

```yaml
settings.system.title: "System"
settings.system.description: "Health and status information"

settings.system.ha_status: "Home Assistant Status"
settings.system.ha_status.help: "Current connection status to Home Assistant"
settings.system.ha_status.connected: "Connected"
settings.system.ha_status.disconnected: "Disconnected"
settings.system.ha_status.error: "Connection Error"

settings.system.last_sync: "Last Sync"
settings.system.last_sync.help: "When SmartDisplay last synchronized with Home Assistant"

settings.system.uptime: "System Uptime"
settings.system.uptime.help: "How long SmartDisplay has been running"

settings.system.storage: "Storage"
settings.system.storage.help: "Available disk space for backups and logs"

settings.system.memory: "Memory"
settings.system.memory.help: "Available RAM for system operations"

settings.system.version: "Version"
settings.system.version.help: "Current SmartDisplay software version"

settings.system.restart: "Restart System"
settings.system.restart.help: "Restart SmartDisplay. Any active requests will be canceled."
```

### 7.5 Advanced Section Keys (18 keys)

```yaml
settings.advanced.title: "Advanced"
settings.advanced.description: "Backup, restore, and system reset"

settings.advanced.backup_create: "Create Backup"
settings.advanced.backup_create.help: "Create an encrypted backup of all settings"
settings.advanced.backup_create.success: "Backup created. Download or transfer to a safe location."

settings.advanced.backup_restore: "Restore Backup"
settings.advanced.backup_restore.help: "Restore all settings from a previous backup"
settings.advanced.backup_restore.success: "Backup restored. System restarting..."

settings.advanced.factory_reset: "Factory Reset"
settings.advanced.factory_reset.help: "Erase all settings and return to first-boot setup"
settings.advanced.factory_reset.danger: "This will erase ALL custom settings. This action cannot be undone."
settings.advanced.factory_reset.success: "Factory reset initiated. System restarting..."
settings.advanced.factory_reset.confirm: "Type 'RESET' to confirm factory reset"
settings.advanced.factory_reset.countdown: "System will reset in {{seconds}} seconds"
```

### 7.6 Action Confirmation Dialog Keys (8 keys)

```yaml
settings.action.confirm_title: "Confirm Action"
settings.action.confirm_button: "Confirm"
settings.action.cancel_button: "Cancel"
settings.action.typing_instruction: "Type '{{word}}' to confirm"
settings.action.countdown_message: "Continuing in {{seconds}} seconds"

settings.action.restart_confirm: "Restart System?"
settings.action.restore_confirm: "Restore Backup?"
settings.action.reset_confirm: "Factory Reset?"
```

### 7.7 Summary: 64 English Keys (all sections)

| Section | Keys | Examples |
|---------|------|----------|
| General | 8 | language, timezone, high_contrast, large_text, reduced_motion |
| Security | 16 | alarm_arm_delay, alarm_sound, guest_max_active, guest_request_timeout, guest_rate_limit, force_ha |
| System | 14 | ha_status, last_sync, uptime, storage, memory, version, restart |
| Advanced | 18 | backup_create, backup_restore, factory_reset, success/help messages |
| Actions | 8 | confirm_title, cancel_button, countdown_message, etc. |
| **Total** | **64** | **Comprehensive settings UI** |

---

## 8. LOGGING & AUDIT

### 8.1 Log Levels

| Action | Level | Example |
|--------|-------|---------|
| Settings change (safe) | INFO | "Language changed from English to Turkish" |
| Settings change (security) | INFO | "Alarm arm delay changed from 30 to 45 seconds" |
| Dangerous action attempt | WARN | "Backup restore requested from: smartdisplay-backup-2026-01-04-15-30-45" |
| Dangerous action confirmed | WARN | "Backup restored (15 settings applied)" |
| Factory reset initiated | WARN | "Factory reset initiated by admin (all settings will be erased on startup)" |
| System restart | WARN | "System restart initiated by admin from Settings" |
| Action error | ERROR | "Backup restore failed: corrupted backup file" |

### 8.2 No Secrets in Logs

**Never log:**
- Full backup file contents
- API keys or tokens (from HA connection)
- Usernames or email addresses
- Full device configurations
- Detailed error messages (abstract them)

**Always log:**
- Action name (what was done)
- Timestamp
- Initiating user (role)
- Result (success/failure)
- High-level impact ("X settings applied", "System restarting")

### 8.3 Audit Trail Structure

```json
{
  "timestamp": "2026-01-04T15:45:32Z",
  "category": "settings",
  "action": "backup_restore",
  "severity": "warning",
  "actor_role": "admin",
  "target": "backup:smartdisplay-backup-2026-01-04-15-30-45",
  "result": "success",
  "details": {
    "settings_applied": 15,
    "changes": ["language: en→tr", "guest_max_active: 1→2"],
    "requires_restart": true
  },
  "message": "Backup restored from 2026-01-04 (15 settings applied)"
}
```

---

## 9. TONE GUIDELINES

### 9.1 Core Tone (All Sections)

| Attribute | Example |
|-----------|---------|
| **Calm** | "Extra time to cancel alarm after voice confirmation" |
| **Factual** | "Alarm will trigger silently if sound is disabled" |
| **Clear** | "5 settings will be restored from this backup" |
| **No Jargon** | "Not: 'JWT token expiration'; Use: 'Guest access time limit'" |
| **Impact-First** | Show consequences before asking for confirmation |

### 9.2 Dangerous Action Tone

| Action | Tone Pattern |
|--------|--------------|
| **Backup Restore** | "This will restore X settings and overwrite your current configuration. Changes: [list]." |
| **System Restart** | "SmartDisplay will restart. Active requests will be canceled. System unavailable for ~30 seconds." |
| **Factory Reset** | "This will erase ALL custom settings and return to first-boot. This action cannot be undone." |

### 9.3 Warning Language

**Good:**
- "Disabling sound means alarm will trigger silently"
- "This will change guest limits from 1 to 2"
- "This action cannot be undone"

**Bad:**
- "Are you really, really sure?" (warning fatigue)
- "This might cause issues" (vague)
- "Invalid configuration" (jargon)

---

## 10. STATE TRANSITIONS

### 10.1 Settings Change State Machine

```
                                   ┌────────────────────┐
                                   │  Settings Fetched  │
                                   └─────────┬──────────┘
                                             │
                                   ┌─────────▼──────────┐
                                   │  Awaiting Input    │
                                   │  (read state)      │
                                   └─────────┬──────────┘
                                             │
                          ┌──────────────────┼──────────────────┐
                          │                  │                  │
                    ┌─────▼────┐      ┌─────▼────┐      ┌──────▼──────┐
                    │  Change  │      │  Action  │      │  No Change  │
                    │ Safe     │      │  Confirm │      │  (Idle)     │
                    │ Field    │      │ Required │      │             │
                    └─────┬────┘      └─────┬────┘      └──────┬──────┘
                          │                │                   │
                    ┌─────▼────────┐      │                   │
                    │ Apply Change │      │                   │
                    │ (no confirm) │      │                   │
                    └─────┬────────┘      │                   │
                          │               │                   │
                          │        ┌──────▼──────┐             │
                          │        │ Show        │             │
                          │        │ Confirm     │             │
                          │        │ Dialog      │             │
                          │        └──────┬──────┘             │
                          │               │                   │
                          │        ┌──────┴──────┐             │
                          │        │             │             │
                          │   ┌────▼────┐  ┌────▼────┐        │
                          │   │Confirmed│  │ Canceled│        │
                          │   └────┬────┘  └────┬────┘        │
                          │        │            │             │
                    ┌─────▼────────▼────┐   │             │
                    │ Apply & Log       │   │             │
                    │ (with confirm)    │   │             │
                    └─────┬─────────────┘   │             │
                          │                │             │
                          └────┬────────────┴─────────────┘
                               │
                        ┌──────▼──────┐
                        │  Restart    │
                        │  Required?  │
                        └──────┬──────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
              ┌─────▼────┐          ┌────▼────┐
              │ Show     │          │ Back to  │
              │ Countdown│          │ Idle     │
              └─────┬────┘          └──────────┘
                    │
              ┌─────▼──────┐
              │ Restart    │
              │ System     │
              └────────────┘
```

### 10.2 Backup Restore State Machine

```
                        ┌──────────────────┐
                        │ Backup List      │
                        │ Displayed        │
                        └────────┬─────────┘
                                 │
                        ┌────────▼────────┐
                        │ User Selects    │
                        │ Backup          │
                        └────────┬────────┘
                                 │
                        ┌────────▼─────────────┐
                        │ Show Restore        │
                        │ Confirmation Dialog │
                        │ (list changes)      │
                        └────────┬────────────┘
                                 │
                          ┌──────┴──────┐
                          │             │
                    ┌─────▼───┐   ┌────▼────┐
                    │ Restore │   │ Cancel  │
                    │ Confirm │   │         │
                    └─────┬───┘   └────┬────┘
                          │            │
                  ┌───────▼────────┐   │
                  │ Apply Settings │   │
                  │ & Start        │   │
                  │ Countdown      │   │
                  └───────┬────────┘   │
                          │            │
                  ┌───────▼──────┐     │
                  │ Restart      │     │
                  │ System       │     │
                  └──────────────┘     │
                                       │
                           ┌───────────▼───┐
                           │ Return to     │
                           │ Settings View │
                           └───────────────┘
```

### 10.3 Factory Reset State Machine

```
                    ┌───────────────────────┐
                    │ Factory Reset Button  │
                    │ Clicked               │
                    └──────────┬────────────┘
                               │
                    ┌──────────▼──────────┐
                    │ Show Danger Dialog  │
                    │ + Typing Field      │
                    └──────────┬──────────┘
                               │
                        ┌──────▼───────┐
                        │ User Types   │
                        │ "RESET"      │
                        └──────┬───────┘
                               │
                    ┌──────────▼──────────┐
                    │ Show Countdown      │
                    │ Dialog              │
                    │ (with Cancel option)│
                    └──────────┬──────────┘
                               │
                        ┌──────┴──────┐
                        │             │
                  ┌─────▼────┐   ┌────▼────┐
                  │ Confirm  │   │ Cancel  │
                  │ Reset    │   │         │
                  └─────┬────┘   └────┬────┘
                        │             │
              ┌─────────▼──────┐      │
              │ Clear Settings │      │
              │ & Start        │      │
              │ Countdown      │      │
              └─────────┬──────┘      │
                        │             │
              ┌─────────▼──────┐      │
              │ Restart to     │      │
              │ First-Boot     │      │
              └────────────────┘      │
                                      │
                        ┌─────────────▼────┐
                        │ Return to        │
                        │ Settings View    │
                        └──────────────────┘
```

---

## 11. TESTING STRATEGY (Behavior Validation)

### 11.1 Confirmation Flows

- [ ] Alarm sound disable shows warning before apply
- [ ] Guest max > 2 shows warning before apply
- [ ] Force HA toggle shows warning before apply
- [ ] Backup restore shows changes before confirmation
- [ ] Factory reset requires "RESET" typing + countdown

### 11.2 Accessibility Variants

- [ ] reduced_motion: Advanced section expands instantly
- [ ] large_text: No collapse/expand (flat list)
- [ ] high_contrast: Danger actions have red background

### 11.3 API Response Contracts

- [ ] GET /api/ui/settings returns all sections with correct field types
- [ ] POST field_change updates setting and returns old/new values
- [ ] POST backup_create returns download URL
- [ ] POST backup_restore requires confirm=true
- [ ] POST factory_reset requires confirm_type and confirm_text

### 11.4 Logging Validation

- [ ] Safe changes log at INFO level
- [ ] Dangerous actions log at WARN level
- [ ] No secrets in any log entry
- [ ] Backup restore logs number of settings applied
- [ ] Factory reset logs with high severity

---

## 12. SUMMARY

**Settings D7 provides:**
1. ✅ 4 settings sections (General, Security, System, Advanced)
2. ✅ Progressive disclosure (Advanced starts collapsed)
3. ✅ Clear warnings only for dangerous actions
4. ✅ 3 dangerous operations with confirmation (Backup Restore, Restart, Factory Reset)
5. ✅ 64 i18n keys (all sections)
6. ✅ Accessibility integration (reduced_motion, large_text, high_contrast)
7. ✅ 2 API endpoints (GET /api/ui/settings, POST /api/ui/settings/action)
8. ✅ Complete logging strategy (info/warn/error)
9. ✅ No visual design, CSS, or animations
10. ✅ Backend-only, deterministic behavior

**Ready for implementation:**
- AlarmManager/SettingsManager in Coordinator
- HTTP endpoints in api/server.go
- i18n integration (64 keys)
- State persistence in RuntimeConfig
- Audit logging

---

## APPENDIX: KEYBOARD/ACCESSIBILITY FLOW

### Text Input Confirmation (Type "RESET")

```
User clicks "Factory Reset"
→ System shows dialog: "Type 'RESET' to confirm"
→ User focuses text input field (automatic focus)
→ User types "R", "E", "S", "E", "T" (keyboard input)
→ System compares input to "RESET" (case-sensitive)
→ On match: Reveal Countdown Dialog
→ On mismatch: Keep input field, no error message
→ On submit (Enter key): Trigger confirmation
→ On focus loss: Clear input (user must retype if needed)
```

### Countdown Interaction

```
User sees "System will restart in 10 seconds"
→ Countdown decrements: 10 → 9 → 8 ... → 1 → 0
→ User can click "Cancel" at any point
→ Keyboard: ESC cancels countdown
→ On 0: System restart (no further user action)
→ Countdown persists even if page loses focus (important)
```

---

**SPECIFICATION COMPLETE**

This D7_SPECIFICATION.md defines a complete, backend-ready Settings UX for SmartDisplay with:
- Clear, calm tone
- Progressive disclosure (basic → advanced)
- Strong confirmations for dangerous actions
- Full accessibility support
- API contracts ready for implementation
- 64 i18n keys
- Comprehensive audit logging
