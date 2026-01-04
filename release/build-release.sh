#!/bin/bash
set -e
# Build and package smartdisplay-core for release (Linux)

VERSION=$(grep 'Version =' internal/version/version.go | awk -F '"' '{print $2}')
DIST=dist
BIN=smartdisplay-core

mkdir -p $DIST

# Build for linux/arm (Raspberry Pi 3)
GOOS=linux GOARCH=arm GOARM=7 go build -o $BIN-linux-arm cmd/smartdisplay/main.go

# Optionally build for linux/arm64
GOOS=linux GOARCH=arm64 go build -o $BIN-linux-arm64 cmd/smartdisplay/main.go

# Prepare package content
for ARCH in linux-arm linux-arm64; do
  PKG=$DIST/${BIN}_${VERSION}_${ARCH}.tar.gz
  mkdir -p _pkg
  cp $BIN-$ARCH _pkg/$BIN
  cp -r web _pkg/
  cp -r deploy _pkg/
  cp -r configs _pkg/
  cp README.md _pkg/ || true
  cp .env.example _pkg/ || true
  tar -czf $PKG -C _pkg .
  rm -rf _pkg $BIN-$ARCH
  echo "Created $PKG"
done
