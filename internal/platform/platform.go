package platform

type Platform interface {
	Init() error
	Name() string
	IsEmbedded() bool
}

type LinuxPlatform struct{}

func (p *LinuxPlatform) Init() error      { return nil }
func (p *LinuxPlatform) Name() string     { return "linux" }
func (p *LinuxPlatform) IsEmbedded() bool { return false }

// RaspberryPiPlatform implements Platform

type RaspberryPiPlatform struct{}

func (p *RaspberryPiPlatform) Init() error      { return nil }
func (p *RaspberryPiPlatform) Name() string     { return "raspberrypi" }
func (p *RaspberryPiPlatform) IsEmbedded() bool { return true }

// DetectPlatform returns the correct Platform implementation
func DetectPlatform() Platform {
	// Deterministic: use GOARCH/GOOS for now
	// In real code, would check /proc/cpuinfo or similar
	if isRaspberryPi() {
		return &RaspberryPiPlatform{}
	}
	return &LinuxPlatform{}
}

func isRaspberryPi() bool {
	// Deterministic stub: always false for now
	return false
}
