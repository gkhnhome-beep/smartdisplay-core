// Package telemetry provides privacy-first, opt-in product improvement telemetry.
// - Opt-in only: opt-in must be explicitly enabled via API
// - No personal data: only aggregated counts and categories
// - No raw events: only bucketized performance metrics
// - Standard library only: no external dependencies
// - Local aggregation: data stays on device until explicitly uploaded (disabled by default)
package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Collector aggregates telemetry data locally without sending it anywhere.
// All data is opt-in and never personally identifiable.
type Collector struct {
	mu                 sync.RWMutex
	optInEnabled       bool
	featureUsage       map[string]int
	errorCategories    map[string]int
	performanceBuckets map[string]int // count of operations in each performance bucket
	lastFlush          time.Time
	dataDir            string
}

// PerformanceBucket represents a time-based bucket for performance metrics.
// Buckets are: <100ms, <500ms, <1s, <5s, >=5s
type PerformanceBucket string

const (
	BucketVeryFast PerformanceBucket = "very_fast" // <100ms
	BucketFast     PerformanceBucket = "fast"      // <500ms
	BucketNormal   PerformanceBucket = "normal"    // <1s
	BucketSlow     PerformanceBucket = "slow"      // <5s
	BucketVerySlow PerformanceBucket = "very_slow" // >=5s
	telemetryFile                    = "telemetry.json"
)

// New creates a new telemetry collector with data directory.
// opt-in is disabled by default.
func New(dataDir string) *Collector {
	return &Collector{
		optInEnabled:       false,
		featureUsage:       make(map[string]int),
		errorCategories:    make(map[string]int),
		performanceBuckets: make(map[string]int),
		lastFlush:          time.Now(),
		dataDir:            dataDir,
	}
}

// SetOptIn enables or disables telemetry opt-in.
func (c *Collector) SetOptIn(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.optInEnabled = enabled
}

// IsOptedIn returns whether telemetry is opted in.
func (c *Collector) IsOptedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.optInEnabled
}

// RecordFeatureUsage records a feature being used (increments count).
func (c *Collector) RecordFeatureUsage(featureName string) {
	if !c.IsOptedIn() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.featureUsage[featureName]++
}

// RecordError records an error by category (not message, just the category/type).
func (c *Collector) RecordError(errorCategory string) {
	if !c.IsOptedIn() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorCategories[errorCategory]++
}

// RecordPerformance records an operation's duration in the appropriate bucket.
func (c *Collector) RecordPerformance(operationName string, duration time.Duration) {
	if !c.IsOptedIn() {
		return
	}
	bucket := getBucket(duration)
	c.mu.Lock()
	defer c.mu.Unlock()
	key := operationName + ":" + string(bucket)
	c.performanceBuckets[key]++
}

// getBucket maps a duration to a performance bucket.
func getBucket(d time.Duration) PerformanceBucket {
	switch {
	case d < 100*time.Millisecond:
		return BucketVeryFast
	case d < 500*time.Millisecond:
		return BucketFast
	case d < 1*time.Second:
		return BucketNormal
	case d < 5*time.Second:
		return BucketSlow
	default:
		return BucketVerySlow
	}
}

// Summary represents aggregated telemetry data.
type Summary struct {
	OptInEnabled       bool           `json:"opt_in_enabled"`
	FeatureUsage       map[string]int `json:"feature_usage,omitempty"`
	ErrorCategories    map[string]int `json:"error_categories,omitempty"`
	PerformanceBuckets map[string]int `json:"performance_buckets,omitempty"`
	CollectedAt        time.Time      `json:"collected_at"`
}

// GetSummary returns a snapshot of current aggregated data without personal information.
func (c *Collector) GetSummary() Summary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Deep copy maps to avoid external modification
	features := make(map[string]int)
	for k, v := range c.featureUsage {
		features[k] = v
	}

	errors := make(map[string]int)
	for k, v := range c.errorCategories {
		errors[k] = v
	}

	perf := make(map[string]int)
	for k, v := range c.performanceBuckets {
		perf[k] = v
	}

	return Summary{
		OptInEnabled:       c.optInEnabled,
		FeatureUsage:       features,
		ErrorCategories:    errors,
		PerformanceBuckets: perf,
		CollectedAt:        time.Now(),
	}
}

// Flush persists aggregated data to disk (for potential future upload).
// Data is never uploaded automatically; manual upload would require user action.
func (c *Collector) Flush() error {
	c.mu.Lock()
	summary := Summary{
		OptInEnabled:       c.optInEnabled,
		FeatureUsage:       c.featureUsage,
		ErrorCategories:    c.errorCategories,
		PerformanceBuckets: c.performanceBuckets,
		CollectedAt:        time.Now(),
	}
	c.lastFlush = time.Now()
	c.mu.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(c.dataDir, 0755); err != nil {
		return err
	}

	// Write to file
	filePath := filepath.Join(c.dataDir, telemetryFile)
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// LoadState restores persisted telemetry state from disk.
func (c *Collector) LoadState() error {
	filePath := filepath.Join(c.dataDir, telemetryFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist yet, that's fine
		return nil
	}

	var summary Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.optInEnabled = summary.OptInEnabled
	c.featureUsage = summary.FeatureUsage
	c.errorCategories = summary.ErrorCategories
	c.performanceBuckets = summary.PerformanceBuckets

	return nil
}

// Reset clears all aggregated data (useful for testing).
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.featureUsage = make(map[string]int)
	c.errorCategories = make(map[string]int)
	c.performanceBuckets = make(map[string]int)
}
