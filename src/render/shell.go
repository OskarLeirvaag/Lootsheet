package render

import (
	"fmt"
	"strings"
)

type confirmState struct {
	Section Section
	ItemKey string
	Action  ItemActionData
}

type handleResult struct {
	Command *Command
	Quit    bool
	Redraw  bool
	Reload  bool
}

// Shell renders the interactive multi-screen TUI.
type Shell struct {
	Data            ShellData
	Section         Section
	scrolls         map[Section]int
	selectedKeys    map[Section]string
	selectedIndexes map[Section]int
	viewHeights     map[Section]int
	status          StatusMessage
	confirm         *confirmState
}

// NewShell constructs the interactive TUI shell state.
func NewShell(data *ShellData) *Shell {
	shell := &Shell{
		Data:            resolveShellData(data),
		Section:         SectionDashboard,
		scrolls:         make(map[Section]int),
		selectedKeys:    make(map[Section]string),
		selectedIndexes: make(map[Section]int),
		viewHeights:     make(map[Section]int),
	}
	shell.reconcileSelections()

	return shell
}

// Reload swaps the shell snapshot while keeping navigation state intact.
func (s *Shell) Reload(data *ShellData) {
	if s == nil {
		return
	}

	s.Data = resolveShellData(data)
	s.confirm = nil
	s.reconcileSelections()
}

// SetStatus updates the transient status line.
func (s *Shell) SetStatus(status StatusMessage) {
	if s == nil {
		return
	}

	s.status = status
}

// HandleAction updates shell state for a semantic key action.
func (s *Shell) HandleAction(action Action) handleResult {
	if s == nil {
		return handleResult{}
	}

	if s.confirm != nil {
		return s.handleConfirmAction(action)
	}

	switch action {
	case ActionNone, ActionConfirm:
		return handleResult{}
	case ActionQuit:
		return handleResult{Quit: true}
	case ActionRedraw:
		return handleResult{Reload: true}
	case ActionNextSection:
		s.Section = s.Section.next()
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionPrevSection:
		s.Section = s.Section.previous()
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowDashboard:
		s.Section = SectionDashboard
		return handleResult{Redraw: true}
	case ActionShowAccounts:
		s.Section = SectionAccounts
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowJournal:
		s.Section = SectionJournal
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowQuests:
		s.Section = SectionQuests
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowLoot:
		s.Section = SectionLoot
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionMoveUp:
		if s.moveSelection(-1) {
			return handleResult{Redraw: true}
		}
	case ActionMoveDown:
		if s.moveSelection(1) {
			return handleResult{Redraw: true}
		}
	case ActionPageUp:
		if s.moveSelection(-s.pageSize()) {
			return handleResult{Redraw: true}
		}
	case ActionPageDown:
		if s.moveSelection(s.pageSize()) {
			return handleResult{Redraw: true}
		}
	case ActionMoveTop:
		if s.moveSelectionTo(0) {
			return handleResult{Redraw: true}
		}
	case ActionMoveBottom:
		if s.moveSelectionTo(1 << 30) {
			return handleResult{Redraw: true}
		}
	case ActionPrimary:
		if s.openPrimaryAction() {
			return handleResult{Redraw: true}
		}
	}

	return handleResult{}
}

