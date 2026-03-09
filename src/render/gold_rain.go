package render

import (
	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/texture/animation"
)

// GoldRain wraps the animation state for the hoard panel.
type GoldRain struct {
	rain *animation.GoldRain
}

// NewGoldRain creates a new gold rain animation.
func NewGoldRain() *GoldRain {
	return &GoldRain{rain: animation.NewGoldRain()}
}

// Update advances the animation by one frame.
func (r *GoldRain) Update() {
	if r == nil {
		return
	}
	r.rain.Update()
}

// Render draws the current rain state into the buffer.
func (r *GoldRain) Render(buffer *Buffer, rect Rect, theme *Theme) {
	if r == nil || buffer == nil || rect.Empty() || theme == nil {
		return
	}

	r.rain.Resize(rect.W, rect.H)

	for _, d := range r.rain.Drops() {
		for t := 0; t <= d.Trail; t++ {
			py := d.Y - t
			if py < 0 || py >= rect.H {
				continue
			}

			sx := rect.X + d.X
			sy := rect.Y + py
			if sx >= rect.X+rect.W {
				continue
			}

			if t == 0 {
				ch := 'O'
				if d.Spark {
					ch = '*'
				}
				buffer.Set(sx, sy, ch, theme.HoardGold)
			} else {
				buffer.Set(sx, sy, '.', trailStyle(t, d.Trail, theme))
			}
		}
	}
}

func trailStyle(position int, length int, theme *Theme) tcell.Style {
	// Interpolate from bright (ink) to dim (near-background).
	const r0, g0, b0 = 244, 239, 228 // ink / bright end
	const r1, g1, b1 = 50, 55, 65    // dim end

	denom := max(1, length-1)
	frac := float64(position-1) / float64(denom)

	cr := int32(float64(r0) + float64(r1-r0)*frac)
	cg := int32(float64(g0) + float64(g1-g0)*frac)
	cb := int32(float64(b0) + float64(b1-b0)*frac)

	return theme.Text.Foreground(tcell.NewRGBColor(cr, cg, cb))
}
