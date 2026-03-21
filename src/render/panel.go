package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/canvas"

// Type aliases re-export canvas panel types.
type PanelTexture = canvas.PanelTexture
type BorderSet = canvas.BorderSet
type Panel = canvas.Panel

const (
	PanelTextureBrick = canvas.PanelTextureBrick
	PanelTextureLeaf  = canvas.PanelTextureLeaf
	PanelTextureNone  = canvas.PanelTextureNone
)

var runicBorders = canvas.RunicBorders

// Scatter glyph aliases.
var (
	scatterGlyphs   = canvas.ScatterGlyphs
	scatterLoot     = canvas.ScatterLoot
	scatterQuests   = canvas.ScatterQuests
	scatterJournal  = canvas.ScatterJournal
	scatterLedger   = canvas.ScatterLedger
	scatterPeople   = canvas.ScatterPeople
	scatterNotes    = canvas.ScatterNotes
	scatterSettings = canvas.ScatterSettings
)

// panelStyle converts theme fields into a canvas.PanelStyle.
func panelStyle(theme *Theme) canvas.PanelStyle {
	return canvas.PanelStyle{
		Background: theme.Panel,
		Texture:    theme.Brick,
		Border:     theme.Border,
		Title:      theme.PanelTitle,
		Text:       theme.Text,
	}
}

// DrawPanel renders a boxed panel, bridging Theme to canvas.PanelStyle.
func DrawPanel(buffer *Buffer, rect Rect, theme *Theme, panel Panel) { //nolint:gocritic // hugeParam: passing by value matches canvas.DrawPanel signature
	canvas.DrawPanel(buffer, rect, panelStyle(theme), panel)
}
