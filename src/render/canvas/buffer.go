package canvas

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Cell is a single drawable terminal cell.
type Cell struct {
	Rune  rune
	Style tcell.Style
}

// Screen captures the small part of tcell used by the TUI.
type Screen interface {
	Init() error
	Fini()
	Size() (int, int)
	SetStyle(tcell.Style)
	Clear()
	HideCursor()
	Show()
	Sync()
	EnableMouse(...tcell.MouseFlags)
	PollEvent() tcell.Event
	PostEvent(tcell.Event) error
	SetContent(x int, y int, primary rune, combining []rune, style tcell.Style)
}

// ScreenFactory creates a terminal screen implementation.
type ScreenFactory func() (Screen, error)

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

// Get returns the cell at the given position, or a zero Cell if out of bounds.
func (b *Buffer) Get(x int, y int) Cell {
	if b == nil || x < 0 || y < 0 || x >= b.width || y >= b.height {
		return Cell{}
	}
	return b.cells[(y*b.width)+x]
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

// FillTexture tiles a repeating text pattern across a rectangle.
func (b *Buffer) FillTexture(rect Rect, pattern [][]rune, style tcell.Style) {
	if b == nil || len(pattern) == 0 {
		return
	}

	visible := rect.Intersect(b.Bounds())
	if visible.Empty() {
		return
	}

	for y := visible.Y; y < visible.Y+visible.H; y++ {
		row := pattern[y%len(pattern)]
		if len(row) == 0 {
			continue
		}
		for x := visible.X; x < visible.X+visible.W; x++ {
			b.Set(x, y, row[x%len(row)], style)
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

	for y := range b.height {
		for x := range b.width {
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
	for y := range b.height {
		runes := make([]rune, b.width)
		for x := range b.width {
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
