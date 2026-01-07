package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/ha/alarmo"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/security"
	"smartdisplay-core/internal/settings"
	"strings"
	"time"
)

// handleHASettingsSave saves Home Assistant credentials securely (admin-only).
// POST /api/settings/homeassistant
// FAZ S2: Accept server address and token, encrypt token, persist.
// Response NEVER includes token.
func (s *Server) handleHASettingsSave(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	role := getRole(r)
	if role != auth.Admin {
		logger.Error("HA credentials save blocked: insufficient role=" + string(role))
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// Parse request
	var req struct {
		ServerURL string `json:"server_url"`
		Token     string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("HA credentials save error: invalid request=" + err.Error())
		s.respondError(w, r, CodeBadRequest, "invalid json")
		return
	}

	// Validate server URL
	if req.ServerURL == "" {
		logger.Error("HA credentials save error: empty server_url")
		s.respondError(w, r, CodeBadRequest, "server_url required")
		return
	}

	// Validate URL format
	if _, err := url.Parse(req.ServerURL); err != nil {
		logger.Error("HA credentials save error: invalid server_url format")
		s.respondError(w, r, CodeBadRequest, "invalid server_url format")
		return
	}

	// Validate token
	if req.Token == "" {
		logger.Error("HA credentials save error: empty token")
		s.respondError(w, r, CodeBadRequest, "token required")
		return
	}

	// Normalize base URL to avoid double slashes in downstream calls
	cleanURL := strings.TrimRight(req.ServerURL, "/")

	// Encrypt token
	encryptedToken, err := settings.Encrypt(req.Token)
	if err != nil {
		logger.Error("HA credentials encryption failed: " + err.Error())
		// Don't log the token or encrypted token
		s.respondError(w, r, CodeInternalError, "failed to encrypt token")
		audit.Record("ha_config_error", "failed to encrypt token")
		return
	}

	// Create config
	cfg := &settings.HAConfig{
		ServerURL:      cleanURL,
		EncryptedToken: encryptedToken,
		ConfiguredAt:   time.Now(),
	}

	// Save config
	if err := settings.SaveHAConfig(cfg); err != nil {
		logger.Error("HA config save failed: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to save configuration")
		audit.Record("ha_config_error", "failed to save configuration")
		return
	}

	// Log event WITHOUT token
	logger.Info("HA configuration updated: server_url=" + cleanURL)
	audit.Record("ha_config_updated", "server_url="+cleanURL+", token="+security.Redact(req.Token))

	// FAZ S4: Reset global connection state when credentials change
	// New credentials must be tested again before marking as connected
	// Also reset sync state - old entity data is now invalid
	s.runtimeCfg.HaConnected = false
	s.runtimeCfg.HaLastTestedAt = nil
	s.runtimeCfg.InitialSyncDone = false
	s.runtimeCfg.InitialSyncAt = nil
	if err := s.saveRuntimeConfig(); err != nil {
		logger.Error("failed to reset HA connection state: " + err.Error())
	}

	// Reinitialize Alarmo adapter immediately with new credentials to avoid restart requirement
	if s.coord != nil {
		s.coord.AlarmoMu.Lock()
		s.coord.AlarmoAdapter = alarmo.New(cleanURL, req.Token)
		s.coord.AlarmoMu.Unlock()
		logger.Info("alarmo adapter reinitialized with new HA configuration")
	}

	// Respond with safe confirmation (NO token, but include server URL for frontend form update)
	s.respond(w, true, map[string]interface{}{
		"success":       true,
		"message":       "Home Assistant credentials saved securely",
		"configured":    true,
		"configured_at": cfg.ConfiguredAt,
		"server_url":    cleanURL, // Include server URL for frontend form update (no sensitive data)
	}, "", http.StatusOK)
}

// handleHASettingsStatus returns safe HA status information (admin-only).
// GET /api/settings/homeassistant/status
// FAZ S2: Returns is_configured and configured_at, NEVER returns token or server_url.
func (s *Server) handleHASettingsStatus(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	role := getRole(r)
	if role != auth.Admin {
		logger.Error("HA status check blocked: insufficient role=" + string(role))
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	// Get status
	status, err := settings.GetHAStatus()
	if err != nil {
		logger.Error("HA status retrieval failed: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to get status")
		return
	}

	logger.Info("HA status retrieved: is_configured=" + (map[bool]string{true: "yes", false: "no"})[status.IsConfigured])

	// FAZ S4: Include global HA connection state in response
	// FAZ S5: Include initial sync metadata and entity counts
	// FAZ S6: Include runtime health state
	response := map[string]interface{}{
		"is_configured":          status.IsConfigured,
		"configured_at":          status.ConfiguredAt,
		"ha_connected":           s.runtimeCfg.HaConnected,
		"ha_last_tested_at":      s.runtimeCfg.HaLastTestedAt,
		"initial_sync_done":      s.runtimeCfg.InitialSyncDone,
		"initial_sync_at":        s.runtimeCfg.InitialSyncAt,
		"ha_runtime_unreachable": s.healthMonitor.IsUnreachable(),
		"ha_last_seen_at":        s.healthMonitor.GetLastSeenAt(),
	}

	// Add server URL if HA is configured (safe to return, no sensitive data)
	if status.IsConfigured {
		if haConfig, err := settings.LoadHAConfig(); err == nil && haConfig != nil {
			if serverURL, err := settings.DecryptServerURL(); err == nil && serverURL != "" {
				response["server_url"] = serverURL
			}
		}
	}

	// Add safe metadata if sync completed
	if s.runtimeCfg.InitialSyncDone {
		response["ha_meta"] = map[string]interface{}{
			"version":       s.runtimeCfg.HaVersion,
			"time_zone":     s.runtimeCfg.HaTimeZone,
			"location_name": s.runtimeCfg.HaLocationName,
		}
		response["entity_counts"] = map[string]interface{}{
			"lights":   s.runtimeCfg.EntityLights,
			"sensors":  s.runtimeCfg.EntitySensors,
			"switches": s.runtimeCfg.EntitySwitches,
			"others":   s.runtimeCfg.EntityOthers,
		}
	}

	s.respond(w, true, response, "", http.StatusOK)
}

// handleHASettingsTest performs a connection test against Home Assistant (admin-only).
// POST /api/settings/homeassistant/test
// FAZ S3: Connection test engine. Returns success/stage/message, never returns token or URL.
// Stops on first failure. Updates last_tested_at on success.
func (s *Server) handleHASettingsTest(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	role := getRole(r)
	if role != auth.Admin {
		logger.Error("HA test blocked: insufficient role=" + string(role))
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// Log test start
	logger.Info("HA connection test started")

	// Execute test
	result, err := settings.TestHAConnection()
	if err != nil {
		logger.Error("HA connection test error: " + err.Error())
		s.respondError(w, r, CodeInternalError, err.Error())
		audit.Record("ha_test_error", err.Error())
		return
	}

	// Log result
	if result.Success {
		logger.Info("HA connection test succeeded: stage=" + result.Stage)
		audit.Record("ha_test_success", "stage="+result.Stage)

		// FAZ S4: Update global HA connection state
		if result.Stage == settings.StageOK {
			// Connection successful - update runtime state
			s.runtimeCfg.HaConnected = true
			testedAt := time.Now().Format(time.RFC3339)
			s.runtimeCfg.HaLastTestedAt = &testedAt
			// Save updated config
			if err := s.saveRuntimeConfig(); err != nil {
				logger.Error("failed to save HA connection state: " + err.Error())
			}
		}
	} else {
		logger.Error("HA connection test failed: stage=" + result.Stage)
		audit.Record("ha_test_failed", "stage="+result.Stage)

		// FAZ S4: Update global HA connection state
		// Connection failed - mark as disconnected, but keep last_tested_at for reference
		s.runtimeCfg.HaConnected = false
		testedAt := time.Now().Format(time.RFC3339)
		s.runtimeCfg.HaLastTestedAt = &testedAt
		if err := s.saveRuntimeConfig(); err != nil {
			logger.Error("failed to save HA connection state: " + err.Error())
		}
	}

	// Return result (SAFE: no token, no URL, no raw HA responses)
	s.respond(w, true, result, "", http.StatusOK)
}

// handleHAInitialSync performs a one-time initial HA synchronization (admin-only).
// POST /api/settings/homeassistant/sync
// FAZ S5: Bootstrap sync that extracts safe metadata, confirms Alarmo, counts entities.
// Runs ONLY when ha_connected=true AND initial_sync_done=false.
func (s *Server) handleHAInitialSync(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	role := getRole(r)
	if role != auth.Admin {
		logger.Error("HA sync blocked: insufficient role=" + string(role))
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// Check if already synced (idempotency)
	if s.runtimeCfg.InitialSyncDone {
		logger.Info("HA initial sync already completed, skipping")
		s.respond(w, true, map[string]interface{}{
			"success":      true,
			"message":      "Initial sync already completed",
			"already_done": true,
		}, "", http.StatusOK)
		return
	}

	// Check if HA is connected
	if !s.runtimeCfg.HaConnected {
		logger.Error("HA sync blocked: HA not connected")
		s.respondError(w, r, CodeBadRequest, "HA must be connected first (run test)")
		return
	}

	// Log sync start
	logger.Info("HA initial sync started")

	// Execute sync
	result, err := settings.PerformInitialSync()
	if err != nil {
		logger.Error("HA initial sync error: " + err.Error())
		s.respondError(w, r, CodeInternalError, err.Error())
		audit.Record("ha_sync_error", err.Error())
		return
	}

	// Log result
	if result.Success {
		logger.Info("HA initial sync succeeded")
		audit.Record("ha_sync_success", "entities_total="+
			fmt.Sprintf("%d", result.Counts.Lights+result.Counts.Sensors+result.Counts.Switches+result.Counts.Others))

		// FAZ S5: Update runtime config with sync metadata
		s.runtimeCfg.InitialSyncDone = true
		syncTime := time.Now().Format(time.RFC3339)
		s.runtimeCfg.InitialSyncAt = &syncTime

		if result.Meta != nil {
			s.runtimeCfg.HaVersion = result.Meta.Version
			s.runtimeCfg.HaTimeZone = result.Meta.TimeZone
			s.runtimeCfg.HaLocationName = result.Meta.LocationName
		}

		if result.Counts != nil {
			s.runtimeCfg.EntityLights = result.Counts.Lights
			s.runtimeCfg.EntitySensors = result.Counts.Sensors
			s.runtimeCfg.EntitySwitches = result.Counts.Switches
			s.runtimeCfg.EntityOthers = result.Counts.Others
		}

		// Save updated config
		if err := s.saveRuntimeConfig(); err != nil {
			logger.Error("failed to save HA sync state: " + err.Error())
		}

		// Return safe result
		s.respond(w, true, map[string]interface{}{
			"success": true,
			"message": result.Message,
			"meta": map[string]interface{}{
				"version":       result.Meta.Version,
				"time_zone":     result.Meta.TimeZone,
				"location_name": result.Meta.LocationName,
			},
			"entity_counts": map[string]interface{}{
				"lights":   result.Counts.Lights,
				"sensors":  result.Counts.Sensors,
				"switches": result.Counts.Switches,
				"others":   result.Counts.Others,
			},
		}, "", http.StatusOK)
	} else {
		logger.Error("HA initial sync failed: " + result.Message)
		audit.Record("ha_sync_failed", result.Message)

		// Failure does NOT change initial_sync_done (remains false)
		s.respond(w, true, map[string]interface{}{
			"success": false,
			"message": result.Message,
		}, "", http.StatusOK)
	}
}
