# Build and package smartdisplay-core for release (Windows)
$ErrorActionPreference = 'Stop'

$versionLine = Get-Content internal/version/version.go | Select-String 'Version ='
$version = ($versionLine -split '"')[1]
$dist = "dist"
$bin = "smartdisplay-core.exe"

if (!(Test-Path $dist)) { New-Item -ItemType Directory -Path $dist | Out-Null }

# Build for windows/amd64
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o $bin cmd/smartdisplay/main.go

# Prepare package content
$pkg = "$dist/smartdisplay-core_${version}_windows-amd64.zip"
$pkgdir = ".\_pkg"
if (Test-Path $pkgdir) { Remove-Item -Recurse -Force $pkgdir }
New-Item -ItemType Directory -Path $pkgdir | Out-Null
Copy-Item $bin $pkgdir\
Copy-Item -Recurse web $pkgdir\web
Copy-Item -Recurse deploy $pkgdir\deploy
Copy-Item -Recurse configs $pkgdir\configs
if (Test-Path README.md) { Copy-Item README.md $pkgdir\ }
if (Test-Path .env.example) { Copy-Item .env.example $pkgdir\ }
Compress-Archive -Path $pkgdir\* -DestinationPath $pkg
Remove-Item -Recurse -Force $pkgdir
Remove-Item $bin
Write-Host "Created $pkg"
