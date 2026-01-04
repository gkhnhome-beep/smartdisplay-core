//go:build linux
// +build linux

package rfid

import (
	"fmt"
	"os"
)

type RPISPIDevice struct {
	id     string
	spidev string
	file   *os.File
	ready  bool
	err    error
}

func NewRPISPIDevice(id, spidev string) *RPISPIDevice {
	return &RPISPIDevice{id: id, spidev: spidev}
}

func (d *RPISPIDevice) ID() string       { return d.id }
func (d *RPISPIDevice) Type() string     { return "rfid" }
func (d *RPISPIDevice) IsReady() bool    { return d.ready }
func (d *RPISPIDevice) LastError() error { return d.err }

func (d *RPISPIDevice) Init() error {
	f, err := os.OpenFile(d.spidev, os.O_RDWR, 0)
	if err != nil {
		d.err = fmt.Errorf("SPI open failed: %w", err)
		d.ready = false
		return d.err
	}
	d.file = f
	d.ready = true
	return nil
}

func (d *RPISPIDevice) Shutdown() error {
	if d.file != nil {
		d.file.Close()
		d.file = nil
	}
	d.ready = false
	return nil
}

func (d *RPISPIDevice) Read() (any, error) {
	if !d.ready || d.file == nil {
		return "", fmt.Errorf("SPI device not ready")
	}
	// Simulate SPI read: just return a deterministic cardID
	buf := make([]byte, 4)
	_, err := d.file.Read(buf)
	if err != nil {
		d.err = fmt.Errorf("SPI read failed: %w", err)
		return "", d.err
	}
	// Simulate: if buf[0] is nonzero, return a fake cardID
	if buf[0] != 0 {
		return fmt.Sprintf("CARD%02X%02X%02X%02X", buf[0], buf[1], buf[2], buf[3]), nil
	}
	return "", nil
}
