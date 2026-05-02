package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func (m Model) isPlaylistSection() bool {
	return m.navIdx == 1 || m.navIdx == 3
}

func (m Model) isTrackSection() bool {
	return m.navIdx == 2 || m.navIdx == 4
}

func (m Model) showingPlaylistTracks() bool {
	return m.playlistTracks && m.playlistTracksSection == m.navIdx
}

func (m Model) currentTracks() []api.Track {
	if m.navIdx == 0 {
		return m.searchTracks
	}
	if m.isTrackSection() {
		if m.navIdx == 2 {
			return m.myTracks
		}
		return m.likedTracks
	}
	if m.showingPlaylistTracks() {
		playlists := m.currentPlaylists()
		if m.playlistIdx >= 0 && m.playlistIdx < len(playlists) {
			return playlists[m.playlistIdx].Tracks
		}
	}
	return nil
}

func (m Model) currentPlaylists() []api.Playlist {
	switch m.navIdx {
	case 1:
		return m.myPlaylists
	case 3:
		return m.likedPlaylists
	case 0:
		return m.searchPlaylists
	default:
		return nil
	}
}

func (m *Model) clearPlaylistTracksIfNavChanged(prev int) {
	if m.navIdx != prev && m.playlistTracks {
		if !m.isPlaylistSection() || m.playlistTracksSection != m.navIdx {
			m.playlistTracks = false
		}
	}
}

func (m Model) listKind() ListKind {
	if m.navIdx == 0 {
		if m.searchMode == SearchTracks {
			return ListTracks
		}
		return ListPlaylists
	}
	if m.isTrackSection() || m.showingPlaylistTracks() {
		return ListTracks
	}
	if m.isPlaylistSection() {
		return ListPlaylists
	}
	return ListNone
}

func (m *Model) moveListDown() {
	if m.listKind() == ListTracks {
		tracks := m.currentTracks()
		if len(tracks) > 0 {
			m.trackIdx = (m.trackIdx + 1) % len(tracks)
		}
		return
	}
	if m.listKind() == ListPlaylists {
		playlists := m.currentPlaylists()
		if len(playlists) > 0 {
			m.playlistIdx = (m.playlistIdx + 1) % len(playlists)
		}
	}
}

func (m *Model) moveListUp() {
	if m.listKind() == ListTracks {
		tracks := m.currentTracks()
		if len(tracks) > 0 {
			m.trackIdx--
			if m.trackIdx < 0 {
				m.trackIdx = len(tracks) - 1
			}
		}
		return
	}
	if m.listKind() == ListPlaylists {
		playlists := m.currentPlaylists()
		if len(playlists) > 0 {
			m.playlistIdx--
			if m.playlistIdx < 0 {
				m.playlistIdx = len(playlists) - 1
			}
		}
	}
}

func (m Model) selectedTrack() *api.Track {
	tracks := m.currentTracks()
	if m.trackIdx < 0 || m.trackIdx >= len(tracks) {
		return nil
	}
	t := tracks[m.trackIdx]
	return &t
}

func (m Model) selectedPlaylist() *api.Playlist {
	playlists := m.currentPlaylists()
	if m.playlistIdx < 0 || m.playlistIdx >= len(playlists) {
		return nil
	}
	p := playlists[m.playlistIdx]
	return &p
}

func (m Model) isOwnedPlaylist(p api.Playlist) bool {
	if m.currentUser.ID == 0 {
		return false
	}
	if p.UserID != 0 {
		return p.UserID == m.currentUser.ID
	}
	return p.User.ID == m.currentUser.ID
}

func (m Model) isTrackLiked(id int) bool {
	if id == 0 {
		return false
	}
	return m.likedTrackIDs[id]
}

func (m Model) isPlaylistLiked(id int) bool {
	if id == 0 {
		return false
	}
	return m.likedPlaylistIDs[id]
}

func (m Model) maybeFetchTrackDetailsCmd() tea.Cmd {
	if m.listKind() != ListTracks {
		return nil
	}
	t := m.selectedTrack()
	if t == nil || t.ID == 0 {
		return nil
	}
	if t.Bpm > 0 {
		return nil
	}
	if m.trackDetailFetched[t.ID] {
		return nil
	}
	m.trackDetailFetched[t.ID] = true
	return func() tea.Msg {
		if m.api == nil {
			return trackDetailMsg{err: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		tr, err := m.api.TrackByID(ctx, m.token, t.ID)
		if err != nil {
			return trackDetailMsg{err: err}
		}
		return trackDetailMsg{track: *tr}
	}
}

func (m *Model) applyTrackDetail(t api.Track) {
	if t.ID == 0 {
		return
	}
	m.myTracks = updateTrackSlice(m.myTracks, t)
	m.likedTracks = updateTrackSlice(m.likedTracks, t)
	m.searchTracks = updateTrackSlice(m.searchTracks, t)
	updateTracksInPlaylists(m.myPlaylists, t)
	updateTracksInPlaylists(m.likedPlaylists, t)
	updateTracksInPlaylists(m.searchPlaylists, t)
	if m.playlistPlayActive {
		for i := range m.playlistPlayTracks {
			if m.playlistPlayTracks[i].ID == t.ID {
				m.playlistPlayTracks[i] = t
			}
		}
	}
}
