# Telemetry System - WOW Phase FAZ 76

Privacy-first, opt-in product improvement telemetry for smartdisplay-core.

## Design Principles

✅ **Opt-in only**: Disabled by default. Explicitly enabled by admin via API.
✅ **No personal data**: Only aggregated counts and categories, never personally identifiable.
✅ **No raw events**: Only bucketized performance metrics.
✅ **Standard library only**: No external dependencies (uses Go stdlib only).
✅ **Local aggregation**: Data stays on device. Upload disabled by default.

## What Gets Captured

### 1. Feature Usage Counts
- Count of times each feature is used (e.g., "guest_approval_requested", "alarm_armed")
- No feature names are logged, only aggregate counts per feature

### 2. Error Categories
- Aggregated by error type/category (e.g., "network_error", "hardware_fault", "timeout")
- No error messages or stack traces are captured
- Just category counts

### 3. Performance Buckets
- Operation durations binned into 5 buckets:
  - `very_fast`: < 100ms
  - `fast`: < 500ms
  - `normal`: < 1 second
  - `slow`: < 5 seconds
  - `very_slow`: ≥ 5 seconds
- Format: `operation_name:bucket_name` with count

## API Endpoints

### Get Telemetry Summary (Admin-Only)
```
GET /api/admin/telemetry/summary
```

Returns aggregated telemetry snapshot:
```json
{
  "ok": true,
  "data": {
    "opt_in_enabled": false,
    "feature_usage": {
      "alarm_armed": 15,
      "alarm_disarmed": 12,
      "guest_approval_requested": 3
    },
    "error_categories": {
      "network_timeout": 1,
      "hardware_fault": 0
    },
    "performance_buckets": {
      "alarm_arm:very_fast": 10,
      "alarm_arm:fast": 5,
      "ha_query:normal": 3
    },
    "collected_at": "2026-01-04T15:30:45.123456Z"
  }
}
```

### Enable/Disable Opt-In (Admin-Only)
```
POST /api/admin/telemetry/optin
Content-Type: application/json

{
  "enabled": true
}
```

Response:
```json
{
  "ok": true,
  "data": {
    "enabled": true,
    "message": "telemetry enabled"
  }
}
```

## How to Use

### From Client Code (Go)

```go
import "smartdisplay-core/internal/telemetry"

// Server already initializes telemetry
// Record feature usage when opt-in is enabled
server.telemetry.RecordFeatureUsage("alarm_armed")

// Record error categories (not error messages)
server.telemetry.RecordError("network_timeout")

// Record operation performance
start := time.Now()
// ... perform operation ...
elapsed := time.Since(start)
server.telemetry.RecordPerformance("ha_query", elapsed)

// Persist state periodically
server.telemetry.Flush() // saves to data/telemetry.json
```

### Check Opt-In Status
```go
isOptedIn := server.telemetry.IsOptedIn()
if isOptedIn {
    server.telemetry.RecordFeatureUsage("some_feature")
}
```

## Data Storage

All telemetry data is stored locally in:
```
data/telemetry.json
```

Format (pretty-printed):
```json
{
  "opt_in_enabled": false,
  "feature_usage": {
    "feature_name": count
  },
  "error_categories": {
    "category": count
  },
  "performance_buckets": {
    "operation:bucket": count
  },
  "collected_at": "ISO8601-timestamp"
}
```

## Default Behavior

- **Opt-in disabled by default**: No telemetry is recorded until explicitly enabled
- **No automatic upload**: Data is persisted locally only
- **No background processes**: Telemetry collection is synchronous and lightweight
- **Admin-only control**: Only users with `Admin` role can enable/disable telemetry

## Audit Trail

All telemetry decisions are recorded in the audit log:
- `telemetry_optin`: Admin enabled/disabled telemetry
- `telemetry_error`: Any errors during telemetry operations

## Future Enhancements

These are NOT implemented yet:
- Background upload of aggregated data
- Compression/encryption of persisted data
- Telemetry data purging after X days
- Custom performance operation names and thresholds

## Compliance Notes

✓ GDPR compliant: No personal data captured
✓ Privacy-first: Opt-in required, off by default
✓ Transparent: All data stored locally, visible via API
✓ Non-invasive: No background processes or network access

## Testing

Reset telemetry data (for testing only):
```go
server.telemetry.Reset()
```

This clears all aggregated data in memory (not persisted state).
