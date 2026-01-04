#!/bin/bash
# Launch Chromium in kiosk mode for SmartDisplay
chromium-browser \
  --kiosk \
  --incognito \
  --disable-pinch \
  --overscroll-history-navigation=0 \
  --check-for-update-interval=31536000 \
  http://localhost:8090 &
