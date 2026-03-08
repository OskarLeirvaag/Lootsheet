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
	main, footer := outer.SplitHorizontal(maxInt(0, outer.H-1), 0)
	header, body := main.SplitHorizontal(4, 1)

	DrawPanel(buffer, header, theme, Panel{
		Title: "LootSheet Dashboard",
		Lines: data.HeaderLines,
	})

	left, right := body.SplitVertical(body.W*2/5, 1)
	accounts, journal := left.SplitHorizontal(left.H/2, 1)
	ledger, lowerRight := right.SplitHorizontal(right.H/2, 1)
	quests, loot := lowerRight.SplitHorizontal(lowerRight.H/2, 1)

	DrawPanel(buffer, accounts, theme, Panel{
		Title: "Accounts",
		Lines: data.AccountsLines,
	})

	DrawPanel(buffer, journal, theme, Panel{
		Title: "Journal",
		Lines: data.JournalLines,
	})

	DrawPanel(buffer, ledger, theme, Panel{
		Title: "Ledger Snapshot",
		Lines: data.LedgerLines,
	})

	DrawPanel(buffer, quests, theme, Panel{
		Title: "Quest Register",
		Lines: data.QuestLines,
	})

	DrawPanel(buffer, loot, theme, Panel{
		Title: "Loot Register",
		Lines: data.LootLines,
	})

	drawFooter(buffer, footer, theme, keymap.HelpText())
}

func renderCompactDashboard(buffer *Buffer, bounds Rect, theme *Theme, keymap KeyMap) {
	panel := bounds.Inset(1)
	DrawPanel(buffer, panel, theme, Panel{
		Title: "LootSheet",
		Lines: []string{
			"Terminal too small for the full dashboard.",
			"Resize the terminal and the boxed layout will redraw cleanly.",
			keymap.HelpText(),
		},
	})
}

func drawFooter(buffer *Buffer, rect Rect, theme *Theme, text string) {
	visible := rect.Intersect(buffer.Bounds())
	if visible.Empty() {
		return
	}

	buffer.FillRect(visible, ' ', theme.Footer)
	buffer.WriteString(visible.X, visible.Y, theme.Footer, clipText(text, visible.W))
}
