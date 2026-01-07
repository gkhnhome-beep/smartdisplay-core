// Package config provides runtime configuration management for smartdisplay-core.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// RuntimeConfig holds persistent runtime configuration.
type RuntimeConfig struct {
	HABaseURL       string `json:"ha_base_url"`
	HAToken         string `json:"ha_token"`
	UIEnabled       bool   `json:"ui_enabled"`
	HardwareProfile string `json:"hardware_profile"` // minimal|standard|full
	WizardCompleted bool   `json:"wizard_completed"`
	BindAddr        string `json:"bind_addr"`
	Port            int    `json:"port"`
	TrustProxy      bool   `json:"trust_proxy"`
	QuietHoursStart string `json:"quiet_hours_start"` // "22:00"
	QuietHoursEnd   string `json:"quiet_hours_end"`   // "06:00"
	Language        string `json:"language"`          // "en"|"tr"

	// Accessibility preferences (FAZ 80)
	HighContrast  bool `json:"high_contrast"`  // High contrast mode
	LargeText     bool `json:"large_text"`     // Large text mode
	ReducedMotion bool `json:"reduced_motion"` // Reduced motion/calmer AI phrasing

	// Voice feedback (FAZ 81)
	VoiceEnabled bool `json:"voice_enabled"` // Voice feedback hooks enabled

	// Home Assistant global state (FAZ S4)
	HaConnected    bool    `json:"ha_connected"`                // true = last test succeeded and reached stage=ok
	HaLastTestedAt *string `json:"ha_last_tested_at,omitempty"` // RFC3339 timestamp of last successful test

	// Home Assistant initial sync (FAZ S5)
	InitialSyncDone bool    `json:"initial_sync_done"`          // true = bootstrap sync completed successfully
	InitialSyncAt   *string `json:"initial_sync_at,omitempty"`  // RFC3339 timestamp of successful sync
	HaVersion       string  `json:"ha_version,omitempty"`       // HA version from initial sync
	HaTimeZone      string  `json:"ha_timezone,omitempty"`      // HA timezone from initial sync
	HaLocationName  string  `json:"ha_location_name,omitempty"` // HA location name from initial sync
	EntityLights    int     `json:"entity_lights,omitempty"`    // Count of light entities
	EntitySensors   int     `json:"entity_sensors,omitempty"`   // Count of sensor entities
	EntitySwitches  int     `json:"entity_switches,omitempty"`  // Count of switch entities
	EntityOthers    int     `json:"entity_others,omitempty"`    // Count of other entities

	// Home Assistant runtime health (FAZ S6)
	HaRuntimeUnreachable  bool    `json:"ha_runtime_unreachable"`            // true = HA became temporarily unreachable after N failures
	HaLastSeenAt          *string `json:"ha_last_seen_at,omitempty"`         // RFC3339 timestamp of last successful HA read
	HaConsecutiveFailures int     `json:"ha_consecutive_failures,omitempty"` // Counter for consecutive read failures (not persisted, runtime only)
}

const RuntimeConfigPath = "data/runtime.json"

// LoadRuntimeConfig loads the runtime config from file, or returns default if not found.
func LoadRuntimeConfig() (*RuntimeConfig, error) {
	f, err := os.Open(RuntimeConfigPath)
	var cfg RuntimeConfig
	if err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		cfg = RuntimeConfig{
			HABaseURL:             "",
			HAToken:               "",
			UIEnabled:             false,
			HardwareProfile:       "minimal",
			WizardCompleted:       false,
			BindAddr:              "0.0.0.0",
			Port:                  8090,
			TrustProxy:            false,
			Language:              "en",
			HighContrast:          false,
			LargeText:             false,
			ReducedMotion:         false,
			VoiceEnabled:          false,
			HaConnected:           false,
			HaLastTestedAt:        nil,
			InitialSyncDone:       false,
			InitialSyncAt:         nil,
			HaRuntimeUnreachable:  false,
			HaLastSeenAt:          nil,
			HaConsecutiveFailures: 0,
		}
	} else {
		return nil, err
	}
	// Env overrides
	if v := os.Getenv("BIND_ADDR"); v != "" {
		cfg.BindAddr = v
	} else if cfg.BindAddr == "" {
		cfg.BindAddr = "0.0.0.0"
	}
	if v := os.Getenv("PORT"); v != "" {
		// ignore error, fallback to default
		var p int
		fmt.Sscanf(v, "%d", &p)
		if p > 0 {
			cfg.Port = p
		}
	} else if cfg.Port == 0 {
		cfg.Port = 8090
	}
	if v := os.Getenv("TRUST_PROXY"); v != "" {
		cfg.TrustProxy = v == "true" || v == "1"
	}
	if v := os.Getenv("LANGUAGE"); v != "" {
		cfg.Language = v
	} else if cfg.Language == "" {
		cfg.Language = "en"
	}
	return &cfg, nil
}

// SaveRuntimeConfig saves the runtime config to file (overwrites).
func SaveRuntimeConfig(cfg *RuntimeConfig) error {
	f, err := os.Create(RuntimeConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}
