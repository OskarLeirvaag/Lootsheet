package render

import "fmt"

// Shell renders the read-only multi-screen TUI.
type Shell struct {
	Data    ShellData
	Section Section
	scrolls map[Section]int
}

// NewShell constructs the read-only TUI shell state.
func NewShell(data *ShellData) *Shell {
	return &Shell{
		Data:    resolveShellData(data),
		Section: SectionDashboard,
		scrolls: make(map[Section]int),
	}
}

// Reload swaps the shell snapshot while keeping navigation state intact.
func (s *Shell) Reload(data *ShellData) {
	if s == nil {
		return
	}

	s.Data = resolveShellData(data)
}

// HandleAction updates shell state for a semantic key action.
func (s *Shell) HandleAction(action Action) bool {
	if s == nil {
		return false
	}

	switch action {
	case ActionNextSection:
		s.Section = s.Section.next()
		return true
	case ActionPrevSection:
		s.Section = s.Section.previous()
		return true
	case ActionShowDashboard:
		s.Section = SectionDashboard
		return true
	case ActionShowAccounts:
		s.Section = SectionAccounts
		return true
	case ActionShowJournal:
		s.Section = SectionJournal
		return true
	case ActionShowQuests:
		s.Section = SectionQuests
		return true
	case ActionShowLoot:
		s.Section = SectionLoot
		return true
	case ActionScrollUp:
		s.adjustScroll(-1)
		return true
	case ActionScrollDown:
		s.adjustScroll(1)
		return true
	case ActionPageUp:
		s.adjustScroll(-8)
		return true
	case ActionPageDown:
		s.adjustScroll(8)
		return true
	case ActionScrollTop:
		s.setScroll(0)
		return true
	case ActionScrollBottom:
		s.setScroll(1 << 30)
		return true
	default:
		return false
	}
}

// Render draws the full shell for the current section.
func (s *Shell) Render(buffer *Buffer, theme *Theme, keymap KeyMap) {
	if buffer == nil {
		return
	}

	bounds := buffer.Bounds()
	if bounds.Empty() {
		return
	}

	if bounds.W < 56 || bounds.H < 14 {
		renderCompactShell(buffer, bounds, theme, keymap, s)
		return
	}

	outer := bounds.Inset(1)
	main, footer := outer.SplitHorizontal(maxInt(0, outer.H-1), 0)
	header, body := main.SplitHorizontal(6, 1)

	DrawPanel(buffer, header, theme, Panel{
		Title: "LootSheet TUI",
		Lines: s.headerLines(),
	})

	switch s.Section {
	case SectionAccounts:
		s.renderListSection(buffer, body, theme, SectionAccounts, &s.Data.Accounts)
	case SectionJournal:
		s.renderListSection(buffer, body, theme, SectionJournal, &s.Data.Journal)
	case SectionQuests:
		s.renderListSection(buffer, body, theme, SectionQuests, &s.Data.Quests)
	case SectionLoot:
		s.renderListSection(buffer, body, theme, SectionLoot, &s.Data.Loot)
	default:
		drawDashboardPanels(buffer, body, theme, &s.Data.Dashboard)
	}

	drawFooter(buffer, footer, theme, keymap.HelpTextFor(s.helpActions()...))
}

func (s *Shell) headerLines() []string {
	source := s.currentHeaderLines()
	lines := make([]string, 0, len(source)+2)
	lines = append(lines,
		fmt.Sprintf("Section: %s", s.Section.Title()),
	)
	lines = append(lines, source...)
	lines = append(lines, "Sections: "+s.tabsLine())

	return lines
}

func (s *Shell) currentHeaderLines() []string {
	switch s.Section {
	case SectionAccounts:
		return append([]string{}, s.Data.Accounts.HeaderLines...)
	case SectionJournal:
		return append([]string{}, s.Data.Journal.HeaderLines...)
	case SectionQuests:
		return append([]string{}, s.Data.Quests.HeaderLines...)
	case SectionLoot:
		return append([]string{}, s.Data.Loot.HeaderLines...)
	default:
		return append([]string{}, resolveDashboardData(&s.Data.Dashboard).HeaderLines...)
	}
}

