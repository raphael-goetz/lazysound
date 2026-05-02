package component

import (
	"strings"

	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

func Nav(s style.Styles, items []string, idx int, focused bool, width, height int) string {
	lines := make([]string, 0, len(items))
	for i, it := range items {
		if i == idx {
			if focused {
				lines = append(lines, s.RowSel.Render("› "+it))
			} else {
				lines = append(lines, s.Accent.Render("› "+it))
			}
		} else {
			lines = append(lines, s.Row.Render("  "+it))
		}
	}

	// pad / clamp to height
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	// fit to width
	for i := range lines {
		lines[i] = fitWidth(lines[i], width)
	}
	return strings.Join(lines, "\n")
}
