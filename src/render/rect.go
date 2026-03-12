package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/canvas"

// Type aliases re-export canvas primitives so render-internal code
// and tests can continue using unqualified names.
type Rect = canvas.Rect

func clampInt(value, low, high int) int            { return canvas.ClampInt(value, low, high) }
func clipText(text string, width int) string       { return canvas.ClipText(text, width) }
func panelContentRect(rect Rect, bounds Rect) Rect { return canvas.PanelContentRect(rect, bounds) }
func modalBounds(parent Rect, lines []string, defaultW, minW, maxW, minH int) Rect {
	return canvas.ModalBounds(parent, lines, defaultW, minW, maxW, minH)
}
