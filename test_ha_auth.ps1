# Test HA authentication
# This script decrypts the token and tests HA connectivity

$haConfigPath = "data\ha_config.json"
$config = Get-Content $haConfigPath | ConvertFrom-Json

Write-Host "HA Server URL: $($config.server_url)"
Write-Host "Encrypted Token Length: $($config.encrypted_token.Length)"

# Now we need to decrypt the token using the same method as Go backend
# Since we can't easily decrypt from PowerShell, let's create a simple Go test program

$testProgram = @"
package main

import (
	"fmt"
	"os"
	"smartdisplay-core/internal/settings"
)

func main() {
	// Decrypt token
	token, err := settings.DecryptToken()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Token (first 20 chars): %.20s\n", token)
	fmt.Printf("Token length: %d\n", len(token))
	
	// Decrypt server URL
	serverURL, err := settings.DecryptServerURL()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Server URL: %s\n", serverURL)
}
"@

# Write test program
Write-Host "`nCreating test program..."
$testProgram | Out-File -FilePath "cmd\test_ha_decrypt\main.go" -Encoding UTF8 -Force

Write-Host "Run: go run cmd/test_ha_decrypt/main.go"
