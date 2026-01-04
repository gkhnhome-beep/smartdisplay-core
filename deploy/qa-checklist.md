# SmartDisplay Operator QA Checklist

## 1. Boot Test
- Power on device and verify service starts without errors.
- Check /health endpoint for status: ok.

## 2. UI Load Test
- Open web UI in browser (http://<device>:8090).
- Confirm UI loads and displays main dashboard.

## 3. HA Connect/Disconnect Behavior
- Verify Home Assistant connection is established (ha_connected: true in /health).
- Disconnect HA network or change token, confirm ha_connected: false.
- Restore connection, confirm ha_connected: true.

## 4. Guest Request Approve/Deny Path
- Trigger guest request from UI or API.
- Approve and deny requests as admin; verify state changes and UI updates.

## 5. Alarm Arm/Disarm Path
- Arm and disarm alarm via UI or API.
- Confirm alarm state changes and UI reflects status.

## 6. Hardware Readiness Path
- Check /health for hardware_ready: true.
- Disconnect or simulate missing hardware, confirm hardware_ready: false.

## 7. Backup/Restore Test
- Use /api/admin/backup to download config backup.
- Use /api/admin/restore to upload and restore config.
- Confirm settings are restored and system remains stable.

---
- Record all test results and issues for each deployment.
- For any failures, consult logs and retry after resolving issues.
