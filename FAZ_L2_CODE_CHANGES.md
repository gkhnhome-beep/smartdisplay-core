# FAZ L2 Backend - Code Changes Summary

## 📋 Overview
Backend implementation of guest access request system with approval flow.
- ✅ In-memory request manager
- ✅ API endpoints for request/approval
- ✅ Alarm armed validation
- ✅ Logging for all operations
- ✅ Thread-safe with sync.RWMutex
- ✅ Zero external dependencies

---

## 📁 Files Changed

### 1. NEW FILE: `internal/guest/request.go` (213 lines)

**Provides**:
- `GuestRequest` struct (ID, TargetUser, Status, timestamps)
- `Manager` struct (in-memory request handler)

**Key Features**:
- Single active request constraint
- Auto-expiration (60s default timeout)
- Async callbacks on approval/rejection
- Thread-safe operations with RWMutex
- Logging on all state changes

**Methods**:
```go
func NewManager(timeout time.Duration) *Manager
func (m *Manager) CreateRequest(targetUser string) (*GuestRequest, error)
func (m *Manager) ApproveRequest(requestID string) error
func (m *Manager) RejectRequest(requestID string) error
func (m *Manager) GetActiveRequest() *GuestRequest
func (m *Manager) ClearRequest()
func (m *Manager) SetApprovedCallback(fn func(*GuestRequest) error)
func (m *Manager) SetRejectedCallback(fn func(*GuestRequest) error)
```

---

### 2. MODIFIED: `internal/system/coordinator.go`

**Line 66**: Added field
```go
GuestRequest *guest.Manager       // FAZ L2: Guest approval flow
```

**Line 210**: Initialize manager
```go
GuestRequest:   guest.NewManager(60 * time.Second),    // FAZ L2: Guest approval flow
```

---

### 3. MODIFIED: `internal/api/server.go`

#### Function 1: `handleGuestRequest()` (Lines 985-1010)
```go
// Create new guest access request
// Auth: guest role only
// Checks: alarm must be armed
// Action: Creates request, returns request_id
```

**Validation**:
- POST method only
- Guest role check
- Alarm armed check (Alarmo state)
- Only one pending request

**Response**: `{request_id, status, expires_at, target_user}`

#### Function 2: `handleGuestApprove()` (Lines 233-280)
```go
// Handle approval/rejection callback
// Payload: {request_id, decision: "approve"|"reject"}
```

**Actions**:
- Approve: Marks approved, calls onApproved callback
- Reject: Marks rejected, calls onRejected callback

#### Function 3: `handleGuestDeny()` (Lines 282-310)
```go
// Legacy rejection endpoint
// Calls RejectRequest() on active request
```

#### Function 4: `handleGuestRequestStatus()` (Lines 1044-1080)
```go
// Polling endpoint for frontend
// GET /api/ui/guest/request/{request_id}
// Returns: {request_id, status, expires_at, target_user}
```

---

### 4. MODIFIED: `internal/api/bootstrap.go`

**Added Routes**:

Line 86:
```go
mux.HandleFunc("/api/ui/guest/request", s.handleGuestRequest)
```

Line 87:
```go
mux.HandleFunc("/api/ui/guest/request/", s.handleGuestRequestStatus) // FAZ L2
```

Line 118:
```go
mux.HandleFunc("/api/guest/approve", s.handleGuestApprove)
```

Line 119:
```go
mux.HandleFunc("/api/guest/deny", s.handleGuestDeny)
```

---

## 🔧 Implementation Details

### Data Flow: Request Creation
```
1. Guest POSTs to /api/guest/request with {ha_user}
2. Auth check (guest role required)
3. Alarm check (must be armed)
4. Manager.CreateRequest(ha_user)
   - Check single request constraint
   - Create GuestRequest with timeout
   - Start expiration timer
   - Log: "guest request created: id=greq-xxx target=user"
5. Return {request_id, status, expires_at}
```

