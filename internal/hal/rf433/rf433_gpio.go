//go:build rpi
// +build rpi

package rf433

import (
	"log"
	"smartdisplay-core/internal/gpio"
	"sync"
	"time"
)

// EdgeEvent represents a GPIO edge event (rising/falling)
type EdgeEvent struct {
	Timestamp time.Time
	Rising    bool
}

// RF433GPIODevice binds RF433 to a GPIO input pin, detects edges, and forwards raw codes
// No protocol decoding, no debounce logic
// Only edge pattern to code conversion

type RF433GPIODevice struct {
	id     string
	pin    *gpio.GPIOPin
	events []EdgeEvent
	lock   sync.Mutex
	ready  bool
	err    error
	quit   chan struct{}
}

func NewRF433GPIODevice(id string, pinNum int) *RF433GPIODevice {
	return &RF433GPIODevice{
		id:   id,
		pin:  &gpio.GPIOPin{PinNumber: pinNum},
		quit: make(chan struct{}),
	}
}

func (d *RF433GPIODevice) ID() string       { return d.id }
func (d *RF433GPIODevice) Type() string     { return "rf433" }
func (d *RF433GPIODevice) IsReady() bool    { return d.ready }
func (d *RF433GPIODevice) LastError() error { return d.err }

func (d *RF433GPIODevice) Init() error {
	if err := d.pin.Export(); err != nil {
		d.err = err
		return err
	}
	if err := d.pin.SetDirection("in"); err != nil {
		d.err = err
		return err
	}
	d.ready = true
	go d.pollEdges()
	return nil
}

func (d *RF433GPIODevice) Shutdown() error {
	close(d.quit)
	d.ready = false
	return d.pin.Unexport()
}

// pollEdges polls the GPIO pin for edge changes and records events
func (d *RF433GPIODevice) pollEdges() {
	var last int
	for {
		select {
		case <-d.quit:
			return
		default:
			val, err := d.pin.Read()
			if err != nil {
				d.err = err
				log.Printf("RF433GPIODevice: GPIO read error: %v", err)
				time.Sleep(10 * time.Millisecond)
				continue
			}
			if val != last {
				d.lock.Lock()
				d.events = append(d.events, EdgeEvent{Timestamp: time.Now(), Rising: val == 1})
				d.lock.Unlock()
				last = val
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// Read returns a slice of edge events as a raw code (timestamped pattern)
func (d *RF433GPIODevice) Read() (any, error) {
	if !d.ready {
		return nil, nil
	}
	d.lock.Lock()
	evs := d.events
	d.events = nil
	d.lock.Unlock()
	if len(evs) == 0 {
		return nil, nil
	}
	return evs, nil
}

var _ InputDevice = (*RF433GPIODevice)(nil)
