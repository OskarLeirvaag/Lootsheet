package render

import (
	"fmt"
	"strings"
)

// Render draws the full shell for the current section.
func (s *Shell) Render(buffer *Buffer, theme *Theme, keymap KeyMap) {
	if buffer == nil {
		return
	}

	bounds := buffer.Bounds()
	if bounds.Empty() {
		return
	}

	if bounds.W < 60 || bounds.H < 18 {
		renderCompactShell(buffer, bounds, theme, keymap, s)
		return
	}

	outer := bounds.Inset(1)
	main, footer := outer.SplitHorizontal(max(0, outer.H-2), 0)
	statusRect, helpRect := footer.SplitHorizontal(1, 0)
	header, body := main.SplitHorizontal(6, 1)
	headerAccent := s.sectionStyle(theme)

	DrawPanel(buffer, header, theme, Panel{
		Title:       "LootSheet TUI",
		Lines:       s.headerLines(),
		BorderStyle: &headerAccent,
		TitleStyle:  &headerAccent,
	})
	s.drawHeaderHighlights(buffer, header, theme)

	switch s.Section {
	case SectionAccounts:
		s.renderListSection(buffer, body, theme, SectionAccounts, &s.Data.Accounts)
	case SectionJournal:
		s.renderListSection(buffer, body, theme, SectionJournal, &s.Data.Journal)
	case SectionQuests:
		s.renderListSection(buffer, body, theme, SectionQuests, &s.Data.Quests)
	case SectionLoot:
		s.renderListSection(buffer, body, theme, SectionLoot, &s.Data.Loot)
	case SectionAssets:
		s.renderListSection(buffer, body, theme, SectionAssets, &s.Data.Assets)
	case SectionCodex:
		s.renderListSection(buffer, body, theme, SectionCodex, &s.Data.Codex)
	case SectionNotes:
		s.renderNotesSection(buffer, body, theme)
	case SectionSettings:
		s.renderSettingsSection(buffer, body, theme)
	default:
		drawDashboardPanels(buffer, body, theme, &s.Data.Dashboard, s.rain)
	}

	drawStatusLine(buffer, statusRect, theme, s.status)
	drawFooter(buffer, helpRect, theme, s.footerHelpText(keymap))

	if s.editor != nil {
		s.renderEditor(buffer, body, theme)
	} else if s.compose != nil {
		s.renderCompose(buffer, body, theme)
	}
	if s.codexPicker != nil {
		s.renderCodexPickerModal(buffer, body, theme)
	}
	if s.search != nil {
		s.renderSearchModal(buffer, body, theme)
	}
	if s.input != nil {
		s.renderInputModal(buffer, body, theme)
	}
	if s.confirm != nil {
		s.renderConfirmModal(buffer, body, theme)
	}
	if s.glossary != nil {
		s.renderGlossaryModal(buffer, body, theme)
	}
	if s.quitConfirm {
		s.renderQuitConfirmModal(buffer, body, theme)
	}
	if s.disconnected {
		s.renderDisconnectModal(buffer, body, theme)
	}
}

func (s *Shell) renderSettingsSection(buffer *Buffer, rect Rect, theme *Theme) {
	if rect.H < 3 {
		s.renderListSection(buffer, rect, theme, s.activeSettingsSection(), s.listDataForSection(s.activeSettingsSection()))
		return
	}

	// Draw tab bar in the first row.
	tabRect := Rect{X: rect.X, Y: rect.Y, W: rect.W, H: 1}
	buffer.FillRect(tabRect, ' ', theme.Muted)

	x := tabRect.X + 1
	for i, tab := range settingsTabs {
		if i > 0 {
			x += buffer.WriteString(x, tabRect.Y, theme.Muted, "  ")
		}
		label := tab.Title()
		if i == s.settingsTab {
			label = "[" + label + "]"
			x += buffer.WriteString(x, tabRect.Y, theme.SectionSettings, label)
		} else {
			label = " " + label + " "
			x += buffer.WriteString(x, tabRect.Y, theme.TabInactive, label)
		}
	}

	// Render the active tab's list section in the remaining space.
	bodyRect := Rect{X: rect.X, Y: rect.Y + 1, W: rect.W, H: rect.H - 1}
	active := s.activeSettingsSection()
	s.renderListSection(buffer, bodyRect, theme, active, s.listDataForSection(active))
}

