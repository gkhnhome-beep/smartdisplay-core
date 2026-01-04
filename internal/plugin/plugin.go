// Package plugin provides a compile-time internal plugin system.
// This system supports:
// - Plugin interface with lifecycle: ID(), Init(), Start(), Stop()
// - Compile-time registration only (no dynamic loading)
// - PluginRegistry for managing registered plugins
// - Coordinator integration for centralized plugin management
//
// NOTE: This is for internal plugins only. No third-party or external plugins supported.
package plugin

import (
	"fmt"
	"sync"
)

// Plugin defines the interface for internal plugins.
// All plugins must implement this interface for registration.
type Plugin interface {
	// ID returns the unique identifier for this plugin.
	// IDs must be unique across all registered plugins.
	ID() string

	// Init is called once during plugin registry initialization.
	// Used for setup operations before the system starts.
	// If Init returns an error, the plugin will not be registered.
	Init() error

	// Start is called when the plugin should begin operations.
	// This is called after all plugins have been initialized.
	// If Start returns an error, the system will log it but continue.
	Start() error

	// Stop is called when the plugin should gracefully shut down.
	// This is called during system shutdown.
	// Stop should be fast and non-blocking.
	Stop() error
}

// Registry manages all registered internal plugins.
// Registration happens at compile-time; no runtime loading is supported.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin // ID -> Plugin
	started map[string]bool   // ID -> started status
	errors  map[string]error  // ID -> init/start errors
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		started: make(map[string]bool),
		errors:  make(map[string]error),
	}
}

// Register registers a plugin with the registry.
// Returns an error if a plugin with the same ID is already registered,
// or if the plugin's Init() method fails.
func (r *Registry) Register(p Plugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	id := p.ID()
	if id == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin with ID %q already registered", id)
	}

	// Initialize the plugin
	if err := p.Init(); err != nil {
		r.errors[id] = err
		return fmt.Errorf("plugin %q init failed: %w", id, err)
	}

	r.plugins[id] = p
	r.started[id] = false
	return nil
}

// Start starts a registered plugin by ID.
// Returns an error if the plugin is not found or if Start() fails.
func (r *Registry) Start(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %q not registered", id)
	}

	if r.started[id] {
		return fmt.Errorf("plugin %q already started", id)
	}

	if err := p.Start(); err != nil {
		r.errors[id] = err
		return fmt.Errorf("plugin %q start failed: %w", id, err)
	}

	r.started[id] = true
	return nil
}

// StartAll starts all registered plugins.
// If any plugin fails to start, the error is logged and StartAll continues with the rest.
// Returns a map of plugin IDs to their start errors (empty if all succeed).
func (r *Registry) StartAll() map[string]error {
	r.mu.RLock()
	pluginIDs := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		pluginIDs = append(pluginIDs, id)
	}
	r.mu.RUnlock()

	errs := make(map[string]error)
	for _, id := range pluginIDs {
		if err := r.Start(id); err != nil {
			errs[id] = err
		}
	}
	return errs
}

// Stop stops a registered plugin by ID.
// Returns an error if the plugin is not found or if Stop() fails.
func (r *Registry) Stop(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %q not registered", id)
	}

	if !r.started[id] {
		return fmt.Errorf("plugin %q not started", id)
	}

	if err := p.Stop(); err != nil {
		r.errors[id] = err
		return fmt.Errorf("plugin %q stop failed: %w", id, err)
	}

	r.started[id] = false
	return nil
}

// StopAll stops all running plugins.
// If any plugin fails to stop, the error is logged and StopAll continues with the rest.
// Returns a map of plugin IDs to their stop errors (empty if all succeed).
func (r *Registry) StopAll() map[string]error {
	r.mu.RLock()
	pluginIDs := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		if r.started[id] {
			pluginIDs = append(pluginIDs, id)
		}
	}
	r.mu.RUnlock()

	errs := make(map[string]error)
	for _, id := range pluginIDs {
		if err := r.Stop(id); err != nil {
			errs[id] = err
		}
	}
	return errs
}

// List returns a list of all registered plugin IDs.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		ids = append(ids, id)
	}
	return ids
}

// GetStatus returns detailed status information about a plugin.
// Returns an error if the plugin is not found.
func (r *Registry) GetStatus(id string) (Status, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[id]
	if !exists {
		return Status{}, fmt.Errorf("plugin %q not registered", id)
	}

	return Status{
		ID:      id,
		Started: r.started[id],
		Error:   r.errors[id],
	}, nil
}

// Status represents the current state of a plugin.
type Status struct {
	ID      string
	Started bool
	Error   error
}

// GetAllStatus returns status information for all registered plugins.
func (r *Registry) GetAllStatus() map[string]Status {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make(map[string]Status)
	for id := range r.plugins {
		statuses[id] = Status{
			ID:      id,
			Started: r.started[id],
			Error:   r.errors[id],
		}
	}
	return statuses
}
