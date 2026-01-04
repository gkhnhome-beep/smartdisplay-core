package health

import (
	"encoding/json"
	"net/http"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/version"
	"time"
)

var startTime = time.Now()

var coordinator *system.Coordinator

// SetCoordinator allows main to provide the coordinator for health info
func SetCoordinator(c *system.Coordinator) {
	coordinator = c
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uptime := int(time.Since(startTime).Seconds())
	haConnected := false
	hardwareReady := false
	if coordinator != nil {
		haConnected = coordinator.HA != nil && coordinator.HA.IsConnected()
		// Hardware ready if all devices are ready
		ready := true
		for _, dev := range coordinator.HALRegistry.DeviceHealthReport() {
			if !dev.Ready {
				ready = false
				break
			}
		}
		hardwareReady = ready
	}
	degraded := false
	if coordinator != nil {
		degraded = coordinator.DegradedMode()
	}
	response := map[string]interface{}{
		"status":         "ok",
		"version":        version.Version,
		"uptime_seconds": uptime,
		"ha_connected":   haConnected,
		"hardware_ready": hardwareReady,
		"degraded_mode":  degraded,
	}
	json.NewEncoder(w).Encode(response)
}
