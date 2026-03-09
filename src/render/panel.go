package render

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/texture"
)

var (
	brickOnce   sync.Once
	brickCached [][]rune
	leafOnce    sync.Once
	leafCached  [][]rune
)

func brickPattern() [][]rune {
	brickOnce.Do(func() {
		brickCached = loadTexturePattern("bricks.ascii")
	})
	return brickCached
}

func leafPattern() [][]rune {
	leafOnce.Do(func() {
		leafCached = loadTexturePattern("tri-leaves.ascii")
	})
	return leafCached
}

func loadTexturePattern(name string) [][]rune {
	data, err := texture.FS.ReadFile(name)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"), "\n")
	pattern := make([][]rune, len(lines))
	maxW := 0
	for i, line := range lines {
		pattern[i] = []rune(line)
		if len(pattern[i]) > maxW {
			maxW = len(pattern[i])
		}
	}
	for i, row := range pattern {
		if len(row) < maxW {
			padded := make([]rune, maxW)
			copy(padded, row)
			for j := len(row); j < maxW; j++ {
				padded[j] = ' '
			}
			pattern[i] = padded
		}
	}
	return pattern
}

// PanelTexture selects the background fill pattern for a panel.
type PanelTexture int

const (
	PanelTextureBrick PanelTexture = iota
	PanelTextureLeaf
)

// Panel describes a boxed panel and its body lines.
type Panel struct {
	Title       string
	Lines       []string
	BorderStyle *tcell.Style
	TitleStyle  *tcell.Style
	Texture     PanelTexture
}

// DrawPanel renders a simple boxed panel into the frame buffer.
func DrawPanel(buffer *Buffer, rect Rect, theme *Theme, panel Panel) {
	if buffer == nil {
		return
	}

	visible := rect.Intersect(buffer.Bounds())
	if visible.Empty() {
		return
	}

	buffer.FillRect(visible, ' ', theme.Panel)
	var pattern [][]rune
	var texStyle tcell.Style
	switch panel.Texture {
	case PanelTextureLeaf:
		pattern = leafPattern()
		texStyle = theme.Leaf
	default:
		pattern = brickPattern()
		texStyle = theme.Brick
	}
	if pattern != nil {
		buffer.FillTexture(visible, pattern, texStyle)
	}

	borderStyle := theme.Border
	if panel.BorderStyle != nil {
		borderStyle = *panel.BorderStyle
	}
	titleStyle := theme.PanelTitle
	if panel.TitleStyle != nil {
		titleStyle = *panel.TitleStyle
	}

	if visible.W < 2 || visible.H < 2 {
		return
	}

	right := visible.X + visible.W - 1
	bottom := visible.Y + visible.H - 1

	for x := visible.X + 1; x < right; x++ {
		buffer.Set(x, visible.Y, '─', borderStyle)
		buffer.Set(x, bottom, '─', borderStyle)
	}
	for y := visible.Y + 1; y < bottom; y++ {
		buffer.Set(visible.X, y, '│', borderStyle)
		buffer.Set(right, y, '│', borderStyle)
	}

	buffer.Set(visible.X, visible.Y, '╔', borderStyle)
	buffer.Set(right, visible.Y, '╗', borderStyle)
	buffer.Set(visible.X, bottom, '╚', borderStyle)
	buffer.Set(right, bottom, '╝', borderStyle)

	title := clipText(panel.Title, maxInt(0, visible.W-4))
	if title != "" {
		buffer.WriteString(visible.X+1, visible.Y, titleStyle, " "+title+" ")
	}

	content := visible.Inset(1)
	if content.Empty() {
		return
	}

	limit := minInt(len(panel.Lines), content.H)
	for index := 0; index < limit; index++ {
		buffer.WriteString(content.X, content.Y+index, theme.Text, clipText(panel.Lines[index], content.W))
	}
}

func clipText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	return string(runes[:width])
}
