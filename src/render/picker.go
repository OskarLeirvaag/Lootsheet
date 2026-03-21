package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// pickerOption is a single selectable item in the generic picker modal.
type pickerOption struct {
	Value string // inserted into the field / editor
	Label string // display name
	Kind  string // category label (account, person, quest, etc.)
}

// pickerState is the shared modal state for searchable selection lists.
type pickerState struct {
	Title         string
	Options       []pickerOption
	Query         string
	Filtered      []pickerOption
	SelectedIndex int
	Scroll        int
}

func newPicker(title string, options []pickerOption) *pickerState {
	return &pickerState{
		Title:    title,
		Options:  options,
		Filtered: options,
	}
}

func (p *pickerState) refilter() {
	if p.Query == "" {
		p.Filtered = p.Options
	} else {
		query := strings.ToLower(p.Query)
		filtered := make([]pickerOption, 0, len(p.Options))
		for _, opt := range p.Options {
			if strings.Contains(strings.ToLower(opt.Value), query) ||
				strings.Contains(strings.ToLower(opt.Label), query) ||
				strings.Contains(strings.ToLower(opt.Kind), query) {
				filtered = append(filtered, opt)
			}
		}
		p.Filtered = filtered
	}
	if p.SelectedIndex >= len(p.Filtered) {
		p.SelectedIndex = max(0, len(p.Filtered)-1)
	}
	p.Scroll = 0
}

// handlePickerKey processes a key event for the picker. Returns (redraw, closed, selectedValue).
// closed=true means the picker was dismissed (Esc or Enter).
// selectedValue is non-empty only when the user pressed Enter on a valid selection.
func handlePickerKey(p *pickerState, event *tcell.EventKey) (closed bool, selectedValue string) {
	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		return true, ""
	case tcell.KeyEnter:
		if len(p.Filtered) > 0 && p.SelectedIndex < len(p.Filtered) {
			return true, p.Filtered[p.SelectedIndex].Value
		}
		return true, ""
	case tcell.KeyUp:
		if p.SelectedIndex > 0 {
			p.SelectedIndex--
		}
	case tcell.KeyDown:
		if p.SelectedIndex < len(p.Filtered)-1 {
			p.SelectedIndex++
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(p.Query) > 0 {
			runes := []rune(p.Query)
			p.Query = string(runes[:len(runes)-1])
			p.refilter()
		}
	case tcell.KeyCtrlU:
		p.Query = ""
		p.refilter()
	case tcell.KeyRune:
		r := event.Rune()
		if r == 'q' && p.Query == "" {
			return true, ""
		}
		p.Query += string(r)
		p.refilter()
	}
	return false, ""
}

// renderPicker draws the picker modal centered in rect.
func renderPicker(p *pickerState, buffer *Buffer, rect Rect, theme *Theme, accent *SectionStyle) {
	if p == nil || rect.Empty() {
		return
	}

	maxVisible := 10
	viewH := min(maxVisible, max(1, len(p.Filtered)))
	totalH := viewH + 6 // border(2) + search(1) + gap(1) + list + help(1) + gap(1)
	width := clampInt(rect.W/2, 36, 56)

	modalRect := Rect{
		X: rect.X + (rect.W-width)/2,
		Y: rect.Y + (rect.H-totalH)/2,
		W: width,
		H: totalH,
	}
	modalRect = modalRect.Intersect(rect)
	if modalRect.Empty() {
		return
	}

	DrawPanel(buffer, modalRect, theme, Panel{
		Title:       p.Title,
		BorderStyle: &accent.Accent,
		TitleStyle:  &accent.Accent,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(modalRect, buffer.Bounds())
	if content.Empty() {
		return
	}

	y := content.Y
	searchText := "/ " + p.Query + "_"
	buffer.WriteString(content.X, y, theme.Text, clipText(searchText, content.W))
	y++

	// Scroll to keep selection visible.
	if p.SelectedIndex < p.Scroll {
		p.Scroll = p.SelectedIndex
	}
	if p.SelectedIndex >= p.Scroll+viewH {
		p.Scroll = p.SelectedIndex - viewH + 1
	}
	maxScroll := max(0, len(p.Filtered)-viewH)
	p.Scroll = clampInt(p.Scroll, 0, maxScroll)

	if len(p.Filtered) == 0 {
		buffer.WriteString(content.X, y, theme.Muted, clipText("  No matches.", content.W))
	} else {
		prevKind := ""
		for row := 0; row < viewH && p.Scroll+row < len(p.Filtered); row++ {
			idx := p.Scroll + row
			opt := p.Filtered[idx]
			lineRect := Rect{X: content.X, Y: y + row, W: content.W, H: 1}

			// Show category header when kind changes.
			if opt.Kind != prevKind && p.Query == "" {
				prevKind = opt.Kind
				buffer.WriteString(content.X, y+row, theme.Muted, clipText("  "+strings.ToUpper(opt.Kind), content.W))
				viewH++ // header takes a row but doesn't count as an item row
				continue
			}
			prevKind = opt.Kind

			style := theme.Text
			prefix := "  "
			if idx == p.SelectedIndex {
				buffer.FillRect(lineRect, ' ', theme.SelectedRow)
				style = theme.SelectedRow
				prefix = "> "
			}
			tag := fmt.Sprintf("[%s] ", opt.Kind)
			buffer.WriteString(content.X, y+row, theme.Muted, clipText(prefix+tag, content.W))
			labelX := content.X + len([]rune(prefix+tag))
			labelW := content.W - len([]rune(prefix+tag))
			if labelW > 0 {
				buffer.WriteString(labelX, y+row, style, clipText(opt.Label, labelW))
			}
		}
	}

	helpY := content.Y + content.H - 1
	buffer.WriteString(content.X, helpY, theme.Muted, clipText("↑↓ select  Enter pick  Esc cancel", content.W))
}
