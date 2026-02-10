package component

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func fitWidth(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s + strings.Repeat(" ", w-lipgloss.Width(s))
	}

	// truncate by runes, respecting visible width
	rs := []rune(s)
	out := ""
	for _, r := range rs {
		if lipgloss.Width(out+string(r)) > w {
			break
		}
		out += string(r)
	}
	if lipgloss.Width(out) < w {
		out += strings.Repeat(" ", w-lipgloss.Width(out))
	}
	return out
}

