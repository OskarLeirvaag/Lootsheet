package render

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/texture"
)

type hoardSegment struct {
	Text  string
	Style tcell.Style
}

func hoardArt(theme *Theme) [][]hoardSegment {
	artBytes, err := texture.FS.ReadFile("hoard.ascii")
	if err != nil {
		return nil
	}
	mapBytes, err := texture.FS.ReadFile("hoard.colormap")
	if err != nil {
		return nil
	}

	artLines := textLines(string(artBytes))
	tags, rules := parseColormap(string(mapBytes))

	styles := hoardStyleMap(theme)
	tagStyles := make(map[string]tcell.Style)
	for tag, name := range tags {
		if s, ok := styles[name]; ok {
			tagStyles[tag] = s
		}
	}

	// Pad lines to uniform width so centering is consistent.
	maxW := 0
	for _, l := range artLines {
		if len(l) > maxW {
			maxW = len(l)
		}
	}
	for i, l := range artLines {
		if len(l) < maxW {
			artLines[i] = l + strings.Repeat(" ", maxW-len(l))
		}
	}

	out := make([][]hoardSegment, len(artLines))
	for i, line := range artLines {
		rule, ok := rules[i+1]
		if !ok {
			out[i] = []hoardSegment{{Text: line, Style: theme.Text}}
			continue
		}
		out[i] = applyColorRule(line, tagStyles[rule.defaultTag], rule.overrides, tagStyles)
	}
	return out
}

func hoardStyleMap(theme *Theme) map[string]tcell.Style {
	return map[string]tcell.Style{
		"HoardDragon": theme.HoardDragon,
		"HoardGold":   theme.HoardGold,
		"HoardGem":    theme.HoardGem,
		"HoardBag":    theme.HoardBag,
	}
}

type colorRule struct {
	defaultTag string
	overrides  []patternOverride
}

type patternOverride struct {
	pattern string
	tag     string
}

func parseColormap(data string) (tags map[string]string, rules map[int]colorRule) {
	tags = make(map[string]string)
	rules = make(map[int]colorRule)

	parts := strings.SplitN(data, "---", 2)
	if len(parts) != 2 {
		return
	}

	for _, line := range textLines(parts[0]) {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		if kv := strings.SplitN(line, "=", 2); len(kv) == 2 {
			tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	for _, line := range textLines(parts[1]) {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		rule := colorRule{defaultTag: fields[1]}
		for i := 2; i+1 < len(fields); i += 2 {
			rule.overrides = append(rule.overrides, patternOverride{
				pattern: fields[i],
				tag:     fields[i+1],
			})
		}

		if dash := strings.Index(fields[0], "-"); dash >= 0 {
			lo, _ := strconv.Atoi(fields[0][:dash])
			hi, _ := strconv.Atoi(fields[0][dash+1:])
			for n := lo; n <= hi; n++ {
				rules[n] = rule
			}
		} else {
			n, _ := strconv.Atoi(fields[0])
			rules[n] = rule
		}
	}
	return
}

func applyColorRule(line string, defaultStyle tcell.Style, overrides []patternOverride, tagStyles map[string]tcell.Style) []hoardSegment {
	if len(overrides) == 0 {
		return []hoardSegment{{Text: line, Style: defaultStyle}}
	}

	type match struct {
		start, end int
		style      tcell.Style
	}
	var matches []match
	for _, ov := range overrides {
		style := tagStyles[ov.tag]
		idx := 0
		for {
			pos := strings.Index(line[idx:], ov.pattern)
			if pos < 0 {
				break
			}
			matches = append(matches, match{idx + pos, idx + pos + len(ov.pattern), style})
			idx += pos + len(ov.pattern)
		}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].start < matches[j].start })

	var segments []hoardSegment
	pos := 0
	for _, m := range matches {
		if m.start > pos {
			segments = append(segments, hoardSegment{Text: line[pos:m.start], Style: defaultStyle})
		}
		segments = append(segments, hoardSegment{Text: line[m.start:m.end], Style: m.style})
		pos = m.end
	}
	if pos < len(line) {
		segments = append(segments, hoardSegment{Text: line[pos:], Style: defaultStyle})
	}
	return segments
}

func textLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func drawHoardPanel(buffer *Buffer, rect Rect, theme *Theme, data *DashboardData) {
	if buffer == nil || theme == nil {
		return
	}

	resolved := resolveDashboardData(data)
	DrawPanel(buffer, rect, theme, Panel{
		Title:       "Party Hoard",
		BorderStyle: &theme.HoardGold,
		TitleStyle:  &theme.HoardGold,
	})

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	art := hoardArt(theme)

	lines := append([]string{}, resolved.HoardLines...)
	lines = append(lines, "")
	lines = append(lines, resolved.QuickEntryLines...)

	if content.W >= 60 {
		artWidth := clampInt((content.W*2)/3, 36, maxInt(36, content.W-20))
		artRect, textRect := content.SplitVertical(artWidth, 1)
		drawHoardArt(buffer, artRect, art)
		drawHoardText(buffer, textRect, theme, lines)
		return
	}

	artHeight := clampInt(content.H/2, 4, len(art))
	artHeight = minInt(artHeight, content.H)
	artRect, textRect := content.SplitHorizontal(artHeight, 1)
	drawHoardArt(buffer, artRect, art)
	drawHoardText(buffer, textRect, theme, lines)
}

func drawHoardArt(buffer *Buffer, rect Rect, art [][]hoardSegment) {
	if buffer == nil || rect.Empty() {
		return
	}

	artHeight := minInt(rect.H, len(art))
	startIndex := maxInt(0, len(art)-artHeight)
	y := rect.Y + maxInt(0, (rect.H-artHeight)/2)
	for lineIndex := 0; lineIndex < artHeight; lineIndex++ {
		drawStyledSegments(buffer, rect, y+lineIndex, art[startIndex+lineIndex])
	}
}

func drawHoardText(buffer *Buffer, rect Rect, theme *Theme, lines []string) {
	if buffer == nil || theme == nil || rect.Empty() {
		return
	}

	y := rect.Y
	for _, line := range lines {
		if y >= rect.Y+rect.H {
			return
		}
		buffer.WriteString(rect.X, y, hoardLineStyle(theme, line), clipText(line, rect.W))
		y++
	}
}

func hoardLineStyle(theme *Theme, line string) tcell.Style {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "To share now:"):
		return theme.HoardShare
	case strings.HasPrefix(trimmed, "Unsold loot:"):
		return theme.HoardUnsold
	case strings.HasPrefix(trimmed, "e  "), strings.HasPrefix(trimmed, "i  "), strings.HasPrefix(trimmed, "a  "):
		return theme.QuickEntry
	case trimmed == "":
		return theme.Text
	default:
		return theme.Text
	}
}

func drawStyledSegments(buffer *Buffer, content Rect, y int, segments []hoardSegment) {
	if buffer == nil || content.Empty() || y < content.Y || y >= content.Y+content.H {
		return
	}

	totalWidth := 0
	for index := range segments {
		totalWidth += len([]rune(segments[index].Text))
	}

	x := content.X + maxInt(0, (content.W-totalWidth)/2)
	for index := range segments {
		x += buffer.WriteString(x, y, segments[index].Style, clipText(segments[index].Text, maxInt(0, content.X+content.W-x)))
	}
}
