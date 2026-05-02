package component

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
	"github.com/raphael-goetz/lazysound/lib/uikit/math"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

func TrackTable(s style.Styles, tracks []api.Track, badges map[int]string, sel int, focused bool, width, height int) string {
	fwidth := float64(width)
	wTitle := int(fwidth * (3.0 / 8.0))
	wArtist := int(fwidth * (2.0 / 8.0))
	wDuration := int(fwidth * (1.0 / 8.0))
	wBPM := int(fwidth * (1.0 / 8.0))
	wBadge := int(fwidth * (1.0 / 8.0))
	headerRow := strings.Join([]string{
		cellDisplay("Title", wTitle),
		cellDisplay("Artist(s)", wArtist),
		cellDisplay("Durtion", wDuration),
		cellDisplay("BPM", wBPM),
		cellDisplay("Status", wBadge),
	}, "  ")
	header := s.Muted.Render(fitWidth(headerRow, width))
	sep := s.Muted.Render(fitWidth(strings.Repeat("─", width), width))

	length := len(tracks)
	maxRows := max(height-2, 3)

	start := math.Clamp(sel-maxRows/2, 0, max(0, length-maxRows))
	end := min(length, start+maxRows)

	lines := []string{header, sep}
	for i := start; i < end; i++ {
		t := tracks[i]
		badge := trackBadge(t, badges)
		bpm := "-"
		if t.Bpm > 0 {
			bpm = strconv.Itoa(int(float64(t.Bpm) + 0.5))
		}
		row := strings.Join([]string{
			cellDisplay(t.Title, wTitle),
			cellDisplay(t.User.Username, wArtist),
			cellDisplay(Time(t.Duration), wDuration),
			cellDisplay(bpm, wBPM),
			cellDisplayStyled(renderBadge(badge, wBadge), wBadge),
		}, "  ")

		if i == sel {
			if focused {
				lines = append(lines, s.RowSel.Render(fitWidth(row, width)))
			} else {
				lines = append(lines, s.Accent.Render(fitWidth(row, width)))
			}
		} else {
			lines = append(lines, s.Row.Render(fitWidth(row, width)))
		}
	}

	// pad remaining space so pane doesn't look empty
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	// fit any empties
	for i := range lines {
		lines[i] = fitWidth(lines[i], width)
	}

	return strings.Join(lines, "\n")
}

func trackBadge(t api.Track, badges map[int]string) string {
	if badge, ok := badges[t.ID]; ok {
		b := strings.ToLower(strings.TrimSpace(badge))
		if b == "playable" || b == "preview" || b == "blocked" {
			return b
		}
	}
	switch strings.ToLower(strings.TrimSpace(t.Access)) {
	case "preview":
		return "preview"
	case "blocked":
		return "blocked"
	case "playable":
		return "playable"
	default:
		return "-"
	}
}

func renderBadge(badge string, maxWidth int) string {
	label := strings.ToLower(strings.TrimSpace(badge))
	if label == "" || label == "-" {
		return "-"
	}

	var color lipgloss.Color
	var text string
	var textColor lipgloss.Color
	switch label {
	case "playable":
		text = "[ok]"
		color = lipgloss.Color("42")
		textColor = lipgloss.Color("42")
	case "preview":
		text = "[pv]"
		color = lipgloss.Color("214")
		textColor = lipgloss.Color("214")
	case "blocked":
		text = "[x]"
		color = lipgloss.Color("196")
		textColor = lipgloss.Color("196")
	default:
		text = "[" + label + "]"
	}

	if lipgloss.Width(text) > maxWidth {
		switch label {
		case "playable":
			text = "ok"
		case "preview":
			text = "pv"
		case "blocked":
			text = "x"
		default:
			text = "-"
		}
	}

	pill := lipgloss.NewStyle().
		Foreground(textColor).
		Bold(true).
		BorderForeground(color).
		Render(text)
	return pill
}
