package render

// Dashboard renders the first read-only placeholder shell.
type Dashboard struct{}

// Render draws the current dashboard frame.
func (Dashboard) Render(buffer *Buffer, theme *Theme, keymap KeyMap) {
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

	outer := bounds.Inset(1)
	main, footer := outer.SplitHorizontal(maxInt(0, outer.H-1), 0)
	header, body := main.SplitHorizontal(4, 1)

	DrawPanel(buffer, header, theme, Panel{
		Title: "LootSheet Dashboard",
		Lines: []string{
			"Accounting shell slice 1: alternate screen, resize-safe layout, boxed panels, and footer help.",
			"Read-only placeholder view. Domain adapters and navigation land in the next slices.",
		},
	})

	left, right := body.SplitVertical(body.W*2/5, 1)
	accounts, journal := left.SplitHorizontal(left.H/2, 1)
	ledger, lowerRight := right.SplitHorizontal(right.H/2, 1)
	quests, loot := lowerRight.SplitHorizontal(lowerRight.H/2, 1)

	DrawPanel(buffer, accounts, theme, Panel{
		Title: "Accounts",
		Lines: []string{
			"Chart of accounts screen comes next.",
			"Codes stay immutable; names remain editable.",
			"Deletion protection stays in the domain layer.",
		},
	})

	DrawPanel(buffer, journal, theme, Panel{
		Title: "Journal",
		Lines: []string{
			"Posted entries remain immutable.",
			"Corrections continue to happen by reversal or adjustment.",
			"Interactive browsing lands after the dashboard shell.",
		},
	})

	DrawPanel(buffer, ledger, theme, Panel{
		Title: "Ledger Snapshot",
		Lines: []string{
			"Read-only data adapters are intentionally deferred.",
			"This slice proves the screen lifecycle before wiring reports.",
			"No raw SQL belongs in src/render.",
		},
	})

	DrawPanel(buffer, quests, theme, Panel{
		Title: "Quest Register",
		Lines: []string{
			"Promised rewards stay off-ledger until earned.",
			"Dashboard drill-down is planned after report adapters.",
		},
	})

	DrawPanel(buffer, loot, theme, Panel{
		Title: "Loot Register",
		Lines: []string{
			"Unrealized appraisals stay off-ledger until recognized.",
			"Sales and losses will remain visible once wired into views.",
		},
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
