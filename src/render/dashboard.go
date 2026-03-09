package render

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

func drawDashboardPanels(buffer *Buffer, body Rect, theme *Theme, data *DashboardData, rain *GoldRain) {
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

	heroHeight := clampInt(body.H/2, 7, 14)
	heroHeight = min(heroHeight, max(0, body.H-8))
	hero, lower := body.SplitHorizontal(heroHeight, 1)
	if lower.H < 4 {
		hero = body
		lower = Rect{}
	}

	drawHoardPanel(buffer, hero, theme, &resolved, rain)

	if lower.Empty() {
		return
	}

	topHeight := clampInt(lower.H/2, 3, 4)
	top, bottom := lower.SplitHorizontal(topHeight, 1)
	if bottom.Empty() {
		top = lower
	}
	topWidth := max(16, (top.W-2)/3)
	accounts, topRest := top.SplitVertical(topWidth, 1)
	journal, ledger := topRest.SplitVertical(topWidth, 1)

	var quests Rect
	var loot Rect
	var assets Rect
	if !bottom.Empty() {
		bottomWidth := max(16, (bottom.W-2)/3)
		var bottomRest Rect
		quests, bottomRest = bottom.SplitVertical(bottomWidth, 1)
		loot, assets = bottomRest.SplitVertical(bottomWidth, 1)
	}

	ssAccounts := SectionAccounts.Style(theme)
	DrawPanel(buffer, accounts, theme, ssAccounts.Panel("Accounts", resolved.AccountsLines))

	ssJournal := SectionJournal.Style(theme)
	DrawPanel(buffer, journal, theme, ssJournal.Panel("Journal", resolved.JournalLines))

	DrawPanel(buffer, ledger, theme, Panel{
		Title:       "Ledger Snapshot",
		Lines:       resolved.LedgerLines,
		BorderStyle: &theme.SectionDashboard,
		TitleStyle:  &theme.SectionDashboard,
	})

	if !quests.Empty() {
		ssQuests := SectionQuests.Style(theme)
		DrawPanel(buffer, quests, theme, ssQuests.Panel("Quest Register", resolved.QuestLines))
	}

	if !loot.Empty() {
		ssLoot := SectionLoot.Style(theme)
		DrawPanel(buffer, loot, theme, ssLoot.Panel("Loot Register", resolved.LootLines))
	}

	if !assets.Empty() {
		ssAssets := SectionAssets.Style(theme)
		DrawPanel(buffer, assets, theme, ssAssets.Panel("Asset Register", resolved.AssetLines))
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
