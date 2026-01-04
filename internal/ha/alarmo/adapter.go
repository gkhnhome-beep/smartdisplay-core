package alarmo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AlarmoState represents normalized alarm state from Home Assistant
// This is the single source of truth for alarm state within SmartDisplay
type AlarmoState struct {
	RawState      string    // Raw state string from HA (e.g., "armed_home")
	Mode          string    // disarmed | arming | armed | triggered
	ArmedMode     string    // home | away | night | "" (empty if not armed)
	EntryDelaySec int       // Entry delay in seconds (0 if not applicable)
	ExitDelaySec  int       // Exit delay in seconds (0 if not applicable)
	Triggered     bool      // True if alarm is triggered
	LastChanged   time.Time // When state last changed
}

// Adapter reads alarm state from Home Assistant Alarmo integration
// Read-only: no arm/disarm operations
type Adapter struct {
	baseURL string
	token   string
	client  *http.Client
}

// haStateResponse represents the JSON response from HA /api/states endpoint
type haStateResponse struct {
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
}

// New creates a new Alarmo adapter
// baseURL: Home Assistant REST API base URL (e.g., "http://localhost:8123")
// decryptedToken: Bearer token (already decrypted from secure storage)
func New(baseURL string, decryptedToken string) *Adapter {
	return &Adapter{
		baseURL: baseURL,
		token:   decryptedToken,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// FetchState retrieves the current alarm state from Home Assistant
// Returns normalized AlarmoState and error if fetch fails
func (a *Adapter) FetchState(ctx context.Context) (AlarmoState, error) {
	// Construct request (do not log URL)
	url := a.baseURL + "/api/states/alarm_control_panel.alarmo"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AlarmoState{}, fmt.Errorf("alarmo: request creation failed: %w", err)
	}

	// Set authorization header (do not log token)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return AlarmoState{}, fmt.Errorf("alarmo: fetch failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP status codes
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return AlarmoState{}, errors.New("alarmo: unauthorized (check HA token)")
	}

	if resp.StatusCode != http.StatusOK {
		return AlarmoState{}, fmt.Errorf("alarmo: http %d", resp.StatusCode)
	}

	// Parse JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AlarmoState{}, fmt.Errorf("alarmo: read body failed: %w", err)
	}

	var haResp haStateResponse
	if err := json.Unmarshal(body, &haResp); err != nil {
		return AlarmoState{}, fmt.Errorf("alarmo: parse failed: %w", err)
	}

	// Map to normalized state
	state := mapAlarmoState(haResp)
	return state, nil
}

// mapAlarmoState converts HA response to normalized AlarmoState
// Mapping is hardcoded per specification:
//
//	HA state          -> Mode       + ArmedMode
//	"disarmed"        -> disarmed
//	"arming"/"pending"-> arming
//	"armed_home"      -> armed      + home
//	"armed_away"      -> armed      + away
//	"armed_night"     -> armed      + night
//	"triggered"       -> triggered  + (triggered=true)
func mapAlarmoState(ha haStateResponse) AlarmoState {
	state := AlarmoState{
		RawState:    ha.State,
		Triggered:   ha.State == "triggered",
		LastChanged: parseLastChanged(ha.LastChanged),
	}

	// Map HA state to normalized mode
	switch ha.State {
	case "disarmed":
		state.Mode = "disarmed"
		state.ArmedMode = ""

	case "arming", "pending":
		state.Mode = "arming"
		state.ArmedMode = ""

	case "armed_home":
		state.Mode = "armed"
		state.ArmedMode = "home"

	case "armed_away":
		state.Mode = "armed"
		state.ArmedMode = "away"

	case "armed_night":
		state.Mode = "armed"
		state.ArmedMode = "night"

	case "triggered":
		state.Mode = "triggered"
		state.ArmedMode = ""
		state.Triggered = true

	default:
		// Unknown state - map as-is
		state.Mode = ha.State
		state.ArmedMode = ""
	}

	// Parse entry delay (optional)
	if entryDelay, ok := ha.Attributes["entry_delay"]; ok {
		if delay, ok := entryDelay.(float64); ok {
			state.EntryDelaySec = int(delay)
		} else if delayStr, ok := entryDelay.(string); ok {
			// Try parsing as string
			fmt.Sscanf(delayStr, "%d", &state.EntryDelaySec)
		}
	}

	// Parse exit delay (optional)
	if exitDelay, ok := ha.Attributes["exit_delay"]; ok {
		if delay, ok := exitDelay.(float64); ok {
			state.ExitDelaySec = int(delay)
		} else if delayStr, ok := exitDelay.(string); ok {
			// Try parsing as string
			fmt.Sscanf(delayStr, "%d", &state.ExitDelaySec)
		}
	}

	return state
}

// parseLastChanged parses HA timestamp string to time.Time
// HA format: "2026-01-04T12:34:56.789012+00:00"
func parseLastChanged(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now()
	}

	// Try RFC3339 with nanoseconds
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		return t
	}

	// Fallback to RFC3339
	t, err = time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t
	}

	// Fallback to now
	return time.Now()
}