func (s *Shell) handleConfirmAction(action Action) handleResult {
	switch action {
	case ActionQuit:
		s.confirm = nil
		return handleResult{Redraw: true}
	case ActionConfirm:
		command := s.pendingCommand()
		s.confirm = nil
		if command == nil {
			return handleResult{Redraw: true}
		}
		return handleResult{Command: command}
	case ActionRedraw:
		s.confirm = nil
		return handleResult{Reload: true}
	default:
		return handleResult{}
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

	if bounds.W < 60 || bounds.H < 18 {
		renderCompactShell(buffer, bounds, theme, keymap, s)
		return
	}

	outer := bounds.Inset(1)
	main, footer := outer.SplitHorizontal(maxInt(0, outer.H-2), 0)
	statusRect, helpRect := footer.SplitHorizontal(1, 0)
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

	drawStatusLine(buffer, statusRect, theme, s.status)
	drawFooter(buffer, helpRect, theme, s.footerHelpText(keymap))

	if s.confirm != nil {
		s.renderConfirmModal(buffer, body, theme)
	}
}

func (s *Shell) headerLines() []string {
	source := s.currentHeaderLines()
	lines := make([]string, 0, len(source)+2)
	lines = append(lines, fmt.Sprintf("Section: %s", s.Section.Title()))
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

func (s *Shell) footerHelpText(keymap KeyMap) string {
	if s.confirm != nil {
		return "Enter confirm  Esc cancel  q cancel"
	}

	help := keymap.HelpTextFor(ActionNextSection, ActionShowDashboard, ActionQuit, ActionRedraw)
	if s.Section.scrollable() {
		help = joinHelp(help, keymap.HelpTextFor(ActionMoveDown))
	}

	if label := s.currentPrimaryActionLabel(); label != "" {
		help = joinHelp(help, label)
	}

	return help
}

func (s *Shell) currentPrimaryActionLabel() string {
	item := s.currentSelectedItem(s.Section)
	if item == nil || item.PrimaryAction == nil {
		return ""
	}

	return item.PrimaryAction.Label
}

func (s *Shell) renderListSection(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData) {
	view := data
	if listScreenDataEmpty(view) {
		fallback := defaultListScreenData(section)
		view = &fallback
	}

	if rect.W < 48 || rect.H < 10 {
		DrawPanel(buffer, rect, theme, Panel{
			Title: section.Title(),
			Lines: []string{
				"Terminal too small for the interactive list view.",
				"Resize to restore selection, detail, and action panels.",
			},
		})
		return
	}

	var summaryRect Rect
	var contentRect Rect
	if rect.W >= 100 {
		summaryWidth := clampInt(rect.W/3, 28, 34)
		summaryWidth = minInt(summaryWidth, maxInt(0, rect.W-22))
		summaryRect, contentRect = rect.SplitVertical(summaryWidth, 1)
	} else {
		summaryHeight := clampInt(rect.H/4, 4, 6)
		summaryHeight = minInt(summaryHeight, maxInt(0, rect.H-9))
		summaryRect, contentRect = rect.SplitHorizontal(summaryHeight, 1)
	}

	listHeight := contentRect.H * 3 / 5
	listHeight = clampInt(listHeight, 4, maxInt(4, contentRect.H-5))
	listRect, detailRect := contentRect.SplitHorizontal(listHeight, 1)

	summaryLines := view.SummaryLines
	if len(summaryLines) == 0 {
		summaryLines = []string{"No summary loaded."}
	}
	DrawPanel(buffer, summaryRect, theme, Panel{
		Title: "Summary",
		Lines: summaryLines,
	})

	selectedIndex := s.currentSelectionIndex(section)
	detailTitle := "Detail"
	detailLines := view.EmptyLines
	if len(detailLines) == 0 {
		detailLines = []string{"No rows loaded."}
	}

	if item := s.currentSelectedItem(section); item != nil {
		if item.DetailTitle != "" {
			detailTitle = item.DetailTitle
		}
		if len(item.DetailLines) > 0 {
			detailLines = item.DetailLines
		} else {
			detailLines = []string{"No detail available."}
		}
	}

	DrawPanel(buffer, detailRect, theme, Panel{
		Title: detailTitle,
		Lines: detailLines,
	})

	s.renderListPanel(buffer, listRect, theme, section, view, selectedIndex)
}

func (s *Shell) renderListPanel(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData, selectedIndex int) {
	items := data.Items
	title := section.Title()
	if len(items) == 0 {
		DrawPanel(buffer, rect, theme, Panel{
			Title: title,
			Lines: data.EmptyLines,
		})
		s.viewHeights[section] = 0
		s.scrolls[section] = 0
		return
	}

	DrawPanel(buffer, rect, theme, Panel{Title: title})

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		s.viewHeights[section] = 0
		return
	}

	s.viewHeights[section] = content.H

	scroll := s.scrolls[section]
	if selectedIndex < scroll {
		scroll = selectedIndex
	}
	if selectedIndex >= scroll+content.H {
		scroll = selectedIndex - content.H + 1
	}

	maxScroll := maxInt(0, len(items)-content.H)
	scroll = clampInt(scroll, 0, maxScroll)
	s.scrolls[section] = scroll

	end := minInt(len(items), scroll+content.H)
	title = fmt.Sprintf("%s %d-%d/%d", section.Title(), scroll+1, end, len(items))
	DrawPanel(buffer, rect, theme, Panel{Title: title})

	for row := 0; row < content.H && scroll+row < len(items); row++ {
		index := scroll + row
		item := items[index]
		lineRect := Rect{X: content.X, Y: content.Y + row, W: content.W, H: 1}

		style := theme.Text
		prefix := "  "
		if index == selectedIndex {
			buffer.FillRect(lineRect, ' ', theme.SelectedRow)
			style = theme.SelectedRow
			prefix = "> "
		}

		line := prefix + item.Row
		buffer.WriteString(content.X, content.Y+row, style, clipText(line, content.W))
	}
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

	width := 56
	for _, line := range lines {
		if candidate := len([]rune(line)) + 4; candidate > width {
			width = candidate
		}
	}

	width = clampInt(width, 36, minInt(64, rect.W))
	height := clampInt(len(lines)+2, 5, rect.H)
	x := rect.X + maxInt(0, (rect.W-width)/2)
	y := rect.Y + maxInt(0, (rect.H-height)/2)
	modal := Rect{X: x, Y: y, W: width, H: height}

	DrawPanel(buffer, modal, theme, Panel{
		Title: s.confirm.Action.ConfirmTitle,
		Lines: lines,
	})
}

func (s *Shell) pageSize() int {
	size := s.viewHeights[s.Section]
	if size <= 1 {
		return 8
	}

	return size - 1
}

func (s *Shell) pendingCommand() *Command {
	if s.confirm == nil {
		return nil
	}

	command := &Command{
		ID:      s.confirm.Action.ID,
		Section: s.confirm.Section,
		ItemKey: s.confirm.ItemKey,
	}

	return command
}

func (s *Shell) moveSelection(delta int) bool {
	if !s.Section.scrollable() || delta == 0 {
		return false
	}

	data := s.listDataForSection(s.Section)
	if data == nil || len(data.Items) == 0 {
		return false
	}

	current := s.currentSelectionIndex(s.Section)
	next := clampInt(current+delta, 0, len(data.Items)-1)
	if next == current {
		return false
	}

	s.setSelection(s.Section, next)
	return true
}

func (s *Shell) moveSelectionTo(index int) bool {
	if !s.Section.scrollable() {
		return false
	}

	data := s.listDataForSection(s.Section)
	if data == nil || len(data.Items) == 0 {
		return false
	}

	if index > len(data.Items)-1 {
		index = len(data.Items) - 1
	}
	index = clampInt(index, 0, len(data.Items)-1)

	if s.currentSelectionIndex(s.Section) == index {
		return false
	}

	s.setSelection(s.Section, index)
	return true
}

func (s *Shell) openPrimaryAction() bool {
	item := s.currentSelectedItem(s.Section)
	if item == nil || item.PrimaryAction == nil {
		return false
	}

	s.confirm = &confirmState{
		Section: s.Section,
		ItemKey: item.Key,
		Action:  *item.PrimaryAction,
	}
	return true
}

func (s *Shell) currentSelectedItem(section Section) *ListItemData {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return nil
	}

	index := s.currentSelectionIndex(section)
	if index < 0 || index >= len(data.Items) {
		return nil
	}

	return &data.Items[index]
}

