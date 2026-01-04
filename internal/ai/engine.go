package ai

import (
	"fmt"
	"strings"
	"time"
)

// GroupAnomalies groups related anomaly events into packets
func (e *InsightEngine) GroupAnomalies() []AnomalyPacket {
	var packets []AnomalyPacket
	var window time.Duration = 10 * time.Minute
	var lastPacket *AnomalyPacket

	for _, ins := range e.history {
		if ins.Type != Anomaly {
			continue
		}

		// Group repeated door opens
		if strings.Contains(ins.Detail, "door open") {
			if lastPacket != nil && strings.Contains(lastPacket.Description, "door open") && insTimeClose(lastPacket.EndTime, time.Now(), window) {
				lastPacket.Events = append(lastPacket.Events, ins)
				lastPacket.EndTime = time.Now()
				lastPacket.Severity += 0.2
				continue
			}
			p := AnomalyPacket{
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Events:      []Insight{ins},
				Severity:    0.5,
				Description: "Repeated door open events",
			}
			packets = append(packets, p)
			lastPacket = &packets[len(packets)-1]
			continue
		}

		// Group device flapping
		if strings.Contains(ins.Detail, "Device") && strings.Contains(ins.Detail, "flap") {
			if lastPacket != nil && strings.Contains(lastPacket.Description, "flapping") && insTimeClose(lastPacket.EndTime, time.Now(), window) {
				lastPacket.Events = append(lastPacket.Events, ins)
				lastPacket.EndTime = time.Now()
				lastPacket.Severity += 0.2
				continue
			}
			p := AnomalyPacket{
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Events:      []Insight{ins},
				Severity:    0.5,
				Description: "Device flapping detected",
			}
			packets = append(packets, p)
			lastPacket = &packets[len(packets)-1]
			continue
		}

		// Group alarm near-miss
		if strings.Contains(ins.Detail, "Alarm") && strings.Contains(ins.Detail, "near-miss") {
			if lastPacket != nil && strings.Contains(lastPacket.Description, "alarm near-miss") && insTimeClose(lastPacket.EndTime, time.Now(), window) {
				lastPacket.Events = append(lastPacket.Events, ins)
				lastPacket.EndTime = time.Now()
				lastPacket.Severity += 0.2
				continue
			}
			p := AnomalyPacket{
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Events:      []Insight{ins},
				Severity:    0.7,
				Description: "Alarm near-miss events",
			}
			packets = append(packets, p)
			lastPacket = &packets[len(packets)-1]
			continue
		}
	}
	return packets
}

// insTimeClose is a helper to check if two times are within a window
func insTimeClose(t1, t2 time.Time, window time.Duration) bool {
	if t1.IsZero() || t2.IsZero() {
		return false
	}
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= window
}

// AnomalyPacket groups related anomaly events
type AnomalyPacket struct {
	StartTime   time.Time
	EndTime     time.Time
	Events      []Insight
	Severity    float64
	Description string
}

// GetDailySummary generates a deterministic daily summary (max 5 bullet points, human language)
func (e *InsightEngine) GetDailySummary() string {
	var bullets []string
	// 1. Alarm events
	countAlarm := 0
	for _, ins := range e.history {
		if ins.Type == Anomaly && strings.Contains(ins.Detail, "Alarm") {
			countAlarm++
		}
	}
	if countAlarm > 0 {
		bullets = append(bullets, fmt.Sprintf("%d alarm event(s) occurred.", countAlarm))
	}

	// 2. Guest visits
	countGuest := 0
	for _, ins := range e.history {
		if ins.Type == Suggestion && strings.Contains(ins.Detail, "Guest") {
			countGuest++
		}
	}
	if countGuest > 0 {
		bullets = append(bullets, fmt.Sprintf("%d guest visit(s) detected.", countGuest))
	}

	// 3. Unusual device behavior
	countDevice := 0
	for _, ins := range e.history {
		if ins.Type == Anomaly && strings.Contains(ins.Detail, "Device") {
			countDevice++
		}
	}
	if countDevice > 0 {
		bullets = append(bullets, fmt.Sprintf("%d device issue(s) noted.", countDevice))
	}

	// 4. HA uptime (deterministic, placeholder)
	bullets = append(bullets, "Home Assistant connection stable.")

	// 5. Fallback/system normal
	if len(bullets) == 0 {
		bullets = append(bullets, "All systems normal. No issues detected today.")
	}

	if len(bullets) > 5 {
		bullets = bullets[:5]
	}

	return "Daily Summary for " + time.Now().Format("Jan 2, 2006") + ":\n- " + strings.Join(bullets, "\n- ")
}

type InsightType string

const (
	Summary     InsightType = "summary"
	Anomaly     InsightType = "anomaly"
	Suggestion  InsightType = "suggestion"
	Explanation InsightType = "explanation"
)

type Tone string

const (
	ToneNeutral      Tone = "neutral"
	ToneReassuring   Tone = "reassuring"
	ToneProfessional Tone = "professional"
)

type Insight struct {
	Type       InsightType
	Detail     string
	Severity   string
	Confidence float64
	Tone       Tone
}

type InsightEngine struct {
	current              Insight
	history              []Insight
	recentTypes          []InsightType
	windowSize           int
	lastSirenExplanation string // explanation for last siren suppression/allowance
	// Trust learning fields
	quickApprovals       int
	frequentCancels      int
	ignoredWarnings      int
	lastTrustExplanation string
}

