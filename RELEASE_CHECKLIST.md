# SmartDisplay Core v1.0.0-rc1 Release Checklist

**Date:** January 4, 2026  
**Version:** v1.0.0-rc1  
**Status:** Ready for Release Candidate

---

## 1. Build Verification

### Build Artifacts
- [ ] `dist/smartdisplay-core-linux-amd64` exists and is executable
  ```bash
  file dist/smartdisplay-core-linux-amd64
  # Expected: ELF 64-bit LSB executable, x86-64
  ls -lh dist/smartdisplay-core-linux-amd64
  ```

- [ ] `dist/smartdisplay-core-linux-arm32v7` exists and is executable
  ```bash
  file dist/smartdisplay-core-linux-arm32v7
  # Expected: ELF 32-bit LSB executable, ARM
  ls -lh dist/smartdisplay-core-linux-arm32v7
  ```

### Version Embedding
- [ ] Build script `build.sh` runs without errors
  ```bash
  chmod +x build.sh
  ./build.sh
  # Expected: Both binaries built successfully
  ```

- [ ] Version metadata embedded correctly
  ```bash
  strings dist/smartdisplay-core-linux-amd64 | grep -i "1.0.0-rc1"
  # Expected: Version string found in binary
  ```

### Reproducibility
- [ ] Build output consistent (same commit hash on re-builds)
  ```bash
  ./build.sh
  # Compare commit hash from multiple builds
  ```

---

## 2. Backend Service Tests

### Health Endpoint
- [ ] `/health` endpoint responds with correct status
  ```bash
  sudo systemctl start smartdisplay-core.service
  curl -s http://localhost:8080/health | jq .
  # Expected: {"status": "healthy", "version": {...}}
  ```

- [ ] Version metadata in health response
  ```bash
  curl -s http://localhost:8080/health | jq .version
  # Expected: {
  #   "version": "1.0.0-rc1",
  #   "commit": "...",
  #   "build_date": "..."
  # }
  ```

### Service Stability
- [ ] Service starts without errors
  ```bash
  sudo systemctl start smartdisplay-core.service
  sleep 2
  sudo systemctl status smartdisplay-core.service
  # Expected: active (running)
  ```

- [ ] Service auto-restarts on failure
  ```bash
  # Kill process
  sudo systemctl kill smartdisplay-core.service
  sleep 3
  # Should restart automatically
  sudo systemctl status smartdisplay-core.service
  # Expected: active (running) with PID changed
  ```

- [ ] No segfaults or panics in logs
  ```bash
  sudo journalctl -u smartdisplay-core.service | grep -i "panic\|fatal\|segfault"
  # Expected: no output
  ```

### Resource Limits
- [ ] Process respects file descriptor limits
  ```bash
  sudo journalctl -u smartdisplay-core.service -n 20 | grep -i "limit"
  # Expected: no resource limit errors
  ```

- [ ] Memory usage stable (no leaks)
  ```bash
  ps aux | grep smartdisplay-core | grep -v grep
  # Check RSS/VSZ columns - should be stable over time
  ```

---

## 3. Systemd Service Configuration

### Backend Service File
- [ ] `/etc/systemd/system/smartdisplay-core.service` syntax valid
  ```bash
  sudo systemd-analyze verify /etc/systemd/system/smartdisplay-core.service
  # Expected: no errors
  ```

- [ ] Service has proper After/Wants dependencies
  ```bash
  grep -E "^(After|Wants)" /etc/systemd/system/smartdisplay-core.service
  # Expected: After=network.target
  ```

- [ ] Restart policy configured
  ```bash
  grep -E "^Restart" /etc/systemd/system/smartdisplay-core.service
  # Expected: Restart=always, RestartSec=5
  ```

- [ ] User/permissions correct
  ```bash
  grep "^User=" /etc/systemd/system/smartdisplay-core.service
  # Expected: User=smartdisplay
  ```

### Kiosk Service File
- [ ] `/home/smartdisplay-ui/.config/systemd/user/smartdisplay-kiosk.service` syntax valid
  ```bash
  sudo -u smartdisplay-ui systemctl --user validate smartdisplay-kiosk.service
  # Expected: no errors
  ```

- [ ] Kiosk waits for backend (Wants=smartdisplay-core.service)
  ```bash
  grep "^Wants=" ~/.config/systemd/user/smartdisplay-kiosk.service
  # Expected: smartdisplay-core.service listed
  ```

- [ ] Network dependency present
  ```bash
  grep "^After=" ~/.config/systemd/user/smartdisplay-kiosk.service
  # Expected: includes network-online.target
  ```

---