func (s *Shell) tabsLine() string {
	line := ""
	for index, section := range orderedSections {
		if index > 0 {
			line += "  "
		}

		label := section.Title()
		if section == s.Section {
			label = "[" + label + "]"
		}
		line += label
	}

	return line
}

func (s *Shell) helpActions() []Action {
	actions := []Action{
		ActionNextSection,
		ActionShowDashboard,
		ActionQuit,
		ActionRedraw,
	}

	if s.Section.scrollable() {
		actions = append(actions, ActionScrollDown)
	}

	return actions
}

func (s *Shell) renderListSection(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData) {
	view := data
	if listScreenDataEmpty(view) {
		fallback := defaultListScreenData(section)
		view = &fallback
	}

	if rect.W < 40 || rect.H < 8 {
		DrawPanel(buffer, rect, theme, Panel{
			Title: section.Title(),
			Lines: []string{
				"Terminal too small for this screen.",
				"Resize to browse the boxed list view.",
			},
		})
		return
	}

	var summaryRect Rect
	var listRect Rect
	if rect.W >= 88 {
		summaryWidth := clampInt(rect.W/3, 26, 34)
		summaryWidth = minInt(summaryWidth, maxInt(0, rect.W-18))
		summaryRect, listRect = rect.SplitVertical(summaryWidth, 1)
	} else {
		summaryHeight := clampInt(rect.H/3, 5, 8)
		summaryHeight = minInt(summaryHeight, maxInt(0, rect.H-4))
		summaryRect, listRect = rect.SplitHorizontal(summaryHeight, 1)
	}

	summaryLines := view.SummaryLines
	if len(summaryLines) == 0 {
		summaryLines = []string{"No summary loaded."}
	}
	DrawPanel(buffer, summaryRect, theme, Panel{
		Title: "Summary",
		Lines: summaryLines,
	})

	rowLines := view.RowLines
	empty := len(rowLines) == 0
	if empty {
		rowLines = view.EmptyLines
		if len(rowLines) == 0 {
			rowLines = []string{"No rows loaded."}
		}
	}

	contentHeight := maxInt(0, listRect.H-2)
	scroll := 0
	title := section.Title()
	visible := rowLines
	if !empty && contentHeight > 0 {
		maxScroll := maxInt(0, len(rowLines)-contentHeight)
		scroll = clampInt(s.scrolls[section], 0, maxScroll)
		s.scrolls[section] = scroll

		end := minInt(len(rowLines), scroll+contentHeight)
		visible = rowLines[scroll:end]
		title = fmt.Sprintf("%s %d-%d/%d", section.Title(), scroll+1, end, len(rowLines))
	} else {
		s.scrolls[section] = 0
	}

	DrawPanel(buffer, listRect, theme, Panel{
		Title: title,
		Lines: visible,
	})
}

func (s *Shell) adjustScroll(delta int) {
	if !s.Section.scrollable() || delta == 0 {
		return
	}

	next := s.scrolls[s.Section] + delta
	if next < 0 {
		next = 0
	}

	s.scrolls[s.Section] = next
}

func (s *Shell) setScroll(value int) {
	if !s.Section.scrollable() {
		return
	}

	if value < 0 {
		value = 0
	}

	s.scrolls[s.Section] = value
}

func renderCompactShell(buffer *Buffer, bounds Rect, theme *Theme, keymap KeyMap, shell *Shell) {
	panel := bounds.Inset(1)
	section := SectionDashboard
	if shell != nil {
		section = shell.Section
	}

	DrawPanel(buffer, panel, theme, Panel{
		Title: "LootSheet",
		Lines: []string{
			"Terminal too small for the full TUI shell.",
			"Current section: " + section.Title(),
			"Resize and the boxed layout will redraw cleanly.",
			keymap.HelpTextFor(ActionNextSection, ActionQuit, ActionRedraw),
		},
	})
}
