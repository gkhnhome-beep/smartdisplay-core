# SmartDisplay Systemd Service Recommendations (Raspberry Pi)

Add or adjust these options in your systemd service file for best stability:

```
[Service]
# File descriptor limit
LimitNOFILE=4096

# Restart delay (seconds)
RestartSec=5

# Set process priority (lower is higher priority, 0 is default)
Nice=5

# Optional: CPU usage limit (50% of one core)
CPUQuota=50%

# Optional: Memory usage limit (256MB)
MemoryMax=256M
```

- These settings help prevent resource exhaustion and improve recovery.
- Adjust values as needed for your hardware and workload.
