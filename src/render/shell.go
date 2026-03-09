package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type confirmState struct {
	Section Section
	ItemKey string
	Action  ItemActionData
}

type inputState struct {
	Section     Section
	ItemKey     string
	Action      ItemActionData
	Title       string
	Prompt      string
	Value       string
	Placeholder string
	ErrorText   string
	HelpLines   []string
}

type glossaryState struct {
	Title string
	Lines []string
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
	input           *inputState
	compose         *composeState
	glossary        *glossaryState
	rain            *GoldRain
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
		rain:            NewGoldRain(),
	}
	shell.reconcileSelections()

	return shell
}

// TickRain advances the gold rain animation by one frame.
func (s *Shell) TickRain() {
	if s == nil || s.rain == nil {
		return
	}
	s.rain.Update()
}

// Reload swaps the shell snapshot while keeping navigation state intact.
func (s *Shell) Reload(data *ShellData) {
	if s == nil {
		return
	}

	s.Data = resolveShellData(data)
	s.confirm = nil
	s.input = nil
	s.compose = nil
	s.glossary = nil
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

	if s.input != nil {
		return s.handleInputAction(action)
	}
	if s.glossary != nil {
		return s.handleGlossaryAction(action)
	}
	if s.compose != nil {
		switch action {
		case ActionNone, ActionConfirm, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowAccounts, ActionShowJournal, ActionShowQuests, ActionShowLoot,
			ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
			ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionRecognize, ActionSell,
			ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose:
			return handleResult{}
		case ActionQuit:
			s.compose = nil
			return handleResult{Redraw: true}
		case ActionRedraw:
			s.compose = nil
			return handleResult{Reload: true}
		}
	}

	if s.confirm != nil {
		return s.handleConfirmAction(action)
	}

	switch action {
	case ActionNone, ActionConfirm, ActionSubmitCompose:
		return handleResult{}
	case ActionQuit:
		return handleResult{Quit: true}
	case ActionRedraw:
		return handleResult{Reload: true}
	case ActionHelp:
		if s.toggleGlossary() {
			return handleResult{Redraw: true}
		}
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
	case ActionNewExpense:
		if s.openComposeForAction(ActionNewExpense) {
			return handleResult{Redraw: true}
		}
	case ActionNewIncome:
		if s.openComposeForAction(ActionNewIncome) {
			return handleResult{Redraw: true}
		}
	case ActionNewCustom:
		if s.openComposeForAction(ActionNewCustom) {
			return handleResult{Redraw: true}
		}
	case ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionRecognize, ActionSell:
		if s.openAction(action) {
			return handleResult{Redraw: true}
		}
	}

	return handleResult{}
}

// HandleKeyEvent updates shell state for raw key input when the shell needs more
// than semantic action mapping, such as text entry inside a modal.
func (s *Shell) HandleKeyEvent(event *tcell.EventKey, keymap KeyMap) handleResult {
	if s == nil {
		return handleResult{}
	}

	action := keymap.Resolve(event)
	if s.input != nil {
		if result, handled := s.handleInputKeyEvent(event, action); handled {
			return result
		}
	}
	if s.compose != nil {
		if result, handled := s.handleComposeKeyEvent(event, action); handled {
			return result
		}
	}

	return s.HandleAction(action)
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

func (s *Shell) handleInputAction(action Action) handleResult {
	switch action {
	case ActionNone, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowAccounts, ActionShowJournal, ActionShowQuests, ActionShowLoot,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionRecognize, ActionSell,
		ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose:
		return handleResult{}
	case ActionQuit:
		s.input = nil
		return handleResult{Redraw: true}
	case ActionRedraw:
		s.input = nil
		return handleResult{Reload: true}
	case ActionConfirm:
		return handleResult{}
	}

	return handleResult{}
}

func (s *Shell) handleGlossaryAction(action Action) handleResult {
	switch action {
	case ActionQuit, ActionHelp:
		s.glossary = nil
		return handleResult{Redraw: true}
	case ActionRedraw:
		s.glossary = nil
		return handleResult{Reload: true}
	default:
		return handleResult{}
	}
}

func (s *Shell) toggleGlossary() bool {
	if s == nil {
		return false
	}
	if s.glossary != nil {
		s.glossary = nil
		return true
	}

	s.glossary = &glossaryState{
		Title: s.glossaryTitle(),
		Lines: s.glossaryLines(),
	}
	return true
}

func (s *Shell) handleInputKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.input == nil || event == nil {
		return handleResult{}, false
	}

	switch action {
	case ActionQuit, ActionRedraw:
		return s.handleInputAction(action), true
	case ActionConfirm:
		if strings.TrimSpace(s.input.Value) == "" {
			s.input.ErrorText = "Sale amount is required."
			return handleResult{Redraw: true}, true
		}

		command := s.pendingCommand()
		if command == nil {
			return handleResult{Redraw: true}, true
		}

		return handleResult{Command: command}, true
	default:
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(s.input.Value)
		if len(runes) == 0 {
			return handleResult{}, true
		}
		s.input.Value = string(runes[:len(runes)-1])
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		if s.input.Value == "" && s.input.ErrorText == "" {
			return handleResult{}, true
		}
		s.input.Value = ""
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		s.input.Value += string(event.Rune())
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	default:
		return handleResult{}, true
	}
}

