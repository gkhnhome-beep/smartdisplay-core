# SmartDisplay Core Build Guide

## Overview

This document describes the deterministic build process for SmartDisplay Core v1.0.0-rc1. The build system creates cross-platform binaries with version metadata embedded via `-ldflags`.

## Prerequisites

- **Go:** 1.19+ (check with `go version`)
- **Git:** For commit hash extraction
- **Make or bash:** For running build scripts

## Build Commands

### Automated Build (Recommended)

```bash
# Make build.sh executable (Linux/macOS)
chmod +x build.sh

# Run automated build
./build.sh
```

This builds both targets and outputs to `dist/`:
- `dist/smartdisplay-core-linux-amd64` - x64 server/desktop
- `dist/smartdisplay-core-linux-arm32v7` - ARMv7 (Raspberry Pi)

### Manual Build

#### linux/amd64 (Server)
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build \
    -ldflags="-X smartdisplay-core/internal/version.Version=1.0.0-rc1 \
              -X smartdisplay-core/internal/version.Commit=$(git rev-parse --short HEAD) \
              -X smartdisplay-core/internal/version.BuildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o dist/smartdisplay-core-linux-amd64 \
    cmd/smartdisplay
```

#### linux/arm (Raspberry Pi ARMv7)
```bash
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build \
    -ldflags="-X smartdisplay-core/internal/version.Version=1.0.0-rc1 \
              -X smartdisplay-core/internal/version.Commit=$(git rev-parse --short HEAD) \
              -X smartdisplay-core/internal/version.BuildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o dist/smartdisplay-core-linux-arm32v7 \
    cmd/smartdisplay
```

## Build Flags Explanation

| Flag | Value | Purpose |
|------|-------|---------|
| `-ldflags` | Version metadata | Embed version info at compile time |
| `GOOS` | linux | Target operating system |
| `GOARCH` | amd64 / arm | Target architecture |
| `GOARM` | 7 | ARM version (v7 = ARMv7, ~ARMv6+) |
| `CGO_ENABLED` | 0 | Disable C dependencies for portability |

## Version Metadata

Each build embeds:
- **Version:** `v1.0.0-rc1`
- **Commit:** Git short hash (7 chars)
- **BuildDate:** ISO 8601 UTC timestamp

Access via `/health` endpoint:
```bash
curl http://localhost:8080/health | jq .version
```

Response:
```json
{
  "version": "1.0.0-rc1",
  "commit": "a1b2c3d",
  "build_date": "2024-01-15T14:32:18Z"
}
```

## Output Verification

After building, verify the binary:
```bash
# Check file exists and size
ls -lh dist/smartdisplay-core-*

# Verify it's executable
file dist/smartdisplay-core-linux-amd64

# Test run (requires config files)
./dist/smartdisplay-core-linux-amd64 --help
```

## Troubleshooting

### Build fails with "command not found: git"
- Install Git or set `COMMIT="unknown"` manually in build.sh

### Build fails with architecture mismatch
- Verify `go version` shows 1.19+
- Check `uname -m` matches target architecture

### Binary doesn't run on target
- Verify target OS: `uname -s`
- Verify architecture: `uname -m` (amd64 = x86_64, arm = armv7l)
- Check libc compatibility (using CGO_ENABLED=0 avoids this)

## Release Checklist

Before release:
- [ ] Run `./build.sh` successfully
- [ ] Verify both binaries exist in `dist/`
- [ ] Test `file` command output shows correct architecture
- [ ] Update CHANGELOG.md with commit hash
- [ ] Tag release: `git tag v1.0.0-rc1`
- [ ] Create GitHub release with binaries

## See Also

- [RELEASE_FREEZE.md](RELEASE_FREEZE.md) - RC freeze policy
- [README.md](README.md) - Project overview
- [internal/version/version.go](internal/version/version.go) - Version package
