package animation

import "math/rand/v2"

// Drop describes one visible drop for rendering.
type Drop struct {
	X     int
	Y     int
	Trail int
	Spark bool
}

type goldDrop struct {
	x     int
	y     int
	speed int
	trail int
	age   int
}

// GoldRain simulates gold coins dripping down like the Matrix.
type GoldRain struct {
	drops  []goldDrop
	width  int
	height int
	frame  int
}

// NewGoldRain creates a new gold rain animation.
func NewGoldRain() *GoldRain {
	return &GoldRain{}
}

// Resize adjusts the animation area, reinitializing drops when dimensions change.
func (r *GoldRain) Resize(width, height int) {
	if r == nil || (width == r.width && height == r.height) {
		return
	}

	r.width = width
	r.height = height
	r.drops = nil

	h := max(1, height)
	for x := 0; x < width; x += 6 {
		r.drops = append(r.drops, goldDrop{
			x:     x,
			y:     -rand.IntN(h) - 1,
			speed: 1 + rand.IntN(5),
			trail: 4 + rand.IntN(5),
		})
	}
}

// Update advances the animation by one frame.
func (r *GoldRain) Update() {
	if r == nil {
		return
	}

	r.frame++
	for i := range r.drops {
		d := &r.drops[i]
		d.age++
		if d.age >= d.speed {
			d.age = 0
			d.y++
			if d.y-d.trail > r.height {
				d.y = -rand.IntN(max(1, r.height/2)) - 1
				d.speed = 1 + rand.IntN(5)
				d.trail = 4 + rand.IntN(5)
			}
		}
	}
}

// Drops returns the current visible drop states for rendering.
func (r *GoldRain) Drops() []Drop {
	if r == nil {
		return nil
	}

	out := make([]Drop, 0, len(r.drops))
	for _, d := range r.drops {
		if d.y+d.trail < 0 || d.y-d.trail >= r.height {
			continue
		}
		out = append(out, Drop{
			X:     d.x,
			Y:     d.y,
			Trail: d.trail,
			Spark: r.frame%7 == (d.x/6)%7,
		})
	}

	return out
}