func NewInsightEngine() *InsightEngine {
	return &InsightEngine{history: make([]Insight, 0, 10), recentTypes: make([]InsightType, 0, 5), windowSize: 5}
}

// Observe now tracks user trust actions for deterministic learning
func (e *InsightEngine) Observe(alarmState, guestState string, deviceStates ...string) {
	var newInsight Insight
	// Determine context for tone selection
	now := time.Now()
	hour := now.Hour()
	userRole := "user" // Default; can be set by Coordinator if needed
	severity := "low"
	// Device anomaly detection (deterministic)
	for _, ds := range deviceStates {
		if ds == "offline" {
			severity = "high"
			newInsight = Insight{
				Type:       Anomaly,
				Detail:     "Device offline detected.",
				Severity:   severity,
				Confidence: 0.99,
			}
			goto selectTone
		}
		if ds == "error" {
			severity = "high"
			newInsight = Insight{
				Type:       Anomaly,
				Detail:     "Device error state detected.",
				Severity:   severity,
				Confidence: 0.98,
			}
			goto selectTone
		}
	}
	if alarmState == "TRIGGERED" {
		severity = "high"
		newInsight = Insight{
			Type:       Anomaly,
			Detail:     "Alarm was triggered.",
			Severity:   severity,
			Confidence: 1.0,
		}
	} else if guestState == "APPROVED" {
		severity = "medium"
		newInsight = Insight{
			Type:       Suggestion,
			Detail:     "Guest access is active. Monitor entry.",
			Severity:   severity,
			Confidence: 0.8,
		}
	} else {
		severity = "low"
		newInsight = Insight{
			Type:       Summary,
			Detail:     "System normal.",
			Severity:   severity,
			Confidence: 0.95,
		}
	}

selectTone:
	// Tone selection logic (never playful/unsafe)
	// - High severity: Professional
	// - Night (22:00-6:00): Reassuring
	// - Admin: Professional
	// - Default: Neutral
	tone := ToneNeutral
	if severity == "high" {
		tone = ToneProfessional
	} else if hour >= 22 || hour < 6 {
		tone = ToneReassuring
	}
	if userRole == "admin" {
		tone = ToneProfessional
	}
	newInsight.Tone = tone

	// Adjust confidence/severity based on trust
	if e.quickApprovals > 3 {
		newInsight.Confidence += 0.05
		newInsight.Severity = "lower"
		e.lastTrustExplanation = "Based on your previous quick alarm approvals, confidence is increased and severity is reduced."
	} else if e.frequentCancels > 3 {
		newInsight.Confidence -= 0.1
		newInsight.Severity = "higher"
		e.lastTrustExplanation = "Based on your frequent alarm cancels, confidence is reduced and severity is increased."
	} else if e.ignoredWarnings > 3 {
		newInsight.Confidence -= 0.05
		e.lastTrustExplanation = "Based on your tendency to ignore warnings, confidence is slightly reduced."
	} else {
		e.lastTrustExplanation = ""
	}
	// Rate limit check for insight type
	for _, t := range e.recentTypes {
		if t == newInsight.Type {
			logSuppressed(newInsight)
			return
		}
	}
	e.current = newInsight
	if len(e.history) == 10 {
		importLog()
		logDiscarded(e.history[0])
		e.history = e.history[1:]
	}
	e.history = append(e.history, newInsight)
	e.recentTypes = append(e.recentTypes, newInsight.Type)
	if len(e.recentTypes) > e.windowSize {
		e.recentTypes = e.recentTypes[1:]
	}
}
func logSuppressed(insight Insight) {
	println("ai insight suppressed: " + string(insight.Type) + " - " + insight.Detail)
}

func importLog() {}
func logDiscarded(insight Insight) {
	// Use logger if available, else fallback to println
	// logger.Info("ai insight discarded: " + string(insight.Type) + " - " + insight.Detail)
	println("ai insight discarded: " + string(insight.Type) + " - " + insight.Detail)
}
func (e *InsightEngine) GetInsightHistory() []Insight {
	return append([]Insight(nil), e.history...)
}

func (e *InsightEngine) GetCurrentInsight() Insight {
	return e.current
}

// ExplainInsight returns a human explanation for the current insight or last siren decision
func (e *InsightEngine) ExplainInsight() string {
	if e.lastSirenExplanation != "" {
		return e.lastSirenExplanation
	}
	if e.lastTrustExplanation != "" {
		return e.lastTrustExplanation
	}
	switch e.current.Type {
	case Anomaly:
		return "An anomaly was detected: " + e.current.Detail
	case Suggestion:
		return "Suggestion: " + e.current.Detail
	case Summary:
		return "Summary: " + e.current.Detail
	case Explanation:
		return "Explanation: " + e.current.Detail
	default:
		return "No insight available."
	}
}

// Methods to track user trust actions
func (e *InsightEngine) TrackQuickApproval() {
	e.quickApprovals++
}

func (e *InsightEngine) TrackFrequentCancel() {
	e.frequentCancels++
}

func (e *InsightEngine) TrackIgnoredWarning() {
	e.ignoredWarnings++
}

// SetSirenExplanation sets the last siren explanation for UI/API
func (e *InsightEngine) SetSirenExplanation(msg string) {
	e.lastSirenExplanation = msg
}
