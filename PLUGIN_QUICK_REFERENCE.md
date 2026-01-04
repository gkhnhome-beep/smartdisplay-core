# FAZ 78 Plugin System - Quick Reference

## Quick Start

### 1. Implement a Plugin

```go
package myfeature

type MyPlugin struct {
    // ... fields ...
}

func (p *MyPlugin) ID() string {
    return "myfeature"
}

func (p *MyPlugin) Init() error {
    // One-time setup
    return nil
}

func (p *MyPlugin) Start() error {
    // Begin operations
    return nil
}

func (p *MyPlugin) Stop() error {
    // Graceful shutdown
    return nil
}
```

### 2. Register with Coordinator

```go
coord := NewCoordinator(...)

// Register plugins
coord.RegisterPlugin(myfeature.NewPlugin())
coord.RegisterPlugin(otherfeature.NewPlugin())
```

### 3. Startup and Shutdown

```go
// During startup
errs := coord.StartPlugins()
if len(errs) > 0 {
    logger.Error("Some plugins failed to start")
}

// ... system running ...

// During shutdown
errs := coord.StopPlugins()
```

## Plugin Lifecycle

```
1. RegisterPlugin() -> Plugin.Init() called
2. StartPlugins() -> Plugin.Start() called for each plugin
3. ... system running ...
4. StopPlugins() -> Plugin.Stop() called for each plugin
```

## API Reference

### Coordinator Methods

```go
// Register plugin (calls Init())
func (c *Coordinator) RegisterPlugin(p plugin.Plugin) error

// Start all registered plugins
func (c *Coordinator) StartPlugins() map[string]error

// Stop all running plugins
func (c *Coordinator) StopPlugins() map[string]error

// Get status of one plugin
func (c *Coordinator) GetPluginStatus(id string) (plugin.Status, error)

// Get status of all plugins
func (c *Coordinator) GetAllPluginStatus() map[string]plugin.Status

// List all plugin IDs
func (c *Coordinator) ListPlugins() []string
```

### Plugin Status

```go
type Status struct {
    ID      string  // Plugin ID
    Started bool    // Is plugin running?
    Error   error   // Last error (if any)
}
```

## Common Patterns

### Partial Failure Handling

```go
// StartAll continues even if some plugins fail
errs := coord.StartPlugins()
for id, err := range errs {
    logger.Error("Plugin failed: " + id + " - " + err.Error())
}
```

### Status Checking

```go
statuses := coord.GetAllPluginStatus()
for id, status := range statuses {
    if !status.Started {
        logger.Warn("Plugin not running: " + id)
    }
    if status.Error != nil {
        logger.Error("Plugin error: " + id + " - " + status.Error.Error())
    }
}
```

### Single Plugin Management

```go
// Start one plugin
if err := coord.StartPlugins(); err != nil {
    // ... handle error ...
}

// Check status
status, err := coord.GetPluginStatus("myfeature")
if err == nil && status.Started {
    // Plugin is running
}

// Stop one plugin via status
statuses := coord.GetAllPluginStatus()
for id, s := range statuses {
    if s.Started {
        logger.Info("Plugin running: " + id)
    }
}
```

## Example: Full Lifecycle

```go
package main

import (
    "smartdisplay-core/internal/system"
    "smartdisplay-core/internal/plugin"
)

// Define a simple plugin
type HealthPlugin struct{}

func (p *HealthPlugin) ID() string { return "health" }
func (p *HealthPlugin) Init() error {
    println("Health check: Init()")
    return nil
}
func (p *HealthPlugin) Start() error {
    println("Health check: Start()")
    return nil
}
func (p *HealthPlugin) Stop() error {
    println("Health check: Stop()")
    return nil
}

func main() {
    // Create coordinator
    coord := system.NewCoordinator(...)
    
    // Register plugins
    coord.RegisterPlugin(&HealthPlugin{})
    
    // Start all plugins
    if errs := coord.StartPlugins(); len(errs) > 0 {
        panic("Failed to start plugins")
    }
    
    // Check plugin status
    status, _ := coord.GetPluginStatus("health")
    println("Health plugin started:", status.Started)
    
    // List all plugins
    plugins := coord.ListPlugins()
    println("Registered plugins:", plugins)
    
    // Shutdown
    coord.StopPlugins()
}
```

## Error Handling

### Register Error
```go
if err := coord.RegisterPlugin(p); err != nil {
    // Plugin not registered (Init failed or duplicate ID)
    logger.Error("Failed to register: " + err.Error())
}
```

### Startup Errors
```go
errs := coord.StartPlugins()
// errs is map[string]error - ID -> error
// Only contains plugins that failed to start
// Other plugins continue running
```

### Status Check
```go
status, err := coord.GetPluginStatus("myfeature")
if err != nil {
    // Plugin not registered
} else if status.Error != nil {
    // Plugin encountered an error during Init/Start/Stop
}
```

## Testing Example

```go
func TestMyPlugin(t *testing.T) {
    reg := plugin.NewRegistry()
    p := &MyPlugin{}
    
    // Register
    if err := reg.Register(p); err != nil {
        t.Fatal(err)
    }
    
    // Start
    if err := reg.Start(p.ID()); err != nil {
        t.Fatal(err)
    }
    
    // Check status
    status, _ := reg.GetStatus(p.ID())
    if !status.Started {
        t.Error("Plugin should be started")
    }
    
    // Stop
    if err := reg.Stop(p.ID()); err != nil {
        t.Fatal(err)
    }
}
```

## Important Notes

### Compile-Time Only
- All plugins must be registered in source code
- No dynamic discovery or external loading
- No reflection used

### Thread-Safe
- All operations safe for concurrent access
- RWMutex protects internal state
- Multiple goroutines can call methods safely

### No Auto-Start
- Plugins must be explicitly registered
- StartPlugins() must be called to start all
- Plugins don't auto-restart if they fail

### Fast Shutdown
- Stop() should complete quickly
- Don't block on locks or I/O
- Should be idempotent (safe to call multiple times)

## Future Enhancements

Potential additions in future phases:
- Plugin dependencies
- Plugin ordering
- Configuration files
- State persistence
- Messaging/communication
- Performance monitoring
- Resource limits

---

For full documentation, see [internal/plugin/README.md](internal/plugin/README.md)