func (s *Shell) renderListSection(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData) {
	view := data
	if listScreenDataEmpty(view) {
		fallback := defaultListScreenData(section)
		view = &fallback
	}
	ss := sectionStyleFor(section, theme)

	if rect.W < 48 || rect.H < 10 {
		DrawPanel(buffer, rect, theme, ss.Panel(section.Title(), []string{
			"Terminal too small for the interactive list view.",
			"Resize to restore selection, detail, and action panels.",
		}))
		return
	}

	var summaryRect Rect
	var contentRect Rect
	if rect.W >= 100 {
		summaryWidth := clampInt(rect.W/3, 28, 34)
		summaryWidth = min(summaryWidth, max(0, rect.W-22))
		summaryRect, contentRect = rect.SplitVertical(summaryWidth, 1)
	} else {
		summaryHeight := clampInt(rect.H/4, 4, 6)
		summaryHeight = min(summaryHeight, max(0, rect.H-9))
		summaryRect, contentRect = rect.SplitHorizontal(summaryHeight, 1)
	}

	listHeight := contentRect.H * 3 / 5
	listHeight = clampInt(listHeight, 4, max(4, contentRect.H-5))
	listRect, detailRect := contentRect.SplitHorizontal(listHeight, 1)

	summaryLines := view.SummaryLines
	if len(summaryLines) == 0 {
		summaryLines = []string{"No summary loaded."}
	}
	DrawPanel(buffer, summaryRect, theme, ss.Panel("Summary", summaryLines))

	selectedIndex := s.currentSelectionIndex(section)
	detailTitle := "Detail"
	detailLines := view.EmptyLines
	if len(detailLines) == 0 {
		detailLines = []string{"No rows loaded."}
	}

	var detailBody string
	if item := s.currentSelectedItem(section); item != nil {
		if item.DetailTitle != "" {
			detailTitle = item.DetailTitle
		}
		if len(item.DetailLines) > 0 {
			detailLines = item.DetailLines
		} else {
			detailLines = []string{"No detail available."}
		}
		detailBody = item.DetailBody
	}

	if detailBody != "" {
		detailContent := panelContentRect(detailRect, buffer.Bounds())
		mdWidth := detailContent.W
		if mdWidth <= 0 {
			mdWidth = 40
		}
		mdLines := parseMarkdownLines(detailBody, mdWidth, theme)
		DrawStyledPanel(buffer, detailRect, theme, detailTitle, detailLines, mdLines, ss.Accent, ss.Accent)
	} else {
		detailContent := panelContentRect(detailRect, buffer.Bounds())
		wrappedLines := wrapDetailLines(detailLines, detailContent.W)
		DrawPanel(buffer, detailRect, theme, ss.Panel(detailTitle, wrappedLines))
	}

	s.renderListPanel(buffer, listRect, theme, section, view, selectedIndex)
}

// --- Notes section: narrow title list + full markdown body ---

