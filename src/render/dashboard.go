package render

import (
	"github.com/OskarLeirvaag/Lootsheet/src/render/dashboard"
	"github.com/OskarLeirvaag/Lootsheet/src/render/goldrain"
)

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
