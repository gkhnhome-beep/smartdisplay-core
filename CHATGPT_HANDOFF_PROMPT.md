You are taking over an advanced Smart Home kiosk project called SmartDisplay.

IMPORTANT:
- Read PROJECT_STATE.md first.
- Do NOT redesign architecture.
- Do NOT move alarm logic out of Home Assistant Alarmo.
- Secure Storage is DONE and must not be weakened.

Current state:
- Version: v1.0.0-rc1
- Phase: Alarmo Integration
- Active task: FAZ A2 – Alarmo Read-Only Adapter

What is expected from you now:
- Implement Alarmo READ-ONLY adapter
- Poll /api/states/alarm_control_panel.alarmo
- Normalize state (disarmed, arming, armed, triggered)
- Feed UI from Alarmo truth
- No arm/disarm yet

Constraints:
- Go backend
- Standard library only
- Token decrypted from secure storage
- No secrets in logs
- PC-first testing, Raspberry Pi later

When in doubt:
- Alarmo is the single source of truth.
- SmartDisplay is UI + coordinator, not an alarm brain.
