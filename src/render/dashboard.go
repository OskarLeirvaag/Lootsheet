package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Dashboard renders the first read-only shell.
type Dashboard struct {
	Data DashboardData
}

// Render draws the current dashboard frame.
func (d *Dashboard) Render(buffer *Buffer, theme *Theme, keymap KeyMap) {
	if buffer == nil {
		return
	}

	bounds := buffer.Bounds()
	if bounds.Empty() {
		return
	}

	if bounds.W < 52 || bounds.H < 16 {
		renderCompactDashboard(buffer, bounds, theme, keymap)
		return
	}

	data := resolveDashboardData(&d.Data)
	outer := bounds.Inset(1)
	main, footer := outer.SplitHorizontal(maxInt(0, outer.H-1), 0)
	header, body := main.SplitHorizontal(4, 1)

	DrawPanel(buffer, header, theme, Panel{
		Title:       "LootSheet Dashboard",
		Lines:       data.HeaderLines,
		BorderStyle: &theme.SectionDashboard,
		TitleStyle:  &theme.SectionDashboard,
	})

	drawDashboardPanels(buffer, body, theme, &data)
	drawFooter(buffer, footer, theme, keymap.HelpTextFor(ActionQuit, ActionRedraw))
}

func renderCompactDashboard(buffer *Buffer, bounds Rect, theme *Theme, keymap KeyMap) {
	panel := bounds.Inset(1)
	DrawPanel(buffer, panel, theme, Panel{
		Title: "LootSheet",
		Lines: []string{
			"Terminal too small for the full dashboard.",
			"Resize the terminal and the boxed layout will redraw cleanly.",
			keymap.HelpTextFor(ActionQuit, ActionRedraw),
		},
	})
}

func drawDashboardPanels(buffer *Buffer, body Rect, theme *Theme, data *DashboardData) {
	if body.Empty() {
		return
	}

	resolved := resolveDashboardData(data)

	if body.W < 52 || body.H < 10 {
		DrawPanel(buffer, body, theme, Panel{
			Title:       "Dashboard",
			Lines:       []string{
				"Terminal too small for the full dashboard panels.",
				"Resize to restore the boxed layout.",
			},
			BorderStyle: &theme.SectionDashboard,
			TitleStyle:  &theme.SectionDashboard,
		})
		return
	}

	heroHeight := clampInt(body.H/2, 7, 9)
	heroHeight = minInt(heroHeight, maxInt(0, body.H-8))
	hero, lower := body.SplitHorizontal(heroHeight, 1)
	if lower.H < 4 {
		hero = body
		lower = Rect{}
	}

	drawHoardPanel(buffer, hero, theme, &resolved)

	if lower.Empty() {
		return
	}

	topHeight := clampInt(lower.H/2, 3, 4)
	top, bottom := lower.SplitHorizontal(topHeight, 1)
	if bottom.Empty() {
		top = lower
	}
	topWidth := maxInt(16, (top.W-2)/3)
	accounts, topRest := top.SplitVertical(topWidth, 1)
	journal, ledger := topRest.SplitVertical(topWidth, 1)

	var quests Rect
	var loot Rect
	if !bottom.Empty() {
		bottomWidth := maxInt(20, (bottom.W-1)/2)
		quests, loot = bottom.SplitVertical(bottomWidth, 1)
	}

	DrawPanel(buffer, accounts, theme, Panel{
		Title:       "Accounts",
		Lines:       resolved.AccountsLines,
		BorderStyle: &theme.SectionAccounts,
		TitleStyle:  &theme.SectionAccounts,
	})

	DrawPanel(buffer, journal, theme, Panel{
		Title:       "Journal",
		Lines:       resolved.JournalLines,
		BorderStyle: &theme.SectionJournal,
		TitleStyle:  &theme.SectionJournal,
	})

	DrawPanel(buffer, ledger, theme, Panel{
		Title:       "Ledger Snapshot",
		Lines:       resolved.LedgerLines,
		BorderStyle: &theme.SectionDashboard,
		TitleStyle:  &theme.SectionDashboard,
	})

	if !quests.Empty() {
		DrawPanel(buffer, quests, theme, Panel{
			Title:       "Quest Register",
			Lines:       resolved.QuestLines,
			BorderStyle: &theme.SectionQuests,
			TitleStyle:  &theme.SectionQuests,
		})
	}

	if !loot.Empty() {
		DrawPanel(buffer, loot, theme, Panel{
			Title:       "Loot Register",
			Lines:       resolved.LootLines,
			BorderStyle: &theme.SectionLoot,
			TitleStyle:  &theme.SectionLoot,
		})
	}
}

