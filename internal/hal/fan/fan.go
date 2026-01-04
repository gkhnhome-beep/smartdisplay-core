package fan

import (
	"errors"
	"smartdisplay-core/internal/hal"
)

type FanDevice struct {
	id    string
	on    bool
	level int
	err   error
}

func (f *FanDevice) IsReady() bool {
	return true
}

func (f *FanDevice) LastError() error {
	return f.err
}

func NewFanDevice(id string) *FanDevice {
	return &FanDevice{id: id}
}

func (f *FanDevice) ID() string {
	return f.id
}

func (f *FanDevice) Type() string {
	return "fan"
}

func (f *FanDevice) Init() error {
	f.on = false
	f.level = 0
	return nil
}

func (f *FanDevice) Shutdown() error {
	f.on = false
	f.level = 0
	return nil
}

func (f *FanDevice) Write(value any) error {
	m, ok := value.(map[string]any)
	if !ok {
		return errors.New("invalid command")
	}
	if cmd, ok := m["cmd"].(string); ok {
		switch cmd {
		case "on":
			f.on = true
			return nil
		case "off":
			f.on = false
			return nil
		case "set_level":
			lvl, ok := m["level"].(int)
			if ok && lvl >= 0 && lvl <= 100 {
				f.level = lvl
				return nil
			}
			return errors.New("invalid level")
		}
	}
	return errors.New("unknown command")
}

var _ hal.OutputDevice = (*FanDevice)(nil)
