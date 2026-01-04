// Package firstboot manages the initial setup flow for SmartDisplay.
// It handles a deterministic, sequential flow through setup steps.
package firstboot

import (
	"fmt"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/logger"
	"sync"
)

// Step represents a single step in the first-boot flow
type Step struct {
	ID    string // "welcome", "language", "ha_check", "alarm_role", "ready"
	Title string // Display title
	Order int    // Step order (1-5)
}

// FirstBootManager manages the first-boot flow state
type FirstBootManager struct {
	mu       sync.RWMutex
	active   bool            // Is first-boot active?
	current  int             // Current step (1-5)
	complete map[string]bool // Track completed steps
	steps    []Step
}

// AllSteps defines the complete first-boot flow sequence
var AllSteps = []Step{
	{ID: "welcome", Title: "Welcome", Order: 1},
	{ID: "language", Title: "Language Confirmation", Order: 2},
	{ID: "ha_check", Title: "Home Assistant Check", Order: 3},
	{ID: "alarm_role", Title: "Alarm Role Explanation", Order: 4},
	{ID: "ready", Title: "Ready", Order: 5},
}

// New creates a new FirstBootManager from runtime config state
func New(wizardCompleted bool) *FirstBootManager {
	mgr := &FirstBootManager{
		active:   !wizardCompleted, // First-boot is active if wizard not completed
		current:  1,                // Always start at step 1
		complete: make(map[string]bool),
		steps:    AllSteps,
	}

	if mgr.active {
		logger.Info("firstboot: mode activated (wizard_completed=false)")
	}

	return mgr
}

// Active returns whether first-boot mode is currently active
func (m *FirstBootManager) Active() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
}

// CurrentStep returns the current step object
func (m *FirstBootManager) CurrentStep() Step {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current >= 1 && m.current <= len(m.steps) {
		return m.steps[m.current-1]
	}
	return Step{} // Should not reach here in normal flow
}

// CurrentStepID returns the ID of the current step
func (m *FirstBootManager) CurrentStepID() string {
	return m.CurrentStep().ID
}

// CurrentStepOrder returns the numeric order of current step (1-5)
func (m *FirstBootManager) CurrentStepOrder() int {
	return m.CurrentStep().Order
}

// AllStepsStatus returns status of all steps including current
func (m *FirstBootManager) AllStepsStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	steps := make([]map[string]interface{}, 0, len(m.steps))
	for i, step := range m.steps {
		stepNum := i + 1
		steps = append(steps, map[string]interface{}{
			"id":        step.ID,
			"title":     step.Title,
			"order":     step.Order,
			"completed": m.complete[step.ID],
			"current":   m.current == stepNum,
		})
	}

	return map[string]interface{}{
		"active": m.active,
		"current_step": map[string]interface{}{
			"id":    m.CurrentStep().ID,
			"order": m.CurrentStep().Order,
			"title": m.CurrentStep().Title,
		},
		"steps": steps,
	}
}

// Next advances to the next step if possible
// Returns (true, nil) on success
// Returns (false, error) if already at last step or invalid state
func (m *FirstBootManager) Next() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return false, fmt.Errorf("firstboot: not in active mode")
	}

	if m.current >= len(m.steps) {
		return false, fmt.Errorf("firstboot: already at final step")
	}

	// Mark current step as completed and advance
	m.complete[m.steps[m.current-1].ID] = true
	m.current++

	logger.Info(fmt.Sprintf("firstboot: advanced to step %d (%s)", m.current, m.steps[m.current-1].ID))

	return true, nil
}

// Back returns to the previous step if possible
// Returns (true, nil) on success
// Returns (false, error) if already at first step
func (m *FirstBootManager) Back() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return false, fmt.Errorf("firstboot: not in active mode")
	}

	if m.current <= 1 {
		return false, fmt.Errorf("firstboot: already at first step")
	}

	m.current--

	logger.Info(fmt.Sprintf("firstboot: returned to step %d (%s)", m.current, m.steps[m.current-1].ID))

	return true, nil
}

// Complete exits first-boot mode and marks wizard as completed
// Returns (true, nil) on success
// Returns (false, error) if not at final step
func (m *FirstBootManager) Complete() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return false, fmt.Errorf("firstboot: not in active mode")
	}

	if m.current != len(m.steps) {
		return false, fmt.Errorf("firstboot: must complete all steps before finishing")
	}

	// Mark final step as completed
	m.complete[m.steps[len(m.steps)-1].ID] = true

	// Exit first-boot mode
	m.active = false

	logger.Info("firstboot: wizard completed, exiting first-boot mode")

	return true, nil
}

// SaveCompletion persists the wizard_completed flag to runtime config
func SaveCompletion(completed bool) error {
	cfg, err := config.LoadRuntimeConfig()
	if err != nil {
		return err
	}

	cfg.WizardCompleted = completed

	if err := config.SaveRuntimeConfig(cfg); err != nil {
		return err
	}

	if completed {
		logger.Info("firstboot: wizard_completed flag saved to config")
	}

	return nil
}