// ApplyInputError updates the open input modal with a validation message.
func (s *Shell) ApplyInputError(message string) {
	if s == nil {
		return
	}
	if s.compose != nil {
		s.applyComposeInputError(message)
		return
	}
	if s.input == nil {
		return
	}
	s.input.ErrorText = strings.TrimSpace(message)
}

// CloseModal closes any currently open modal.
func (s *Shell) CloseModal() {
	if s == nil {
		return
	}
	s.confirm = nil
	s.input = nil
	s.compose = nil
	s.glossary = nil
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
	default:
		drawDashboardPanels(buffer, body, theme, &s.Data.Dashboard, s.rain)
	}

	drawStatusLine(buffer, statusRect, theme, s.status)
	drawFooter(buffer, helpRect, theme, s.footerHelpText(keymap))

	if s.compose != nil {
		s.renderCompose(buffer, body, theme)
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
	width := maxSectionTitleWidth() + 2
	line := ""
	for index, section := range orderedSections {
		if index > 0 {
			line += "  "
		}

		label := section.Title()
		if section == s.Section {
			label = "[" + label + "]"
		} else {
			label = " " + label + " "
		}
		line += fmt.Sprintf("%-*s", width, label)
	}

	return line
}

func (s *Shell) drawHeaderHighlights(buffer *Buffer, rect Rect, theme *Theme) {
	if buffer == nil || theme == nil {
		return
	}

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	sectionLabel := "Section: "
	buffer.WriteString(content.X, content.Y, theme.HeaderLabel, sectionLabel)
	buffer.WriteString(content.X+len([]rune(sectionLabel)), content.Y, s.sectionStyle(theme), s.Section.Title())

	if content.H < 2 {
		return
	}

	prefix := "Sections: "
	tabY := content.Y + content.H - 1
	buffer.WriteString(content.X, tabY, theme.HeaderLabel, prefix)
	x := content.X + len([]rune(prefix))
	for index, section := range orderedSections {
		if index > 0 {
			x += buffer.WriteString(x, tabY, theme.Muted, "  ")
		}

		label := section.Title()
		style := theme.TabInactive
		if section == s.Section {
			label = "[" + label + "]"
			style = s.styleForSection(theme, section)
		} else {
			label = " " + label + " "
		}
		width := maxSectionTitleWidth() + 2
		x += buffer.WriteString(x, tabY, style, fmt.Sprintf("%-*s", width, label))
	}
}

func (s *Shell) sectionStyle(theme *Theme) tcell.Style {
	return s.styleForSection(theme, s.Section)
}

func (s *Shell) styleForSection(theme *Theme, section Section) tcell.Style {
	switch section {
	case SectionAccounts:
		return theme.SectionAccounts
	case SectionJournal:
		return theme.SectionJournal
	case SectionQuests:
		return theme.SectionQuests
	case SectionLoot:
		return theme.SectionLoot
	default:
		return theme.SectionDashboard
	}
}

