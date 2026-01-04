package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestValidatePackageSuccess verifies correct checksum acceptance.
func TestValidatePackageSuccess(t *testing.T) {
	mgr := New("1.0.0", "tmp", &StubAuditLogger{})

	// Create test data
	testData := []byte("test package content")
	hash := sha256.Sum256(testData)
	checksumHex := hex.EncodeToString(hash[:])

	// Validate with correct checksum
	err := mgr.ValidatePackage(bytes.NewReader(testData), checksumHex)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

// TestValidatePackageChecksumMismatch verifies rejection of invalid checksums.
func TestValidatePackageChecksumMismatch(t *testing.T) {
	mgr := New("1.0.0", "tmp", &StubAuditLogger{})

	testData := []byte("test package content")
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	// Validate with wrong checksum
	err := mgr.ValidatePackage(bytes.NewReader(testData), wrongChecksum)
	if err == nil {
		t.Fatal("validation should fail with wrong checksum")
	}
}

// TestStageUpdateWithoutValidation verifies staging works independently.
func TestStageUpdateWithoutValidation(t *testing.T) {
	mgr := New("1.0.0", "tmp/test_stage", &StubAuditLogger{})
	defer mgr.ClearStaged()

	testData := []byte("test package")
	pkg := PackageInfo{
		Version:  "2.0.0",
		BuildID:  "build123",
		Checksum: "abc123",
		Size:     int64(len(testData)),
	}

	// Stage update
	path, err := mgr.StageUpdate(bytes.NewReader(testData), pkg)
	if err != nil {
		t.Fatalf("staging failed: %v", err)
	}

	if path == "" {
		t.Fatal("staging should return path")
	}

	// Verify staged package is recorded
	status := mgr.GetStatus()
	if status.Staged == nil {
		t.Fatal("staged package should be recorded in status")
	}

	if status.Staged.Version != "2.0.0" {
		t.Fatalf("staged version mismatch: %s", status.Staged.Version)
	}
}

// TestActivateOnRebootRequiresStaged verifies reboot activation needs staged package.
func TestActivateOnRebootRequiresStaged(t *testing.T) {
	mgr := New("1.0.0", "tmp", &StubAuditLogger{})

	// Try to activate without staging
	err := mgr.ActivateOnReboot()
	if err == nil {
		t.Fatal("activation should fail without staged package")
	}
}

// TestActivateOnRebootWithStaged verifies successful reboot scheduling.
func TestActivateOnRebootWithStaged(t *testing.T) {
	mgr := New("1.0.0", "tmp/test_activate", &StubAuditLogger{})
	defer mgr.ClearStaged()

	// Stage update first
	pkg := PackageInfo{
		Version: "2.0.0",
		BuildID: "build123",
	}
	_, _ = mgr.StageUpdate(bytes.NewReader([]byte("data")), pkg)

	// Activate for reboot
	err := mgr.ActivateOnReboot()
	if err != nil {
		t.Fatalf("activation failed: %v", err)
	}

	status := mgr.GetStatus()
	if !status.PendingReboot {
		t.Fatal("pending reboot should be set")
	}
}

// TestCancelActivation verifies reboot activation can be cancelled.
func TestCancelActivation(t *testing.T) {
	mgr := New("1.0.0", "tmp/test_cancel", &StubAuditLogger{})
	defer mgr.ClearStaged()

	// Stage and activate
	pkg := PackageInfo{Version: "2.0.0", BuildID: "build123"}
	_, _ = mgr.StageUpdate(bytes.NewReader([]byte("data")), pkg)
	_ = mgr.ActivateOnReboot()

	// Verify pending
	status := mgr.GetStatus()
	if !status.PendingReboot {
		t.Fatal("should have pending reboot")
	}

	// Cancel activation
	err := mgr.CancelActivation()
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	status = mgr.GetStatus()
	if status.PendingReboot {
		t.Fatal("pending reboot should be cleared")
	}
}

// TestCheckAvailableStub verifies stub returns nil (no remote implemented).
func TestCheckAvailableStub(t *testing.T) {
	mgr := New("1.0.0", "tmp", &StubAuditLogger{})

	available, err := mgr.CheckAvailable()
	if err != nil {
		t.Fatalf("check available failed: %v", err)
	}

	if available != nil {
		t.Fatal("stub should return nil")
	}
}

// TestAuditLogging verifies all actions are logged.
func TestAuditLogging(t *testing.T) {
	logger := &StubAuditLogger{}
	mgr := New("1.0.0", "tmp/test_audit", logger)
	defer mgr.ClearStaged()

	// Perform operations
	testData := []byte("data")
	hash := sha256.Sum256(testData)
	checksum := hex.EncodeToString(hash[:])

	mgr.ValidatePackage(bytes.NewReader(testData), checksum)
	pkg := PackageInfo{Version: "2.0.0", BuildID: "build123"}
	mgr.StageUpdate(bytes.NewReader(testData), pkg)
	mgr.ActivateOnReboot()

	// Check audit log
	auditLog := mgr.GetAuditLog()
	if len(auditLog) == 0 {
		t.Fatal("audit log should have entries")
	}

	// Verify entries contain expected actions
	hasValidate := false
	hasStage := false
	hasActivate := false

	for _, entry := range auditLog {
		if bytes.Contains([]byte(entry), []byte("update_validate")) {
			hasValidate = true
		}
		if bytes.Contains([]byte(entry), []byte("update_staged")) {
			hasStage = true
		}
		if bytes.Contains([]byte(entry), []byte("update_activate")) {
			hasActivate = true
		}
	}

	if !hasValidate || !hasStage || !hasActivate {
		t.Fatal("audit log missing expected entries")
	}
}

// TestGetStatus verifies status snapshot.
func TestGetStatus(t *testing.T) {
	mgr := New("1.0.0", "tmp/test_status", &StubAuditLogger{})
	defer mgr.ClearStaged()

	status := mgr.GetStatus()

	if status.CurrentVersion != "1.0.0" {
		t.Fatalf("version mismatch: %s", status.CurrentVersion)
	}

	if status.Available != nil {
		t.Fatal("available should be nil initially")
	}

	if status.Staged != nil {
		t.Fatal("staged should be nil initially")
	}

	if status.PendingReboot {
		t.Fatal("pending reboot should be false initially")
	}
}

// ExampleWorkflow demonstrates typical update workflow.
func ExampleWorkflow(t *testing.T) {
	logger := &StubAuditLogger{}
	mgr := New("1.0.0", "tmp/example", logger)
	defer mgr.ClearStaged()

	// 1. Check for available updates (stub, always returns nil)
	available, _ := mgr.CheckAvailable()
	if available != nil {
		// In real scenario, would fetch metadata
		testData := []byte("downloaded package data")

		// 2. Validate package integrity
		if err := mgr.ValidatePackage(bytes.NewReader(testData), available.Checksum); err == nil {
			// 3. Stage the update
			if path, err := mgr.StageUpdate(bytes.NewReader(testData), *available); err == nil {
				// 4. Schedule for reboot
				_ = mgr.ActivateOnReboot()
				_ = path // staged path
			}
		}
	}

	// 5. Check status
	status := mgr.GetStatus()
	_ = status // would show pending update
}
