#!/usr/bin/env pwsh
# Direct HA Alarmo PIN test

# HA config dosyasından bilgileri alalım
$configFile = "e:\SmartDisplayV3\data\ha_config.json"

Write-Host "======================================"  -ForegroundColor Cyan
Write-Host "Direct HA Alarmo PIN Test" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

if (-not (Test-Path $configFile)) {
    Write-Host "`nERROR: HA config file not found" -ForegroundColor Red
    Write-Host "File: $configFile" -ForegroundColor Red
    exit 1
}

try {
    $config = Get-Content $configFile | ConvertFrom-Json
    $haUrl = $config.base_url
    $haToken = $config.token
    
    Write-Host "`nHA Configuration:" -ForegroundColor Yellow
    Write-Host "  URL: $haUrl" -ForegroundColor Gray
    Write-Host "  Token: Present (length: $($haToken.Length))" -ForegroundColor Gray
}
catch {
    Write-Host "`nERROR: Failed to parse HA config" -ForegroundColor Red
    exit 1
}

# Test disarm with PIN
Write-Host "`n[TEST] Disarm Alarmo with PIN 2606" -ForegroundColor Yellow

$headers = @{
    "Authorization" = "Bearer $haToken"
    "Content-Type" = "application/json"
}

$body = @{
    entity_id = "alarm_control_panel.alarmo"
    code = "2606"
} | ConvertTo-Json

Write-Host "  URL: $haUrl/api/services/alarm_control_panel/alarm_disarm" -ForegroundColor Gray
Write-Host "  Body: $body" -ForegroundColor Gray

try {
    $response = Invoke-WebRequest `
        -Uri "$haUrl/api/services/alarm_control_panel/alarm_disarm" `
        -Method POST `
        -Headers $headers `
        -Body $body `
        -UseBasicParsing
    
    Write-Host "`nResponse Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response Body:" -ForegroundColor Green
    Write-Host $response.Content -ForegroundColor Gray
}
catch {
    Write-Host "`nError occurred:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    if ($_.Exception.Response) {
        $streamReader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $streamReader.ReadToEnd()
        Write-Host "Error Body: $errorBody" -ForegroundColor Red
    }
}

Write-Host "`n======================================" -ForegroundColor Cyan
Write-Host "Check if Alarmo state changed in HA" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
