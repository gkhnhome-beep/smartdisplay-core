// Package settings provides secure credential handling for Home Assistant integration.
// hasync.go: Initial HA synchronization engine for FAZ S5
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

// HaMetadata represents safe, read-only HA system metadata from initial sync.
// FAZ S5: Contains only safe metadata, never credentials or secrets.
type HaMetadata struct {
	Version      string `json:"version,omitempty"`       // e.g., "2024.1.0"
	TimeZone     string `json:"time_zone,omitempty"`     // e.g., "America/New_York"
	LocationName string `json:"location_name,omitempty"` // e.g., "Home"
}

// EntityCounts represents aggregated entity counts by domain.
// FAZ S5: Counts only, no entity IDs or names.
type EntityCounts struct {
	Lights   int `json:"lights,omitempty"`
	Sensors  int `json:"sensors,omitempty"`
	Switches int `json:"switches,omitempty"`
	Others   int `json:"others,omitempty"`
}

// InitialSyncResult represents the outcome of initial HA synchronization.
// FAZ S5: One-time bootstrap sync after successful connection.
type InitialSyncResult struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Meta    *HaMetadata   `json:"meta,omitempty"`
	Counts  *EntityCounts `json:"counts,omitempty"`
}

// PerformInitialSync performs a one-time read-only bootstrap synchronization with HA.
// FAZ S5: Executes in strict order, stops on first failure.
//
// Steps:
// 1. GET /api/config - extract metadata (version, timezone, location)
// 2. GET /api/states/alarm_control_panel.alarmo - confirm Alarmo presence
// 3. GET /api/states - count entities by domain
//
// Returns aggregated result. Does NOT persist automatically - caller must save.
func PerformInitialSync() (*InitialSyncResult, error) {
	// Load encrypted HA config
	cfg, err := LoadHAConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load HA config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("HA not configured")
	}

	// Decrypt credentials
	serverURL, err := DecryptServerURL()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt server URL: %w", err)
	}
	if serverURL == "" {
		return nil, fmt.Errorf("server URL not configured")
	}

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

	// STEP 1: Fetch HA core config metadata
	meta, err := fetchHAConfig(client, serverURL, token)
	if err != nil {
		return &InitialSyncResult{
			Success: false,
			Message: "Failed to fetch HA configuration: " + err.Error(),
		}, nil
	}

	// STEP 2: Confirm Alarmo presence
	alarmoPresent, err := confirmAlarmoPresence(client, serverURL, token)
	if err != nil {
		return &InitialSyncResult{
			Success: false,
			Message: "Failed to confirm Alarmo integration: " + err.Error(),
		}, nil
	}

	if !alarmoPresent {
		return &InitialSyncResult{
			Success: false,
			Message: "Alarmo integration not found in Home Assistant",
		}, nil
	}

	// STEP 3: Count entities by domain
	counts, err := countEntities(client, serverURL, token)
	if err != nil {
		// Non-fatal: we have config and Alarmo is present
		// Return success with what we have
		logger.Error("failed to count entities: " + err.Error())
		counts = &EntityCounts{Others: 0}
	}

	return &InitialSyncResult{
		Success: true,
		Message: "Initial HA synchronization completed successfully",
		Meta:    meta,
		Counts:  counts,
	}, nil
}

// fetchHAConfig retrieves safe HA system metadata.
// FAZ S5: Read-only, extracts only safe fields.
func fetchHAConfig(client *http.Client, serverURL, token string) (*HaMetadata, error) {
	configURL := strings.TrimSuffix(serverURL, "/") + "/api/config"

	req, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(body, &configData); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	// Extract safe metadata
	meta := &HaMetadata{}

	if version, ok := configData["version"].(string); ok {
		meta.Version = version
	}

	if tz, ok := configData["time_zone"].(string); ok {
		meta.TimeZone = tz
	}

	if loc, ok := configData["location_name"].(string); ok {
		meta.LocationName = loc
	}

	return meta, nil
}

// confirmAlarmoPresence verifies that Alarmo integration is reachable.
// FAZ S5: Read-only check, does not store alarm state.
func confirmAlarmoPresence(client *http.Client, serverURL, token string) (bool, error) {
	alarmoURL := strings.TrimSuffix(serverURL, "/") + "/api/states/alarm_control_panel.alarmo"

	req, err := http.NewRequest("GET", alarmoURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	// 200 = found, 404 = not found
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

// countEntities counts entities by domain from HA states.
// FAZ S5: Aggregation only, no entity IDs or names stored.
func countEntities(client *http.Client, serverURL, token string) (*EntityCounts, error) {
	statesURL := strings.TrimSuffix(serverURL, "/") + "/api/states"

	req, err := http.NewRequest("GET", statesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var states []map[string]interface{}
	if err := json.Unmarshal(body, &states); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	counts := &EntityCounts{}

	// Count by domain (entity_id format: "domain.entity_name")
	for _, state := range states {
		entityID, ok := state["entity_id"].(string)
		if !ok {
			continue
		}

		domain := strings.Split(entityID, ".")[0]

		switch domain {
		case "light":
			counts.Lights++
		case "sensor":
			counts.Sensors++
		case "switch":
			counts.Switches++
		default:
			counts.Others++
		}
	}

	return counts, nil
}
