#!/bin/bash
# SmartDisplay Core - Build Script
# Builds deterministic cross-platform binaries

set -e

# Configuration
PROJECT="smartdisplay-core"
VERSION="1.0.0-rc1"
DIST_DIR="dist"
CMD_PATH="cmd/smartdisplay"
OUTPUT_NAME="smartdisplay-core"

# Get git info for embedding
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Ensure dist directory exists
mkdir -p "$DIST_DIR"

# Build flags
LD_FLAGS="-X smartdisplay-core/internal/version.Version=$VERSION"
LD_FLAGS="$LD_FLAGS -X smartdisplay-core/internal/version.Commit=$COMMIT"
LD_FLAGS="$LD_FLAGS -X smartdisplay-core/internal/version.BuildDate=$BUILD_DATE"

echo "[Build] SmartDisplay Core $VERSION"
echo "[Build] Commit: $COMMIT"
echo "[Build] Date: $BUILD_DATE"
echo ""

# Supported targets
TARGETS=(
  "linux:amd64:smartdisplay-core-linux-amd64"
  "linux:arm:smartdisplay-core-linux-arm32v7"
)

# Build for each target
for target in "${TARGETS[@]}"; do
  IFS=':' read -r OS ARCH OUTPUT <<<"$target"
  
  echo "[Build] Building $OS/$ARCH -> $OUTPUT"
  
  GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 \
    go build \
      -ldflags "$LD_FLAGS" \
      -o "$DIST_DIR/$OUTPUT" \
      "$CMD_PATH"
  
  if [ $? -eq 0 ]; then
    SIZE=$(stat -f%z "$DIST_DIR/$OUTPUT" 2>/dev/null || stat -c%s "$DIST_DIR/$OUTPUT" 2>/dev/null)
    echo "  ✓ Success ($(numfmt --to=iec-i --suffix=B $SIZE 2>/dev/null || echo $SIZE bytes))"
  else
    echo "  ✗ Failed"
    exit 1
  fi
done

echo ""
echo "[Build] Complete"
echo "[Build] Artifacts:"
ls -lh "$DIST_DIR/"
