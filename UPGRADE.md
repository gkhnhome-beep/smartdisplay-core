# SmartDisplay Upgrade & Rollback Guide

Complete procedures for upgrading and rolling back SmartDisplay Core v1.0.0-rc1 with zero-downtime verification.

---

## Table of Contents

1. [Pre-Upgrade Checklist](#pre-upgrade-checklist)
2. [Upgrade Procedure](#upgrade-procedure)
3. [Rollback Procedure](#rollback-procedure)
4. [Failure Handling](#failure-handling)
5. [Version Management](#version-management)
6. [Troubleshooting](#troubleshooting)

---

## Pre-Upgrade Checklist

Before upgrading, verify system state:

### Health Verification

```bash
# 1. Backend health check
curl -s http://localhost:8080/health | jq .

# Expected: HTTP 200, "status": "healthy"

# 2. Kiosk responsive check
curl -s http://localhost:8090/ | head -20

# Expected: HTTP 200, HTML content returned

# 3. Service status
sudo systemctl status smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service

# Expected: active (running)

# 4. Disk space (need ~50MB minimum for binary + backup)
df -h /opt /home

# Expected: both >100MB available
```

### Documentation

Record current version before proceeding:

```bash
# 1. Current version
curl -s http://localhost:8080/health | jq .version.version

# 2. Current commit
curl -s http://localhost:8080/health | jq .version.commit

# 3. Current build date
curl -s http://localhost:8080/health | jq .version.build_date

# Example output to record:
# Version: 1.0.0-rc1
# Commit: a1b2c3d
# Date: 2024-01-04T14:32:18Z
```

### Configuration Backup

Even though upgrade only replaces binary, backup config:

```bash
# Backup config
sudo cp -r /opt/smartdisplay-core /opt/smartdisplay-core.backup.$(date +%s)

# List backups
sudo ls -la /opt/smartdisplay-core.backup.*
```

---

## Upgrade Procedure

### Step 1: Announce Maintenance Window

If user-facing, notify kiosk users:

```bash
# Option 1: Display message on kiosk
# (Manual: have admin access kiosk and display notice)

# Option 2: Check who's using system
sudo journalctl -u smartdisplay-kiosk.service -n 20 | grep "Chromium"
```

### Step 2: Stop Services (Ordered)

Stop in reverse dependency order:

```bash
# 1. Stop kiosk first (UI can tolerate downtime)
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service

# Verify it stopped
sleep 2
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service

# Expected: inactive (dead)

# 2. Stop backend
sudo systemctl stop smartdisplay-core.service

# Verify it stopped
sleep 2
sudo systemctl status smartdisplay-core.service

# Expected: inactive (dead)

# 3. Verify processes terminated
ps aux | grep smartdisplay-core | grep -v grep

# Expected: no output (process gone)
```

### Step 3: Backup Current Binary

Create rollback point:

```bash
# Backup current binary with timestamp
sudo cp /opt/smartdisplay-core/smartdisplay-core \
         /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup

# Verify backup exists
ls -lh /opt/smartdisplay-core/smartdisplay-core*

# Expected output:
# smartdisplay-core
# smartdisplay-core.v1.0.0-rc1.backup

# Record backup location (for rollback)
echo "Backup: /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup" > /tmp/upgrade.log
```

### Step 4: Replace Binary

Deploy new version:

```bash
# Verify new binary exists (from dist/ or download location)
ls -lh dist/smartdisplay-core-linux-arm32v7

# Copy new binary
sudo cp dist/smartdisplay-core-linux-arm32v7 /opt/smartdisplay-core/smartdisplay-core

# Set permissions (must match original)
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core/smartdisplay-core
sudo chmod 755 /opt/smartdisplay-core/smartdisplay-core

# Verify new binary
ls -lh /opt/smartdisplay-core/smartdisplay-core
file /opt/smartdisplay-core/smartdisplay-core

# Expected:
# -rwxr-xr-x smartdisplay smartdisplay
# ELF 32-bit LSB executable (for ARM) or ELF 64-bit (for amd64)
```

### Step 5: Start Services (Ordered)

Start in dependency order (backend → kiosk):

```bash
# 1. Start backend
sudo systemctl start smartdisplay-core.service

# Wait for startup
sleep 3

# 2. Verify backend is running
sudo systemctl status smartdisplay-core.service

# Expected: active (running)

# 3. Verify backend health
curl -s http://localhost:8080/health | jq .

# Expected: HTTP 200
# {
#   "status": "healthy",
#   "version": {
#     "version": "1.0.0-rc1",
#     ...
#   }
# }
```

### Step 6: Verify Upgrade Success

```bash
# 1. Check new version is running
curl -s http://localhost:8080/health | jq .version

# Record new version info
UPGRADE_VERSION=$(curl -s http://localhost:8080/health | jq -r .version.version)
UPGRADE_COMMIT=$(curl -s http://localhost:8080/health | jq -r .version.commit)
echo "Upgraded to: $UPGRADE_VERSION ($UPGRADE_COMMIT)" >> /tmp/upgrade.log

# 2. Check backend responsiveness (full health check)
curl -v http://localhost:8080/health 2>&1 | grep -E "< HTTP|\"status\""

# Expected: HTTP/1.1 200 OK, "status": "healthy"

# 3. Start kiosk
sudo systemctl --user -M smartdisplay-ui start smartdisplay-kiosk.service

# Wait for startup
sleep 3

# 4. Verify kiosk is responsive
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service

# Expected: active (running)

# 5. Test UI endpoint
curl -s http://localhost:8090/ | head -5

# Expected: HTML5 doctype and structure
```

### Step 7: Post-Upgrade Monitoring

Monitor for 10-15 minutes:

```bash
# 1. Watch backend logs for errors
sudo journalctl -u smartdisplay-core.service -f &

# 2. In another terminal, watch kiosk logs
sudo journalctl --user-unit=smartdisplay-kiosk.service -M smartdisplay-ui -f &

# 3. Spot-check health every 2 minutes
for i in {1..5}; do
  echo "Health check $i..."
  curl -s http://localhost:8080/health | jq .version.version
  sleep 120
done

# 4. Check system resources
free -h
df -h /opt
ps aux | grep smartdisplay
```

### Step 8: Log Upgrade Completion

```bash
# Record final state
sudo tee -a /tmp/upgrade.log << EOF
Upgrade Status: SUCCESS
Timestamp: $(date -Iseconds)
Duration: $(date -d "$(head -1 /tmp/upgrade.log)" +%s) seconds ago
Backend Health: $(curl -s http://localhost:8080/health | jq -r .status)
Kiosk Status: $(sudo systemctl --user -M smartdisplay-ui is-active smartdisplay-kiosk.service)
EOF

# Display summary
cat /tmp/upgrade.log
```

---

## Rollback Procedure

Perform immediate rollback if upgrade fails.

### Step 1: Verify Failure

```bash
# Check if backend crashed
sudo systemctl status smartdisplay-core.service

# Check logs for errors
sudo journalctl -u smartdisplay-core.service -n 30 | tail -10

# Try health check
curl -s http://localhost:8080/health 2>&1 | head -20

# If any timeout/connection error → proceed to rollback
```

### Step 2: Stop Current Services

```bash
# Stop kiosk
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service

# Stop backend (may be already stopped)
sudo systemctl stop smartdisplay-core.service || true

# Wait and verify
sleep 2
sudo systemctl status smartdisplay-core.service || true
```

### Step 3: Restore Previous Binary

```bash
# Find backup (created during upgrade)
ls -lh /opt/smartdisplay-core/smartdisplay-core*

# Expected output:
# smartdisplay-core (broken)
# smartdisplay-core.v1.0.0-rc1.backup (good)

# Restore from backup
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core

# Verify
ls -lh /opt/smartdisplay-core/smartdisplay-core
file /opt/smartdisplay-core/smartdisplay-core
```

### Step 4: Restart Services

```bash
# Start backend with previous binary
sudo systemctl start smartdisplay-core.service

# Wait for startup
sleep 3

# Verify it's running
sudo systemctl status smartdisplay-core.service

# Check health (should succeed now)
curl -s http://localhost:8080/health | jq .

# Start kiosk
sudo systemctl --user -M smartdisplay-ui start smartdisplay-kiosk.service

# Verify
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service
```

### Step 5: Verify Rollback Success

```bash
# Confirm version is back to previous
ROLLED_VERSION=$(curl -s http://localhost:8080/health | jq -r .version.version)
ROLLED_COMMIT=$(curl -s http://localhost:8080/health | jq -r .version.commit)

echo "Rolled back to: $ROLLED_VERSION ($ROLLED_COMMIT)"

# Test full functionality
curl -s http://localhost:8090/ | head -5

# Expected: HTML response, services responsive
```

### Step 6: Log Rollback

```bash
# Record rollback
sudo tee -a /tmp/rollback.log << EOF
Rollback Status: SUCCESS
Timestamp: $(date -Iseconds)
Rolled Back To: $ROLLED_VERSION ($ROLLED_COMMIT)
Backend Health: $(curl -s http://localhost:8080/health | jq -r .status)
EOF

cat /tmp/rollback.log
```

---

## Failure Handling

### Scenario 1: Binary Won't Start

**Symptoms:** `systemctl status` shows failed or timeout

**Steps:**

```bash
# 1. Check what happened
sudo journalctl -u smartdisplay-core.service -n 50 | tail -20

# 2. Check binary is correct architecture
file /opt/smartdisplay-core/smartdisplay-core

# 3. Check permissions
ls -l /opt/smartdisplay-core/smartdisplay-core

# Should be: -rwxr-xr-x smartdisplay:smartdisplay

# 4. Fix permissions if needed
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core/smartdisplay-core
sudo chmod 755 /opt/smartdisplay-core/smartdisplay-core

# 5. Try manual start (as smartdisplay user)
sudo -u smartdisplay /opt/smartdisplay-core/smartdisplay-core

# If crashes, note error and ROLLBACK immediately

# 6. Rollback
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core
sudo systemctl restart smartdisplay-core.service
```

### Scenario 2: Binary Crashes on Startup

**Symptoms:** Process exits with error code

**Steps:**

```bash
# 1. Check error logs
sudo journalctl -u smartdisplay-core.service | grep -i "panic\|fatal\|error"

# 2. Check for missing files/config
ls -l /opt/smartdisplay-core/

# May need: configs/*.json, data/runtime.json

# 3. If config is corrupted, restore from backup
sudo cp /opt/smartdisplay-core.backup.*/configs/ /opt/smartdisplay-core/

# 4. If still fails, ROLLBACK
sudo systemctl stop smartdisplay-core.service
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core
sudo systemctl start smartdisplay-core.service
```

### Scenario 3: Health Check Fails

**Symptoms:** `curl http://localhost:8080/health` returns timeout or 5xx

**Steps:**

```bash
# 1. Check if process is running
ps aux | grep smartdisplay-core | grep -v grep

# 2. Check if port is listening
sudo lsof -i :8080 | grep smartdisplay || echo "Not listening on 8080"

# 3. Check logs for startup errors
sudo journalctl -u smartdisplay-core.service -n 100 | tail -30

# 4. If process is running but health fails, try restart
sudo systemctl restart smartdisplay-core.service
sleep 5
curl -s http://localhost:8080/health

# 5. If still fails after 3 restarts, ROLLBACK
sudo systemctl stop smartdisplay-core.service
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core
sudo systemctl start smartdisplay-core.service
```

### Scenario 4: Kiosk Won't Display

**Symptoms:** Chromium doesn't launch or closes immediately

**Steps:**

```bash
# 1. Check kiosk service status
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service

# 2. Check kiosk logs
sudo journalctl --user-unit=smartdisplay-kiosk.service -M smartdisplay-ui -n 50

# 3. Verify backend is healthy
curl -s http://localhost:8080/health | jq .status

# 4. Test Chromium manually
su - smartdisplay-ui
chromium-browser --version
chromium-browser --kiosk http://localhost:8090 &
# Should open window; close with Ctrl+C

# 5. Restart kiosk service
sudo systemctl --user -M smartdisplay-ui restart smartdisplay-kiosk.service

# 6. Check if backend binary needs rollback
# (kiosk failure usually isn't due to backend upgrade, but verify)
curl -s http://localhost:8080/health | jq .version.version
```

### Scenario 5: Data Corruption During Upgrade

**Symptoms:** Config files unreadable or logs show data errors

**Steps:**

```bash
# 1. Stop services immediately
sudo systemctl stop smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service

# 2. Check integrity
sudo ls -la /opt/smartdisplay-core/

# 3. Restore full config backup (created in Step 3 of upgrade)
sudo rm -rf /opt/smartdisplay-core
sudo cp -r /opt/smartdisplay-core.backup.TIMESTAMP /opt/smartdisplay-core

# 4. Restore binary
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core

# 5. Restart
sudo systemctl restart smartdisplay-core.service
```

---

## Version Management

### Versioning Scheme

SmartDisplay uses **semantic versioning** with RC tagging:

```
v1.0.0-rc1      Release Candidate 1
v1.0.0-rc2      Release Candidate 2 (if needed)
v1.0.0          Stable Release
v1.0.1          Patch Release
v1.1.0          Minor Release
v2.0.0          Major Release
```

### Version File Location

```
internal/version/version.go
```

**Current version constants:**
```go
const (
    Version = "1.0.0-rc1"
    Commit = "..."
    BuildDate = "..."
)
```

### Checking Version

```bash
# From health endpoint
curl -s http://localhost:8080/health | jq .version

# From binary directly (if --version flag exists)
/opt/smartdisplay-core/smartdisplay-core --version

# From logs at startup
sudo journalctl -u smartdisplay-core.service -n 1 | grep -i version
```

### Backup Naming

Always use version in backup names:

```bash
# Good naming:
smartdisplay-core.v1.0.0-rc1.backup
smartdisplay-core.v1.0.0-rc1.backup.2024-01-04T14-32-18Z

# Avoid:
smartdisplay-core.old
smartdisplay-core.backup
```

---

## Troubleshooting

### "Cannot restore: backup not found"

```bash
# List available backups
ls -lh /opt/smartdisplay-core.backup.*
ls -lh /opt/smartdisplay-core/smartdisplay-core*

# If no backups exist, try last git tag
git describe --tags

# Or restore from release tarball
tar -xzf smartdisplay-core-v1.0.0-rc1.tar.gz -C /opt/smartdisplay-core
```

### "systemctl fails to start service"

```bash
# Check systemd configuration syntax
sudo systemd-analyze verify smartdisplay-core.service

# Check if service file exists
sudo ls -l /etc/systemd/system/smartdisplay-core.service

# Reload systemd daemon
sudo systemctl daemon-reload

# Try manual start to see actual error
sudo -u smartdisplay /opt/smartdisplay-core/smartdisplay-core
```

### "Health check times out"

```bash
# Check if port is actually open
sudo netstat -tlnp | grep 8080
sudo ss -tlnp | grep 8080

# Check backend service status
sudo systemctl status smartdisplay-core.service

# Check firewall
sudo iptables -L -n | grep 8080
sudo ufw status | grep 8080

# If firewall blocking, allow port
sudo ufw allow 8080
```

### "Multiple backup files, not sure which is current"

```bash
# Check file timestamps
ls -lhtr /opt/smartdisplay-core/smartdisplay-core*

# Last modified is most recent
# Check binary size matches expected
stat /opt/smartdisplay-core/smartdisplay-core
stat /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup

# Run both and check version
/opt/smartdisplay-core/smartdisplay-core --version 2>&1
/opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup --version 2>&1
```

---

## Quick Reference

### Upgrade (Normal Path)

```bash
# Verify health
curl http://localhost:8080/health | jq .status

# Stop services
sudo systemctl stop smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service

# Backup
sudo cp /opt/smartdisplay-core/smartdisplay-core \
        /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup

# Deploy
sudo cp dist/smartdisplay-core-linux-arm32v7 /opt/smartdisplay-core/smartdisplay-core
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core/smartdisplay-core
sudo chmod 755 /opt/smartdisplay-core/smartdisplay-core

# Start
sudo systemctl start smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui start smartdisplay-kiosk.service

# Verify
curl http://localhost:8080/health | jq .version.version
```

### Rollback (Emergency)

```bash
# Stop
sudo systemctl stop smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service

# Restore
sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
        /opt/smartdisplay-core/smartdisplay-core

# Start
sudo systemctl start smartdisplay-core.service
sudo systemctl --user -M smartdisplay-ui start smartdisplay-kiosk.service

# Verify
curl http://localhost:8080/health | jq .status
```

---

## See Also

- [INSTALL.md](INSTALL.md) - Initial setup
- [BUILD.md](BUILD.md) - How to build binaries
- [RELEASE_FREEZE.md](RELEASE_FREEZE.md) - RC policy
- [smartdisplay-core.service](smartdisplay-core.service) - Backend systemd unit
- [deploy/smartdisplay-kiosk.service](deploy/smartdisplay-kiosk.service) - Kiosk systemd unit
