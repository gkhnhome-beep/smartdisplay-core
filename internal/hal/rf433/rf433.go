package rf433

import (
	"smartdisplay-core/internal/hal"
)

type RFDevice struct {
	id    string
	code  string
	ready bool
	err   error
}

func (r *RFDevice) IsReady() bool {
	return r.ready
}

func (r *RFDevice) LastError() error {
	return r.err
}

func NewRFDevice(id string) *RFDevice {
	return &RFDevice{id: id}
}

func (r *RFDevice) ID() string {
	return r.id
}

func (r *RFDevice) Type() string {
	return "rf433"
}

func (r *RFDevice) Init() error {
	r.ready = true
	r.code = ""
	return nil
}

func (r *RFDevice) Shutdown() error {
	r.ready = false
	return nil
}

func (r *RFDevice) Read() (any, error) {
	if !r.ready {
		return "", nil
	}
	if r.code != "" {
		code := r.code
		r.code = ""
		return code, nil
	}
	return "", nil
}

// SimulateCode sets a code for deterministic test
func (r *RFDevice) SimulateCode(code string) {
	r.code = code
}

var _ hal.InputDevice = (*RFDevice)(nil)
