// Package settings provides secure credential handling for Home Assistant integration.
// hatestor.go: Connection test engine for Home Assistant (FAZ S3)
package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smartdisplay-core/internal/logger"
	"strings"
	"time"
)

// ConnectionTestResult represents the outcome of a HA connection test.
// FAZ S3: Deterministic test result model.
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Stage   string `json:"stage"` // server_unreachable, auth_failed, api_unavailable, alarmo_missing, ok
	Message string `json:"message"`
}

// TestConnectionStages (internal enum)
const (
	StageServerUnreachable = "server_unreachable"
	StageAuthFailed        = "auth_failed"
	StageAPIUnavailable    = "api_unavailable"
	StageAlarmoMissing     = "alarmo_missing"
	StageOK                = "ok"
)

// TestHAConnection performs a deterministic connection test against Home Assistant.
// STRICT ORDER:
// 1. Server Reachability: GET /api/
// 2. Authentication: GET /api/config with Bearer token
// 3. Core API Sanity: Validate JSON response
// 4. Alarmo Presence: GET /api/states/alarm_control_panel.alarmo (404 is OK)
//
// Stops on first failure. Returns result and updates last_tested_at on success.
// FAZ S3: Connection test engine, callable from API or programmatically.
func TestHAConnection() (*ConnectionTestResult, error) {
	// Load encrypted config
	cfg, err := LoadHAConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load HA config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("HA not configured")
	}

	// Decrypt server URL
	serverURL, err := DecryptServerURL()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt server URL: %w", err)
	}
	if serverURL == "" {
		return nil, fmt.Errorf("server URL not configured")
	}

	// Decrypt token
	token, err := DecryptToken()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("token not configured")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// STAGE 1: Server Reachability
	// GET /api/ - check if server responds with authentication
	reachURL := strings.TrimSuffix(serverURL, "/") + "/api/"
	req, err := http.NewRequest("GET", reachURL, nil)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageServerUnreachable,
			Message: "Failed to prepare request",
		}, nil
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageServerUnreachable,
			Message: "Cannot reach Home Assistant server",
		}, nil
	}
	if resp.Body != nil {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		// Different status codes indicate different issues
		stage := StageServerUnreachable
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			stage = StageAuthFailed
		}
		return &ConnectionTestResult{
			Success: false,
			Stage:   stage,
			Message: fmt.Sprintf("Home Assistant returned %d", resp.StatusCode),
		}, nil
	}

	// STAGE 2: Authentication Validity
	// GET /api/config with Bearer token
	configURL := strings.TrimSuffix(serverURL, "/") + "/api/config"
	req2, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAPIUnavailable,
			Message: "Failed to prepare authentication request",
		}, nil
	}
	req2.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req2)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAuthFailed,
			Message: "Failed to authenticate with Home Assistant",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAuthFailed,
			Message: "Authentication failed. Please check token",
		}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAPIUnavailable,
			Message: "Home Assistant API not accessible",
		}, nil
	}

	// STAGE 3: Core API Sanity
	// Validate response is valid JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAPIUnavailable,
			Message: "Failed to read Home Assistant response",
		}, nil
	}

	var configData interface{}
	if err := json.Unmarshal(body, &configData); err != nil {
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAPIUnavailable,
			Message: "Home Assistant API returned invalid data",
		}, nil
	}

	// STAGE 4: Alarmo Presence (READ ONLY)
	// GET /api/states/alarm_control_panel.alarmo
	// 200 = found, 404 = not found (both OK for now)
	alarmoURL := strings.TrimSuffix(serverURL, "/") + "/api/states/alarm_control_panel.alarmo"
	req, err = http.NewRequest("GET", alarmoURL, nil)
	if err != nil {
		// Non-fatal: Alarmo detection failed, but connection is OK
		return &ConnectionTestResult{
			Success: true,
			Stage:   StageOK,
			Message: "Connected to Home Assistant (Alarmo integration not checked)",
		}, nil
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		// Non-fatal: timeout or network issue, but connection is OK
		return &ConnectionTestResult{
			Success: true,
			Stage:   StageOK,
			Message: "Connected to Home Assistant (Alarmo integration not checked)",
		}, nil
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	// Handle Alarmo detection
	if resp.StatusCode == http.StatusNotFound {
		// Alarmo not found, but HA is accessible
		return &ConnectionTestResult{
			Success: false,
			Stage:   StageAlarmoMissing,
			Message: "Alarmo integration not found in Home Assistant",
		}, nil
	}

	if resp.StatusCode == http.StatusOK {
		// Alarmo found - all checks passed
		result := &ConnectionTestResult{
			Success: true,
			Stage:   StageOK,
			Message: "Successfully connected to Home Assistant with Alarmo integration",
		}

		// Update last_tested_at on success
		now := time.Now()
		cfg.LastTestedAt = &now
		if err := SaveHAConfig(cfg); err != nil {
			// Log but don't fail the test result
			logger.Error("failed to update last_tested_at: " + err.Error())
		}

		return result, nil
	}

	// Alarmo check returned something other than 200/404
	return &ConnectionTestResult{
		Success: true,
		Stage:   StageOK,
		Message: "Connected to Home Assistant (Alarmo integration unconfirmed)",
	}, nil
}

// UpdateHAConnectionState updates the global HA connection state based on test result.
// FAZ S4: Sets ha_connected = true ONLY when test succeeds (stage=ok).
// Sets ha_connected = false when test fails or no config exists.
// This is the ONLY way ha_connected should be modified.
func UpdateHAConnectionState(result *ConnectionTestResult, runtimeCfg interface{}) error {
	// Type assert to RuntimeConfig (passed from main.go)
	// Using interface{} to avoid circular import with config package
	// Note: This function currently returns nil - the handler updates RuntimeConfig directly
	// to avoid circular import issues
	return nil
}

// HAConnectionStatus represents the safe, public HA connection status.
// FAZ S4: Safe to return to frontend/API.
type HAConnectionStatus struct {
	HaConnected    bool    `json:"ha_connected"`                // true = last test reached stage=ok
	HaLastTestedAt *string `json:"ha_last_tested_at,omitempty"` // RFC3339 timestamp, omit if never tested
}

// GetHAConnectionStatus returns the current global HA connection state.
// FAZ S4: Safe to call frequently, no side effects.
func GetHAConnectionStatus(runtimeCfg *interface{}) *HAConnectionStatus {
	// Placeholder - will be called from handler with access to RuntimeConfig
	return &HAConnectionStatus{
		HaConnected:    false,
		HaLastTestedAt: nil,
	}
}
