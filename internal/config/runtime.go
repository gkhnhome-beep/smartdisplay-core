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
			HABaseURL:       "",
			HAToken:         "",
			UIEnabled:       false,
			HardwareProfile: "minimal",
			WizardCompleted: false,
			BindAddr:        "0.0.0.0",
			Port:            8090,
			TrustProxy:      false,
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
