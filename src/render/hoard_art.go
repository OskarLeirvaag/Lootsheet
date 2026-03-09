package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

func drawHoardPanel(buffer *Buffer, rect Rect, theme *Theme, data *DashboardData, rain *GoldRain) {
	if buffer == nil || theme == nil {
		return
	}

	resolved := resolveDashboardData(data)
	DrawPanel(buffer, rect, theme, Panel{
		Title:       "Party Hoard",
		BorderStyle: &theme.HoardGold,
		TitleStyle:  &theme.HoardGold,
	})

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	lines := append([]string{}, resolved.HoardLines...)
	lines = append(lines, "")
	lines = append(lines, resolved.QuickEntryLines...)

	if content.W >= 60 {
		rainWidth := clampInt((content.W*2)/3, 20, max(20, content.W-20))
		rainRect, textRect := content.SplitVertical(rainWidth, 1)
		if rain != nil {
			rain.Render(buffer, rainRect, theme)
		}
		drawHoardText(buffer, textRect, theme, lines)
		return
	}

	rainHeight := clampInt(content.H/2, 4, max(4, content.H-len(lines)-1))
	rainHeight = min(rainHeight, content.H)
	rainRect, textRect := content.SplitHorizontal(rainHeight, 1)
	if rain != nil {
		rain.Render(buffer, rainRect, theme)
	}
	drawHoardText(buffer, textRect, theme, lines)
}

func drawHoardText(buffer *Buffer, rect Rect, theme *Theme, lines []string) {
	if buffer == nil || theme == nil || rect.Empty() {
		return
	}

	y := rect.Y
	for _, line := range lines {
		if y >= rect.Y+rect.H {
			return
		}
		buffer.WriteString(rect.X, y, hoardLineStyle(theme, line), clipText(line, rect.W))
		y++
	}
}

func hoardLineStyle(theme *Theme, line string) tcell.Style {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "To share now:"):
		return theme.HoardShare
	case strings.HasPrefix(trimmed, "Unsold loot:"):
		return theme.HoardUnsold
	case strings.HasPrefix(trimmed, "e  "), strings.HasPrefix(trimmed, "i  "), strings.HasPrefix(trimmed, "a  "):
		return theme.QuickEntry
	default:
		return theme.Text
	}
}
