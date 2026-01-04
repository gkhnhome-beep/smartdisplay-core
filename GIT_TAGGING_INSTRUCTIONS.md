# Git Tagging Instructions for v1.0.0-rc1

**Perform these steps only after RELEASE_CHECKLIST.md is fully verified.**

---

## Prerequisites

- [ ] All items in `RELEASE_CHECKLIST.md` are checked
- [ ] `git status` shows clean working tree
- [ ] You are on the main/master branch
- [ ] You have git configured with name/email

---

## Step 1: Verify Repository State

```bash
# Check for uncommitted changes
git status

# Expected output:
# On branch main
# nothing to commit, working tree clean

# If there are changes, commit them first:
git add .
git commit -m "docs: prepare RC1 release"
```

---

## Step 2: Verify No Existing Tag

```bash
# Check if v1.0.0-rc1 already exists
git tag -l | grep v1.0.0-rc1

# Expected: no output (tag doesn't exist yet)

# If tag exists, check its commit:
git rev-list -n 1 v1.0.0-rc1

# If it matches HEAD, tag already created (skip to Step 5)
# If it doesn't match, you need to delete it first:
# git tag -d v1.0.0-rc1
# git push origin :refs/tags/v1.0.0-rc1
```

---

## Step 3: Create Annotated Tag

Create the release tag with comprehensive message:

```bash
git tag -a v1.0.0-rc1 \
  -m "Release Candidate 1

SmartDisplay Core v1.0.0-rc1

## Features Included

- Kiosk-safe vanilla JS/HTML/CSS frontend with 6 views
- Systemd service integration (backend + UI)
- Cross-platform builds (linux/amd64 + linux/arm/v7)
- Version metadata embedded via ldflags
- Complete accessibility support (WCAG 2.1 AA)
- Installation, build, and upgrade guides
- Backup/rollback procedures with failure handling

## RC Policy

This is a Release Candidate - bugfix-only release. See RELEASE_FREEZE.md.

## Known Limitations

- Docker support deferred
- Plugin framework deferred
- Advanced diagnostics deferred
- No automated UI tests yet

## Testing Required

- All items in RELEASE_CHECKLIST.md verified
- Services start/stop/restart working
- Health endpoint responding with version
- Both systemd units functional
- Upgrade/rollback tested successfully

## Documentation

See the following for complete information:
- RELEASE_NOTES.md - Scope, features, limitations
- RELEASE_CHECKLIST.md - Verification items
- INSTALL.md - Installation steps
- BUILD.md - Build instructions
- UPGRADE.md - Upgrade/rollback procedures
- RELEASE_FREEZE.md - RC policy

## Next Steps

1. Tag pushed to origin
2. Create GitHub Release from this tag
3. Attach binary artifacts
4. Link to RELEASE_NOTES.md
5. Announce RC1 for community testing

## Community Testing Period

RC1 is open for community feedback. Please report issues:
- Version: curl localhost:8080/health | jq .version
- OS: uname -a
- Exact reproduction steps
- Relevant logs: journalctl -u smartdisplay-core.service

Timeline: Jan 4 - Jan 31, 2026"
```

---

## Step 4: Verify Tag Was Created

```bash
# List the tag
git tag -l v1.0.0-rc1

# Show tag details
git show v1.0.0-rc1

# Expected: Tag message, commit hash, commit message

# Verify it points to current HEAD
git rev-list -n 1 v1.0.0-rc1

# Should match:
git rev-parse HEAD
```

---

## Step 5: Push Tag to Remote

Push the tag to GitHub/GitLab:

```bash
# Push just the tag
git push origin v1.0.0-rc1

# Or push all tags
git push origin --tags

# Verify it was pushed
git ls-remote --tags origin | grep v1.0.0-rc1
```

---

## Step 6: Create GitHub Release (Manual)

After tag is pushed:

1. Go to **GitHub** → **Releases** → **Draft a new release**

2. Select tag: **v1.0.0-rc1**

