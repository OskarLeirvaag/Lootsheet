package render

import "github.com/gdamore/tcell/v2"

// Theme stores the small style palette needed for the first dashboard slice.
type Theme struct {
	Base             tcell.Style
	Header           tcell.Style
	Panel            tcell.Style
	Border           tcell.Style
	PanelTitle       tcell.Style
	Text             tcell.Style
	Muted            tcell.Style
	Footer           tcell.Style
	SelectedRow      tcell.Style
	StatusInfo       tcell.Style
	StatusError      tcell.Style
	StatusOK         tcell.Style
	HeaderLabel      tcell.Style
	SectionDashboard tcell.Style
	SectionAccounts  tcell.Style
	SectionJournal   tcell.Style
	SectionQuests    tcell.Style
	SectionLoot      tcell.Style
	SectionAssets    tcell.Style
	TabInactive      tcell.Style
	QuickEntry       tcell.Style
	HoardShare       tcell.Style
	HoardUnsold      tcell.Style
	HoardGold        tcell.Style
	HoardBag         tcell.Style
	HoardGem         tcell.Style
	Brick            tcell.Style
	ScatterAssets    tcell.Style
	ScatterAccounts  tcell.Style
	ScatterJournal   tcell.Style
	ScatterQuests    tcell.Style
	ScatterLoot      tcell.Style
	SectionCodex     tcell.Style
	ScatterCodex     tcell.Style
	SectionNotes     tcell.Style
	ScatterNotes     tcell.Style

	// Editor chrome styles.
	EditorLineNumber  tcell.Style
	EditorCursor      tcell.Style
	EditorStatusBar   tcell.Style
	EditorCommandLine tcell.Style

	// Markdown rendering styles.
	MarkdownHeading    tcell.Style
	MarkdownBold       tcell.Style
	MarkdownItalic     tcell.Style
	MarkdownCode       tcell.Style
	MarkdownBlockquote tcell.Style
	MarkdownReference  tcell.Style
}

