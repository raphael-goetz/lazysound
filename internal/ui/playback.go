package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func (m *Model) applyDaemonState(st *daemon.State) {
	if st == nil {
		return
	}
	prevID := 0
	if m.nowPlaying != nil {
		prevID = m.nowPlaying.ID
	}
	if st.Track != nil {
		m.nowPlaying = st.Track
	} else if !st.Playing {
		m.nowPlaying = nil
	}
	m.volume = st.Volume
	m.shuffleEnabled = st.Shuffle
	m.repeatEnabled = st.Repeat
	m.playPaused = st.Paused
	if st.Playing {
		if st.DurationMs > 0 {
			m.playTotal = st.DurationMs
		}
		if st.ElapsedMs >= 0 {
			m.playElapsed = st.ElapsedMs
		}
		if m.playStart.IsZero() || (st.Track != nil && st.Track.ID != prevID) {
			m.playStart = time.Now()
		}
	} else {
		m.playStart = time.Time{}
		m.playElapsed = 0
		m.playTotal = 0
	}
}

func (m Model) daemonStatusCmd() tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonStatusMsg{err: fmt.Errorf("daemon not available")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		st, err := m.daemon.Status(ctx)
		return daemonStatusMsg{err: err, state: st}
	}
}

func (m Model) playSelectedCmd(session int) tea.Cmd {
	t := m.selectedTrack()
	if t == nil {
		return nil
	}
	return m.playTrackCmd(*t, session)
}

func (m Model) stopCmd() tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "stop", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "stop", err: err, session: m.playSession}
		}
		state, err := m.daemon.Stop(ctx)
		return daemonResultMsg{action: "stop", err: err, state: state, session: m.playSession}
	}
}

func (m Model) restartCmd(session int) tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "restart", err: fmt.Errorf("daemon not available"), session: session}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "restart", err: err, session: session}
		}
		state, err := m.daemon.Restart(ctx)
		return daemonResultMsg{action: "restart", err: err, state: state, session: session}
	}
}

func (m Model) volumeCmd(delta int) tea.Cmd {
	return func() tea.Msg {
		vol := m.volume + delta
		if vol < 0 {
			vol = 0
		}
		if vol > 100 {
			vol = 100
		}
		if m.daemon == nil {
			return daemonResultMsg{action: "volume", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "volume", err: err, session: m.playSession}
		}
		state, err := m.daemon.SetVolume(ctx, vol)
		return daemonResultMsg{action: "volume", err: err, state: state, session: m.playSession}
	}
}

func (m Model) setShuffleCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "shuffle", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "shuffle", err: err, session: m.playSession}
		}
		state, err := m.daemon.SetShuffle(ctx, enabled)
		return daemonResultMsg{action: "shuffle", err: err, state: state, session: m.playSession}
	}
}

func (m Model) setRepeatCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "repeat", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "repeat", err: err, session: m.playSession}
		}
		state, err := m.daemon.SetRepeat(ctx, enabled)
		return daemonResultMsg{action: "repeat", err: err, state: state, session: m.playSession}
	}
}

func (m Model) playTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return playTickMsg{} })
}

func (m Model) pauseCmd() tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "pause", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "pause", err: err, session: m.playSession}
		}
		state, err := m.daemon.TogglePause(ctx)
		return daemonResultMsg{action: "pause", err: err, state: state, session: m.playSession}
	}
}

func (m Model) seekCmd(delta int) tea.Cmd {
	return func() tea.Msg {
		if delta > 0 {
			m.playElapsed += int64(delta) * 1000
		} else {
			m.playElapsed += int64(delta) * 1000
			if m.playElapsed < 0 {
				m.playElapsed = 0
			}
		}
		if m.daemon == nil {
			return daemonResultMsg{action: "seek", err: fmt.Errorf("daemon not available"), session: m.playSession}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "seek", err: err, session: m.playSession}
		}
		state, err := m.daemon.Seek(ctx, delta)
		return daemonResultMsg{action: "seek", err: err, state: state, session: m.playSession}
	}
}

func (m Model) playTrackCmd(t api.Track, session int) tea.Cmd {
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "play", err: fmt.Errorf("daemon not available"), session: session, trackID: t.ID}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "play", err: err, session: session, trackID: t.ID}
		}
		state, err := m.daemon.PlayTrack(ctx, m.token, t)
		return daemonResultMsg{action: "play", err: err, state: state, session: session, trackID: t.ID}
	}
}

func (m Model) playQueueCmd(tracks []api.Track, start int, session int) tea.Cmd {
	trackID := 0
	if start >= 0 && start < len(tracks) {
		trackID = tracks[start].ID
	}
	return func() tea.Msg {
		if m.daemon == nil {
			return daemonResultMsg{action: "play", err: fmt.Errorf("daemon not available"), session: session, trackID: trackID}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.daemon.EnsureRunning(ctx); err != nil {
			return daemonResultMsg{action: "play", err: err, session: session, trackID: trackID}
		}
		state, err := m.daemon.PlayQueue(ctx, m.token, tracks, start)
		return daemonResultMsg{action: "play", err: err, state: state, session: session, trackID: trackID}
	}
}

func (m *Model) playAnyCmd(session int) tea.Cmd {
	if m.listKind() == ListTracks {
		m.playlistPlayActive = false
		return m.playSelectedCmd(session)
	}
	if m.listKind() == ListPlaylists {
		p := m.selectedPlaylist()
		if p == nil || len(p.Tracks) == 0 {
			return nil
		}
		m.playlistPlayTracks = p.Tracks
		if m.shuffleEnabled {
			m.playlistPlayTracks = shuffleTracks(p.Tracks)
		}
		m.playlistPlayIdx = 0
		m.playlistPlayActive = true
		m.trackIdx = 0
		m.playlistTracks = true
		m.playlistTracksSection = m.navIdx
		return m.playQueueCmd(m.playlistPlayTracks, 0, session)
	}
	if m.showingPlaylistTracks() {
		tracks := m.currentTracks()
		if len(tracks) == 0 {
			return nil
		}
		m.playlistPlayTracks = tracks
		if m.shuffleEnabled {
			m.playlistPlayTracks = shuffleTracks(tracks)
		}
		m.playlistPlayIdx = m.trackIdx
		m.playlistPlayActive = true
		return m.playQueueCmd(m.playlistPlayTracks, m.playlistPlayIdx, session)
	}
	return nil
}
