package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tea.KeyMsg:
		if m.actionPromptActive {
			switch msg.String() {
			case "esc":
				m.actionPromptActive = false
				m.actionPrompt.Blur()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.actionPrompt.Value())
				m.actionPromptActive = false
				m.actionPrompt.Blur()
				if val == "" {
					m.setStatus("empty input")
					return m, nil
				}
				return m, m.runActionWithInput(m.actionPromptKind, val)
			}
			var cmd tea.Cmd
			m.actionPrompt, cmd = m.actionPrompt.Update(msg)
			return m, cmd
		}
		if m.actionConfirmActive {
			switch msg.String() {
			case "y", "Y":
				m.actionConfirmActive = false
				return m, m.runAction(m.actionConfirmKind)
			case "n", "N", "esc":
				m.actionConfirmActive = false
				return m, nil
			}
			return m, nil
		}
		if m.actionPickActive {
			switch msg.String() {
			case "esc":
				m.actionPickActive = false
				return m, nil
			case "j", "down":
				if len(m.myPlaylists) > 0 {
					m.actionPickIdx = (m.actionPickIdx + 1) % len(m.myPlaylists)
				}
			case "k", "up":
				if len(m.myPlaylists) > 0 {
					m.actionPickIdx--
					if m.actionPickIdx < 0 {
						m.actionPickIdx = len(m.myPlaylists) - 1
					}
				}
			case "enter":
				m.actionPickActive = false
				return m, m.runActionPick(m.actionPickKind, m.actionPickIdx)
			}
			return m, nil
		}
		if m.actionMenuActive {
			switch msg.String() {
			case "esc":
				m.actionMenuActive = false
				return m, nil
			case "j", "down":
				total := len(m.actionItems) + 1
				m.actionIdx = (m.actionIdx + 1) % total
			case "k", "up":
				total := len(m.actionItems) + 1
				m.actionIdx--
				if m.actionIdx < 0 {
					m.actionIdx = total - 1
				}
			case "enter":
				if m.actionIdx == len(m.actionItems) {
					m.actionMenuActive = false
					return m, nil
				}
				item := m.actionItems[m.actionIdx]
				m.actionMenuActive = false
				switch item.Kind {
				case ActionCreatePlaylist, ActionRenamePlaylist:
					m.actionPromptKind = item.Kind
					m.actionPromptActive = true
					m.actionPrompt.Focus()
					if item.Kind == ActionRenamePlaylist {
						if p := m.selectedPlaylist(); p != nil {
							m.actionPrompt.SetValue(p.Title)
							m.actionPrompt.CursorEnd()
						}
					} else {
						m.actionPrompt.SetValue("")
					}
					return m, nil
				case ActionDeletePlaylist, ActionRemoveFromPlaylist:
					m.actionConfirmKind = item.Kind
					m.actionConfirmActive = true
					return m, nil
				case ActionAddToPlaylist:
					m.actionPickKind = item.Kind
					m.actionPickIdx = 0
					m.actionPickActive = true
					return m, nil
				default:
					return m, m.runAction(item.Kind)
				}
			}
			return m, nil
		}

		if m.searchActive {
			switch msg.String() {
			case "esc":
				m.searchActive = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				query := strings.TrimSpace(m.searchInput.Value())
				if query == "" {
					m.searchActive = false
					m.searchInput.Blur()
					return m, nil
				}
				m.searchActive = false
				m.searchInput.Blur()
				m.searchQuery = query
				m.searchLoading = true
				m.searchErr = ""
				m.trackIdx = 0
				m.playlistIdx = 0
				return m, m.searchCmd(query)
			}
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		key := msg.String()
		if keyIs(key, m.keymap.Quit) {
			return m, tea.Quit
		}
		if keyIs(key, m.keymap.Tab) {
			m.focus = (m.focus + 1) % 2
			return m, nil
		}
		if keyIs(key, m.keymap.Left) {
			if m.focus > 0 {
				m.focus--
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Right) {
			if m.focus < 1 {
				m.focus++
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Search) {
			if m.navIdx == 0 {
				m.searchActive = true
				m.searchInput.Focus()
				m.searchInput.SetValue(m.searchQuery)
				m.searchInput.CursorEnd()
			}
			return m, nil
		}
		if keyIs(key, m.keymap.ActionMenu) {
			if m.focus == FocusList {
				m.openActionMenu()
			}
			return m, nil
		}
		if keyIs(key, m.keymap.SearchTracks) {
			if m.navIdx == 0 {
				m.searchMode = SearchTracks
				m.playlistTracks = false
				m.trackIdx = 0
			}
			return m, nil
		}
		if keyIs(key, m.keymap.SearchPlay) {
			if m.navIdx == 0 {
				m.searchMode = SearchPlaylists
				m.playlistTracks = false
				m.playlistIdx = 0
				return m, nil
			}
		}
		if keyIs(key, m.keymap.Down) {
			switch m.focus {
			case FocusNav:
				prev := m.navIdx
				if len(m.navItems) > 0 {
					m.navIdx = (m.navIdx + 1) % len(m.navItems)
				}
				m.clearPlaylistTracksIfNavChanged(prev)
			case FocusList:
				m.moveListDown()
				if cmd := m.maybeFetchTrackDetailsCmd(); cmd != nil {
					return m, cmd
				}
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Up) {
			switch m.focus {
			case FocusNav:
				prev := m.navIdx
				if len(m.navItems) > 0 {
					m.navIdx--
					if m.navIdx < 0 {
						m.navIdx = len(m.navItems) - 1
					}
				}
				m.clearPlaylistTracksIfNavChanged(prev)
			case FocusList:
				m.moveListUp()
				if cmd := m.maybeFetchTrackDetailsCmd(); cmd != nil {
					return m, cmd
				}
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Open) {
			if m.focus == FocusList && m.listKind() == ListPlaylists && !m.showingPlaylistTracks() && len(m.currentPlaylists()) > 0 {
				m.playlistTracks = true
				m.playlistTracksSection = m.navIdx
				m.trackIdx = 0
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Back) {
			if m.playlistTracks {
				m.playlistTracks = false
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Play) {
			if m.focus == FocusList && m.navIdx != 0 {
				if m.playPending {
					m.setStatus("player busy")
					return m, nil
				}
				m.playSession++
				m.playPending = true
				return m, m.playAnyCmd(m.playSession)
			}
			return m, nil
		}
		if keyIs(key, m.keymap.Stop) {
			return m, m.pauseCmd()
		}
		if keyIs(key, m.keymap.Restart) {
			if m.playPending {
				m.setStatus("player busy")
				return m, nil
			}
			m.playSession++
			m.playPending = true
			return m, m.restartCmd(m.playSession)
		}
		if keyIs(key, m.keymap.VolumeUp) {
			return m, m.volumeCmd(5)
		}
		if keyIs(key, m.keymap.VolumeDown) {
			return m, m.volumeCmd(-5)
		}
		if keyIs(key, m.keymap.PauseToggle) {
			return m, m.pauseCmd()
		}
		if keyIs(key, m.keymap.SeekForward) {
			return m, m.seekCmd(m.seekStepSeconds)
		}
		if keyIs(key, m.keymap.SeekBack) {
			return m, m.seekCmd(-m.seekStepSeconds)
		}
		if keyIs(key, m.keymap.Shuffle) {
			m.shuffleEnabled = !m.shuffleEnabled
			if m.shuffleEnabled {
				m.setStatus("shuffle on")
			} else {
				m.setStatus("shuffle off")
			}
			return m, m.setShuffleCmd(m.shuffleEnabled)
		}
		if keyIs(key, m.keymap.Repeat) {
			m.repeatEnabled = !m.repeatEnabled
			if m.repeatEnabled {
				m.setStatus("repeat on")
			} else {
				m.setStatus("repeat off")
			}
			return m, m.setRepeatCmd(m.repeatEnabled)
		}
	case searchResultMsg:
		m.searchLoading = false
		m.searchTracks = msg.tracks
		m.searchPlaylists = msg.playlists
		m.trackIdx = 0
		m.playlistIdx = 0
		if msg.errTracks != nil || msg.errPlaylists != nil {
			errText := "search failed"
			if msg.errTracks != nil {
				errText = msg.errTracks.Error()
			} else if msg.errPlaylists != nil {
				errText = msg.errPlaylists.Error()
			}
			m.searchErr = errText
			m.setStatus("search failed")
		}
	case actionResultMsg:
		m.actionPromptActive = false
		m.actionConfirmActive = false
		m.actionPickActive = false
		if msg.errAction != nil {
			m.setStatus("action failed: " + msg.errAction.Error())
			return m, nil
		}
		switch msg.kind {
		case ActionCreatePlaylist:
			if msg.playlist != nil {
				m.myPlaylists = updatePlaylistSlice(m.myPlaylists, *msg.playlist)
				if m.navIdx == 1 {
					m.playlistIdx = len(m.myPlaylists) - 1
				}
				m.setStatus("playlist created")
			}
		case ActionRenamePlaylist, ActionAddToPlaylist, ActionRemoveFromPlaylist:
			if msg.playlist != nil {
				m.myPlaylists = updatePlaylistSlice(m.myPlaylists, *msg.playlist)
				m.searchPlaylists = updatePlaylistSlice(m.searchPlaylists, *msg.playlist)
				m.likedPlaylists = updatePlaylistSlice(m.likedPlaylists, *msg.playlist)
				m.setStatus("playlist updated")
			}
		case ActionDeletePlaylist:
			if msg.playlistID > 0 {
				m.myPlaylists = removePlaylistByID(m.myPlaylists, msg.playlistID)
				m.likedPlaylists = removePlaylistByID(m.likedPlaylists, msg.playlistID)
				m.searchPlaylists = removePlaylistByID(m.searchPlaylists, msg.playlistID)
				m.setStatus("playlist deleted")
			}
		case ActionLikeTrack:
			if msg.trackID > 0 {
				m.likedTrackIDs[msg.trackID] = true
				m.setStatus("track liked")
			}
		case ActionUnlikeTrack:
			if msg.trackID > 0 {
				delete(m.likedTrackIDs, msg.trackID)
				if m.navIdx == 4 {
					m.likedTracks = removeTrackByID(m.likedTracks, msg.trackID)
				}
				m.setStatus("track unliked")
			}
		case ActionLikePlaylist:
			if msg.playlistID > 0 {
				m.likedPlaylistIDs[msg.playlistID] = true
				m.setStatus("playlist liked")
			}
		case ActionUnlikePlaylist:
			if msg.playlistID > 0 {
				delete(m.likedPlaylistIDs, msg.playlistID)
				if m.navIdx == 3 {
					m.likedPlaylists = removePlaylistByID(m.likedPlaylists, msg.playlistID)
				}
				m.setStatus("playlist unliked")
			}
		}
		if msg.refreshTracks {
			m.likedTracks = msg.likedTracks
			m.likedTrackIDs = likedTrackMap(msg.likedTracks)
		}
		if msg.refreshPlaylists {
			m.likedPlaylists = msg.likedPlaylists
			m.likedPlaylistIDs = likedPlaylistMap(msg.likedPlaylists)
		}
		if msg.errRefresh != nil {
			m.setStatus("refresh failed: " + msg.errRefresh.Error())
		}
	case daemonResultMsg:
		if msg.session != m.playSession && msg.session != 0 {
			return m, nil
		}
		m.playPending = false
		if msg.err != nil {
			m.setStatus("Playback error: " + msg.err.Error())
			return m, nil
		}
		if msg.state != nil {
			m.applyDaemonState(msg.state)
		}
		switch msg.action {
		case "play":
			m.setStatus("playing")
			return m, m.playTickCmd()
		case "stop":
			m.setStatus("stopped")
		case "restart":
			m.setStatus("restarted")
		case "volume":
			m.setStatus("volume " + fmt.Sprintf("%d", m.volume))
		case "pause":
			if m.playPaused {
				m.playPaused = false
				m.setStatus("playing")
			} else {
				m.playPaused = true
				m.setStatus("paused")
			}
		case "seek":
			m.setStatus("seek")
		case "shuffle":
			if m.shuffleEnabled {
				m.setStatus("shuffle on")
			} else {
				m.setStatus("shuffle off")
			}
		case "repeat":
			if m.repeatEnabled {
				m.setStatus("repeat on")
			} else {
				m.setStatus("repeat off")
			}
		}
	case playTickMsg:
		return m, m.daemonStatusCmd()
	case daemonStatusMsg:
		if msg.err != nil {
			return m, nil
		}
		if msg.state != nil {
			m.applyDaemonState(msg.state)
			if msg.state.Playing || msg.state.Paused {
				return m, m.playTickCmd()
			}
		}
	case trackDetailMsg:
		if msg.err != nil {
			return m, nil
		}
		m.applyTrackDetail(msg.track)
	}
	return m, nil
}
