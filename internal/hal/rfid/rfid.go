package rfid

import (
	"smartdisplay-core/internal/hal"
)

type RFIDDevice struct {
	id    string
	ready bool
	card  string
	err   error
}

func (r *RFIDDevice) IsReady() bool {
	return r.ready
}

func (r *RFIDDevice) LastError() error {
	return r.err
}

func NewRFIDDevice(id string) *RFIDDevice {
	return &RFIDDevice{id: id}
}

func (r *RFIDDevice) ID() string {
	return r.id
}

func (r *RFIDDevice) Type() string {
	return "rfid"
}

func (r *RFIDDevice) Init() error {
	r.ready = true
	r.card = ""
	return nil
}

func (r *RFIDDevice) Shutdown() error {
	r.ready = false
	return nil
}

func (r *RFIDDevice) Read() (any, error) {
	if !r.ready {
		return "", nil
	}
	// Simulate deterministic card scan
	if r.card != "" {
		card := r.card
		r.card = "" // clear after read
		return card, nil
	}
	return "", nil
}

// SimulateCard sets a cardID for deterministic test
func (r *RFIDDevice) SimulateCard(cardID string) {
	r.card = cardID
}

var _ hal.InputDevice = (*RFIDDevice)(nil)