### Data Flow: Approval
```
1. HA automation POSTs to /api/guest/approve
   {request_id: "greq-xxx", decision: "approve"}
2. Manager.ApproveRequest(request_id)
   - Check request exists
   - Set status to approved
   - Stop expiration timer
   - Call onApproved() async
   - Log: "guest request approved: id=greq-xxx"
3. Return {result: "ok"}
```

### Data Flow: Status Polling
```
1. Guest GETs /api/ui/guest/request/greq-xxx
2. Manager.GetActiveRequest()
   - Return current request
   - Return nil if expired
3. Return {request_id, status, expires_at}
```

---

## 🔒 Thread Safety

All operations protected by `sync.RWMutex`:

```go
type Manager struct {
    mu            sync.RWMutex           // Protects activeRequest
    activeRequest *GuestRequest          // Single active request
    timeout       time.Duration          // Timeout duration
    onApproved    func(*GuestRequest) error
    onRejected    func(*GuestRequest) error
    expireTimer   *time.Timer            // Expiration timer
}
```

**Lock Strategy**:
- Write operations: Full lock (Lock/Unlock)
- Read operations: Read lock (RLock/RUnlock)
- Timer operations: Locked during setup/cleanup

---

## 📊 Request Lifecycle

```
Created (pending)
    ↓
    ├→ [Approved by HA] → Approved → Callback fired
    ├→ [Rejected by HA] → Rejected → Callback fired
    └→ [60s timeout] → Expired → No callback
```

---

## 📝 Logging

**On Create**:
```
[INFO] guest request created: id=greq-1704416123456 target=username
```

**On Approve**:
```
[INFO] guest request approved: id=greq-1704416123456
```

**On Reject**:
```
[INFO] guest request rejected: id=greq-1704416123456
```

**On Expire**:
```
[INFO] guest request expired: id=greq-1704416123456
```

**On Callback Error**:
```
[ERROR] approval callback failed: error message
```

---

## ✅ Testing Checklist

- [ ] Create request as guest role
  ```bash
  curl -X POST http://localhost:8090/api/guest/request \
    -H "X-SmartDisplay-PIN: <guest-pin>" \
    -H "Content-Type: application/json" \
    -d '{"ha_user":"admin"}'
  ```

- [ ] Verify single request constraint
  - Create request 1 → Success
  - Create request 2 → Error: "guest request already pending"

- [ ] Approve request
  ```bash
  curl -X POST http://localhost:8090/api/guest/approve \
    -H "Content-Type: application/json" \
    -d '{"request_id":"greq-xxx","decision":"approve"}'
  ```

- [ ] Reject request
  ```bash
  curl -X POST http://localhost:8090/api/guest/approve \
    -H "Content-Type: application/json" \
    -d '{"request_id":"greq-xxx","decision":"reject"}'
  ```

- [ ] Check status (polling)
  ```bash
  curl http://localhost:8090/api/ui/guest/request/greq-xxx
  ```

- [ ] Verify request expires after 60s
- [ ] Test with alarm disarmed (should reject)
- [ ] Verify logs for all operations

---

## 🎯 Implementation Quality

✅ **Zero External Dependencies**
- Uses stdlib only (sync, time, fmt, errors)

✅ **Thread Safety**
- RWMutex protects all shared state
- Safe for concurrent requests

✅ **Proper Error Handling**
- Validation checks
- Error messages
- Graceful failures

✅ **Logging**
- All state changes logged
- Debug-friendly messages
- Timestamps on operations

✅ **Code Organization**
- Single package responsibility
- Clear method names
- Documented functions

---

## 📦 Compilation

**Build Status**: ✅ SUCCESS

```
$ go build ./internal/guest
$ go build ./cmd/smartdisplay
```

Both build successfully with zero errors.

---

**Implementation Date**: 2026-01-05
**Status**: COMPLETE
**Ready For**: Integration, Testing, Frontend Development
