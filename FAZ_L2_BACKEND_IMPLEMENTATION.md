# FAZ L2 - Guest Request Backend Implementation

## Overview
Implemented in-memory guest access request system with approval flow.

---

## STEP 1 ✅ Guest Request Data Model

**File**: `internal/guest/request.go`

```go
type GuestRequest struct {
    ID         string    // Unique request ID
    TargetUser string    // HA user to approve request
    Status     string    // pending | approved | rejected | expired
    RequestedAt time.Time
    ExpiresAt   time.Time
}
```

**Status Constants**:
- `StatusPending` = "pending"
- `StatusApproved` = "approved"
- `StatusRejected` = "rejected"
- `StatusExpired` = "expired"

---

## STEP 2 ✅ In-Memory Manager

**Type**: `Manager`

**Methods**:
- `NewManager(timeout time.Duration)` - Create manager with 60s default timeout
- `CreateRequest(targetUser string)` - Create new request (enforces single active request)
- `ApproveRequest(requestID string)` - Mark request as approved, triggers callback
- `RejectRequest(requestID string)` - Mark request as rejected, triggers callback
- `GetActiveRequest()` - Get current active request (nil if expired)
- `ClearRequest()` - Clear active request
- `SetApprovedCallback(fn)` - Set callback on approval
- `SetRejectedCallback(fn)` - Set callback on rejection

**Constraints**:
- Only ONE active request allowed at a time
- Automatic expiration after timeout
- Thread-safe with sync.RWMutex
- No persistence (memory only)

---

## STEP 3 ✅ API Endpoints

### POST /api/guest/request
Create new guest access request

**Auth**: Guest role only
**Payload**:
```json
{
    "ha_user": "username"
}
```

**Response**:
```json
{
    "ok": true,
    "data": {
        "request_id": "greq-1736086123456",
        "status": "pending",
        "expires_at": "2026-01-05T00:05:23Z",
        "target_user": "username"
    }
}
```

**Checks**:
- Guest role required
- Alarm must be armed (armed_home or armed_away)
- Only one pending request allowed

---

### POST /api/guest/approve
Approval callback (internal use by HA automation)

**Payload**:
```json
{
    "request_id": "greq-1736086123456",
    "decision": "approve"  // or "reject"
}
```

**Response**:
```json
{
    "ok": true,
    "data": {
        "result": "ok"
    }
}
```

**Actions on Approval**:
- Triggers onApproved callback
- Callback can request Alarmo disarm

**Actions on Rejection**:
- Triggers onRejected callback
- No alarm action

---

### GET /api/ui/guest/request/{request_id}
Check request status (for frontend polling)

**Response**:
```json
{
    "ok": true,
    "data": {
        "request_id": "greq-1736086123456",
        "status": "pending",
        "expires_at": "2026-01-05T00:05:23Z",
        "target_user": "username"
    }
}
```

---

## STEP 4 ✅ Alarm Interaction

**Current Implementation**: Placeholder
- Creates request with alarm armed check
- Callbacks ready for alarm disarm integration
- No alarm state manipulation in SmartDisplay (read-only from Alarmo)

**Future**: Set approval callback to call Alarmo DISARM when approved

---

## STEP 5 ✅ Logging

All operations logged via `logger.Info()` and `logger.Error()`:

**On Request Creation**:
```
[INFO] guest request created: id=greq-1736086123456 target=username
```

**On Approval**:
```
[INFO] guest request approved: id=greq-1736086123456
```

**On Rejection**:
```
[INFO] guest request rejected: id=greq-1736086123456
```

**On Expiration**:
```
[INFO] guest request expired: id=greq-1736086123456
```

**On Callback Error**:
```
[ERROR] approval callback failed: error message
```

---

## Integration Points

### Coordinator (`internal/system/coordinator.go`)
- Field: `GuestRequest *guest.Manager`
- Initialized with 60-second timeout
- Available to all API handlers via `s.coord.GuestRequest`

### API Server (`internal/api/server.go`)
- `handleGuestRequest()` - Create new request
- `handleGuestApprove()` - Handle approval/rejection
- `handleGuestRequestStatus()` - Check status (for polling)

### Routes (`internal/api/bootstrap.go`)
- `POST /api/guest/request` → handleGuestRequest
- `POST /api/guest/approve` → handleGuestApprove
- `GET /api/ui/guest/request/{request_id}` → handleGuestRequestStatus
- `POST /api/guest/deny` → handleGuestDeny (legacy)

---

## Compilation Status

✅ No build errors
✅ All imports correct
✅ Thread safety verified (sync.RWMutex)
✅ Error handling complete

---

## Testing Checklist

- [ ] Create request as guest role
- [ ] Verify single request constraint
- [ ] Approve request
- [ ] Reject request
- [ ] Request expires after 60s
- [ ] Check status via polling
- [ ] Verify logs for all operations
- [ ] Test with alarm disarmed (should reject)

---

## Notes

- All data in memory (no database)
- Expires after 60 seconds by default
- Callbacks are async (goroutines)
- Thread-safe for concurrent access
- Ready for frontend integration
- Alarm integration callbacks waiting for HA adapter configuration
