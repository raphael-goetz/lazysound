package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
	"github.com/raphael-goetz/lazysound/lib/uikit/component"
	"github.com/raphael-goetz/lazysound/lib/uikit/math"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

// ---------- Header / Player / Cmd ----------

func renderHeader(s style.Styles, m Model, width int) string {
	left := "LazySound"
	right := s.Muted.Render(m.greeting)

	space := max(width-lipgloss.Width(left)-lipgloss.Width(right)-2, 1)
	line := left + strings.Repeat(" ", space) + right
	return s.Header.Width(width).Render(line)
}

func renderPlayerBar(s style.Styles, m Model, width int) string {
	t, ok := currentTrackForPlayer(m)
	if !ok {
		line1 := "NOW PLAYING: -"
		status := m.status
		if status == "" {
			status = "stopped"
		}
		line2 := fmt.Sprintf("Status: %s    🔊 %d%%", status, m.volume)
		return s.Player.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2))
	}

	title := s.Accent.Render(math.Truncate(t.Title, 28))
	artist := math.Truncate(t.User.Username, 22)
	status := "stopped"
	if m.nowPlaying != nil && !m.playStart.IsZero() {
		status = "playing"
	}
	if m.playPaused {
		status = "paused"
	}
	if m.status != "" {
		status = m.status
	}

	elapsed := m.playElapsed
	total := m.playTotal
	if total <= 0 {
		total = t.Duration
	}
	if total < 0 {
		total = 0
	}
	if elapsed > total && total > 0 {
		elapsed = total
	}
	p := 0.0
	if total > 0 {
		p = float64(elapsed) / float64(total)
	}
	barW := math.Clamp(width-68, 12, 40)
	bar := progressBar(s, p, barW)

	line1 := fmt.Sprintf("NOW PLAYING: %s — %s", title, artist)
	line2 := fmt.Sprintf("%s  %s / %s  Status: %s    🔊 %d%%", bar, component.Time(elapsed), component.Time(total), status, m.volume)

	return s.Player.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2))
}

func currentTrackForPlayer(m Model) (api.Track, bool) {
	// Player should always reflect the active playback, not the hovered row.
	if m.nowPlaying == nil {
		return api.Track{}, false
	}
	return *m.nowPlaying, true
}

func wrapText(s string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	words := strings.Fields(strings.ReplaceAll(s, "\n", " "))
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	line := words[0]
	for _, w := range words[1:] {
		if lipgloss.Width(line)+1+lipgloss.Width(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line = line + " " + w
		}
	}
	lines = append(lines, line)
	return lines
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
func progressBar(s style.Styles, p float64, w int) string {
	if w < 3 {
		w = 3
	}
	filled := min(max(int(p*float64(w)), 0), w)
	return s.Progress.Render(strings.Repeat(" ", filled)) + strings.Repeat(" ", w-filled)
}

func renderInspectInfo(s style.Styles, m Model, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := []string{}
	if m.listKind() == ListTracks {
		tracks := m.currentTracks()
		if len(tracks) == 0 || m.trackIdx < 0 || m.trackIdx >= len(tracks) {
			lines = append(lines, s.Muted.Render("No track selected."))
		} else {
			t := tracks[m.trackIdx]
			titleLines := wrapText(t.Title, width)
			if len(titleLines) == 0 {
				titleLines = []string{""}
			}
			lines = append(lines, s.Accent.Render(titleLines[0]))
			for i := 1; i < len(titleLines); i++ {
				lines = append(lines, s.Accent.Render(titleLines[i]))
			}
			artistLines := wrapText(t.User.Username, width)
			for _, l := range artistLines {
				lines = append(lines, s.Muted.Render(l))
			}
			lines = append(lines, "")
			lines = append(lines, s.Muted.Render("Details"))
			lines = append(lines, fmt.Sprintf("Duration: %s", component.Time(t.Duration)))
			if t.Bpm > 0 {
				lines = append(lines, fmt.Sprintf("BPM:      %d", int(float64(t.Bpm)+0.5)))
			} else {
				lines = append(lines, "BPM:      -")
			}
			if strings.TrimSpace(t.TagList) != "" {
				lines = append(lines, "")
				lines = append(lines, s.Muted.Render("Tags"))
				for _, l := range wrapText(t.TagList, width) {
					lines = append(lines, l)
				}
			}
			if strings.TrimSpace(t.ArtworkURL) != "" {
				lines = append(lines, "")
				lines = append(lines, s.Muted.Render("Artwork"))
				for _, l := range wrapText(t.ArtworkURL, width) {
					lines = append(lines, l)
				}
			}
			if strings.TrimSpace(t.WaveformURL) != "" {
				lines = append(lines, "")
				lines = append(lines, s.Muted.Render("Waveform"))
				for _, l := range wrapText(t.WaveformURL, width) {
					lines = append(lines, l)
				}
			}
			if strings.TrimSpace(t.Description) != "" {
				lines = append(lines, "")
				lines = append(lines, s.Muted.Render("Description"))
				for _, l := range wrapText(t.Description, width) {
					lines = append(lines, l)
				}
			}
		}
	} else if m.listKind() == ListPlaylists {
		playlists := m.currentPlaylists()
		if len(playlists) == 0 || m.playlistIdx < 0 || m.playlistIdx >= len(playlists) {
			lines = append(lines, s.Muted.Render("No playlist selected."))
		} else {
			p := playlists[m.playlistIdx]
			titleLines := wrapText(p.Title, width)
			if len(titleLines) == 0 {
				titleLines = []string{""}
			}
			lines = append(lines, s.Accent.Render(titleLines[0]))
			for i := 1; i < len(titleLines); i++ {
				lines = append(lines, s.Accent.Render(titleLines[i]))
			}
			for _, l := range wrapText(p.User.Username, width) {
				lines = append(lines, s.Muted.Render(l))
			}
			lines = append(lines, "")
			lines = append(lines, s.Muted.Render("Details"))
			lines = append(lines, fmt.Sprintf("Duration: %s", component.Time(p.Duration)))
			lines = append(lines, fmt.Sprintf("Tracks:   %d", p.TrackCount))
			if strings.TrimSpace(p.Description) != "" {
				lines = append(lines, "")
				lines = append(lines, s.Muted.Render("Description"))
				for _, l := range wrapText(p.Description, width) {
					lines = append(lines, l)
				}
			}
		}
	} else {
		lines = append(lines, s.Muted.Render("Select an item."))
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = fitWidth(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func fitWidth(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s + strings.Repeat(" ", w-lipgloss.Width(s))
	}

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
