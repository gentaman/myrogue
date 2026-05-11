package animation

type Projectile struct {
	StartX, StartY float64
	EndX, EndY     float64
	Frame          int
	TotalFrames    int
	ColorHex       string
	IsFlash        bool
}

type Queue struct {
	Projectiles []Projectile
	Speed       float64
	Accumulator float64
}

func NewQueue() *Queue {
	return &Queue{Speed: 1.0}
}

func (q *Queue) IsPlaying() bool {
	return len(q.Projectiles) > 0
}

func (q *Queue) Add(p Projectile) {
	q.Projectiles = append(q.Projectiles, p)
}

func (q *Queue) Tick() {
	q.TickWithDelta(q.Speed)
}

func (q *Queue) TickWithDelta(delta float64) {
	q.Accumulator += delta
	for q.Accumulator >= 1.0 {
		q.Accumulator -= 1.0
		alive := q.Projectiles[:0]
		for i := range q.Projectiles {
			q.Projectiles[i].Frame++
			if q.Projectiles[i].Frame < q.Projectiles[i].TotalFrames {
				alive = append(alive, q.Projectiles[i])
			}
		}
		q.Projectiles = alive
	}
}

func (q *Queue) SkipAll() {
	q.Projectiles = q.Projectiles[:0]
	q.Accumulator = 0
}
