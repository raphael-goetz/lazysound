package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/raphael-goetz/lazysound/internal/math"
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

// ---------- Header / Player / Cmd ----------

func renderHeader(s style.Styles, width int) string {
	left := "Test"
	right := s.Muted.Render("user@soundcloud")

	space := max(width-lipgloss.Width(left)-lipgloss.Width(right)-2, 1)
	line := left + strings.Repeat(" ", space) + right
	return s.Header.Width(width).Render(line)
}

func renderPlayerBar(s style.Styles, m Model, width int) string {
	t := m.tracks[m.trackIdx]
	title := s.Accent.Render(math.Truncate(t.Title, 28))
	artist := math.Truncate(t.User.Username, 22)

	pos := "02:21"
	total := t.Duration

	barW := math.Clamp(width-62, 12, 40)
	bar := progressBar(0.35, barW)

	line1 := fmt.Sprintf("NOW PLAYING: %s — %s   ⏯  ◀◀  ▶▶", title, artist)
	line2 := fmt.Sprintf("%s  %s / %s    🔊 65%%    ♡  +Q", bar, pos, total)

	return s.Player.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2))
}

func joinPanes(panes ...string) string {
	out := []string{}
	for _, p := range panes {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, p)
	}
	// One space gap between panes (lazygit-ish)
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}
func progressBar(p float64, w int) string {
	if w < 3 {
		w = 3
	}
	filled := min(max(int(p*float64(w)), 0), w)
	return lipgloss.NewStyle().Background(lipgloss.Color("#000000")).Render(strings.Repeat(" ", filled)) +
		strings.Repeat(" ", w-filled)
}
