package render

// Panel describes a boxed panel and its body lines.
type Panel struct {
	Title string
	Lines []string
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

	if visible.W < 2 || visible.H < 2 {
		return
	}

	right := visible.X + visible.W - 1
	bottom := visible.Y + visible.H - 1

	for x := visible.X + 1; x < right; x++ {
		buffer.Set(x, visible.Y, '─', theme.Border)
		buffer.Set(x, bottom, '─', theme.Border)
	}
	for y := visible.Y + 1; y < bottom; y++ {
		buffer.Set(visible.X, y, '│', theme.Border)
		buffer.Set(right, y, '│', theme.Border)
	}

	buffer.Set(visible.X, visible.Y, '┌', theme.Border)
	buffer.Set(right, visible.Y, '┐', theme.Border)
	buffer.Set(visible.X, bottom, '└', theme.Border)
	buffer.Set(right, bottom, '┘', theme.Border)

	title := clipText(panel.Title, maxInt(0, visible.W-4))
	if title != "" {
		buffer.WriteString(visible.X+1, visible.Y, theme.PanelTitle, " "+title+" ")
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
