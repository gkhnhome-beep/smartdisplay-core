Write-Host "============================================" -ForegroundColor Cyan
Write-Host "ALARMO ALARM KURMA TESTI" -ForegroundColor Cyan
Write-Host "SmartDisplay PIN: 1234" -ForegroundColor Yellow
Write-Host "Alarmo PIN: 2606" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8090"

# STEP 1: SmartDisplay giriş (PIN: 1234)
Write-Host "`n[1] SmartDisplay girisini yap (PIN: 1234)" -ForegroundColor Yellow
$testResp = Invoke-WebRequest "$baseUrl/api/ui/home/state" -Method GET `
    -Headers @{"X-SmartDisplay-PIN"="1234"; "Accept"="application/json"} `
    -UseBasicParsing 2>&1
    
if ($testResp.StatusCode -eq 200) {
    Write-Host "OK - Login basarili" -ForegroundColor Green
} else {
    Write-Host "ERROR - Login basarisiz: $($testResp.StatusCode)" -ForegroundColor Red
    exit 1
}

# STEP 2: Mevcut Alarmo durumunu kontrol et
Write-Host "`n[2] Mevcut Alarmo durumunu kontrol et" -ForegroundColor Yellow
$statusResp = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET `
    -Headers @{"X-SmartDisplay-PIN"="1234"; "Accept"="application/json"} `
    -UseBasicParsing
$statusData = $statusResp.Content | ConvertFrom-Json

Write-Host "OK - Alarmo baglantiyor: $($statusData.alarmo_connected)" -ForegroundColor Green

# STEP 3: Disarida Alarm Kur (armed_away) - PIN: 2606
Write-Host "`n[3] Disarida alarm kur (armed_away) - PIN: 2606" -ForegroundColor Yellow
$armBody = @{
    mode = "armed_away"
    code = "2606"
} | ConvertTo-Json

$armResp = Invoke-WebRequest "$baseUrl/api/ui/alarmo/arm" -Method POST `
    -Headers @{"X-SmartDisplay-PIN"="1234"; "Content-Type"="application/json"} `
    -Body $armBody -UseBasicParsing 2>&1

if ($armResp.StatusCode -eq 200) {
    Write-Host "OK - Arm komutu gonderildi" -ForegroundColor Green
    $armData = $armResp.Content | ConvertFrom-Json
    Write-Host "Response: $($armResp.Content)" -ForegroundColor Gray
} else {
    Write-Host "ERROR - Arm komutu basarisiz: $($armResp.StatusCode)" -ForegroundColor Red
    Write-Host "Response: $($armResp.Content)" -ForegroundColor Red
}

# STEP 4: Sonuc kontrol et
Write-Host "`n[4] 2 saniye bekle ve kontrol et" -ForegroundColor Yellow
Start-Sleep -Seconds 2

$statusResp2 = Invoke-WebRequest "$baseUrl/api/ui/alarmo/status" -Method GET `
    -Headers @{"X-SmartDisplay-PIN"="1234"; "Accept"="application/json"} `
    -UseBasicParsing
$statusData2 = $statusResp2.Content | ConvertFrom-Json

Write-Host "OK - Alarmo baglantisi kontrol: $($statusData2.alarmo_connected)" -ForegroundColor Green

# Arm komutu donusu kontrol et
Write-Host "`n[5] Arm komutu sonucunu kontrol et" -ForegroundColor Yellow
if ($armResp -and $armResp.StatusCode -eq 200) {
    $armRespData = $armResp.Content | ConvertFrom-Json
    if ($armRespData.response.ok -eq $true) {
        Write-Host "OK - Arm komutu basarili" -ForegroundColor Green
        Write-Host "  Mode: $($armRespData.response.data.mode)" -ForegroundColor Green
        Write-Host "  Status: $($armRespData.response.data.status)" -ForegroundColor Green
        
        if ($armRespData.response.data.status -like "*armed*") {
            Write-Host "`nSUCCESS - TEST PASSED!" -ForegroundColor Green
            Write-Host "Alarmo basarili sekilde kuruldu!" -ForegroundColor Green
        } else {
            Write-Host "`nFAILED - Alarm status yanlis" -ForegroundColor Red
        }
    } else {
        Write-Host "ERROR - Arm komutu hata dondurdu" -ForegroundColor Red
    }
} else {
    Write-Host "ERROR - Arm request basarisiz" -ForegroundColor Red
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "TEST TAMAMLANDI" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
