Write-Host "TEST: Alarmo Alarm Kurma Senaryosu" -ForegroundColor Cyan
Write-Host "" 

$baseUrl = "http://localhost:8090"
$pin = "2606"

# STEP 1: Login
Write-Host "STEP 1: Login (PIN: $pin)" -ForegroundColor Yellow
$body = @{pin = $pin} | ConvertTo-Json
$resp = Invoke-WebRequest "$baseUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body -UseBasicParsing
$data = $resp.Content | ConvertFrom-Json
$token = $data.token
Write-Host "LOGIN SUCCESS - Role: $($data.role)" -ForegroundColor Green

# STEP 2: Check current alarm status
Write-Host "`nSTEP 2: Mevcut Alarm Durumunu Kontrol Et" -ForegroundColor Yellow
$resp2 = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET -Headers @{"Authorization"="Bearer $token"} -UseBasicParsing
$status = $resp2.Content | ConvertFrom-Json
Write-Host "Current State: $($status.state)" -ForegroundColor Gray

# STEP 3: Arm Away
Write-Host "`nSTEP 3: Disarida Alarm Kur (arm_away)" -ForegroundColor Yellow
$armBody = @{state = "armed_away"; code = $pin} | ConvertTo-Json
$resp3 = Invoke-WebRequest "$baseUrl/api/ui/alarmo/arm" -Method POST -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} -Body $armBody -UseBasicParsing
$armData = $resp3.Content | ConvertFrom-Json
Write-Host "Arm Request Sent - Result: $($armData.message)" -ForegroundColor Green

# STEP 4: Wait and check
Write-Host "`nSTEP 4: Bekle ve Kontrol Et (2 saniye)" -ForegroundColor Yellow
Start-Sleep -Seconds 2

$resp4 = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET -Headers @{"Authorization"="Bearer $token"} -UseBasicParsing
$status2 = $resp4.Content | ConvertFrom-Json
Write-Host "Updated State: $($status2.state)" -ForegroundColor Gray

# STEP 5: Result
Write-Host "`nSTEP 5: SONUC" -ForegroundColor Yellow
if ($status2.state -like "*armed*") {
    Write-Host "SUCCESS: Alarm basariyla kuruldu!" -ForegroundColor Green
    Write-Host "Onceki: $($status.state) -> Yeni: $($status2.state)" -ForegroundColor Green
} else {
    Write-Host "RESULT: $($status2.state)" -ForegroundColor Yellow
}

Write-Host "`nTest Tamamlandi." -ForegroundColor Cyan
