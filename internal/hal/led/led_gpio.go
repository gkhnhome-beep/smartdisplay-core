package led

import (
	"errors"
	"log"
	"runtime"
	"smartdisplay-core/internal/gpio"
	"time"
)

type GPIORGBLed struct {
	id    string
	pins  [3]*gpio.GPIOPin // R, G, B
	color [3]uint8
	ready bool
	err   error
}

// NewGPIORGBLed creates a new GPIO RGB LED on 3 output pins
func NewGPIORGBLed(id string, rPin, gPin, bPin int) *GPIORGBLed {
	return &GPIORGBLed{
		id: id,
		pins: [3]*gpio.GPIOPin{
			&gpio.GPIOPin{PinNumber: rPin},
			&gpio.GPIOPin{PinNumber: gPin},
			&gpio.GPIOPin{PinNumber: bPin},
		},
	}
}

func (l *GPIORGBLed) ID() string       { return l.id }
func (l *GPIORGBLed) Type() string     { return "rgb_led" }
func (l *GPIORGBLed) IsReady() bool    { return l.ready }
func (l *GPIORGBLed) LastError() error { return l.err }

func (l *GPIORGBLed) Init() error {
	if runtime.GOOS != "linux" {
		l.err = errors.New("GPIO RGB LED only supported on Linux")
		l.ready = false
		return l.err
	}
	for _, pin := range l.pins {
		if err := pin.Export(); err != nil {
			l.err = err
			return err
		}
		if err := pin.SetDirection("out"); err != nil {
			l.err = err
			return err
		}
		pin.Write(0)
	}
	l.ready = true
	return nil
}

func (l *GPIORGBLed) Shutdown() error {
	for _, pin := range l.pins {
		pin.Write(0)
		pin.Unexport()
	}
	l.ready = false
	return nil
}

// Write expects map[string]any{"cmd":"set_color", "r":uint8, "g":uint8, "b":uint8}
func (l *GPIORGBLed) Write(value any) error {
	m, ok := value.(map[string]any)
	if !ok {
		return errors.New("invalid command")
	}
	if cmd, ok := m["cmd"].(string); ok && cmd == "set_color" {
		r, rok := m["r"].(uint8)
		g, gok := m["g"].(uint8)
		b, bok := m["b"].(uint8)
		if !rok || !gok || !bok {
			return errors.New("invalid color values")
		}
		l.color = [3]uint8{r, g, b}
		return l.setColor(r, g, b)
	}
	return errors.New("unknown command")
}

// setColor bit-bangs the RGB value to the pins (solid color only)
func (l *GPIORGBLed) setColor(r, g, b uint8) error {
	if !l.ready {
		return errors.New("device not ready")
	}
	// Best-effort: just set pin high/low for each color
	vals := [3]uint8{r, g, b}
	for i, pin := range l.pins {
		if vals[i] > 0 {
			pin.Write(1)
		} else {
			pin.Write(0)
		}
		// Simulate bit-bang delay
		time.Sleep(100 * time.Microsecond)
	}
	log.Printf("GPIORGBLed: set color R=%d G=%d B=%d", r, g, b)
	return nil
}

var _ OutputDevice = (*GPIORGBLed)(nil)
