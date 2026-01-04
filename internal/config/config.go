package config

import (
	"encoding/json"
	"os"
	"smartdisplay-core/internal/logger"
)

type Config struct {
	AIEnabled    bool `json:"ai_enabled"`
	GuestAccess  bool `json:"guest_access"`
	AlarmControl bool `json:"alarm_control"`
	// Runtime fields (optional)
	HABaseURL       string `json:"ha_base_url,omitempty"`
	HAToken         string `json:"ha_token,omitempty"`
	UIEnabled       bool   `json:"ui_enabled,omitempty"`
	HardwareProfile string `json:"hardware_profile,omitempty"`
	WizardCompleted bool   `json:"wizard_completed,omitempty"`
}

// Load loads the features.json config (legacy, no runtime fields)
func Load() Config {
	data, err := os.ReadFile("configs/features.json")
	if err != nil {
		logger.Error("failed to read config file: " + err.Error())
		os.Exit(1)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		logger.Error("failed to parse config JSON: " + err.Error())
		os.Exit(1)
	}
	return cfg
}

// LoadMerged loads features.json and merges with runtime.json (if present)
func LoadMerged() Config {
	cfg := Load()
	runtimeCfg, err := LoadRuntimeConfig()
	if err == nil && runtimeCfg != nil {
		// Merge runtime fields
		if runtimeCfg.HABaseURL != "" {
			cfg.HABaseURL = runtimeCfg.HABaseURL
		}
		if runtimeCfg.HAToken != "" {
			cfg.HAToken = runtimeCfg.HAToken
		}
		cfg.UIEnabled = runtimeCfg.UIEnabled
		if runtimeCfg.HardwareProfile != "" {
			cfg.HardwareProfile = runtimeCfg.HardwareProfile
		}
		cfg.WizardCompleted = runtimeCfg.WizardCompleted
	}
	return cfg
}
