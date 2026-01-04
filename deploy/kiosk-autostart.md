# SmartDisplay Kiosk Autostart (Raspberry Pi)

## 1. Install Chromium
```
sudo apt update
sudo apt install -y chromium-browser
```

## 2. Create dedicated user
```
sudo adduser --disabled-password --gecos "" smartdisplay-ui
```

## 3. Enable autologin for smartdisplay-ui
- Edit `/etc/lightdm/lightdm.conf` (or your display manager config):
  - Set:
    ```
    [Seat:*]
    autologin-user=smartdisplay-ui
    ```

## 4. Deploy kiosk scripts and service
```
sudo cp -r deploy /home/smartdisplay-ui/
sudo chown -R smartdisplay-ui:smartdisplay-ui /home/smartdisplay-ui/deploy
```

## 5. Enable systemd user service
```
su - smartdisplay-ui
systemctl --user enable /home/smartdisplay-ui/deploy/smartdisplay-kiosk.service
systemctl --user start smartdisplay-kiosk.service
```

## 6. To stop kiosk
```
systemctl --user stop smartdisplay-kiosk.service
```

---
- Kiosk will launch Chromium in fullscreen on login.
- To revert, disable the service and remove the user if desired.
