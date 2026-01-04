# Telemetry Quick Reference

## Files Created
- `internal/telemetry/telemetry.go` - Core collector (170 lines)
- `internal/telemetry/telemetry_test.go` - Tests & examples (95 lines)
- `internal/telemetry/README.md` - Full documentation
- Modified `internal/api/server.go` - API integration

## API Endpoints

### Check Status
```bash
curl -H "X-User-Role: admin" http://localhost:8090/api/admin/telemetry/summary
```

### Enable Telemetry
```bash
curl -X POST -H "X-User-Role: admin" -H "Content-Type: application/json" \
  -d '{"enabled": true}' \
  http://localhost:8090/api/admin/telemetry/optin
```

### Disable Telemetry
```bash
curl -X POST -H "X-User-Role: admin" -H "Content-Type: application/json" \
  -d '{"enabled": false}' \
  http://localhost:8090/api/admin/telemetry/optin
```

## Code Integration

```go
// Recording feature usage
server.telemetry.RecordFeatureUsage("alarm_armed")

// Recording errors (category only, not messages)
server.telemetry.RecordError("network_timeout")

// Recording performance (automatically buckets by time)
start := time.Now()
// ... do work ...
server.telemetry.RecordPerformance("ha_query", time.Since(start))

// Check if opted in
if server.telemetry.IsOptedIn() {
    // Recording is enabled
}

// Get current data
summary := server.telemetry.GetSummary()
// summary.FeatureUsage, summary.ErrorCategories, 
// summary.PerformanceBuckets, summary.CollectedAt

// Save to disk
err := server.telemetry.Flush()
```

## Performance Buckets

| Bucket | Duration | Example |
|--------|----------|---------|
| `very_fast` | < 100ms | `alarm_arm:very_fast` |
| `fast` | < 500ms | `ha_query:fast` |
| `normal` | < 1s | `config_load:normal` |
| `slow` | < 5s | `backup:slow` |
| `very_slow` | ≥ 5s | `restore:very_slow` |

## Default Settings

- **Opt-in**: Disabled (off by default)
- **Storage**: `data/telemetry.json`
- **Upload**: None (disabled)
- **Admin Only**: Both API endpoints require `X-User-Role: admin`

## No Personal Data

✓ No user IDs, names, emails, timestamps, IP addresses
✓ Only counts and category names
✓ No individual events, only aggregated buckets
✓ No raw data collection

## Opt-In Workflow

1. User/Admin sees telemetry is disabled by default
2. Admin chooses to enable via `POST /api/admin/telemetry/optin` with `{"enabled": true}`
3. State persists to `data/telemetry.json`
4. Features start recording when opt-in enabled
5. Data aggregates locally only
6. Admin can view summary via `GET /api/admin/telemetry/summary`
7. Admin can disable anytime via same endpoint with `{"enabled": false}`

## File Format

```json
{
  "opt_in_enabled": true,
  "feature_usage": {
    "feature_name": count
  },
  "error_categories": {
    "error_type": count
  },
  "performance_buckets": {
    "operation:bucket": count
  },
  "collected_at": "2026-01-04T15:30:45.123456Z"
}
```

---

For full documentation, see `internal/telemetry/README.md`
