# SmartDisplay – Project State Flag

## Current Status
- Version: v1.0.0-rc1
- Phase: Alarmo Integration
- Active Subphase: FAZ A2 – Alarmo Read-Only Adapter
- Secure Storage: PASS
- UI Kiosk: STABLE (PC)

## Last Completed
- Secure Storage (AES-256-GCM, OS-level master key)
- UI Sprint 3.x
- RC Sprint 4.x
- Health, Graceful Shutdown, Integration Tests

## Current Task (DO NOT SKIP)
FAZ A2 – Alarmo Read-Only Adapter
- No arm/disarm
- Read-only polling
- State mapping only
- UI reflects Alarmo truth

## Next Tasks
- FAZ A3 – Alarm Screen ↔ Alarmo Sync
- FAZ A4 – Write Actions (Arm/Disarm)
- FAZ A5 – Triggered Handling

## Hard Rules
- Alarm logic lives ONLY in Alarmo
- Tokens are never logged or returned
- Master key is OS-level only