func (s *Shell) renderNotesSection(buffer *Buffer, rect Rect, theme *Theme) {
	ss := sectionStyleFor(SectionNotes, theme)
	data := &s.Data.Notes
	selectedIndex := s.currentSelectionIndex(SectionNotes)

	// Layout: narrow left column (list + refs) and large body on right.
	leftW := clampInt(rect.W/4, 20, 36)
	leftRect, bodyRect := rect.SplitVertical(leftW, 1)

	// Split left column: title list (75%) and references (25%).
	refsH := clampInt(leftRect.H/4, 4, 10)
	titleListRect, refsRect := leftRect.SplitHorizontal(leftRect.H-refsH, 0)

	// --- Title list panel ---
	DrawPanel(buffer, titleListRect, theme, Panel{
		Title:         "Notes",
		BorderStyle:   &ss.Accent,
		TitleStyle:    &ss.Accent,
		Texture:       ss.Texture,
		Borders:       ss.Borders,
		ScatterGlyphs: ss.ScatterGlyphs,
		ScatterStyle:  ss.ScatterStyle,
	})

	listContent := panelContentRect(titleListRect, buffer.Bounds())
	if !listContent.Empty() && len(data.Items) > 0 {
		scroll := min(s.scrolls[SectionNotes], selectedIndex)
		if selectedIndex >= scroll+listContent.H {
			scroll = selectedIndex - listContent.H + 1
		}
		scroll = clampInt(scroll, 0, max(0, len(data.Items)-listContent.H))
		s.scrolls[SectionNotes] = scroll

		for row := 0; row < listContent.H && scroll+row < len(data.Items); row++ {
			index := scroll + row
			item := data.Items[index]
			lineRect := Rect{X: listContent.X, Y: listContent.Y + row, W: listContent.W, H: 1}
			style := theme.Text
			prefix := " "
			if index == selectedIndex {
				buffer.FillRect(lineRect, ' ', theme.SelectedRow)
				style = theme.SelectedRow
				prefix = ">"
			}
			buffer.WriteString(listContent.X, listContent.Y+row, style, clipText(prefix+item.DetailTitle, listContent.W))
		}
	} else if !listContent.Empty() {
		for i, line := range data.EmptyLines {
			if i >= listContent.H {
				break
			}
			buffer.WriteString(listContent.X, listContent.Y+i, theme.Muted, clipText(line, listContent.W))
		}
	}

	// --- References panel ---
	var refLines []string
	if item := s.currentSelectedItem(SectionNotes); item != nil {
		for _, line := range item.DetailLines {
			if strings.HasPrefix(line, "  @") || strings.HasPrefix(line, "Linked from:") || strings.HasPrefix(line, "References:") {
				refLines = append(refLines, line)
			}
		}
	}
	if len(refLines) == 0 {
		refLines = []string{"No references."}
	}
	DrawPanel(buffer, refsRect, theme, Panel{
		Title:       "References",
		Lines:       refLines,
		BorderStyle: &ss.Accent,
		TitleStyle:  &ss.Accent,
		Texture:     PanelTextureNone,
	})

	// --- Body panel (markdown render of selected note) ---
	bodyTitle := "Note"
	var metaLines []string
	var bodyText string

	if item := s.currentSelectedItem(SectionNotes); item != nil {
		bodyTitle = item.DetailTitle
		// Only pass non-reference detail lines as meta.
		for _, line := range item.DetailLines {
			if !strings.HasPrefix(line, "  @") && !strings.HasPrefix(line, "References:") && !strings.HasPrefix(line, "Linked from:") {
				metaLines = append(metaLines, line)
			}
		}
		bodyText = item.DetailBody
	}

	if bodyText != "" {
		bodyContent := panelContentRect(bodyRect, buffer.Bounds())
		mdWidth := bodyContent.W
		if mdWidth <= 0 {
			mdWidth = 40
		}
		mdLines := parseMarkdownLines(bodyText, mdWidth, theme)
		DrawStyledPanel(buffer, bodyRect, theme, bodyTitle, metaLines, mdLines, ss.Accent, ss.Accent)
	} else {
		DrawPanel(buffer, bodyRect, theme, Panel{
			Title:       bodyTitle,
			Lines:       metaLines,
			BorderStyle: &ss.Accent,
			TitleStyle:  &ss.Accent,
			Texture:     PanelTextureNone,
		})
	}
}

// wrapDetailLines wraps each plain-text detail line to fit within width.
func wrapDetailLines(lines []string, width int) []string {
	if width <= 0 {
		return lines
	}
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapPlainText(line, width)...)
	}
	return wrapped
}

func (s *Shell) renderListPanel(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData, selectedIndex int) {
	items := data.Items
	title := section.Title()
	ss := sectionStyleFor(section, theme)
	if len(items) == 0 {
		DrawPanel(buffer, rect, theme, ss.Panel(title, data.EmptyLines))
		s.viewHeights[section] = 0
		s.scrolls[section] = 0
		return
	}

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		DrawPanel(buffer, rect, theme, ss.Panel(title, nil))
		s.viewHeights[section] = 0
		return
	}

	// Reserve one row for the column header if present.
	listContent := content
	hasHeader := data.ListHeaderRow != "" && content.H > 2
	if hasHeader {
		listContent = Rect{X: content.X, Y: content.Y + 1, W: content.W, H: content.H - 1}
	}

	s.viewHeights[section] = listContent.H

	scroll := min(selectedIndex, s.scrolls[section])
	if selectedIndex >= scroll+listContent.H {
		scroll = selectedIndex - listContent.H + 1
	}

	maxScroll := max(0, len(items)-listContent.H)
	scroll = clampInt(scroll, 0, maxScroll)
	s.scrolls[section] = scroll

	end := min(len(items), scroll+listContent.H)
	title = fmt.Sprintf("%s %d-%d/%d", section.Title(), scroll+1, end, len(items))
	DrawPanel(buffer, rect, theme, ss.Panel(title, nil))

	if hasHeader {
		buffer.WriteString(content.X, content.Y, theme.Muted, clipText("  "+data.ListHeaderRow, content.W))
	}

	for row := 0; row < listContent.H && scroll+row < len(items); row++ {
		index := scroll + row
		item := items[index]
		lineRect := Rect{X: listContent.X, Y: listContent.Y + row, W: listContent.W, H: 1}

		style := theme.Text
		prefix := "  "
		if index == selectedIndex {
			buffer.FillRect(lineRect, ' ', theme.SelectedRow)
			style = theme.SelectedRow
			prefix = "> "
		}

		line := prefix + item.Row
		buffer.WriteString(listContent.X, listContent.Y+row, style, clipText(line, listContent.W))
	}
}

