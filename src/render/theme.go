package render

import "github.com/gdamore/tcell/v2"

// Theme stores the small style palette needed for the first dashboard slice.
type Theme struct {
	Base       tcell.Style
	Header     tcell.Style
	Panel      tcell.Style
	Border     tcell.Style
	PanelTitle tcell.Style
	Text       tcell.Style
	Muted      tcell.Style
	Footer     tcell.Style
}

// DefaultTheme returns the boxed-dashboard palette used by the first slice.
func DefaultTheme() Theme {
	background := tcell.NewRGBColor(24, 28, 31)
	panelBackground := tcell.NewRGBColor(34, 39, 44)
	bronze := tcell.NewRGBColor(181, 150, 92)
	ink := tcell.NewRGBColor(233, 226, 213)
	muted := tcell.NewRGBColor(159, 166, 173)
	footerBackground := tcell.NewRGBColor(50, 60, 42)

	return Theme{
		Base:       tcell.StyleDefault.Foreground(ink).Background(background),
		Header:     tcell.StyleDefault.Foreground(ink).Background(panelBackground).Bold(true),
		Panel:      tcell.StyleDefault.Foreground(ink).Background(panelBackground),
		Border:     tcell.StyleDefault.Foreground(bronze).Background(panelBackground),
		PanelTitle: tcell.StyleDefault.Foreground(bronze).Background(panelBackground).Bold(true),
		Text:       tcell.StyleDefault.Foreground(ink).Background(panelBackground),
		Muted:      tcell.StyleDefault.Foreground(muted).Background(panelBackground),
		Footer:     tcell.StyleDefault.Foreground(ink).Background(footerBackground).Bold(true),
	}
}

func resolveTheme(theme *Theme) Theme {
	if theme == nil || *theme == (Theme{}) {
		return DefaultTheme()
	}
	return *theme
}
