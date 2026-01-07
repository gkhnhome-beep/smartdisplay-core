package main

import (
	"fmt"
	"os"
	"smartdisplay-core/internal/settings"
)

func main() {
	fmt.Println("=== SmartDisplay HA Token Test ===\n")

	// Test 1: Load HA Config
	fmt.Println("1. Loading HA config...")
	cfg, err := settings.LoadHAConfig()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	if cfg == nil {
		fmt.Println("ERROR: HA config not found")
		os.Exit(1)
	}

	fmt.Printf("   Server URL: %s\n", cfg.ServerURL)
	fmt.Printf("   Encrypted token length: %d\n", len(cfg.EncryptedToken))
	fmt.Printf("   Configured at: %s\n\n", cfg.ConfiguredAt)

	// Test 2: Decrypt Server URL
	fmt.Println("2. Decrypting server URL...")
	serverURL, err := settings.DecryptServerURL()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Server URL: %s\n\n", serverURL)

	// Test 3: Decrypt Token
	fmt.Println("3. Decrypting token...")
	token, err := settings.DecryptToken()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Token (first 30 chars): %.30s...\n", token)
	fmt.Printf("   Token length: %d\n", len(token))
	fmt.Printf("   Token is valid: %v\n\n", len(token) > 0)

	// Test 4: Check IsConfigured
	fmt.Println("4. Checking if HA is configured...")
	isConfigured, err := settings.IsConfigured()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Is configured: %v\n\n", isConfigured)

	// Test 5: Get HA Status
	fmt.Println("5. Getting HA status...")
	status, err := settings.GetHAStatus()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Is configured: %v\n", status.IsConfigured)
	fmt.Printf("   Configured at: %v\n\n", status.ConfiguredAt)

	fmt.Println("✓ All tests passed!")
}
