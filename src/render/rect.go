package render

// Rect is a terminal-space rectangle measured in cells.
type Rect struct {
	X int
	Y int
	W int
	H int
}

// Empty reports whether the rectangle has no visible area.
func (r Rect) Empty() bool {
	return r.W <= 0 || r.H <= 0
}

// Inset shrinks the rectangle on all sides.
func (r Rect) Inset(padding int) Rect {
	if padding <= 0 {
		return r
	}

	width := r.W - (padding * 2)
	height := r.H - (padding * 2)
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return Rect{
		X: r.X + padding,
		Y: r.Y + padding,
		W: width,
		H: height,
	}
}

// Intersect returns the visible overlap of the two rectangles.
func (r Rect) Intersect(other Rect) Rect {
	left := max(r.X, other.X)
	top := max(r.Y, other.Y)
	right := min(r.X+r.W, other.X+other.W)
	bottom := min(r.Y+r.H, other.Y+other.H)

	width := right - left
	height := bottom - top
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return Rect{
		X: left,
		Y: top,
		W: width,
		H: height,
	}
}

// SplitHorizontal splits a rectangle into top and bottom regions with a gap.
func (r Rect) SplitHorizontal(topHeight int, gap int) (Rect, Rect) {
	if r.Empty() {
		return Rect{}, Rect{}
	}

	if gap < 0 {
		gap = 0
	}

	topHeight = clampInt(topHeight, 0, r.H)
	used := topHeight
	if used < r.H {
		used += min(gap, r.H-used)
	}

	return Rect{
			X: r.X,
			Y: r.Y,
			W: r.W,
			H: topHeight,
		}, Rect{
			X: r.X,
			Y: r.Y + used,
			W: r.W,
			H: r.H - used,
		}
}

// SplitVertical splits a rectangle into left and right regions with a gap.
func (r Rect) SplitVertical(leftWidth int, gap int) (Rect, Rect) {
	if r.Empty() {
		return Rect{}, Rect{}
	}

	if gap < 0 {
		gap = 0
	}

	leftWidth = clampInt(leftWidth, 0, r.W)
	used := leftWidth
	if used < r.W {
		used += min(gap, r.W-used)
	}

	return Rect{
			X: r.X,
			Y: r.Y,
			W: leftWidth,
			H: r.H,
		}, Rect{
			X: r.X + used,
			Y: r.Y,
			W: r.W - used,
			H: r.H,
		}
}

func clampInt(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
