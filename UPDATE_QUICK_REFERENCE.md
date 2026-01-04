# OTA Update Quick Reference

## Files Created
- `internal/update/manager.go` - Core UpdateManager (235 lines)
- `internal/update/manager_test.go` - Test suite (215 lines)
- `internal/update/README.md` - Technical documentation

## Files Modified
- `internal/api/server.go` - API integration

## API Endpoints

### Check Update Status
```bash
curl -H "X-User-Role: admin" http://localhost:8090/api/admin/update/status
```

### Stage Update (STUB)
```bash
curl -X POST -H "X-User-Role: admin" -H "Content-Type: application/json" \
  -d '{
    "package": {
      "version": "2.0.0",
      "build_id": "build123",
      "checksum": "sha256hex...",
      "size": 1024000
    }
  }' \
  http://localhost:8090/api/admin/update/stage
```

## Code Integration

### Initialize UpdateManager
```go
auditLogger := &UpdateAuditLogger{}
updateMgr := update.New("1.0.0", "data/staging", auditLogger)

// Server now has updateMgr field
server.updateMgr.GetStatus()
```

### Validate Package
```go
import "smartdisplay-core/internal/update"

data := // ... package bytes
err := updateMgr.ValidatePackage(bytes.NewReader(data), expectedChecksum)
if err == nil {
    // Checksum verified
}
```

### Stage Package
```go
pkg := update.PackageInfo{
    Version: "2.0.0",
    BuildID: "build123",
    Checksum: "abc...",
    Size: 1024000,
}
path, err := updateMgr.StageUpdate(bytes.NewReader(data), pkg)
if err == nil {
    // Package staged at: path
}
```

### Schedule Reboot
```go
err := updateMgr.ActivateOnReboot()
if err == nil {
    // Scheduled for next reboot
}
```

### Check Status
```go
status := updateMgr.GetStatus()
// status.CurrentVersion
// status.Staged
// status.PendingReboot
// status.LastError
```

### Cancel Activation
```go
err := updateMgr.CancelActivation()
if err == nil {
    // Reboot activation cancelled
}
```

### Clear Staged Package
```go
err := updateMgr.ClearStaged()
if err == nil {
    // Staged package removed
}
```

## Status Response Format

```json
{
  "current_version": "1.0.0",
  "available": {
    "version": "2.0.0",
    "build_id": "build123",
    "checksum": "abc...",
    "size": 1024000,
    "release_notes": "..."
  },
  "staged": {
    "version": "2.0.0",
    "build_id": "build123",
    "checksum": "abc...",
    "size": 1024000,
    "staged_at": "2026-01-04T15:30:45Z",
    "stage_path": "/path/to/update-2.0.0-build123.bin"
  },
  "pending_reboot": false,
  "last_check_time": "2026-01-04T15:30:45Z",
  "last_error": ""
}
```

## Safety Features

✅ Admin-only endpoints
✅ Checksum validation before staging
✅ No automatic execution
✅ Manual reboot required
✅ Can cancel before reboot
✅ Full audit logging
✅ Isolated staging directory

## What's NOT Implemented Yet

❌ Remote package download
❌ Digital signature verification
❌ Automatic package execution
❌ Forced reboot
❌ Rollback capability
❌ Delta updates
❌ Scheduled reboot windows

All can be added without breaking this API.

## Audit Log Entries

```
update_check - Checked for available updates
update_validate_success - Package validated
update_validate_failed - Checksum mismatch
update_staged - Package written to staging
update_activate_scheduled - Reboot activation scheduled
update_activation_cancelled - Pending reboot cancelled
update_cleared - Staged package removed
update_status_check - Admin checked status
update_stage_stub - Stage endpoint called (current stub)
```

## Default Settings

- **Current Version**: 1.0.0 (configured in NewServer)
- **Staging Dir**: data/staging/ (must be writable)
- **Auto-Update**: Disabled (not implemented)
- **Check Interval**: No automatic checks
- **Reboot Behavior**: Manual only

## Typical Workflow

1. Admin checks status: `GET /api/admin/update/status`
2. System shows current version and no pending updates
3. Admin stages new version: `POST /api/admin/update/stage` (stub)
4. Status shows staged package (future: after real download)
5. Admin manually reboots system
6. On reboot: system applies update (future: not yet implemented)

## Testing

```bash
cd internal/update
go test -v
```

Run in terminal to verify all tests pass.

---

For full documentation, see `internal/update/README.md`
For implementation details, see `FAZ77_IMPLEMENTATION_REPORT.md`
