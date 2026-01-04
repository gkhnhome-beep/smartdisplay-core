# SmartDisplay Core Installation Guide

Complete step-by-step setup for SmartDisplay Core v1.0.0-rc1 on Raspberry Pi or x86-64 Linux.

## System Requirements

- **OS:** Debian 11+ (Raspberry Pi OS or Ubuntu 20.04+)
- **Hardware:** Raspberry Pi 4+ (or x86-64 with 2GB+ RAM)
- **Browser:** Chromium (for kiosk UI)
- **Display Manager:** LightDM (for autologin)

---

## Part 1: Backend Service Setup

### Step 1.1: Create Service User

```bash
# Create unprivileged user for backend service
sudo adduser --system --group --home /opt/smartdisplay-core smartdisplay
```

### Step 1.2: Prepare Directory Structure

```bash
# Create service directory
sudo mkdir -p /opt/smartdisplay-core

# Build or download binary (from dist/)
sudo cp dist/smartdisplay-core-linux-arm32v7 /opt/smartdisplay-core/smartdisplay-core

# Set permissions
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core
sudo chmod 755 /opt/smartdisplay-core/smartdisplay-core

# Copy config files (if any)
sudo cp configs/*.json /opt/smartdisplay-core/
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core/*.json
```

### Step 1.3: Install Backend Service

```bash
# Copy systemd service file
sudo cp smartdisplay-core.service /etc/systemd/system/

# Reload systemd daemon
sudo systemctl daemon-reload

# Enable service to start at boot
sudo systemctl enable smartdisplay-core.service

# Start service
sudo systemctl start smartdisplay-core.service

# Verify it's running
sudo systemctl status smartdisplay-core.service
```

### Step 1.4: Check Backend Health

```bash
# Monitor logs in real-time
sudo journalctl -u smartdisplay-core.service -f

# Check health endpoint (after service starts)
curl http://localhost:8080/health | jq .
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0-rc1",
  "commit": "a1b2c3d",
  "build_date": "2024-01-15T14:32:18Z"
}
```

---

## Part 2: Kiosk UI Setup

### Step 2.1: Install Display Server & Browser

```bash
# Update package lists
sudo apt update

# Install Chromium browser
sudo apt install -y chromium-browser

# Install X11 and display tools
sudo apt install -y xserver-xorg xinit lightdm

# Verify Chromium installation
chromium-browser --version
```

### Step 2.2: Create Kiosk UI User

```bash
# Create unprivileged user for kiosk display
sudo adduser --disabled-password --gecos "" smartdisplay-ui

# Verify user created
id smartdisplay-ui
```

### Step 2.3: Deploy Kiosk Scripts

```bash
# Copy kiosk directory to UI user home
sudo cp -r deploy /home/smartdisplay-ui/

# Set permissions
sudo chown -R smartdisplay-ui:smartdisplay-ui /home/smartdisplay-ui/deploy
sudo chmod 755 /home/smartdisplay-ui/deploy/*.sh

# Verify scripts are executable
ls -l /home/smartdisplay-ui/deploy/
```

Expected files:
- `kiosk-start.sh` - Launch Chromium
- `kiosk-stop.sh` - Stop Chromium
- `smartdisplay-kiosk.service` - systemd user service

### Step 2.4: Enable Display Autologin

Edit LightDM configuration:
```bash
sudo nano /etc/lightdm/lightdm.conf
```

Find and update:
```ini
[Seat:*]
autologin-user=smartdisplay-ui
autologin-user-timeout=0
```

Save (Ctrl+O, Enter, Ctrl+X).

### Step 2.5: Enable Kiosk Service (User-Level)

```bash
# Switch to kiosk user
su - smartdisplay-ui

# Create systemd user service directory
mkdir -p ~/.config/systemd/user/

# Copy service file to user directory
cp ~/deploy/smartdisplay-kiosk.service ~/.config/systemd/user/

# Reload user systemd
systemctl --user daemon-reload

# Enable service to start on user login
systemctl --user enable smartdisplay-kiosk.service

# Start service manually (will auto-start on next login)
systemctl --user start smartdisplay-kiosk.service

# Check status
systemctl --user status smartdisplay-kiosk.service

# View logs
journalctl --user -u smartdisplay-kiosk.service -f

# Return to root
exit
```

---

## Part 3: Verify Full Stack

### Step 3.1: Check Both Services

```bash
# Backend (system-level)
sudo systemctl status smartdisplay-core.service

# Kiosk (as user)
sudo -u smartdisplay-ui systemctl --user status smartdisplay-kiosk.service
```

### Step 3.2: Monitor Logs

Terminal 1 (Backend):
```bash
sudo journalctl -u smartdisplay-core.service -f
```

Terminal 2 (Kiosk):
```bash
sudo journalctl --user -u smartdisplay-kiosk.service -e -f
```

### Step 3.3: Test Health & UI

```bash
# Test backend health (should return 200 with version)
curl -s http://localhost:8080/health | jq .

# Test UI is serving (should return HTML)
curl -s http://localhost:8090/ | head -20
```

### Step 3.4: Reboot & Verify Auto-Start

```bash
# Reboot system
sudo reboot

# After reboot, check services started automatically
sudo systemctl status smartdisplay-core.service
sudo -u smartdisplay-ui systemctl --user status smartdisplay-kiosk.service

# Verify backend is responsive
curl http://localhost:8080/health
```

---

## Part 4: Service Management

### View Service Status

```bash
# Backend
sudo systemctl status smartdisplay-core.service

# Kiosk
sudo systemctl --user -M smartdisplay-ui status smartdisplay-kiosk.service
```

### Restart Services

