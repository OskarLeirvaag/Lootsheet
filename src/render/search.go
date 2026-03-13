package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const maxSearchResults = 50

// SearchHandler performs server-side search for a given section and query.
// Return nil items to fall back to client-side filtering.
type SearchHandler func(section Section, query string) ([]ListItemData, error)

type searchResult struct {
	Section Section
	ItemKey string
	Row     string
}

type searchState struct {
	Query         string
	FilterIndex   int // 0=All, 1..N=searchableSections[i-1]
	Results       []searchResult
	SelectedIndex int
	Scroll        int
}

func (s *Shell) openSearch() bool {
	if s == nil {
		return false
	}
	s.search = &searchState{}
	s.computeSearchResults()
	return true
}

func (s *Shell) computeSearchResults() {
	if s.search == nil {
		return
	}

	query := strings.ToLower(strings.TrimSpace(s.search.Query))
	var sections []Section
	if s.search.FilterIndex == 0 {
		sections = searchableSections
	} else {
		sections = []Section{searchableSections[s.search.FilterIndex-1]}
	}

	results := make([]searchResult, 0, maxSearchResults)
	for _, section := range sections {
		if len(results) >= maxSearchResults {
			break
		}

		// Try server-side search when a handler is available and there is a query.
		// On error or nil items, fall back silently to client-side filtering so
		// search remains usable even if the backend is temporarily unavailable.
		if s.searchHandler != nil && query != "" {
			if items, err := s.searchHandler(section, query); err == nil && items != nil {
				for i := range items {
					if len(results) >= maxSearchResults {
						break
					}
					results = append(results, searchResult{
						Section: section,
						ItemKey: items[i].Key,
						Row:     items[i].Row,
					})
				}
				continue
			}
		}

		// Fall back to client-side filtering.
		data := s.listDataForSection(section)
		if data == nil {
			continue
		}
		for i := range data.Items {
			if len(results) >= maxSearchResults {
				break
			}
			item := &data.Items[i]
			if query == "" || matchesSearch(item, query) {
				results = append(results, searchResult{
					Section: section,
					ItemKey: item.Key,
					Row:     item.Row,
				})
			}
		}
	}

	s.search.Results = results
	if s.search.SelectedIndex >= len(results) {
		s.search.SelectedIndex = max(0, len(results)-1)
	}
	if s.search.Scroll > s.search.SelectedIndex {
		s.search.Scroll = s.search.SelectedIndex
	}
}

func matchesSearch(item *ListItemData, query string) bool {
	if strings.Contains(strings.ToLower(item.Row), query) {
		return true
	}
	if strings.Contains(strings.ToLower(item.DetailTitle), query) {
		return true
	}
	for _, line := range item.DetailLines {
		if strings.Contains(strings.ToLower(line), query) {
			return true
		}
	}
	return false
}

func (s *Shell) handleSearchKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.search == nil || event == nil {
		return handleResult{}, false
	}

	switch action {
	case ActionQuit:
		s.search = nil
		return handleResult{Redraw: true}, true
	case ActionRedraw:
		s.search = nil
		return handleResult{Reload: true}, true
	default:
	}

	switch event.Key() {
	case tcell.KeyEsc:
		s.search = nil
		return handleResult{Redraw: true}, true
	case tcell.KeyEnter:
		if len(s.search.Results) > 0 && s.search.SelectedIndex < len(s.search.Results) {
			result := s.search.Results[s.search.SelectedIndex]
			s.search = nil
			s.Navigate(result.Section, result.ItemKey)
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyUp:
		if s.search.SelectedIndex > 0 {
			s.search.SelectedIndex--
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyDown:
		if s.search.SelectedIndex < len(s.search.Results)-1 {
			s.search.SelectedIndex++
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyLeft:
		s.search.FilterIndex = (s.search.FilterIndex + len(searchableSections) + 1 - 1) % (len(searchableSections) + 1)
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyRight:
		s.search.FilterIndex = (s.search.FilterIndex + 1) % (len(searchableSections) + 1)
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyTab:
		s.search.FilterIndex = (s.search.FilterIndex + 1) % (len(searchableSections) + 1)
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyBacktab:
		s.search.FilterIndex = (s.search.FilterIndex + len(searchableSections) + 1 - 1) % (len(searchableSections) + 1)
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(s.search.Query)
		if len(runes) == 0 {
			return handleResult{}, true
		}
		s.search.Query = string(runes[:len(runes)-1])
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		if s.search.Query == "" {
			return handleResult{}, true
		}
		s.search.Query = ""
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		s.search.Query += string(event.Rune())
		s.computeSearchResults()
		return handleResult{Redraw: true}, true
	default:
		return handleResult{}, true
	}
}

func (s *Shell) renderSearchModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.search == nil || rect.Empty() {
		return
	}

	// Modal dimensions.
	modalW := clampInt(rect.W*3/4, 40, min(80, rect.W))
	modalH := clampInt(rect.H*3/4, 10, min(30, rect.H))
	x := rect.X + max(0, (rect.W-modalW)/2)
	y := rect.Y + max(0, (rect.H-modalH)/2)
	modalRect := Rect{X: x, Y: y, W: modalW, H: modalH}

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalRect, theme, Panel{
		Title:       "Search",
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(modalRect, buffer.Bounds())
	if content.Empty() {
		return
	}

	row := content.Y

	// Filter tabs line.
	if row < content.Y+content.H {
		var tabs strings.Builder
		filterCount := len(searchableSections) + 1
		for i := range filterCount {
			if i > 0 {
				tabs.WriteString("  ")
			}
			label := "All"
			if i > 0 {
				label = searchableSections[i-1].Title()
			}
			if i == s.search.FilterIndex {
				label = "[" + label + "]"
			} else {
				label = " " + label + " "
			}
			tabs.WriteString(label)
		}
		buffer.WriteString(content.X, row, theme.Muted, clipText(tabs.String(), content.W))
		row++
	}

	// Input line.
	if row < content.Y+content.H {
		input := fmt.Sprintf("Search: %s_", s.search.Query)
		buffer.WriteString(content.X, row, theme.Text, clipText(input, content.W))
		row++
	}

	// Blank separator.
	if row < content.Y+content.H {
		row++
	}

	// Results list.
	listH := content.Y + content.H - row
	if listH <= 0 {
		return
	}

	results := s.search.Results
	sel := s.search.SelectedIndex

	// Adjust scroll.
	scroll := min(s.search.Scroll, sel)
	if sel >= scroll+listH {
		scroll = sel - listH + 1
	}
	maxScroll := max(0, len(results)-listH)
	scroll = clampInt(scroll, 0, maxScroll)
	s.search.Scroll = scroll

	if len(results) == 0 {
		if s.search.Query != "" {
			buffer.WriteString(content.X, row, theme.Muted, clipText("No results.", content.W))
		}
		return
	}

	for i := 0; i < listH && scroll+i < len(results); i++ {
		r := results[scroll+i]
		style := theme.Text
		prefix := "  "
		if scroll+i == sel {
			lineRect := Rect{X: content.X, Y: row + i, W: content.W, H: 1}
			buffer.FillRect(lineRect, ' ', theme.SelectedRow)
			style = theme.SelectedRow
			prefix = "> "
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, r.Section.Title(), r.Row)
		buffer.WriteString(content.X, row+i, style, clipText(line, content.W))
	}
}
