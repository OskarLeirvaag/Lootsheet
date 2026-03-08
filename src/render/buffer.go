package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Cell is a single drawable terminal cell.
type Cell struct {
	Rune  rune
	Style tcell.Style
}

// Buffer stores a full frame before it is flushed to the screen.
type Buffer struct {
	width  int
	height int
	cells  []Cell
}

// NewBuffer creates a buffer prefilled with spaces using the provided style.
func NewBuffer(width int, height int, fillStyle tcell.Style) *Buffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	buffer := &Buffer{
		width:  width,
		height: height,
		cells:  make([]Cell, width*height),
	}

	fill := Cell{Rune: ' ', Style: fillStyle}
	for index := range buffer.cells {
		buffer.cells[index] = fill
	}

	return buffer
}

// Bounds returns the full buffer rectangle.
func (b *Buffer) Bounds() Rect {
	if b == nil {
		return Rect{}
	}
	return Rect{W: b.width, H: b.height}
}

// Set updates a single cell if it is inside the buffer.
func (b *Buffer) Set(x int, y int, value rune, style tcell.Style) {
	if b == nil || x < 0 || y < 0 || x >= b.width || y >= b.height {
		return
	}
	if value == 0 {
		value = ' '
	}
	b.cells[(y*b.width)+x] = Cell{Rune: value, Style: style}
}

// FillRect fills a visible sub-rectangle.
func (b *Buffer) FillRect(rect Rect, value rune, style tcell.Style) {
	if b == nil {
		return
	}

	visible := rect.Intersect(b.Bounds())
	if visible.Empty() {
		return
	}

	for y := visible.Y; y < visible.Y+visible.H; y++ {
		for x := visible.X; x < visible.X+visible.W; x++ {
			b.Set(x, y, value, style)
		}
	}
}

// WriteString draws a single-line string clipped to the visible buffer width.
func (b *Buffer) WriteString(x int, y int, style tcell.Style, text string) int {
	if b == nil || y < 0 || y >= b.height {
		return 0
	}

	drawn := 0
	column := x
	for _, value := range text {
		if column >= b.width {
			break
		}
		if column >= 0 {
			b.Set(column, y, value, style)
			drawn++
		}
		column++
	}

	return drawn
}

// Flush writes the entire buffer to the target screen.
func (b *Buffer) Flush(screen Screen) {
	if b == nil || screen == nil {
		return
	}

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			cell := b.cells[(y*b.width)+x]
			screen.SetContent(x, y, cell.Rune, nil, cell.Style)
		}
	}
}

// PlainText returns a trimmed text snapshot useful for deterministic tests.
func (b *Buffer) PlainText() string {
	if b == nil {
		return ""
	}

	lines := make([]string, 0, b.height)
	for y := 0; y < b.height; y++ {
		runes := make([]rune, b.width)
		for x := 0; x < b.width; x++ {
			value := b.cells[(y*b.width)+x].Rune
			if value == 0 {
				value = ' '
			}
			runes[x] = value
		}
		lines = append(lines, strings.TrimRight(string(runes), " "))
	}

	return strings.Join(lines, "\n")
}
