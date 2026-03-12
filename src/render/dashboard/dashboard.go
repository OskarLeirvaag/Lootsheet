package dashboard

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/render/canvas"
	"github.com/OskarLeirvaag/Lootsheet/src/render/goldrain"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// Styles carries resolved style fields for dashboard rendering,
// decoupled from the render.Theme type to avoid an import cycle.
type Styles struct {
	PanelStyle canvas.PanelStyle

	// Section accents for sub-panels.
	Dashboard tcell.Style
	Accounts  tcell.Style
	Journal   tcell.Style
	Quests    tcell.Style
	Loot      tcell.Style
	Assets    tcell.Style

	// Hoard panel.
	HoardGold   tcell.Style
	HoardShare  tcell.Style
	HoardUnsold tcell.Style
	QuickEntry  tcell.Style

	// General.
	Text tcell.Style
}

// DrawPanels renders the dashboard panel grid: hoard hero, accounts, journal,
// ledger, quests, loot, and assets.
func DrawPanels(buffer *canvas.Buffer, body canvas.Rect, styles *Styles, data *model.DashboardData, rain *goldrain.GoldRain) {
	if body.Empty() {
		return
	}

	resolved := ResolveData(data)

	if body.W < 52 || body.H < 10 {
		canvas.DrawPanel(buffer, body, styles.PanelStyle, canvas.Panel{
			Title: "Dashboard",
			Lines: []string{
				"Terminal too small for the full dashboard panels.",
				"Resize to restore the boxed layout.",
			},
			BorderStyle: &styles.Dashboard,
			TitleStyle:  &styles.Dashboard,
		})
		return
	}

	heroHeight := canvas.ClampInt(body.H/2, 7, 14)
	heroHeight = min(heroHeight, max(0, body.H-8))
	hero, lower := body.SplitHorizontal(heroHeight, 1)
	if lower.H < 4 {
		hero = body
		lower = canvas.Rect{}
	}

	DrawHoard(buffer, hero, styles, &resolved, rain)

	if lower.Empty() {
		return
	}

	topHeight := canvas.ClampInt(lower.H/2, 3, 4)
	top, bottom := lower.SplitHorizontal(topHeight, 1)
	if bottom.Empty() {
		top = lower
	}
	topWidth := max(16, (top.W-2)/3)
	accounts, topRest := top.SplitVertical(topWidth, 1)
	journal, ledger := topRest.SplitVertical(topWidth, 1)

	var quests canvas.Rect
	var loot canvas.Rect
	var assets canvas.Rect
	if !bottom.Empty() {
		bottomWidth := max(16, (bottom.W-2)/3)
		var bottomRest canvas.Rect
		quests, bottomRest = bottom.SplitVertical(bottomWidth, 1)
		loot, assets = bottomRest.SplitVertical(bottomWidth, 1)
	}

	canvas.DrawPanel(buffer, accounts, styles.PanelStyle, accentPanel("Accounts", resolved.AccountsLines, styles.Accounts))
	canvas.DrawPanel(buffer, journal, styles.PanelStyle, accentPanel("Journal", resolved.JournalLines, styles.Journal))
	canvas.DrawPanel(buffer, ledger, styles.PanelStyle, canvas.Panel{
		Title:       "Ledger Snapshot",
		Lines:       resolved.LedgerLines,
		BorderStyle: &styles.Dashboard,
		TitleStyle:  &styles.Dashboard,
	})

	if !quests.Empty() {
		canvas.DrawPanel(buffer, quests, styles.PanelStyle, accentPanel("Quest Register", resolved.QuestLines, styles.Quests))
	}

	if !loot.Empty() {
		canvas.DrawPanel(buffer, loot, styles.PanelStyle, accentPanel("Loot Register", resolved.LootLines, styles.Loot))
	}

	if !assets.Empty() {
		canvas.DrawPanel(buffer, assets, styles.PanelStyle, accentPanel("Asset Register", resolved.AssetLines, styles.Assets))
	}
}

// DrawHoard renders the hoard hero panel with gold rain and summary text.
func DrawHoard(buffer *canvas.Buffer, rect canvas.Rect, styles *Styles, data *model.DashboardData, rain *goldrain.GoldRain) {
	if buffer == nil || styles == nil {
		return
	}

	resolved := ResolveData(data)
	canvas.DrawPanel(buffer, rect, styles.PanelStyle, canvas.Panel{
		Title:       "Party Hoard",
		BorderStyle: &styles.HoardGold,
		TitleStyle:  &styles.HoardGold,
	})

	content := canvas.PanelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	lines := append([]string{}, resolved.HoardLines...)
	lines = append(lines, "")
	lines = append(lines, resolved.QuickEntryLines...)

	if content.W >= 60 {
		rainWidth := canvas.ClampInt((content.W*2)/3, 20, max(20, content.W-20))
		rainRect, textRect := content.SplitVertical(rainWidth, 1)
		if rain != nil {
			rain.Render(buffer, rainRect, styles.HoardGold, styles.Text)
		}
		drawHoardText(buffer, textRect, styles, lines)
		return
	}

	rainHeight := canvas.ClampInt(content.H/2, 4, max(4, content.H-len(lines)-1))
	rainHeight = min(rainHeight, content.H)
	rainRect, textRect := content.SplitHorizontal(rainHeight, 1)
	if rain != nil {
		rain.Render(buffer, rainRect, styles.HoardGold, styles.Text)
	}
	drawHoardText(buffer, textRect, styles, lines)
}

func drawHoardText(buffer *canvas.Buffer, rect canvas.Rect, styles *Styles, lines []string) {
	if buffer == nil || styles == nil || rect.Empty() {
		return
	}

	y := rect.Y
	for _, line := range lines {
		if y >= rect.Y+rect.H {
			return
		}
		buffer.WriteString(rect.X, y, hoardLineStyle(styles, line), canvas.ClipText(line, rect.W))
		y++
	}
}

func hoardLineStyle(styles *Styles, line string) tcell.Style {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "To share now:"):
		return styles.HoardShare
	case strings.HasPrefix(trimmed, "Unsold loot:"):
		return styles.HoardUnsold
	case strings.HasPrefix(trimmed, "e  "), strings.HasPrefix(trimmed, "i  "), strings.HasPrefix(trimmed, "a  "):
		return styles.QuickEntry
	default:
		return styles.Text
	}
}

func accentPanel(title string, lines []string, accent tcell.Style) canvas.Panel {
	return canvas.Panel{
		Title:       title,
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
	}
}
