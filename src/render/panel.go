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
	PanelTextureNone
)

// BorderSet defines the characters used to draw panel borders.
type BorderSet struct {
	Horizontal rune
	Vertical   rune
	TopLeft    rune
	TopRight   rune
	BotLeft    rune
	BotRight   rune
	// RuneBar, if set, fills horizontal edges with repeating runes instead
	// of a single Horizontal character.
	RuneBar []rune
}

var (
	defaultBorders = BorderSet{
		Horizontal: '─',
		Vertical:   '│',
		TopLeft:    '╔',
		TopRight:   '╗',
		BotLeft:    '╚',
		BotRight:   '╝',
	}
	runicBorders = BorderSet{
		Horizontal: '─',
		Vertical:   '│',
		TopLeft:    'ᛟ',
		TopRight:   'ᛟ',
		BotLeft:    'ᛟ',
		BotRight:   'ᛟ',
		RuneBar:    []rune{'ᚠ', 'ᚢ', 'ᚦ', 'ᚨ'},
	}
)

// Panel describes a boxed panel and its body lines.
type Panel struct {
	Title       string
	Lines       []string
	BorderStyle *tcell.Style
	TitleStyle  *tcell.Style
	Texture     PanelTexture
	Borders     *BorderSet
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
	switch panel.Texture {
	case PanelTextureLeaf:
		if pattern := leafPattern(); pattern != nil {
			buffer.FillTexture(visible, pattern, theme.Brick)
		}
		scatterRunes(buffer, visible, theme.Leaf)
	case PanelTextureNone:
		// clean background, no texture
	default:
		if pattern := brickPattern(); pattern != nil {
			buffer.FillTexture(visible, pattern, theme.Brick)
		}
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

	bs := defaultBorders
	if panel.Borders != nil {
		bs = *panel.Borders
	}

	right := visible.X + visible.W - 1
	bottom := visible.Y + visible.H - 1

	spread := len(bs.RuneBar)
	for x := visible.X + 1; x < right; x++ {
		ch := bs.Horizontal
		distLeft := x - visible.X - 1
		distRight := right - x - 1
		if distLeft < spread {
			ch = bs.RuneBar[distLeft]
		} else if distRight < spread {
			ch = bs.RuneBar[distRight]
		}
		buffer.Set(x, visible.Y, ch, borderStyle)
		buffer.Set(x, bottom, ch, borderStyle)
	}
	for y := visible.Y + 1; y < bottom; y++ {
		buffer.Set(visible.X, y, bs.Vertical, borderStyle)
		buffer.Set(right, y, bs.Vertical, borderStyle)
	}

	buffer.Set(visible.X, visible.Y, bs.TopLeft, borderStyle)
	buffer.Set(right, visible.Y, bs.TopRight, borderStyle)
	buffer.Set(visible.X, bottom, bs.BotLeft, borderStyle)
	buffer.Set(right, bottom, bs.BotRight, borderStyle)

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

var scatterGlyphs = []rune{'ᚠ', 'ᚢ', 'ᚦ', 'ᚨ', 'ᚱ', 'ᚲ', 'ᚷ', 'ᚹ', 'ᚺ', 'ᚾ', 'ᛃ', 'ᛈ', 'ᛇ', 'ᛉ', 'ᛊ', 'ᛏ', 'ᛒ', 'ᛚ', 'ᛗ', 'ᛞ', 'ᛟ'}

// scatterRunes places a few random Elder Futhark runes into the texture area.
// Uses a simple position-seeded hash so the result is deterministic.
func scatterRunes(buffer *Buffer, rect Rect, style tcell.Style) {
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		for x := rect.X; x < rect.X+rect.W; x++ {
			if buffer.Get(x, y).Rune != ' ' {
				continue
			}
			h := uint32(x*7919 + y*104729 + 277)
			h ^= h >> 16
			h *= 0x45d9f3b
			h ^= h >> 16
			if h%97 < 3 {
				ch := scatterGlyphs[h%uint32(len(scatterGlyphs))]
				buffer.Set(x, y, ch, style)
			}
		}
	}
}
