#!/usr/bin/env pwsh
# Test disarm with PIN

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "SmartDisplay Alarmo DISARM PIN Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8090"
$adminPin = "1234"
$alarmoPin = "2606"

Write-Host "`n[1] Auth" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest "$baseUrl/api/ui/home/state" -Method GET `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin} `
        -UseBasicParsing
    Write-Host "OK" -ForegroundColor Green
}
catch {
    Write-Host "FAILED" -ForegroundColor Red
    exit 1
}

Write-Host "`n[2] Disarm with PIN '$alarmoPin'" -ForegroundColor Yellow
$payload = @{
    code = $alarmoPin
} | ConvertTo-Json

Write-Host "   Payload: $payload" -ForegroundColor Gray

try {
    $disarmResp = Invoke-WebRequest "$baseUrl/api/ui/alarmo/disarm" -Method POST `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin; "Content-Type"="application/json"} `
        -Body $payload `
        -UseBasicParsing
    Write-Host "OK - Disarm sent" -ForegroundColor Green
    Write-Host "   Response: $($disarmResp.Content)" -ForegroundColor Gray
}
catch {
    Write-Host "FAILED" -ForegroundColor Red
    exit 1
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "Check logs to see HA response" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
