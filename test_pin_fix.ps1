#!/usr/bin/env pwsh
# Test script for PIN code fix

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "SmartDisplay Alarmo PIN Code Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8090"
$adminPin = "1234"
$alarmoPin = "2606"

Write-Host "`n[1] Test authentication" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest "$baseUrl/api/ui/home/state" -Method GET `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin} `
        -UseBasicParsing
    Write-Host "OK - Auth passed" -ForegroundColor Green
}
catch {
    Write-Host "ERROR - Auth failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n[2] Check current alarm state" -ForegroundColor Yellow
try {
    $statusResp = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin} `
        -UseBasicParsing
    $status = $statusResp.Content | ConvertFrom-Json
    Write-Host "OK - Current state: $($status.state)" -ForegroundColor Green
}
catch {
    Write-Host "ERROR - Get status failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3] Arm with mode 'armed_away' and PIN '$alarmoPin'" -ForegroundColor Yellow
$payload = @{
    mode = "armed_away"
    code = $alarmoPin
} | ConvertTo-Json

Write-Host "   Payload: $payload" -ForegroundColor Gray

try {
    $armResp = Invoke-WebRequest "$baseUrl/api/ui/alarmo/arm" -Method POST `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin; "Content-Type"="application/json"} `
        -Body $payload `
        -UseBasicParsing
    Write-Host "OK - Arm command accepted" -ForegroundColor Green
}
catch {
    Write-Host "ERROR - Arm command failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n[4] Wait 2 seconds for state update" -ForegroundColor Yellow
Start-Sleep -Seconds 2

Write-Host "`n[5] Check if alarm was armed" -ForegroundColor Yellow
try {
    $statusResp2 = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET `
        -Headers @{"X-SmartDisplay-PIN"=$adminPin} `
        -UseBasicParsing
    $status2 = $statusResp2.Content | ConvertFrom-Json
    Write-Host "OK - Updated state: $($status2.state)" -ForegroundColor Green
    
    if (($status2.state -like "*armed*") -or ($status2.state -eq "armed_away")) {
        Write-Host "" -ForegroundColor Green
        Write-Host "SUCCESS! Alarm was armed with PIN code!" -ForegroundColor Green
        Write-Host "  Before: $($status.state)" -ForegroundColor Green
        Write-Host "  After:  $($status2.state)" -ForegroundColor Green
    }
    else {
        Write-Host "" -ForegroundColor Red
        Write-Host "FAILED - Alarm state unchanged" -ForegroundColor Red
        Write-Host "  Current: $($status2.state)" -ForegroundColor Red
    }
}
catch {
    Write-Host "ERROR - Get updated status failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "Test completed" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
