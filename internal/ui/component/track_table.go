package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/raphael-goetz/lazysound/internal/api"
	"github.com/raphael-goetz/lazysound/internal/math"
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

func TrackTable(s style.Styles, tracks []api.Track, sel int, focused bool, width, height int) string {
	fwidth := float64(width)
	wTitle := int(fwidth * (4.0 / 8.0))
	wArtist := int(fwidth * (2.0 / 8.0))
	wDuration := int(fwidth * (1.0 / 8.0))
	wBPM := int(fwidth * (1.0 / 8.0))

	format := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds ",
		wTitle,
		wArtist,
		wDuration,
		wBPM,
	)

	header := s.Muted.Render(fitWidth(fmt.Sprintf(format, "Title", "Artist(s)", "Durtion", "BPM"), width))
	sep := s.Muted.Render(fitWidth(strings.Repeat("─", width), width))

	length := len(tracks)
	maxRows := max(height-2, 3)

	start := math.Clamp(sel-maxRows/2, 0, max(0, length-maxRows))
	end := min(length, start+maxRows)

	lines := []string{header, sep}
	for i := start; i < end; i++ {
		t := tracks[i]
		row := fmt.Sprintf(format,
			math.Truncate(t.Title, wTitle),
			math.Truncate(t.User.Username, wArtist),
			math.Truncate(Time(t.Duration), wDuration),
			math.Truncate(strconv.Itoa(t.Bpm), wBPM),
		)

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
