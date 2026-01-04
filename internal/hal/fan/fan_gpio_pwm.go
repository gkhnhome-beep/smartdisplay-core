package fan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"smartdisplay-core/internal/gpio"
	"strconv"
	"strings"
)

type GPIOPWMFan struct {
	id      string
	pwmPath string
	gpioPin *gpio.GPIOPin
	usePWM  bool
	level   int
	ready   bool
	err     error
}

// NewGPIOPWMFan tries to detect sysfs PWM, falls back to GPIO
func NewGPIOPWMFan(id string, pwmchip, pwmnum, gpioPin int) *GPIOPWMFan {
	pwmPath := fmt.Sprintf("/sys/class/pwm/pwmchip%d/pwm%d", pwmchip, pwmnum)
	usePWM := false
	if runtime.GOOS == "linux" {
		if _, err := os.Stat(pwmPath); err == nil {
			usePWM = true
		}
	}
	return &GPIOPWMFan{
		id:      id,
		pwmPath: pwmPath,
		gpioPin: &gpio.GPIOPin{PinNumber: gpioPin},
		usePWM:  usePWM,
	}
}

func (f *GPIOPWMFan) ID() string       { return f.id }
func (f *GPIOPWMFan) Type() string     { return "fan" }
func (f *GPIOPWMFan) IsReady() bool    { return f.ready }
func (f *GPIOPWMFan) LastError() error { return f.err }

func (f *GPIOPWMFan) Init() error {
	if f.usePWM {
		// Enable PWM if not already enabled
		enablePath := filepath.Join(f.pwmPath, "enable")
		if err := writeString(enablePath, "1"); err != nil {
			f.err = err
			return err
		}
		f.ready = true
		return nil
	}
	// Fallback: GPIO output
	if err := f.gpioPin.Export(); err != nil {
		f.err = err
		return err
	}
	if err := f.gpioPin.SetDirection("out"); err != nil {
		f.err = err
		return err
	}
	f.gpioPin.Write(0)
	f.ready = true
	return nil
}

func (f *GPIOPWMFan) Shutdown() error {
	if f.usePWM {
		enablePath := filepath.Join(f.pwmPath, "enable")
		writeString(enablePath, "0")
		f.ready = false
		return nil
	}
	f.gpioPin.Write(0)
	f.gpioPin.Unexport()
	f.ready = false
	return nil
}

// Write expects map[string]any{"cmd":"on"/"off"/"set_level", "level":int}
func (f *GPIOPWMFan) Write(value any) error {
	m, ok := value.(map[string]any)
	if !ok {
		return errors.New("invalid command")
	}
	if cmd, ok := m["cmd"].(string); ok {
		switch cmd {
		case "on":
			return f.setLevel(100)
		case "off":
			return f.setLevel(0)
		case "set_level":
			lvl, ok := m["level"].(int)
			if !ok || lvl < 0 || lvl > 100 {
				return errors.New("invalid level")
			}
			return f.setLevel(lvl)
		}
	}
	return errors.New("unknown command")
}

func (f *GPIOPWMFan) setLevel(level int) error {
	if !f.ready {
		return errors.New("device not ready")
	}
	f.level = level
	if f.usePWM {
		// Set duty cycle (best effort, assume 25kHz, 0-100%)
		dutyPath := filepath.Join(f.pwmPath, "duty_cycle")
		periodPath := filepath.Join(f.pwmPath, "period")
		period := 40000 // ns (25kHz)
		duty := period * level / 100
		writeString(periodPath, strconv.Itoa(period))
		writeString(dutyPath, strconv.Itoa(duty))
		return nil
	}
	// Fallback: GPIO on/off
	if level > 0 {
		f.gpioPin.Write(1)
	} else {
		f.gpioPin.Write(0)
	}
	return nil
}

// writeString is a local helper for sysfs
func writeString(path, val string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(strings.TrimSpace(val))
	return err
}

var _ OutputDevice = (*GPIOPWMFan)(nil)
