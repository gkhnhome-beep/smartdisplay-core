package led

import (
	"errors"
	"smartdisplay-core/internal/hal"
)

type RGBLed struct {
	id    string
	color [3]uint8
	mode  string
	err   error
}

func (l *RGBLed) IsReady() bool {
	return true
}

func (l *RGBLed) LastError() error {
	return l.err
}

func NewRGBLed(id string) *RGBLed {
	return &RGBLed{id: id, mode: "solid"}
}

func (l *RGBLed) ID() string {
	return l.id
}

func (l *RGBLed) Type() string {
	return "rgb_led"
}

func (l *RGBLed) Init() error {
	l.color = [3]uint8{0, 0, 0}
	l.mode = "solid"
	return nil
}

func (l *RGBLed) Shutdown() error {
	return nil
}

func (l *RGBLed) Write(value any) error {
	m, ok := value.(map[string]any)
	if !ok {
		return errors.New("invalid command")
	}
	if cmd, ok := m["cmd"].(string); ok {
		switch cmd {
		case "set_color":
			r, rok := m["r"].(uint8)
			g, gok := m["g"].(uint8)
			b, bok := m["b"].(uint8)
			if rok && gok && bok {
				l.color = [3]uint8{r, g, b}
				return nil
			}
			return errors.New("invalid color values")
		case "set_mode":
			mode, mok := m["mode"].(string)
			if mok && (mode == "solid" || mode == "blink" || mode == "pulse") {
				l.mode = mode
				return nil
			}
			return errors.New("invalid mode")
		}
	}
	return errors.New("unknown command")
}

var _ hal.OutputDevice = (*RGBLed)(nil)