```bash
# Restart backend (waits for health, auto-restart if fails)
sudo systemctl restart smartdisplay-core.service

# Restart kiosk (waits for network before starting)
sudo systemctl --user -M smartdisplay-ui restart smartdisplay-kiosk.service
```

### Stop Services

```bash
# Stop backend
sudo systemctl stop smartdisplay-core.service

# Stop kiosk
sudo systemctl --user -M smartdisplay-ui stop smartdisplay-kiosk.service
```

### View Logs

```bash
# Last 100 lines of backend logs
sudo journalctl -u smartdisplay-core.service -n 100

# Real-time backend logs
sudo journalctl -u smartdisplay-core.service -f

# Kiosk logs (as root viewing user service)
sudo journalctl --user-unit=smartdisplay-kiosk.service -M smartdisplay-ui -f

# All logs since system boot
sudo journalctl -b -n 200
```

---

## Part 5: Troubleshooting

### Backend Won't Start

```bash
# Check file permissions
ls -l /opt/smartdisplay-core/smartdisplay-core

# Should be owned by smartdisplay:smartdisplay and executable
# If not:
sudo chown smartdisplay:smartdisplay /opt/smartdisplay-core/smartdisplay-core
sudo chmod 755 /opt/smartdisplay-core/smartdisplay-core

# Check systemd error
sudo journalctl -u smartdisplay-core.service -n 50

# Test binary directly
sudo -u smartdisplay /opt/smartdisplay-core/smartdisplay-core --help
```

### Kiosk Won't Launch

```bash
# Check Chromium is installed
chromium-browser --version

# Check script permissions
ls -l /home/smartdisplay-ui/deploy/kiosk-start.sh

# Test script manually
/home/smartdisplay-ui/deploy/kiosk-start.sh

# Check user service logs
sudo journalctl --user-unit=smartdisplay-kiosk.service -M smartdisplay-ui -n 50

# Verify X11 is running
ps aux | grep X
```

### Chromium Crashes or Won't Fullscreen

```bash
# Check if another Chromium instance is running
ps aux | grep chromium

# Kill all instances
sudo pkill -9 chromium-browser

# Check Chromium versions (app vs system)
chromium-browser --version
/snap/bin/chromium --version  # if installed as snap

# Try launching manually to test
chromium-browser --kiosk --incognito http://localhost:8090

# If crashes, check system logs
dmesg | tail -50
```

### Network Issues (Kiosk Waits for Network)

The kiosk service includes `network-online.target` dependency. Check:
```bash
# Wait for network
sudo systemctl status systemd-networkd-wait-online.service

# Check connectivity
ping google.com

# Check localhost route
ip route | grep localhost
```

---

## Part 6: Configuration Variables

All paths are configurable:

| Path | Default | User | Purpose |
|------|---------|------|---------|
| Binary | `/opt/smartdisplay-core/smartdisplay-core` | smartdisplay | Backend executable |
| Working Dir | `/opt/smartdisplay-core` | smartdisplay | Config/data location |
| Kiosk Scripts | `/home/smartdisplay-ui/deploy/` | smartdisplay-ui | Launch/stop scripts |
| Kiosk URL | `http://localhost:8090` | smartdisplay-ui | Web UI endpoint |
| Backend URL | `http://localhost:8080` | frontend | API endpoint |
| Logs | journalctl (systemd) | root | Service logs |

To modify, edit service files:
- Backend: `/etc/systemd/system/smartdisplay-core.service`
- Kiosk: `~smartdisplay-ui/.config/systemd/user/smartdisplay-kiosk.service`

Then reload:
```bash
sudo systemctl daemon-reload                              # backend
sudo -u smartdisplay-ui systemctl --user daemon-reload  # kiosk
```

---

## Part 7: Uninstall (Complete Removal)

To fully remove SmartDisplay:

```bash
# Stop services
sudo systemctl stop smartdisplay-core.service
sudo systemctl stop smartdisplay-kiosk.service

# Disable auto-start
sudo systemctl disable smartdisplay-core.service
sudo systemctl --user disable smartdisplay-kiosk.service

# Remove systemd files
sudo rm /etc/systemd/system/smartdisplay-core.service
sudo rm /home/smartdisplay-ui/.config/systemd/user/smartdisplay-kiosk.service

# Reload systemd
sudo systemctl daemon-reload

# Remove users
sudo deluser smartdisplay
sudo deluser smartdisplay-ui

# Remove directories
sudo rm -rf /opt/smartdisplay-core
sudo rm -rf /home/smartdisplay-ui/deploy

# Revert autologin (edit /etc/lightdm/lightdm.conf)
sudo nano /etc/lightdm/lightdm.conf
# Remove autologin-user and autologin-user-timeout lines
```

---

## Service Dependency Tree

```
multi-user.target
└── smartdisplay-core.service (Restart=always)
    After: network.target

default.target (user session)
└── smartdisplay-kiosk.service (Restart=always)
    After: graphical-session.target, network-online.target
    Wants: smartdisplay-core.service
```

**Ordering:**
1. Network becomes available
2. Backend service starts
3. Display manager starts
4. smartdisplay-ui auto-logs in
5. Kiosk service starts (waits for network + graphical session)
6. Chromium opens http://localhost:8090

---

## See Also

- [smartdisplay-core.service](smartdisplay-core.service) - Backend systemd unit
- [deploy/smartdisplay-kiosk.service](deploy/smartdisplay-kiosk.service) - Kiosk systemd user unit
- [deploy/kiosk-autostart.md](deploy/kiosk-autostart.md) - Original kiosk setup notes
- [BUILD.md](BUILD.md) - How to build the binary
- [README.md](README.md) - Project overview
- [RELEASE_FREEZE.md](RELEASE_FREEZE.md) - Release candidate policy
