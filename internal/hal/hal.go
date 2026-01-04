package hal

type Device interface {
	ID() string
	Type() string
	Init() error
	Shutdown() error
	IsReady() bool
	LastError() error
}

type InputDevice interface {
	Device
	Read() (any, error)
}

type OutputDevice interface {
	Device
	Write(value any) error
}
