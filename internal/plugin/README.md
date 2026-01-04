# SmartDisplay Plugin System (FAZ 78)

## Overview

The plugin system provides a **compile-time** internal plugin architecture for SmartDisplay. This system allows core components to register and manage internal plugins without external dependencies or dynamic loading.

### Design Principles

- **No Dynamic Loading**: All plugins are registered at compile time via direct code references
- **No External Plugins**: Plugins must be part of the SmartDisplay codebase
- **Lifecycle Management**: Plugins have defined Init, Start, and Stop phases
- **Centralized Registry**: Single PluginRegistry manages all plugin state
- **Thread-Safe**: All registry operations are protected by RWMutex

## Architecture

### Plugin Interface

Each plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    // ID returns the unique identifier for this plugin
    ID() string
    
    // Init is called once during registry initialization
    // Use for one-time setup before system starts
    Init() error
    
    // Start is called when plugin should begin operations
    // Called after all plugins initialized
    Start() error
    
    // Stop is called during system shutdown
    // Must be fast and non-blocking
    Stop() error
}
```

### PluginRegistry

The `Registry` type manages all registered plugins:

```go
registry := plugin.NewRegistry()

// Register plugins at startup
registry.Register(myPlugin)

// Start all plugins
if errs := registry.StartAll(); len(errs) > 0 {
    // Handle partial failures
}

// Stop all plugins during shutdown
registry.StopAll()
```

## Usage Example

### Define a Plugin

```go
package myfeature

type MyPlugin struct {
    config Config
    state  State
}

func (p *MyPlugin) ID() string {
    return "myfeature"
}

func (p *MyPlugin) Init() error {
    // Load config, validate resources, etc.
    return nil
}

func (p *MyPlugin) Start() error {
    // Begin operations
    p.state.running = true
    return nil
}

func (p *MyPlugin) Stop() error {
    // Graceful shutdown
    p.state.running = false
    return nil
}
```

### Register with Coordinator

In `internal/system/coordinator.go`:

```go
func NewCoordinator(...) *Coordinator {
    coord := &Coordinator{...}
    
    // Create plugin registry
    coord.pluginRegistry = plugin.NewRegistry()
    
    // Register internal plugins
    coord.pluginRegistry.Register(myfeature.NewPlugin())
    coord.pluginRegistry.Register(otherfeature.NewPlugin())
    
    return coord
}
```

### Start Plugins

```go
func (c *Coordinator) StartPlugins() error {
    if errs := c.pluginRegistry.StartAll(); len(errs) > 0 {
        logger.Error("some plugins failed to start:")
        for id, err := range errs {
            logger.Error("  " + id + ": " + err.Error())
        }
        // Continue with partial startup (some plugins may fail)
    }
    return nil
}
```

### Stop Plugins

```go
func (c *Coordinator) StopPlugins() {
    if errs := c.pluginRegistry.StopAll(); len(errs) > 0 {
        logger.Error("some plugins failed to stop:")
        for id, err := range errs {
            logger.Error("  " + id + ": " + err.Error())
        }
    }
}
```

## API Reference

### Registry Methods

| Method | Purpose |
|--------|---------|
| `NewRegistry()` | Create new empty registry |
| `Register(p Plugin)` | Register and init a plugin |
| `Start(id string)` | Start a specific plugin |
| `StartAll()` | Start all registered plugins, return error map |
| `Stop(id string)` | Stop a specific plugin |
| `StopAll()` | Stop all running plugins, return error map |
| `List()` | Get list of all plugin IDs |
| `GetStatus(id)` | Get status for one plugin |
| `GetAllStatus()` | Get status for all plugins |

### Status Type

```go
type Status struct {
    ID      string  // Plugin ID
    Started bool    // Is plugin running?
    Error   error   // Last error, if any
}
```

## Thread Safety

All Registry methods are thread-safe:
- Concurrent reads to `List()` and `GetStatus()` are allowed
- Concurrent calls to `Register()`, `Start()`, `Stop()` are safe
- Internal RWMutex protects plugin map and state

## Error Handling

### Init Errors

If a plugin's `Init()` fails:
- Plugin is NOT registered
- Returned error wraps the original error
- Registry continues without the plugin

### Start Errors

If a plugin's `Start()` fails:
- Error is stored in `Status.Error`
- Other plugins continue starting
- Use `GetAllStatus()` to find failed plugins

### Stop Errors

If a plugin's `Stop()` fails:
- Error is stored in `Status.Error`
- Other plugins continue stopping
- Use `GetAllStatus()` to audit failures

## Compile-Time Registration

All plugins are registered explicitly in code at startup:

```go
// in main.go or Coordinator init
coord.pluginRegistry.Register(logger.NewPlugin())
coord.pluginRegistry.Register(health.NewPlugin())
coord.pluginRegistry.Register(metrics.NewPlugin())
```

**No dynamic discovery, no reflection, no external loading.**

## Future Roadmap

Potential enhancements (for later phases):
- Plugin dependencies (A depends on B)
- Plugin ordering constraints
- Plugin configuration from files
- Plugin state persistence
- Plugin communication/messaging
- External plugin sandboxing (separate package)

## Testing

See `plugin_test.go` for comprehensive examples:
- `TestRegisterPlugin` - Basic registration
- `TestStartAll` - Lifecycle management
- `TestStartAllPartialFailure` - Error handling
- `ExampleRegistry_workflow` - Full example

## Files

- `plugin.go` - Plugin interface, Registry implementation
- `plugin_test.go` - Test suite with examples
- `README.md` - This documentation
