# OTA Update System - WOW Phase FAZ 77

Safe, skeleton implementation of OTA (Over-The-Air) updates.

## Design Principles

✅ **No auto-update**: Updates never happen automatically
✅ **No downloads yet**: Package download logic not implemented
✅ **No execution**: No code execution or reboot forced
✅ **Manual control**: Admin must explicitly stage and activate updates
✅ **Checksum validation**: Package integrity verified before staging
✅ **Full audit logging**: Every action is logged for security review

## What's Implemented

### UpdateManager API

```go
// Check for available updates (STUB - returns nil)
available, err := updateMgr.CheckAvailable()

// Validate package integrity (SHA256 checksum only)
err := updateMgr.ValidatePackage(packageData, expectedChecksum)

// Stage package to disk (write only, no execution)
path, err := updateMgr.StageUpdate(packageData, packageInfo)

// Schedule activation on next reboot (requires staged package)
err := updateMgr.ActivateOnReboot()

// Cancel pending reboot activation (before reboot occurs)
err := updateMgr.CancelActivation()

// Remove staged package from disk
err := updateMgr.ClearStaged()

// Get current system status
status := updateMgr.GetStatus()
```

### Status Information

```json
{
  "current_version": "1.0.0",
  "available": null,
  "staged": {
    "version": "2.0.0",
    "build_id": "build123",
    "checksum": "abc123...",
    "size": 1024000,
    "staged_at": "2026-01-04T15:30:45Z",
    "stage_path": "/path/to/staging/update-2.0.0-build123.bin"
  },
  "pending_reboot": true,
  "last_check_time": "2026-01-04T15:30:45Z",
  "last_error": ""
}
```

## API Endpoints

### Get Update Status (Admin-Only)
```
GET /api/admin/update/status
```

Returns current update system state:
- Current version
- Available updates (if any)
- Staged packages (if any)
- Pending reboot status
- Last check time
- Last error (if any)

**Example Response:**
```json
{
  "ok": true,
  "data": {
    "current_version": "1.0.0",
    "available": null,
    "staged": null,
    "pending_reboot": false,
    "last_check_time": "2026-01-04T15:30:45Z",
    "last_error": ""
  }
}
```

### Stage Update (Admin-Only, STUB)
```
POST /api/admin/update/stage
Content-Type: application/json

{
  "package": {
    "version": "2.0.0",
    "build_id": "build123",
    "checksum": "abc123...",
    "size": 1024000
  }
}
```

**Current Behavior (STUB):**
- Accepts request structure
- Validates JSON format
- Logs action to audit trail
- Returns success message
- Does NOT actually download or write package data

**Future Implementation:**
- Download package from server
- Validate checksum
- Write to staging directory
- Return staged path

**Example Response:**
```json
{
  "ok": true,
  "data": {
    "message": "update staging is a no-op stub in this phase",
    "version": "2.0.0",
    "status": "stub_response"
  }
}
```

## Workflow

### Typical Update Flow (Future)

```
1. Admin checks status
   GET /api/admin/update/status
   ↓
2. Server fetches available updates (CheckAvailable)
   - Remote check (not implemented yet)
   ↓
3. Admin stages update
   POST /api/admin/update/stage {package metadata}
   ↓
4. System validates package
   - Checksum verification (IMPLEMENTED)
   ↓
5. Package written to staging directory
   - StagePath: data/staging/update-VERSION-BUILDID.bin
   ↓
6. System waits for admin confirmation
   ↓
7. Admin schedules for reboot
   POST /api/admin/update/activate (future endpoint)
   ↓
8. System sets boot flag for next reboot
   ↓
9. Admin manually reboots system
   ↓
10. On boot: system applies staged update
    (NOT IMPLEMENTED in this phase)
```

## Data Storage

### Staging Directory
```
data/staging/
  ├── update-2.0.0-build123.bin    (staged package)
  └── ...
```

### Status (In-Memory Only)
- Current version
- Available package info
- Staged package info
- Pending reboot flag
- Audit log entries

## Safety Features

✅ **Admin-Only Control**: Both endpoints require admin role
✅ **Checksum Validation**: SHA256 integrity check before staging
✅ **No Auto-Execution**: Requires explicit admin action
✅ **Manual Reboot**: Admin must manually reboot
✅ **Audit Trail**: Every action logged with timestamp
✅ **Staged Rollback**: Can clear staged package before reboot
✅ **Activation Cancel**: Can cancel pending reboot before it occurs

## What's NOT Implemented

❌ **Auto-Update**: Disabled completely
❌ **Download**: Package fetching not implemented
❌ **Signature Verification**: Only checksum validation
❌ **Compression**: Package handling as-is
❌ **Rollback**: Previous version recovery not implemented
❌ **Execution**: Package application not implemented
❌ **Delta Updates**: Full package download only
❌ **Scheduled Reboot**: Only manual reboot supported

All can be added in future phases without breaking this API.

## Security Considerations

### What's Protected
- ✅ Checksum verification before staging
- ✅ Package staging to isolated directory
- ✅ No execution capability
- ✅ Admin-only endpoints
- ✅ Full audit trail
- ✅ Explicit manual activation required

### What's Not Protected (Future Phases)
- ❌ Package source authentication
- ❌ Signature verification
- ❌ Encrypted storage
- ❌ Secure boot integration
- ❌ Rollback capability

## Audit Logging

All update actions are logged with format: `update_ACTION: detail`

**Examples:**
- `update_check: checked for available updates`
- `update_validate_success: package validated with checksum abc123...`
- `update_validate_failed: checksum mismatch: got X, expected Y`
- `update_staged: version 2.0.0 staged at data/staging/update-2.0.0-build123.bin`
- `update_activate_scheduled: update 2.0.0 scheduled for next reboot`
- `update_activation_cancelled: pending reboot activation cancelled`
- `update_status_check: admin checked update status`
- `update_stage_stub: admin requested staging of version 2.0.0 (no-op stub)`

## Testing

```bash
# Build update package
go build ./internal/update

# Run tests
go test -v ./internal/update

# Test coverage
go test -cover ./internal/update
```

## Code Examples

### Check and Stage Update (Future Flow)

```go
import "smartdisplay-core/internal/update"

// Check available
available, _ := updateMgr.CheckAvailable()
if available != nil {
    // Download package data from server (future)
    packageData := // ... downloaded from server
    
    // Validate checksum
    if err := updateMgr.ValidatePackage(packageData, available.Checksum); err == nil {
        // Stage the package
        path, err := updateMgr.StageUpdate(packageData, *available)
        if err == nil {
            // Schedule for reboot
            updateMgr.ActivateOnReboot()
        }
    }
}

// Check status
status := updateMgr.GetStatus()
if status.PendingReboot {
    // Notify admin of pending update
}
```

### Rollback (Before Reboot)

```go
// If update hasn't been applied yet
err := updateMgr.CancelActivation()
err = updateMgr.ClearStaged()
```

## Compliance

✓ **Transparent**: All operations visible to admin via API
✓ **Controllable**: Admin explicitly approves each step
✓ **Auditable**: Full audit trail of all actions
✓ **Safe**: No automatic execution or forced behavior
✓ **Recoverable**: Can cancel before reboot occurs

## Future Phases

### FAZ 78: Package Download
- Implement CheckAvailable() remote fetch
- Download from configured server
- Verify signature

### FAZ 79: Package Execution
- Implement boot-time package application
- Verify boot flag
- Apply staged package
- Rollback on failure

### FAZ 80: Smart Reboot
- Scheduled reboot windows
- Delay reboot if system busy
- Notification to users before reboot
