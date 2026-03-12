package render

import (
	"github.com/OskarLeirvaag/Lootsheet/src/render/dashboard"
	"github.com/OskarLeirvaag/Lootsheet/src/render/goldrain"
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
	main, footer := outer.SplitHorizontal(max(0, outer.H-1), 0)
	header, body := main.SplitHorizontal(4, 1)

	DrawPanel(buffer, header, theme, Panel{
		Title:       "LootSheet Dashboard",
		Lines:       data.HeaderLines,
		BorderStyle: &theme.SectionDashboard,
		TitleStyle:  &theme.SectionDashboard,
	})

	drawDashboardPanels(buffer, body, theme, &data, nil)
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

func dashboardStyles(theme *Theme) *dashboard.Styles {
	return &dashboard.Styles{
		PanelStyle:  panelStyle(theme),
		Dashboard:   theme.SectionDashboard,
		Accounts:    sectionStyleFor(SectionAccounts, theme).Accent,
		Journal:     sectionStyleFor(SectionJournal, theme).Accent,
		Quests:      sectionStyleFor(SectionQuests, theme).Accent,
		Loot:        sectionStyleFor(SectionLoot, theme).Accent,
		Assets:      sectionStyleFor(SectionAssets, theme).Accent,
		HoardGold:   theme.HoardGold,
		HoardShare:  theme.HoardShare,
		HoardUnsold: theme.HoardUnsold,
		QuickEntry:  theme.QuickEntry,
		Text:        theme.Text,
	}
}

func drawDashboardPanels(buffer *Buffer, body Rect, theme *Theme, data *DashboardData, rain *goldrain.GoldRain) {
	dashboard.DrawPanels(buffer, body, dashboardStyles(theme), data, rain)
}

func drawFooter(buffer *Buffer, rect Rect, theme *Theme, text string) {
	visible := rect.Intersect(buffer.Bounds())
	if visible.Empty() {
		return
	}

	buffer.FillRect(visible, ' ', theme.Footer)
	buffer.WriteString(visible.X, visible.Y, theme.Footer, clipText(text, visible.W))
}