## 4. UI & Frontend Tests

### Web Assets
- [ ] `/web/index.html` present and valid HTML5
  ```bash
  file web/index.html
  # Expected: HTML document
  head -5 web/index.html | grep "<!DOCTYPE html"
  # Expected: DOCTYPE present
  ```

- [ ] All JavaScript files present
  ```bash
  ls -lh web/js/*.js
  # Expected: bootstrap.js, api.js, store.js, *Controller.js, viewManager.js
  ```

- [ ] All CSS files present
  ```bash
  ls -lh web/styles/*.css
  # Expected: main.css
  ```

### UI Endpoint
- [ ] UI responds on `http://localhost:8090`
  ```bash
  curl -s http://localhost:8090/ | head -5
  # Expected: HTML5 doctype
  ```

- [ ] All JavaScript loads without console errors
  ```bash
  # Open in kiosk browser
  # Open DevTools (F12)
  # Expected: No red error messages in console
  ```

### View Functionality
- [ ] FirstBootView loads and navigates
  ```bash
  # Manually test in browser or monitor API calls
  # Expected: Backend-driven step navigation works
  ```

- [ ] HomeView displays clock and state
  ```bash
  # Check home screen renders
  # Expected: Clock updates every second
  ```

- [ ] AlarmView responds to state changes
  ```bash
  # Trigger alarm state change
  # Expected: UI reflects alarm mode correctly
  ```

- [ ] GuestView handles 6-state flow
  ```bash
  # Test guest request → approve/deny → timeout
  # Expected: All states display and auto-transition
  ```

- [ ] MenuView shows role-based items
  ```bash
  # Change backend role
  # Expected: Menu items update accordingly
  ```

### Accessibility
- [ ] Reduced-motion respected (no animations)
  ```bash
  # Enable in browser: Preferences > Accessibility > "Reduce motion"
  # Expected: Animations disabled
  ```

- [ ] High-contrast mode works
  ```bash
  # Enable: Preferences > Accessibility > "Increase contrast"
  # Expected: Text readable, colors high-contrast
  ```

- [ ] Large-text support active
  ```bash
  # Inspect computed font size at 150dpi
  # Expected: 1.5x scaling applied
  ```

---

## 5. Systemd Service Lifecycle Tests

### Service Startup
- [ ] Backend starts before kiosk
  ```bash
  sudo systemctl restart smartdisplay-core.service
  sudo systemctl --user -M smartdisplay-ui restart smartdisplay-kiosk.service
  sleep 3
  # Check both are running
  sudo systemctl status smartdisplay-core.service
  sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service
  # Expected: both active (running)
  ```

- [ ] Kiosk waits for backend health
  ```bash
  # Stop backend
  sudo systemctl stop smartdisplay-core.service
  # Kiosk should still run but show connection error
  # Restart backend
  sudo systemctl start smartdisplay-core.service
  # Kiosk should reconnect automatically
  # Expected: Auto-recovery works
  ```

### Service Stopping
- [ ] Graceful shutdown (no forceful kills)
  ```bash
  sudo systemctl stop smartdisplay-core.service
  # Check logs for clean shutdown
  sudo journalctl -u smartdisplay-core.service -n 5 | grep -i "exit\|stop"
  # Expected: Graceful shutdown logged
  ```

- [ ] Kiosk stops cleanly
  ```bash
  sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service
  sleep 2
  ps aux | grep chromium | grep -v grep
  # Expected: Chromium process gone
  ```

### Service Restart
- [ ] `systemctl restart` works both directions
  ```bash
  sudo systemctl restart smartdisplay-core.service
  sleep 3
  curl -s http://localhost:8080/health | jq .status
  # Expected: healthy
  ```

- [ ] Auto-restart on crash (Restart=always)
  ```bash
  PID=$(pgrep -f smartdisplay-core)
  sudo kill -9 $PID
  sleep 3
  NEW_PID=$(pgrep -f smartdisplay-core)
  # Expected: PID changed, service restarted
  ```

---

## 6. Rollback Testing

### Backup Creation
- [ ] Backup binary created before upgrade
  ```bash
  sudo cp /opt/smartdisplay-core/smartdisplay-core \
          /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup
  ls -lh /opt/smartdisplay-core/smartdisplay-core*
  # Expected: Both files exist
  ```

### Rollback Execution
- [ ] Rollback restores previous binary
  ```bash
  # Simulate failed upgrade
  sudo cp /opt/smartdisplay-core/smartdisplay-core.v1.0.0-rc1.backup \
          /opt/smartdisplay-core/smartdisplay-core
  
  sudo systemctl restart smartdisplay-core.service
  sleep 3
  
  curl -s http://localhost:8080/health | jq .version.version
  # Expected: v1.0.0-rc1 (or previous version)
  ```

