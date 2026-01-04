package plugin

import (
	"errors"
	"testing"
)

// MockPlugin is a test plugin that tracks its lifecycle.
type MockPlugin struct {
	id         string
	initErr    error
	startErr   error
	stopErr    error
	initCalls  int
	startCalls int
	stopCalls  int
}

func NewMockPlugin(id string) *MockPlugin {
	return &MockPlugin{id: id}
}

func (m *MockPlugin) ID() string {
	return m.id
}

func (m *MockPlugin) Init() error {
	m.initCalls++
	return m.initErr
}

func (m *MockPlugin) Start() error {
	m.startCalls++
	return m.startErr
}

func (m *MockPlugin) Stop() error {
	m.stopCalls++
	return m.stopErr
}

func TestRegisterPlugin(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("test-plugin")

	if err := reg.Register(p); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if p.initCalls != 1 {
		t.Errorf("Expected Init to be called once, got %d", p.initCalls)
	}

	// Check that plugin is in the list
	plugins := reg.List()
	if len(plugins) != 1 || plugins[0] != "test-plugin" {
		t.Errorf("Expected plugin list to contain 'test-plugin', got %v", plugins)
	}
}

func TestRegisterDuplicatePlugin(t *testing.T) {
	reg := NewRegistry()
	p1 := NewMockPlugin("duplicate")
	p2 := NewMockPlugin("duplicate")

	if err := reg.Register(p1); err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	if err := reg.Register(p2); err == nil {
		t.Error("Second register should have failed with duplicate ID")
	}
}

func TestRegisterNilPlugin(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(nil); err == nil {
		t.Error("Should reject nil plugin")
	}
}

func TestRegisterEmptyID(t *testing.T) {
	reg := NewRegistry()
	p := &MockPlugin{id: ""}
	if err := reg.Register(p); err == nil {
		t.Error("Should reject plugin with empty ID")
	}
}

func TestRegisterInitError(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("failing-plugin")
	p.initErr = errors.New("init failed")

	if err := reg.Register(p); err == nil {
		t.Error("Should propagate init error")
	}

	// Plugin should not be registered
	plugins := reg.List()
	if len(plugins) != 0 {
		t.Errorf("Plugin should not be registered after init failure, got %v", plugins)
	}
}

func TestStartPlugin(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("test-plugin")

	reg.Register(p)
	if err := reg.Start("test-plugin"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if p.startCalls != 1 {
		t.Errorf("Expected Start to be called once, got %d", p.startCalls)
	}

	status, _ := reg.GetStatus("test-plugin")
	if !status.Started {
		t.Error("Plugin should be marked as started")
	}
}

func TestStartNonexistentPlugin(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Start("nonexistent"); err == nil {
		t.Error("Should fail to start nonexistent plugin")
	}
}

func TestStartPluginError(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("failing-plugin")
	p.startErr = errors.New("start failed")

	reg.Register(p)
	if err := reg.Start("failing-plugin"); err == nil {
		t.Error("Should propagate start error")
	}

	status, _ := reg.GetStatus("failing-plugin")
	if status.Started {
		t.Error("Plugin should not be marked as started after error")
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("test-plugin")

	reg.Register(p)
	reg.Start("test-plugin")

	// Try to start again
	if err := reg.Start("test-plugin"); err == nil {
		t.Error("Should fail to start already started plugin")
	}

	if p.startCalls != 1 {
		t.Errorf("Start should only be called once, got %d", p.startCalls)
	}
}

func TestStopPlugin(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("test-plugin")

	reg.Register(p)
	reg.Start("test-plugin")

	if err := reg.Stop("test-plugin"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if p.stopCalls != 1 {
		t.Errorf("Expected Stop to be called once, got %d", p.stopCalls)
	}

	status, _ := reg.GetStatus("test-plugin")
	if status.Started {
		t.Error("Plugin should be marked as stopped")
	}
}

func TestStopNotStarted(t *testing.T) {
	reg := NewRegistry()
	p := NewMockPlugin("test-plugin")

	reg.Register(p)

	if err := reg.Stop("test-plugin"); err == nil {
		t.Error("Should fail to stop plugin that was never started")
	}

	if p.stopCalls != 0 {
		t.Errorf("Stop should not be called, got %d", p.stopCalls)
	}
}

func TestStartAll(t *testing.T) {
	reg := NewRegistry()
	p1 := NewMockPlugin("plugin1")
	p2 := NewMockPlugin("plugin2")

	reg.Register(p1)
	reg.Register(p2)

	errs := reg.StartAll()
	if len(errs) != 0 {
		t.Errorf("StartAll should succeed with no errors, got %v", errs)
	}

	if p1.startCalls != 1 || p2.startCalls != 1 {
		t.Errorf("Expected both plugins to be started once")
	}

	status1, _ := reg.GetStatus("plugin1")
	status2, _ := reg.GetStatus("plugin2")
	if !status1.Started || !status2.Started {
		t.Error("Both plugins should be marked as started")
	}
}

func TestStartAllPartialFailure(t *testing.T) {
	reg := NewRegistry()
	p1 := NewMockPlugin("plugin1")
	p2 := NewMockPlugin("plugin2")
	p2.startErr = errors.New("plugin2 start failed")

	reg.Register(p1)
	reg.Register(p2)

	errs := reg.StartAll()
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errs))
	}

	status1, _ := reg.GetStatus("plugin1")
	status2, _ := reg.GetStatus("plugin2")
	if !status1.Started {
		t.Error("plugin1 should be started")
	}
	if status2.Started {
		t.Error("plugin2 should not be started")
	}
}