3. Release title: **SmartDisplay Core v1.0.0-rc1**

4. Release description (copy from RELEASE_NOTES.md):
   ```
   SmartDisplay Core v1.0.0-rc1 is the first Release Candidate.
   
   **Features:**
   - Kiosk-safe frontend with 6 views
   - Systemd integration
   - Cross-platform binaries
   - Complete documentation
   
   **See [RELEASE_NOTES.md](RELEASE_NOTES.md) for full details.**
   ```

5. **Attach Binaries:**
   ```
   dist/smartdisplay-core-linux-amd64
   dist/smartdisplay-core-linux-arm32v7
   ```

6. **Mark as pre-release:** ✅ Check "This is a pre-release"

7. **Publish release:** Click "Publish release"

---

## Step 7: Verify Release

```bash
# Check tag exists on remote
git ls-remote --tags origin v1.0.0-rc1

# Check GitHub Release page
# https://github.com/[owner]/smartdisplay-core/releases/tag/v1.0.0-rc1

# Expected: Release visible, binaries attached
```

---

## Troubleshooting

### Tag Already Exists

```bash
# If tag exists at different commit
git tag -d v1.0.0-rc1                    # Delete local
git push origin :refs/tags/v1.0.0-rc1   # Delete remote

# Then create new tag from Step 3
```

### Push Rejected (Permission Denied)

```bash
# Check git credentials
git config --list | grep credential

# Update credentials in Git config or SSH
# For HTTPS: Store token in ~/.git-credentials
# For SSH: Check ~/.ssh/id_rsa permissions

# Retry push
git push origin v1.0.0-rc1
```

### Tag Pushed But GitHub Release Missing

```bash
# Go to GitHub Releases page
# Click "Draft a new release"
# Select v1.0.0-rc1 from dropdown
# Fill in details manually
```

### Need to Update Tag Message

```bash
# Delete tag
git tag -d v1.0.0-rc1
git push origin :refs/tags/v1.0.0-rc1

# Recreate with correct message (Step 3)
# Push again (Step 5)
```

---

## Verification Checklist

After tagging and release:

- [ ] Tag created locally: `git tag -l | grep v1.0.0-rc1`
- [ ] Tag pushed to remote: `git ls-remote --tags origin | grep v1.0.0-rc1`
- [ ] GitHub Release published and visible
- [ ] Pre-release flag checked on GitHub
- [ ] Binaries attached to release
- [ ] RELEASE_NOTES.md linked in release description
- [ ] Release announce sent to community

---

## Next Steps After Release

1. **Announce RC1:**
   - Post on GitHub Discussions
   - Update project website/docs
   - Notify maintainers/stakeholders

2. **Start RC Testing Period:**
   - Collect community feedback (Jan 4 - Jan 31)
   - Log issues on GitHub
   - Only critical bugfixes (see RELEASE_FREEZE.md)

3. **Monitor for Issues:**
   - Watch GitHub Issues
   - Check logs from users
   - Verify installation on multiple platforms

4. **Prepare for rc2 or Stable:**
   - If issues found → Create v1.0.0-rc2 tag
   - If no issues → Proceed to v1.0.0 stable (Feb 28 target)

---

## Rollback Tag (If Needed)

If RC1 needs to be rolled back:

```bash
# Delete local tag
git tag -d v1.0.0-rc1

# Delete remote tag
git push origin :refs/tags/v1.0.0-rc1

# Delete GitHub Release
# Go to GitHub → Releases → v1.0.0-rc1 → Delete

# Create new tag for rc2 (or fix and retag)
git tag -a v1.0.0-rc1 -m "Fixed version"
git push origin v1.0.0-rc1
```

---

## See Also

- [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md) - Pre-release verification
- [RELEASE_NOTES.md](RELEASE_NOTES.md) - RC1 scope and features
- [RELEASE_FREEZE.md](RELEASE_FREEZE.md) - RC policy
- Git Documentation: https://git-scm.com/book/en/v2/Git-Basics-Tagging
