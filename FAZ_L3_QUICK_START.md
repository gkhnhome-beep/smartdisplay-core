# FAZ L3 Quick Setup Guide

## 🚀 Quick Start (5 Minutes)

### Step 1: Configure Home Assistant Automation

1. Copy `configs/homeassistant_guest_automation.yaml` to your HA config directory
2. Edit the file and replace:
   - `<SmartDisplay_IP>`: Your SmartDisplay device IP (e.g., `192.168.1.100`)
   - `<YOUR_HA_TOKEN>`: Your HA long-lived access token

**Generate HA Token**:
- Open Home Assistant
- Click your profile (bottom left)
- Scroll to "Long-Lived Access Tokens"
- Click "Create Token"
- Name it "SmartDisplay Guest Approval"
- Copy the token

### Step 2: Restart Home Assistant

```bash
# Via UI: Settings → System → Restart
# Or via CLI:
ha core restart
```

### Step 3: Verify Mobile App

Ensure your mobile device has the Home Assistant Companion app installed and registered.

**Check registration**:
- Developer Tools → States
- Search for `notify.mobile_app_`
- Note your device name (e.g., `mobile_app_johns_phone`)

### Step 4: Test the Flow

1. **Arm your alarm** in Home Assistant
2. On SmartDisplay:
   - Tap "Request Guest Access"
   - Select target user (use your mobile app notify service name)
3. **Check your phone** for notification with Approve/Reject buttons
4. Tap **Approve**
5. Verify:
   - Alarm disarms
   - SmartDisplay grants access
   - You receive confirmation notification

---

## 🔧 Troubleshooting

### No Notification Received

**Check**:
```yaml
# In HA automation, verify service name matches:
notify.mobile_app_johns_phone  # Your actual service
```

**Fix**: Update target user in SmartDisplay to match your notify service name (without the `notify.` prefix).

### Buttons Don't Work

**Check**:
- HA: Settings → Automations & Scenes
- Verify "SmartDisplay Guest Access" is **enabled**

**Test**: Developer Tools → Events → Listen for `mobile_app_notification_action`

### SmartDisplay Doesn't Respond

**Test endpoint**:
```bash
curl -X POST http://<SmartDisplay_IP>:8090/api/guest/approve \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"request_id":"test","decision":"reject"}'
```

**Expected response**: `{"success":true,"data":{"result":"ok"}}`

**If fails**:
- Check SmartDisplay IP is correct
- Verify port 8090 is accessible
- Check token is valid

### Alarm Not Disarmed

**Check SmartDisplay logs**:
```
guest approval: alarmo disarm requested successfully
```

**If error**:
- Verify Alarmo integration configured in SmartDisplay
- Check HA/Alarmo is reachable from SmartDisplay
- Verify HA token has permissions to control alarm

---

## 📋 Configuration Summary

### Home Assistant

**Required**:
- Home Assistant Companion mobile app installed
- Long-lived access token created
- Automation configured with correct SmartDisplay IP
- REST command configured with valid token

### SmartDisplay

**Required**:
- Environment variables set:
  - `HA_BASE_URL`: Your Home Assistant URL
  - `HA_TOKEN`: Long-lived access token
- Alarmo adapter configured
- Port 8090 accessible from HA

---

## 🔐 Security Checklist

- [ ] HA token stored in `secrets.yaml` (not hardcoded)
- [ ] SmartDisplay only accessible on local network
- [ ] Firewall rules restrict access to SmartDisplay
- [ ] Token rotated periodically
- [ ] HTTPS used if exposed outside local network

---

## 📞 Support

**View detailed documentation**: `FAZ_L3_COMPLETE.md`

**Common log locations**:
- SmartDisplay: Check terminal output or logs directory
- Home Assistant: Settings → System → Logs