func TestStopAll(t *testing.T) {
	reg := NewRegistry()
	p1 := NewMockPlugin("plugin1")
	p2 := NewMockPlugin("plugin2")

	reg.Register(p1)
	reg.Register(p2)
	reg.StartAll()

	errs := reg.StopAll()
	if len(errs) != 0 {
		t.Errorf("StopAll should succeed with no errors, got %v", errs)
	}

	if p1.stopCalls != 1 || p2.stopCalls != 1 {
		t.Errorf("Expected both plugins to be stopped once")
	}

	status1, _ := reg.GetStatus("plugin1")
	status2, _ := reg.GetStatus("plugin2")
	if status1.Started || status2.Started {
		t.Error("Both plugins should be marked as stopped")
	}
}

func TestGetAllStatus(t *testing.T) {
	reg := NewRegistry()
	p1 := NewMockPlugin("plugin1")
	p2 := NewMockPlugin("plugin2")

	reg.Register(p1)
	reg.Register(p2)
	reg.Start("plugin1")

	statuses := reg.GetAllStatus()
	if len(statuses) != 2 {
		t.Errorf("Expected 2 plugin statuses, got %d", len(statuses))
	}

	if !statuses["plugin1"].Started {
		t.Error("plugin1 should be marked as started")
	}
	if statuses["plugin2"].Started {
		t.Error("plugin2 should not be marked as started")
	}
}

// ExamplePlugin demonstrates a simple internal plugin implementation.
type ExamplePlugin struct {
	name        string
	initialized bool
	running     bool
}

func (e *ExamplePlugin) ID() string {
	return e.name
}

func (e *ExamplePlugin) Init() error {
	e.initialized = true
	return nil
}

func (e *ExamplePlugin) Start() error {
	e.running = true
	return nil
}

func (e *ExamplePlugin) Stop() error {
	e.running = false
	return nil
}

func ExampleRegistry_workflow() {
	// Create registry and plugins
	reg := NewRegistry()
	auth := &ExamplePlugin{name: "auth"}
	logger := &ExamplePlugin{name: "logger"}

	// Register plugins (calls Init)
	reg.Register(auth)
	reg.Register(logger)

	// Start all plugins
	reg.StartAll()

	// Check status
	status := reg.GetAllStatus()
	for id, s := range status {
		println(id, "started:", s.Started)
	}

	// Stop all plugins
	reg.StopAll()
}