func (s *Shell) footerHelpText(keymap KeyMap) string {
	if s.input != nil {
		return "Enter submit  Backspace delete  Ctrl+U clear  Esc cancel  q cancel"
	}
	if s.glossary != nil {
		return "? close  Esc cancel  q cancel"
	}
	if s.compose != nil {
		return s.composeHelpText()
	}

	if s.confirm != nil {
		return "Enter confirm  Esc cancel  q cancel"
	}

	help := keymap.HelpTextFor(ActionNextSection, ActionShowDashboard)
	if s.Section.scrollable() {
		help = joinHelp(help, keymap.HelpTextFor(ActionMoveDown))
	}

	if labels := s.currentActionLabels(); labels != "" {
		help = joinHelp(help, labels)
	}

	help = joinHelp(help, s.sectionLauncherHelpText())
	help = joinHelp(help, "? terms", "q quit", "Ctrl+L refresh")

	return help
}

func (s *Shell) sectionLauncherHelpText() string {
	switch s.Section {
	case SectionAccounts:
		return "a add"
	case SectionJournal:
		return "e/i entry"
	case SectionQuests:
		return "a add  u edit"
	case SectionLoot:
		return "a add  u edit"
	default:
		return "e/i/a entry"
	}
}

func (s *Shell) glossaryTitle() string {
	return s.Section.Title() + " Terms"
}

func (s *Shell) glossaryLines() []string {
	switch s.Section {
	case SectionAccounts:
		return []string{
			"Assets: what the party owns or is owed.",
			"Liabilities: what the party owes to others.",
			"Equity: the party's accumulated net worth.",
			"Income: rewards, fees, and gains earned by the party.",
			"Expenses: costs like supplies, rations, inns, and repairs.",
			"Active: usable in new entries.",
			"Inactive: kept for history, blocked for new entries.",
			"Postings: journal lines already recorded against an account.",
		}
	case SectionJournal:
		return []string{
			"Debit / Credit: the two sides of every balanced entry.",
			"Posted: final and immutable.",
			"Reversed: corrected by a later entry instead of editing history.",
			"Reversal: a new entry that cancels an earlier posted entry.",
			"Description: the plain-language reason for the entry.",
		}
	case SectionQuests:
		return []string{
			"Off-ledger promise: discussed reward, not earned yet.",
			"Collectible: earned and ready to be paid.",
			"Receivable: reward still owed to the party.",
			"Collected so far: cash already received against the reward.",
			"Write off: accept that the remaining reward will not be collected.",
		}
	case SectionLoot:
		return []string{
			"Held: physically owned, but not yet on the books.",
			"Appraisal: estimated value of a loot item.",
			"Recognize: move an appraisal onto the ledger as inventory and gain.",
			"Recognized value: the inventory basis currently on the books.",
			"Sell: turn recognized loot into cash and record any gain or loss.",
		}
	default:
		return []string{
			"To share now: current Party Cash balance available to split.",
			"Unsold loot: recognized inventory not yet sold for cash.",
			"Off-ledger: tracked in a register, but not yet in the ledger.",
			"Register: operational tracking that may later create entries.",
			"Ledger: the formal bookkeeping record.",
		}
	}
}

func (s *Shell) currentActionLabels() string {
	item := s.currentSelectedItem(s.Section)
	if item == nil || len(item.Actions) == 0 {
		return ""
	}

	labels := make([]string, 0, len(item.Actions))
	for index := range item.Actions {
		action := item.Actions[index]
		if strings.TrimSpace(action.Label) == "" {
			continue
		}
		labels = append(labels, action.Label)
	}

	return strings.Join(labels, "  ")
}

