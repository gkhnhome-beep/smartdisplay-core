package telemetry

import (
	"testing"
	"time"
)

// ExampleUsage demonstrates how to use the telemetry collector.
func ExampleUsage() {
	// Create a new collector
	collector := New("data")

	// Enable opt-in
	collector.SetOptIn(true)

	// Record feature usage
	collector.RecordFeatureUsage("alarm_armed")
	collector.RecordFeatureUsage("alarm_armed")
	collector.RecordFeatureUsage("guest_approval")

	// Record errors
	collector.RecordError("network_timeout")
	collector.RecordError("network_timeout")

	// Record performance metrics
	collector.RecordPerformance("ha_query", 50*time.Millisecond)
	collector.RecordPerformance("ha_query", 300*time.Millisecond)
	collector.RecordPerformance("alarm_arm", 1500*time.Millisecond)

	// Get summary
	_ = collector.GetSummary()
	// summary.FeatureUsage["alarm_armed"] == 2
	// summary.FeatureUsage["guest_approval"] == 1
	// summary.ErrorCategories["network_timeout"] == 2
	// summary.PerformanceBuckets["ha_query:very_fast"] == 1
	// summary.PerformanceBuckets["ha_query:fast"] == 1
	// summary.PerformanceBuckets["alarm_arm:slow"] == 1

	// Persist to disk
	_ = collector.Flush()

	// Load state later
	collector2 := New("data")
	_ = collector2.LoadState()
	// collector2.IsOptedIn() == true (loaded from disk)
}

// TestCollectorOptInOnly verifies that nothing is recorded when opt-in is disabled.
func TestCollectorOptInOnly(t *testing.T) {
	collector := New("data")

	// Opt-in is disabled by default
	if collector.IsOptedIn() {
		t.Fatal("opt-in should be disabled by default")
	}

	// Try to record something
	collector.RecordFeatureUsage("test_feature")
	collector.RecordError("test_error")
	collector.RecordPerformance("test_op", 100*time.Millisecond)

	// Nothing should be recorded
	summary := collector.GetSummary()
	if len(summary.FeatureUsage) > 0 {
		t.Fatal("feature usage should be empty when opt-in is disabled")
	}
	if len(summary.ErrorCategories) > 0 {
		t.Fatal("error categories should be empty when opt-in is disabled")
	}
	if len(summary.PerformanceBuckets) > 0 {
		t.Fatal("performance buckets should be empty when opt-in is disabled")
	}
}

// TestPerformanceBuckets verifies correct bucketing of durations.
func TestPerformanceBuckets(t *testing.T) {
	collector := New("data")
	collector.SetOptIn(true)

	// Test each bucket
	collector.RecordPerformance("op", 50*time.Millisecond)  // very_fast
	collector.RecordPerformance("op", 300*time.Millisecond) // fast
	collector.RecordPerformance("op", 800*time.Millisecond) // normal
	collector.RecordPerformance("op", 3*time.Second)        // slow
	collector.RecordPerformance("op", 10*time.Second)       // very_slow

	summary := collector.GetSummary()

	if summary.PerformanceBuckets["op:very_fast"] != 1 {
		t.Fatalf("expected very_fast count 1, got %d", summary.PerformanceBuckets["op:very_fast"])
	}
	if summary.PerformanceBuckets["op:fast"] != 1 {
		t.Fatalf("expected fast count 1, got %d", summary.PerformanceBuckets["op:fast"])
	}
	if summary.PerformanceBuckets["op:normal"] != 1 {
		t.Fatalf("expected normal count 1, got %d", summary.PerformanceBuckets["op:normal"])
	}
	if summary.PerformanceBuckets["op:slow"] != 1 {
		t.Fatalf("expected slow count 1, got %d", summary.PerformanceBuckets["op:slow"])
	}
	if summary.PerformanceBuckets["op:very_slow"] != 1 {
		t.Fatalf("expected very_slow count 1, got %d", summary.PerformanceBuckets["op:very_slow"])
	}
}

// TestPersistence verifies that state can be saved and loaded.
func TestPersistence(t *testing.T) {
	collector := New("data")
	collector.SetOptIn(true)
	collector.RecordFeatureUsage("test_feature")

	// Save state
	if err := collector.Flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	// Load in a new collector
	collector2 := New("data")
	if err := collector2.LoadState(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify state was restored
	if !collector2.IsOptedIn() {
		t.Fatal("opt-in should be restored")
	}

	summary := collector2.GetSummary()
	if summary.FeatureUsage["test_feature"] != 1 {
		t.Fatalf("expected feature usage 1, got %d", summary.FeatureUsage["test_feature"])
	}
}
