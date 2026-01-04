# RC Freeze Policy for SmartDisplay v1.0.0-rc1

**Effective Date:** January 4, 2026  
**Freeze Status:** ACTIVE  
**Release Target:** v1.0.0 (stable)

---

## Overview

This document establishes the freezing policy for the Release Candidate phase. No new features or behavioral changes are permitted. Only critical bugfixes are accepted.

---

## RC Naming Convention

All releases follow semantic versioning with RC suffix:

```
v1.0.0-rc1    (first release candidate)
v1.0.0-rc2    (second release candidate, if needed)
v1.0.0        (final stable release)
```

---

## What IS Allowed

✅ **Critical Bugfixes**
- Security vulnerabilities
- Data loss prevention
- Crash/panic fixes
- Backend/frontend desynchronization

✅ **Configuration Changes**
- Tuning timeouts or thresholds
- Adjusting polling intervals
- Fixing incorrect defaults
- Language/i18n corrections

✅ **Metadata Updates**
- Version bumps (rc1 → rc2)
- Build date updates
- Documentation corrections

✅ **Documentation Only**
- README updates
- Deployment guides
- Known issues listing

---

## What IS NOT Allowed

❌ **No New Features**
- Device integrations
- New UI screens
- New subsystems
- Plugin framework expansions

❌ **No Behavioral Changes**
- API endpoint modifications
- State machine additions
- Role/permission additions
- Alarm logic changes

❌ **No Refactoring**
- Package reorganization
- Function signature changes
- Internal API modifications

❌ **No Dependency Updates**
- Go version bumps (unless security)
- Library version upgrades
- Unless critical security patch

---

## Approval Process

### For Bugfixes
1. Create bugfix PR
2. Link to issue (if exists)
3. Include exact reproduction steps
4. Request review from maintainer
5. Merge only after approval

### For RC Versions
1. Bump version.go (rc1 → rc2, etc.)
2. Update RELEASE_FREEZE.md with changes
3. Create tagged release
4. Publish release notes

---

## Transition to Stable

Once all RC issues are resolved:

1. Final verification testing
2. Set version to v1.0.0 (no rc suffix)
3. Create stable release
4. Archive RELEASE_FREEZE.md

---

## RC1 Known Issues

None documented yet. Will be updated as found.

---

## Current RC Status

**RC1 Freeze:** ACTIVE  
**Expected Stable Release:** January 2026  
**Maintainer:** SmartDisplay Team
