@echo off
title SmartDisplay - Backend + UI

echo === SmartDisplay starting ===

REM Backend
echo Starting backend...

start "SmartDisplay Backend" cmd /k ^
cd /d C:\SmartDisplayV3 ^& ^
go run ./cmd/smartdisplay

REM UI
echo Starting UI...

start "SmartDisplay UI" cmd /k ^
cd /d C:\SmartDisplayV3\web ^& ^
python -m http.server 5500

echo === All services started ===
exit