func (s *Shell) renderListSection(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData) {
	view := data
	if listScreenDataEmpty(view) {
		fallback := defaultListScreenData(section)
		view = &fallback
	}
	accent := s.styleForSection(theme, section)

	if rect.W < 48 || rect.H < 10 {
		DrawPanel(buffer, rect, theme, Panel{
			Title: section.Title(),
			Lines: []string{
				"Terminal too small for the interactive list view.",
				"Resize to restore selection, detail, and action panels.",
			},
			BorderStyle: &accent,
			TitleStyle:  &accent,
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
		Title:       "Summary",
		Lines:       summaryLines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
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
		Title:       detailTitle,
		Lines:       detailLines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
	})

	s.renderListPanel(buffer, listRect, theme, section, view, selectedIndex)
}

func (s *Shell) renderListPanel(buffer *Buffer, rect Rect, theme *Theme, section Section, data *ListScreenData, selectedIndex int) {
	items := data.Items
	title := section.Title()
	accent := s.styleForSection(theme, section)
	if len(items) == 0 {
		DrawPanel(buffer, rect, theme, Panel{
			Title:       title,
			Lines:       data.EmptyLines,
			BorderStyle: &accent,
			TitleStyle:  &accent,
		})
		s.viewHeights[section] = 0
		s.scrolls[section] = 0
		return
	}

	DrawPanel(buffer, rect, theme, Panel{
		Title:       title,
		BorderStyle: &accent,
		TitleStyle:  &accent,
	})

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
	DrawPanel(buffer, rect, theme, Panel{
		Title:       title,
		BorderStyle: &accent,
		TitleStyle:  &accent,
	})

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
	accent := s.sectionStyle(theme)

	DrawPanel(buffer, modal, theme, Panel{
		Title:       s.confirm.Action.ConfirmTitle,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
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

	width := 60
	for _, line := range lines {
		if candidate := len([]rune(line)) + 4; candidate > width {
			width = candidate
		}
	}

	width = clampInt(width, 40, minInt(70, rect.W))
	height := clampInt(len(lines)+2, 6, rect.H)
	x := rect.X + maxInt(0, (rect.W-width)/2)
	y := rect.Y + maxInt(0, (rect.H-height)/2)
	modal := Rect{X: x, Y: y, W: width, H: height}
	accent := s.sectionStyle(theme)

	DrawPanel(buffer, modal, theme, Panel{
		Title:       s.input.Title,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
	})
}

func (s *Shell) renderGlossaryModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.glossary == nil || rect.Empty() {
		return
	}

	lines := append([]string{}, s.glossary.Lines...)
	lines = append(lines, "", "? close  Esc/q cancel")

	width := 74
	for _, line := range lines {
		if candidate := len([]rune(line)) + 4; candidate > width {
			width = candidate
		}
	}

	width = clampInt(width, 46, minInt(86, rect.W))
	height := clampInt(len(lines)+2, 8, rect.H)
	x := rect.X + maxInt(0, (rect.W-width)/2)
	y := rect.Y + maxInt(0, (rect.H-height)/2)
	modal := Rect{X: x, Y: y, W: width, H: height}
	accent := s.sectionStyle(theme)

	DrawPanel(buffer, modal, theme, Panel{
		Title:       s.glossary.Title,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
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
	if s.compose != nil {
		command, ok := s.composeCommand()
		if !ok {
			return nil
		}
		return command
	}
	if s.confirm == nil {
		if s.input == nil {
			return nil
		}

		command := &Command{
			ID:      s.input.Action.ID,
			Section: s.input.Section,
			ItemKey: s.input.ItemKey,
		}
		if strings.TrimSpace(s.input.Value) != "" {
			command.Fields = map[string]string{
				"amount": s.input.Value,
			}
		}
		return command
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

func (s *Shell) openAction(trigger Action) bool {
	item := s.currentSelectedItem(s.Section)
	if item == nil || len(item.Actions) == 0 {
		return false
	}

	for index := range item.Actions {
		action := item.Actions[index]
		if action.Trigger != trigger {
			continue
		}

		switch action.Mode {
		case ItemActionModeCompose:
			return s.openComposeFromAction(item.Key, &action)
		case ItemActionModeInput:
			s.input = &inputState{
				Section:     s.Section,
				ItemKey:     item.Key,
				Action:      action,
				Title:       action.InputTitle,
				Prompt:      action.InputPrompt,
				Placeholder: action.Placeholder,
				HelpLines:   append([]string{}, action.InputHelp...),
			}
		default:
			s.confirm = &confirmState{
				Section: s.Section,
				ItemKey: item.Key,
				Action:  action,
			}
		}
		return true
	}

	return false
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

// Navigate switches to the requested section and optionally selects the given item key.
func (s *Shell) Navigate(section Section, selectedKey string) {
	if s == nil {
		return
	}

	if section.scrollable() {
		s.Section = section
		if strings.TrimSpace(selectedKey) != "" {
			s.selectedKeys[section] = strings.TrimSpace(selectedKey)
		}
		s.reconcileSelection(section)
		return
	}

	s.Section = section
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
