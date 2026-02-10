package component

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func BorderedBox(borderStyle, labelStyle lipgloss.Style, label, content string, width, height int) string {
	// Rounded box-drawing characters
	tl, tr, bl, br := "╭", "╮", "╰", "╯"
	hz, vt := "─", "│"

	innerW := width - 2
	innerH := height - 2

	// label is baked into the top border line
	label = strings.ReplaceAll(label, "\n", " ")
	lbl := labelStyle.Render(" " + label + " ")

	used := 1 + lipgloss.Width(lbl) // one leading hz + label
	remaining := max(innerW-used, 0)

	topLine := borderStyle.Render(tl) +
		borderStyle.Render(hz) +
		lbl +
		borderStyle.Render(strings.Repeat(hz, remaining)) +
		borderStyle.Render(tr)

	lines := make([]string, 0, height)
	lines = append(lines, topLine)

	// clamp content to innerH and fit each line to innerW
	contentLines := strings.Split(content, "\n")
	for len(contentLines) < innerH {
		contentLines = append(contentLines, "")
	}
	if len(contentLines) > innerH {
		contentLines = contentLines[:innerH]
	}

	for _, line := range contentLines {
		line = fitWidth(line, innerW)
		lines = append(lines, borderStyle.Render(vt)+line+borderStyle.Render(vt))
	}

	botLine := borderStyle.Render(bl) +
		borderStyle.Render(strings.Repeat(hz, innerW)) +
		borderStyle.Render(br)

	lines = append(lines, botLine)
	return strings.Join(lines, "\n")
}
