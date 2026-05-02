package component

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func tableSafeASCII(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			b.WriteByte(' ')
		case r < 32 || r == 127:
			// drop control chars
		default:
			w := runewidth.RuneWidth(r)
			if w <= 0 {
				// combining/zero-width marks can skew layout in some terminals
				// skip them for stable columns
				continue
			}
			// keep visible unicode, including wide emoji/CJK glyphs
			b.WriteRune(r)
		}
	}
	return b.String()
}

func truncateDisplay(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	const ellipsis = "…"
	if max == 1 {
		return ellipsis
	}
	out := ""
	for _, r := range s {
		next := out + string(r)
		if lipgloss.Width(next)+lipgloss.Width(ellipsis) > max {
			break
		}
		out = next
	}
	if strings.TrimSpace(out) == "" {
		return ellipsis
	}
	return out + ellipsis
}

func padDisplayRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func cellDisplay(s string, width int) string {
	return padDisplayRight(truncateDisplay(tableSafeASCII(s), width), width)
}

func cellDisplayStyled(s string, width int) string {
	if width <= 0 {
		return ""
	}
	// For pre-styled strings (ANSI), do not sanitize characters; just pad by display width.
	if lipgloss.Width(s) > width {
		return padDisplayRight("-", width)
	}
	return padDisplayRight(s, width)
}