func renderCompactShell(buffer *Buffer, bounds Rect, theme *Theme, keymap KeyMap, shell *Shell) {
	panel := bounds.Inset(1)
	section := SectionDashboard
	actionLabels := ""
	if shell != nil {
		section = shell.Section
		actionLabels = shell.currentActionLabels()
	}

	lines := []string{
		"Terminal too small for the full interactive shell.",
		"Current section: " + section.Title(),
		"Resize and the boxed layout will redraw cleanly.",
		keymap.HelpTextFor(ActionNextSection, ActionShowDashboard, ActionQuit, ActionRedraw),
	}
	if actionLabels != "" {
		lines = append(lines, actionLabels)
	}

	DrawPanel(buffer, panel, theme, Panel{
		Title: "LootSheet",
		Lines: lines,
	})
}

func (s *Shell) renderConfirmModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.confirm == nil || rect.Empty() {
		return
	}

	lines := append([]string{}, s.confirm.Action.ConfirmLines...)
	if len(lines) == 0 {
		lines = []string{"Confirm this action."}
	}
	lines = append(lines, "", "Enter confirm  Esc/q cancel")

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalBounds(rect, lines, 56, 36, 64, 5), theme, Panel{
		Title:       s.confirm.Action.ConfirmTitle,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})
}

func (s *Shell) renderInputModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.input == nil || rect.Empty() {
		return
	}

	lines := make([]string, 0, len(s.input.HelpLines)+5)
	if prompt := strings.TrimSpace(s.input.Prompt); prompt != "" {
		lines = append(lines, prompt+": "+s.input.displayValue())
	} else {
		lines = append(lines, s.input.displayValue())
	}

	if strings.TrimSpace(s.input.ErrorText) != "" {
		lines = append(lines, "Error: "+s.input.ErrorText)
	}
	if len(s.input.HelpLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.input.HelpLines...)
	}
	lines = append(lines, "", "Enter submit  Esc/q cancel")

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalBounds(rect, lines, 60, 40, 70, 6), theme, Panel{
		Title:       s.input.Title,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})
}

func (s *Shell) renderGlossaryModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.glossary == nil || rect.Empty() {
		return
	}

	lines := append([]string{}, s.glossary.Lines...)
	lines = append(lines, "", "? close  Esc/q cancel")

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalBounds(rect, lines, 74, 46, 86, 8), theme, Panel{
		Title:       s.glossary.Title,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})
}

func drawStatusLine(buffer *Buffer, rect Rect, theme *Theme, status StatusMessage) {
	style := theme.StatusInfo
	switch status.Level {
	case StatusInfo:
		style = theme.StatusInfo
	case StatusError:
		style = theme.StatusError
	case StatusSuccess:
		style = theme.StatusOK
	}

	if status.Empty() {
		style = theme.Footer
	}

	visible := rect.Intersect(buffer.Bounds())
	if visible.Empty() {
		return
	}

	buffer.FillRect(visible, ' ', style)
	if status.Empty() {
		return
	}

	buffer.WriteString(visible.X, visible.Y, style, clipText(status.Text, visible.W))
}

func (s *Shell) renderQuitConfirmModal(buffer *Buffer, rect Rect, theme *Theme) {
	if rect.Empty() {
		return
	}

	lines := []string{
		"",
		"Are you sure you want",
		"to quit LootSheet?",
		"",
		"Enter quit  Esc cancel",
	}

	DrawPanel(buffer, modalBounds(rect, lines, 30, 26, 36, 5), theme, Panel{
		Title:   "Quit?",
		Lines:   lines,
		Texture: PanelTextureNone,
	})
}

func (s *Shell) renderDisconnectModal(buffer *Buffer, rect Rect, theme *Theme) {
	if rect.Empty() {
		return
	}

	lines := []string{
		"",
		"The server has disconnected.",
		"",
		"Press any key to exit.",
	}

	DrawPanel(buffer, modalBounds(rect, lines, 36, 30, 44, 6), theme, Panel{
		Title:       "Server Disconnected",
		Lines:       lines,
		BorderStyle: &theme.StatusError,
		TitleStyle:  &theme.StatusError,
		Texture:     PanelTextureNone,
	})
}

func (s *inputState) displayValue() string {
	if s == nil {
		return ""
	}
	if strings.TrimSpace(s.Value) != "" {
		return s.Value
	}
	if strings.TrimSpace(s.Placeholder) != "" {
		return "[" + s.Placeholder + "]"
	}
	return ""
}
