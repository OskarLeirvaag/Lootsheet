package canvas

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

	topHeight = ClampInt(topHeight, 0, r.H)
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

	leftWidth = ClampInt(leftWidth, 0, r.W)
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

// ClampInt restricts a value to the given range.
func ClampInt(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// ClipText truncates text to fit within the given width.
func ClipText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	return string(runes[:width])
}

// PanelContentRect returns the inner content area of a panel.
func PanelContentRect(rect Rect, bounds Rect) Rect {
	visible := rect.Intersect(bounds)
	if visible.Empty() {
		return Rect{}
	}

	return visible.Inset(1)
}

// ModalBounds calculates centered modal dimensions within a parent rectangle.
func ModalBounds(parent Rect, lines []string, defaultW, minW, maxW, minH int) Rect {
	width := defaultW
	for _, line := range lines {
		if candidate := len([]rune(line)) + 4; candidate > width {
			width = candidate
		}
	}
	width = ClampInt(width, minW, min(maxW, parent.W))
	height := ClampInt(len(lines)+2, minH, parent.H)
	x := parent.X + max(0, (parent.W-width)/2)
	y := parent.Y + max(0, (parent.H-height)/2)
	return Rect{X: x, Y: y, W: width, H: height}
}