func (s *Shell) currentSelectionIndex(section Section) int {
	if !section.scrollable() {
		return -1
	}

	s.reconcileSelection(section)

	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		return -1
	}

	index := s.selectedIndexes[section]
	if index < 0 || index >= len(data.Items) {
		index = 0
		s.setSelection(section, index)
	}

	return index
}

func (s *Shell) setSelection(section Section, index int) {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		delete(s.selectedIndexes, section)
		delete(s.selectedKeys, section)
		return
	}

	index = clampInt(index, 0, len(data.Items)-1)
	s.selectedIndexes[section] = index
	s.selectedKeys[section] = data.Items[index].Key
}

func (s *Shell) reconcileSelections() {
	for _, section := range orderedSections {
		if !section.scrollable() {
			continue
		}
		s.reconcileSelection(section)
	}
}

func (s *Shell) reconcileSelection(section Section) {
	data := s.listDataForSection(section)
	if data == nil || len(data.Items) == 0 {
		delete(s.selectedIndexes, section)
		delete(s.selectedKeys, section)
		s.scrolls[section] = 0
		return
	}

	if key := s.selectedKeys[section]; key != "" {
		if index := listItemIndexByKey(data.Items, key); index >= 0 {
			s.selectedIndexes[section] = index
			return
		}
	}

	index := s.selectedIndexes[section]
	if index < 0 || index >= len(data.Items) {
		index = 0
	}

	s.setSelection(section, index)
}

func (s *Shell) listDataForSection(section Section) *ListScreenData {
	switch section {
	case SectionAccounts:
		return &s.Data.Accounts
	case SectionJournal:
		return &s.Data.Journal
	case SectionQuests:
		return &s.Data.Quests
	case SectionLoot:
		return &s.Data.Loot
	default:
		return nil
	}
}

func listItemIndexByKey(items []ListItemData, key string) int {
	for index, item := range items {
		if item.Key == key {
			return index
		}
	}

	return -1
}

func renderCompactShell(buffer *Buffer, bounds Rect, theme *Theme, keymap KeyMap, shell *Shell) {
	panel := bounds.Inset(1)
	section := SectionDashboard
	actionLabel := ""
	if shell != nil {
		section = shell.Section
		actionLabel = shell.currentPrimaryActionLabel()
	}

	lines := []string{
		"Terminal too small for the full interactive shell.",
		"Current section: " + section.Title(),
		"Resize and the boxed layout will redraw cleanly.",
		keymap.HelpTextFor(ActionNextSection, ActionShowDashboard, ActionQuit, ActionRedraw),
	}
	if actionLabel != "" {
		lines = append(lines, actionLabel)
	}

	DrawPanel(buffer, panel, theme, Panel{
		Title: "LootSheet",
		Lines: lines,
	})
}

func panelContentRect(rect Rect, bounds Rect) Rect {
	visible := rect.Intersect(bounds)
	if visible.Empty() {
		return Rect{}
	}

	return visible.Inset(1)
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

func joinHelp(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		filtered = append(filtered, part)
	}

	return strings.Join(filtered, "  ")
}
