package countdown

type Countdown struct {
	durationSeconds  int
	remainingSeconds int
	active           bool
}

func New(durationSeconds int) *Countdown {
	return &Countdown{
		durationSeconds:  durationSeconds,
		remainingSeconds: durationSeconds,
		active:           false,
	}
}

func (c *Countdown) Start() {
	c.active = true
	c.remainingSeconds = c.durationSeconds
}

func (c *Countdown) Stop() {
	c.active = false
}

func (c *Countdown) Reset() {
	c.remainingSeconds = c.durationSeconds
}

func (c *Countdown) Tick() bool {
	if c.active && c.remainingSeconds > 0 {
		c.remainingSeconds--
		if c.remainingSeconds == 0 {
			return true
		}
	}
	return false
}