### Service Recovery
- [ ] Services recover after rollback
  ```bash
  sudo systemctl status smartdisplay-core.service
  sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service
  # Expected: both active (running)
  ```

- [ ] Health endpoint responsive post-rollback
  ```bash
  curl -v http://localhost:8080/health 2>&1 | grep "< HTTP"
  # Expected: HTTP/1.1 200 OK
  ```

---

## 7. Documentation Review

### Installation Guide
- [ ] `INSTALL.md` covers all steps
  ```bash
  grep -c "^## Part" INSTALL.md
  # Expected: at least 5 major sections
  ```

- [ ] Commands are accurate
  ```bash
  # Spot-check usernames, paths, service names
  grep "smartdisplay-core.service" INSTALL.md
  grep "/opt/smartdisplay-core" INSTALL.md
  # Expected: no typos or wrong paths
  ```

### Build Guide
- [ ] `BUILD.md` documents cross-compilation
  ```bash
  grep -E "linux/amd64|linux/arm" BUILD.md
  # Expected: Both targets documented
  ```

- [ ] Version embedding documented
  ```bash
  grep -c "ldflags" BUILD.md
  # Expected: At least one reference
  ```

### Upgrade Guide
- [ ] `UPGRADE.md` covers pre-checks
  ```bash
  grep "Pre-Upgrade" UPGRADE.md
  # Expected: Section present
  ```

- [ ] Rollback procedure documented
  ```bash
  grep "Rollback Procedure" UPGRADE.md
  # Expected: Full section with steps
  ```

### Release Policy
- [ ] `RELEASE_FREEZE.md` defines RC rules
  ```bash
  grep "What IS\|What IS NOT" RELEASE_FREEZE.md
  # Expected: Clear allowed/disallowed changes
  ```

---

## 8. Git Repository State

### Commits
- [ ] All changes committed
  ```bash
  git status
  # Expected: working tree clean
  ```

- [ ] No uncommitted files
  ```bash
  git diff --name-only
  # Expected: no output
  ```

- [ ] Relevant commits have descriptive messages
  ```bash
  git log --oneline -20
  # Expected: Clear, semantic commit messages
  ```

### Tags
- [ ] No conflicting tags
  ```bash
  git tag -l | grep v1.0.0
  # Expected: no existing v1.0.0 or v1.0.0-rc1 tag
  ```

- [ ] Repository ready for tagging
  ```bash
  git tag -a v1.0.0-rc1 -m "Release Candidate 1"
  # (Run this only after checklist approval)
  ```

---

## 9. Documentation Completeness

- [ ] `RELEASE_CHECKLIST.md` (this file) present
- [ ] `RELEASE_NOTES.md` documents scope and limitations
- [ ] Git tagging instructions documented
- [ ] All files referenced in documentation exist

---

## Sign-Off

### Checklist Owner

**Verified by:** ________________  
**Date:** ________________  
**Time:** ________________  

### Release Approval

**Release Manager:** ________________  
**Approval Date:** ________________  
**Release Candidate Tag:** v1.0.0-rc1  

---

## Next Steps

1. **All items checked?**
   - If NO: Fix issues and re-verify
   - If YES: Proceed to git tagging

2. **Create git tag:**
   ```bash
   git tag -a v1.0.0-rc1 -m "Release Candidate 1

   This RC includes:
   - Kiosk-safe vanilla JS/HTML/CSS frontend
   - Systemd service configuration
   - Installation and upgrade guides
   - Full accessibility support
   
   Known Limitations:
   - See RELEASE_NOTES.md"
   
   git push origin v1.0.0-rc1
   ```

3. **Publish release:**
   - [ ] Create GitHub Release from tag
   - [ ] Attach binary artifacts
   - [ ] Link to RELEASE_NOTES.md
   - [ ] Announce RC1 publicly

4. **Start RC feedback cycle:**
   - [ ] Wait for community testing
   - [ ] Collect issues in GitHub Issues
   - [ ] Fix critical bugs only (see RELEASE_FREEZE.md)
   - [ ] Bump to rc2 if issues found

---

## See Also

- [RELEASE_NOTES.md](RELEASE_NOTES.md) - RC1 scope and limitations
- [RELEASE_FREEZE.md](RELEASE_FREEZE.md) - RC policy
- [INSTALL.md](INSTALL.md) - Installation steps
- [UPGRADE.md](UPGRADE.md) - Upgrade/rollback procedures
- [BUILD.md](BUILD.md) - Build instructions
