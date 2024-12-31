package animations

type Animation interface {
	Update()
	Frame() int
}

type SingleAnimation struct {
	First        int
	Last         int
	Step         int     // How many indices do we move per frame
	SpeedInTps   float32 // How many ticks before next frame
	frameCounter float32
	frame        int
}

func (a *SingleAnimation) Update() {
	a.frameCounter -= 1.0
	if a.frameCounter < 0.0 {
		a.frameCounter = a.SpeedInTps
		a.frame += a.Step
		if a.frame > a.Last {
			// loop back to the beginning
			a.frame = a.First
		}
	}
}

func (a *SingleAnimation) Frame() int {
	return a.frame
}

func NewAnimation(first, last, step int, speed float32) *SingleAnimation {
	return &SingleAnimation{
		first,
		last,
		step,
		speed,
		speed,
		first,
	}
}

type AnimationStep struct {
	Animation SingleAnimation
	Delay     int
}

type ComposeAnimation struct {
	animations          []AnimationStep
	currentIdxAnimation int
	counter             int
}

func (c *ComposeAnimation) Update() {
	c.counter += 1
	if c.counter > c.animations[c.currentIdxAnimation].Delay {
		c.counter = 0
		c.currentIdxAnimation += 1
		c.currentIdxAnimation %= len(c.animations)
	}
	c.animations[c.currentIdxAnimation].Animation.Update()
}

func (c *ComposeAnimation) Frame() int {
	return c.animations[c.currentIdxAnimation].Animation.Frame()
}

func NewComposeAnimation(animations []AnimationStep) *ComposeAnimation {
	return &ComposeAnimation{
		animations,
		0,
		0,
	}
}
