package render

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/texture"
)

var (
	brickOnce    sync.Once
	brickCached  [][]rune
)

func brickPattern() [][]rune {
	brickOnce.Do(func() {
		data, err := texture.FS.ReadFile("bricks.ascii")
		if err != nil {
			return
		}
		lines := strings.Split(strings.TrimRight(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"), "\n")
		brickCached = make([][]rune, len(lines))
		for i, line := range lines {
			brickCached[i] = []rune(line)
		}
	})
	return brickCached
}

// Panel describes a boxed panel and its body lines.
type Panel struct {
	Title       string
	Lines       []string
	BorderStyle *tcell.Style
	TitleStyle  *tcell.Style
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
	if pattern := brickPattern(); pattern != nil {
		buffer.FillTexture(visible, pattern, theme.Brick)
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
