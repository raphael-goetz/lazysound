package component

import (
	"strconv"
	"strings"

	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
	"github.com/raphael-goetz/lazysound/lib/uikit/math"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

func PlaylistTable(s style.Styles, playlists []api.Playlist, sel int, focused bool, width, height int) string {

	fwidth := float64(width)
	wTitle := int(fwidth * (4.0 / 8.0))
	wArtist := int(fwidth * (2.0 / 8.0))
	wDuration := int(fwidth * (1.0 / 8.0))
	wCount := int(fwidth * (1.0 / 8.0))

	headerRow := strings.Join([]string{
		cellDisplay("Title", wTitle),
		cellDisplay("Creator", wArtist),
		cellDisplay("Duration", wDuration),
		cellDisplay("Tracks", wCount),
	}, "  ")
	header := s.Muted.Render(fitWidth(headerRow, width))
	sep := s.Muted.Render(fitWidth(strings.Repeat("─", width), width))

	length := len(playlists)
	maxRows := max(height-2, 3)

	start := math.Clamp(sel-maxRows/2, 0, max(0, length-maxRows))
	end := min(length, start+maxRows)

	lines := []string{header, sep}
	for i := start; i < end; i++ {
		t := playlists[i]
		row := strings.Join([]string{
			cellDisplay(t.Title, wTitle),
			cellDisplay(t.User.Username, wArtist),
			cellDisplay(Time(t.Duration), wDuration),
			cellDisplay(strconv.Itoa(int(t.TrackCount)), wCount),
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
