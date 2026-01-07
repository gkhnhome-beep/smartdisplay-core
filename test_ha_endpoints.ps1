# Test SmartDisplay HA endpoints
$pin = "1234"
$baseUrl = "http://localhost:8090"

$headers = @{
    "X-SmartDisplay-PIN" = $pin
    "Content-Type" = "application/json"
}

Write-Host "Testing SmartDisplay HA Endpoints" -ForegroundColor Cyan
Write-Host "PIN: $pin`n"

# Test 1: Check HA Status
Write-Host "1. GET /api/settings/homeassistant/status" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/settings/homeassistant/status" `
        -Headers $headers -Method GET -ErrorAction Stop
    $body = $response.Content | ConvertFrom-Json
    Write-Host "Status: OK (HTTP $($response.StatusCode))" -ForegroundColor Green
    $body | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "Error: HTTP $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $body = $reader.ReadToEnd()
    $body | Write-Host -ForegroundColor Red
}

Write-Host "`n---`n"

# Test 2: Test HA Connection
Write-Host "2. POST /api/settings/homeassistant/test" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/settings/homeassistant/test" `
        -Headers $headers -Method POST -Body '{}' -ErrorAction Stop
    $body = $response.Content | ConvertFrom-Json
    Write-Host "Status: OK (HTTP $($response.StatusCode))" -ForegroundColor Green
    $body | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "Error: HTTP $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $body = $reader.ReadToEnd()
    $body | Write-Host -ForegroundColor Red
}