// DefaultTheme returns the boxed-dashboard palette used by the first slice.
func DefaultTheme() Theme {
	panelBackground := tcell.NewRGBColor(26, 31, 39)
	bronze := tcell.NewRGBColor(246, 188, 78)
	ink := tcell.NewRGBColor(244, 239, 228)
	muted := tcell.NewRGBColor(151, 164, 180)
	slateBlue := tcell.NewRGBColor(103, 206, 255)
	amber := tcell.NewRGBColor(230, 100, 90)
	moss := tcell.NewRGBColor(110, 221, 124)
	footerBackground := tcell.NewRGBColor(32, 61, 88)
	selectedBackground := tcell.NewRGBColor(88, 69, 24)
	errorBackground := tcell.NewRGBColor(138, 43, 56)
	okBackground := tcell.NewRGBColor(34, 96, 64)
	violet := tcell.NewRGBColor(200, 140, 255)
	gold := tcell.NewRGBColor(255, 216, 74)
	bag := tcell.NewRGBColor(214, 158, 96)
	gem := tcell.NewRGBColor(72, 229, 217)
	brick := tcell.NewRGBColor(38, 43, 52)
	leafTint := tcell.NewRGBColor(110, 78, 148)
	scatterAccountsTint := tcell.NewRGBColor(55, 100, 130)
	scatterJournalTint := tcell.NewRGBColor(120, 50, 45)
	scatterQuestsTint := tcell.NewRGBColor(55, 110, 65)
	scatterLootTint := tcell.NewRGBColor(110, 100, 30)
	copper := tcell.NewRGBColor(220, 160, 100)
	scatterCodexTint := tcell.NewRGBColor(100, 75, 50)
	sageGreen := tcell.NewRGBColor(140, 190, 160)
	scatterNotesTint := tcell.NewRGBColor(70, 95, 80)

	return Theme{
		Base:             tcell.StyleDefault.Foreground(ink).Background(panelBackground),
		Header:           tcell.StyleDefault.Foreground(ink).Background(panelBackground).Bold(true),
		Panel:            tcell.StyleDefault.Foreground(ink).Background(panelBackground),
		Border:           tcell.StyleDefault.Foreground(bronze).Background(panelBackground),
		PanelTitle:       tcell.StyleDefault.Foreground(bronze).Background(panelBackground).Bold(true),
		Text:             tcell.StyleDefault.Foreground(ink).Background(panelBackground),
		Muted:            tcell.StyleDefault.Foreground(muted).Background(panelBackground),
		Footer:           tcell.StyleDefault.Foreground(ink).Background(footerBackground).Bold(true),
		SelectedRow:      tcell.StyleDefault.Foreground(ink).Background(selectedBackground).Bold(true),
		StatusInfo:       tcell.StyleDefault.Foreground(ink).Background(footerBackground).Bold(true),
		StatusError:      tcell.StyleDefault.Foreground(ink).Background(errorBackground).Bold(true),
		StatusOK:         tcell.StyleDefault.Foreground(ink).Background(okBackground).Bold(true),
		HeaderLabel:      tcell.StyleDefault.Foreground(bronze).Background(panelBackground).Bold(true),
		SectionDashboard: tcell.StyleDefault.Foreground(bronze).Background(panelBackground).Bold(true),
		SectionAccounts:  tcell.StyleDefault.Foreground(slateBlue).Background(panelBackground).Bold(true),
		SectionJournal:   tcell.StyleDefault.Foreground(amber).Background(panelBackground).Bold(true),
		SectionQuests:    tcell.StyleDefault.Foreground(moss).Background(panelBackground).Bold(true),
		SectionLoot:      tcell.StyleDefault.Foreground(gold).Background(panelBackground).Bold(true),
		SectionAssets:    tcell.StyleDefault.Foreground(violet).Background(panelBackground).Bold(true),
		TabInactive:      tcell.StyleDefault.Foreground(muted).Background(panelBackground),
		QuickEntry:       tcell.StyleDefault.Foreground(bronze).Background(panelBackground).Bold(true),
		HoardShare:       tcell.StyleDefault.Foreground(gold).Background(panelBackground).Bold(true),
		HoardUnsold:      tcell.StyleDefault.Foreground(gem).Background(panelBackground).Bold(true),
		HoardGold:        tcell.StyleDefault.Foreground(gold).Background(panelBackground).Bold(true),
		HoardBag:         tcell.StyleDefault.Foreground(bag).Background(panelBackground).Bold(true),
		HoardGem:         tcell.StyleDefault.Foreground(gem).Background(panelBackground).Bold(true),
		Brick:            tcell.StyleDefault.Foreground(brick).Background(panelBackground),
		ScatterAssets:    tcell.StyleDefault.Foreground(leafTint).Background(panelBackground),
		ScatterAccounts:  tcell.StyleDefault.Foreground(scatterAccountsTint).Background(panelBackground),
		ScatterJournal:   tcell.StyleDefault.Foreground(scatterJournalTint).Background(panelBackground),
		ScatterQuests:    tcell.StyleDefault.Foreground(scatterQuestsTint).Background(panelBackground),
		ScatterLoot:      tcell.StyleDefault.Foreground(scatterLootTint).Background(panelBackground),
		SectionCodex:     tcell.StyleDefault.Foreground(copper).Background(panelBackground).Bold(true),
		ScatterCodex:     tcell.StyleDefault.Foreground(scatterCodexTint).Background(panelBackground),
		SectionNotes:     tcell.StyleDefault.Foreground(sageGreen).Background(panelBackground).Bold(true),
		ScatterNotes:     tcell.StyleDefault.Foreground(scatterNotesTint).Background(panelBackground),

		EditorLineNumber:  tcell.StyleDefault.Foreground(muted).Background(panelBackground),
		EditorCursor:      tcell.StyleDefault.Foreground(panelBackground).Background(ink),
		EditorStatusBar:   tcell.StyleDefault.Foreground(ink).Background(footerBackground).Bold(true),
		EditorCommandLine: tcell.StyleDefault.Foreground(ink).Background(footerBackground).Bold(true),

		MarkdownHeading:    tcell.StyleDefault.Foreground(sageGreen).Background(panelBackground).Bold(true),
		MarkdownBold:       tcell.StyleDefault.Foreground(ink).Background(panelBackground).Bold(true),
		MarkdownItalic:     tcell.StyleDefault.Foreground(ink).Background(panelBackground).Italic(true),
		MarkdownCode:       tcell.StyleDefault.Foreground(muted).Background(panelBackground),
		MarkdownBlockquote: tcell.StyleDefault.Foreground(muted).Background(panelBackground).Italic(true),
		MarkdownReference:  tcell.StyleDefault.Foreground(sageGreen).Background(panelBackground),
	}
}

func resolveTheme(theme *Theme) Theme {
	if theme == nil || *theme == (Theme{}) {
		return DefaultTheme()
	}
	return *theme
}