type hoardSegment struct {
	Text  string
	Style tcell.Style
}

func drawHoardPanel(buffer *Buffer, rect Rect, theme *Theme, data *DashboardData) {
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

	art := [][]hoardSegment{
		{
			{Text: "                         __                         ", Style: theme.HoardBag},
		},
		{
			{Text: "                       .-''-.                      ", Style: theme.HoardBag},
		},
		{
			{Text: "                      / .--. \\\\                     ", Style: theme.HoardBag},
		},
		{
			{Text: "                     / /    \\\\ \\\\                    ", Style: theme.HoardBag},
		},
		{
			{Text: "                     | | $$ | |                    ", Style: theme.HoardBag},
		},
		{
			{Text: "               <>    | |____| |    <>              ", Style: theme.HoardGem},
		},
		{
			{Text: "         o o o o o  /________\\\\  o o o o o         ", Style: theme.HoardGold},
		},
		{
			{Text: "      o <> o o o o <> o o o o <> o o o o <> o      ", Style: theme.HoardGold},
		},
		{
			{Text: "   o o o o <> o o o o o <> o o o o o <> o o o o    ", Style: theme.HoardGold},
		},
		{
			{Text: " o <> o o o o o <> o o o o o <> o o o o <> o o o o ", Style: theme.HoardGold},
		},
		{
			{Text: "^^^^^^^", Style: theme.HoardGold},
			{Text: " <> <> <> <>", Style: theme.HoardGem},
			{Text: " ^^^^^ ", Style: theme.HoardGold},
			{Text: "<> <> <>", Style: theme.HoardGem},
			{Text: " ^^^^^^^", Style: theme.HoardGold},
		},
	}

	lines := append([]string{}, resolved.HoardLines...)
	lines = append(lines, "")
	lines = append(lines, resolved.QuickEntryLines...)

	if content.W >= 60 {
		artWidth := clampInt((content.W*2)/3, 36, maxInt(36, content.W-20))
		artRect, textRect := content.SplitVertical(artWidth, 1)
		drawHoardArt(buffer, artRect, art)
		drawHoardText(buffer, textRect, theme, lines)
		return
	}

	artHeight := clampInt(content.H/2, 4, len(art))
	artHeight = minInt(artHeight, content.H)
	artRect, textRect := content.SplitHorizontal(artHeight, 1)
	drawHoardArt(buffer, artRect, art)
	drawHoardText(buffer, textRect, theme, lines)
}

func drawHoardArt(buffer *Buffer, rect Rect, art [][]hoardSegment) {
	if buffer == nil || rect.Empty() {
		return
	}

	artHeight := minInt(rect.H, len(art))
	startIndex := maxInt(0, len(art)-artHeight)
	y := rect.Y + maxInt(0, (rect.H-artHeight)/2)
	for lineIndex := 0; lineIndex < artHeight; lineIndex++ {
		drawStyledSegments(buffer, rect, y+lineIndex, art[startIndex+lineIndex])
	}
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
	case trimmed == "":
		return theme.Text
	default:
		return theme.Text
	}
}

func drawStyledSegments(buffer *Buffer, content Rect, y int, segments []hoardSegment) {
	if buffer == nil || content.Empty() || y < content.Y || y >= content.Y+content.H {
		return
	}

	totalWidth := 0
	for index := range segments {
		totalWidth += len([]rune(segments[index].Text))
	}

	x := content.X + maxInt(0, (content.W-totalWidth)/2)
	for index := range segments {
		x += buffer.WriteString(x, y, segments[index].Style, clipText(segments[index].Text, maxInt(0, content.X+content.W-x)))
	}
}

func drawFooter(buffer *Buffer, rect Rect, theme *Theme, text string) {
	visible := rect.Intersect(buffer.Bounds())
	if visible.Empty() {
		return
	}

	buffer.FillRect(visible, ' ', theme.Footer)
	buffer.WriteString(visible.X, visible.Y, theme.Footer, clipText(text, visible.W))
}
