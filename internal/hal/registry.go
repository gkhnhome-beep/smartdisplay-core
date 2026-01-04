package hal

import "sync"

type DeviceHealth struct {
	ID    string
	Type  string
	Ready bool
	Error string
}

type Registry struct {
	mu      sync.RWMutex
	devices map[string]Device
}

func NewRegistry() *Registry {
	return &Registry{devices: make(map[string]Device)}
}

func (r *Registry) RegisterDevice(device Device) {
	r.mu.Lock()
	r.devices[device.ID()] = device
	r.mu.Unlock()
}

func (r *Registry) GetDevice(id string) Device {
	r.mu.RLock()
	dev := r.devices[id]
	r.mu.RUnlock()
	return dev
}

func (r *Registry) ListDevices() []Device {
	r.mu.RLock()
	list := make([]Device, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	r.mu.RUnlock()
	return list
}

func (r *Registry) DeviceHealthReport() []DeviceHealth {
	r.mu.RLock()
	list := make([]DeviceHealth, 0, len(r.devices))
	for _, d := range r.devices {
		health := DeviceHealth{
			ID:    d.ID(),
			Type:  d.Type(),
			Ready: d.IsReady(),
			Error: "",
		}
		if err := d.LastError(); err != nil {
			health.Error = err.Error()
		}
		list = append(list, health)
	}
	r.mu.RUnlock()
	return list
}
