package ui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func (m Model) badgeForTrack(t api.Track) string {
	if badge, ok := m.trackBadges[t.ID]; ok {
		return normalizeBadge(badge)
	}
	return normalizeBadge(t.Access)
}

func normalizeBadge(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "playable":
		return "playable"
	case "preview":
		return "preview"
	case "blocked":
		return "blocked"
	default:
		return ""
	}
}

func (m *Model) probeTracksCmd(tracks []api.Track) tea.Cmd {
	if m.daemon == nil || len(tracks) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(tracks))
	for _, t := range tracks {
		if t.ID == 0 {
			continue
		}
		if _, ok := m.probeInFlight[t.ID]; ok {
			continue
		}
		if _, ok := m.trackBadges[t.ID]; ok {
			continue
		}
		m.probeInFlight[t.ID] = true
		cmds = append(cmds, m.probeTrackCmd(t))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) probeTrackCmd(t api.Track) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		pr, err := m.daemon.ProbeTrack(ctx, m.token, t)
		if err != nil {
			return trackProbeMsg{trackID: t.ID, err: err}
		}
		if pr == nil {
			return trackProbeMsg{trackID: t.ID}
		}
		return trackProbeMsg{trackID: t.ID, badge: normalizeBadge(pr.Badge)}
	}
}
