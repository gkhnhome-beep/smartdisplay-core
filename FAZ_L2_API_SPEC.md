# FAZ L2 API Specification

## Endpoints

### 1. Create Guest Request
```
POST /api/guest/request
```

**Authentication**: Guest role required

**Request Body**:
```json
{
    "ha_user": "username"
}
```

**Response (Success - 200)**:
```json
{
    "ok": true,
    "data": {
        "request_id": "greq-1704416123456789",
        "status": "pending",
        "target_user": "username",
        "expires_at": "2026-01-05T00:05:23Z"
    }
}
```

**Response (Error)**:
- 400: "guest request already pending" - Another request is active
- 400: "alarm must be armed to request guest access" - Alarm not armed
- 400: "ha_user required" - Missing payload field
- 403: "guest role required" - Auth check failed

**Constraints**:
- Only guest role can create requests
- Alarm must be armed (armed_home or armed_away)
- Only one active (pending) request allowed
- Auto-expires after 60 seconds

**Logging**:
```
[INFO] guest request created: id=greq-1704416123456789 target=username
```

---

### 2. Guest Approval (HA Callback)
```
POST /api/guest/approve
```

**Authentication**: Internal (typically called by HA automation)

**Request Body**:
```json
{
    "request_id": "greq-1704416123456789",
    "decision": "approve"
}
```

**Alternative**:
```json
{
    "request_id": "greq-1704416123456789",
    "decision": "reject"
}
```

**Response (Success - 200)**:
```json
{
    "ok": true,
    "data": {
        "result": "ok"
    }
}
```

**Response (Error)**:
- 400: "request not found" - Request ID doesn't match
- 400: "request is not pending" - Request already decided
- 400: "invalid decision" - Decision not approve/reject

**Callback Actions**:
- On "approve":
  - Sets status to approved
  - Calls onApproved() handler (async)
  - Stops expiration timer
  - Logs: `[INFO] guest request approved: id=greq-...`
  
- On "reject":
  - Sets status to rejected
  - Calls onRejected() handler (async)
  - Stops expiration timer
  - Logs: `[INFO] guest request rejected: id=greq-...`

---

### 3. Check Request Status (Polling)
```
GET /api/ui/guest/request/{request_id}
```

**Authentication**: Any (typically guest polling their own request)

**Response (Success - 200)**:
```json
{
    "ok": true,
    "data": {
        "request_id": "greq-1704416123456789",
        "status": "pending",
        "target_user": "username",
        "expires_at": "2026-01-05T00:05:23Z"
    }
}
```

**Response (Error)**:
- 404: "request not found or expired" - Request doesn't exist or expired

**Possible Statuses**:
- "pending" - Awaiting approval
- "approved" - Approved, callback was triggered
- "rejected" - Rejected, callback was triggered
- "expired" - Auto-expired (60s timeout reached)

---

## Data Model

### GuestRequest
```json
{
    "id": "greq-1704416123456789",
    "target_user": "username",
    "status": "pending",
    "requested_at": "2026-01-05T00:00:23Z",
    "expires_at": "2026-01-05T00:05:23Z"
}
```

### Status Values
- `pending` - Request created, awaiting decision
- `approved` - HA user approved the request
- `rejected` - HA user rejected the request
- `expired` - 60-second timeout reached without decision

---

## Error Response Format

**Standard Error Response**:
```json
{
    "ok": false,
    "data": null,
    "failsafe": {...},
    "error": {
        "code": "forbidden",
        "msg": "guest role required"
    }
}
```

**Error Codes**:
- `method_not_allowed` - Wrong HTTP method
- `bad_request` - Invalid input
- `forbidden` - Auth check failed
- `not_found` - Resource not found
- `internal_error` - Server error

---

## Flow Examples

### Successful Approval Flow
```
1. Guest creates request
   POST /api/guest/request → request_id = greq-123

2. Guest polls for approval
   GET /api/guest/request/greq-123 → status = pending

3. HA user approves in their app
   HA sends: POST /api/guest/approve 
   {request_id: greq-123, decision: approve}

4. Guest checks status again
   GET /api/guest/request/greq-123 → status = approved

5. Guest logs in as approved
```

### Timeout Flow
```
1. Guest creates request
   POST /api/guest/request → request_id = greq-123
   
2. Guest polls for 60+ seconds
   GET /api/guest/request/greq-123
   → status = expired (no response after 60s)
   
3. Request auto-cleared, guest must create new one
```

---

## Implementation Details

### Thread Safety
- All operations protected by sync.RWMutex
- Safe for concurrent access

### In-Memory Storage
- Single active request stored in memory
- No persistence
- Cleared on restart
- Callbacks can be set for integration

### Expiration
- Automatic 60-second timeout
- Cancellable timer stops on approval/rejection
- Checked on each GetActiveRequest() call

### Logging
- All status changes logged
- Failed callbacks logged as errors
- Timestamps in logs

---

## Integration Points

### Approved Callback
```go
m.SetApprovedCallback(func(req *GuestRequest) error {
    // Called when guest is approved
    // Can request Alarmo DISARM here
    // Runs async in goroutine
    return nil
})
```

### Rejected Callback
```go
m.SetRejectedCallback(func(req *GuestRequest) error {
    // Called when guest is rejected
    // Can clean up resources here
    // Runs async in goroutine
    return nil
})
```

---

## Testing

### Create Request (Guest Role)
```bash
curl -X POST http://localhost:8090/api/guest/request \
  -H "X-SmartDisplay-PIN: 1234" \
  -H "Content-Type: application/json" \
  -d '{"ha_user":"admin"}'
```

### Approve Request
```bash
curl -X POST http://localhost:8090/api/guest/approve \
  -H "Content-Type: application/json" \
  -d '{"request_id":"greq-123456","decision":"approve"}'
```

### Check Status
```bash
curl http://localhost:8090/api/ui/guest/request/greq-123456
```

---

**Status**: Backend Implementation Complete
**Last Updated**: 2026-01-05
**Frontend**: Pending Implementation
