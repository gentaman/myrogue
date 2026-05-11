package clock

type Clock interface {
	Delta() float64
	Scale() float64
	SetScale(s float64)
	Paused() bool
	SetPaused(p bool)
	StepFrame()
	Tick()
}

type ScaledClock struct {
	scale       float64
	paused      bool
	delta       float64
	stepPending bool
}

func New() *ScaledClock {
	return &ScaledClock{scale: 1.0}
}

func (c *ScaledClock) Delta() float64 {
	return c.delta
}

func (c *ScaledClock) Scale() float64 {
	return c.scale
}

func (c *ScaledClock) SetScale(s float64) {
	if s < 0 {
		s = 0
	}
	if s > 4.0 {
		s = 4.0
	}
	c.scale = s
}

func (c *ScaledClock) Paused() bool {
	return c.paused
}

func (c *ScaledClock) SetPaused(p bool) {
	c.paused = p
}

func (c *ScaledClock) StepFrame() {
	c.stepPending = true
}

func (c *ScaledClock) Tick() {
	if c.stepPending {
		c.delta = 1.0
		c.stepPending = false
		return
	}
	if c.paused {
		c.delta = 0
		return
	}
	c.delta = c.scale
}
