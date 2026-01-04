// Package trust manages the internal TrustScore for the system-user relationship.
// TrustScore is never exposed to the user, only used to adjust tone, explanation depth, and alert urgency.
package trust

import (
	"log"
	"sync"
)

const (
	minTrustScore  = 0
	maxTrustScore  = 100
	trustScoreStep = 5
)

var (
	trustScore = 50 // Start at neutral
	mu         sync.Mutex
)

// ConfirmWarning should be called when the user confirms a warning.
func ConfirmWarning(reason string) {
	adjustTrustScore(trustScoreStep, "Confirmed warning: "+reason)
}

// IgnoreWarning should be called when the user ignores a warning.
func IgnoreWarning(reason string) {
	adjustTrustScore(-trustScoreStep, "Ignored warning: "+reason)
}

func adjustTrustScore(delta int, logReason string) {
	mu.Lock()
	defer mu.Unlock()
	old := trustScore
	trustScore += delta
	if trustScore > maxTrustScore {
		trustScore = maxTrustScore
	}
	if trustScore < minTrustScore {
		trustScore = minTrustScore
	}
	if old != trustScore {
		log.Printf("[TrustScore] %s | Score: %d -> %d", logReason, old, trustScore)
	}
}

// getTrustScore returns the current TrustScore (internal use only).
func getTrustScore() int {
	mu.Lock()
	defer mu.Unlock()
	return trustScore
}
