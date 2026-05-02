package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/raphael-goetz/lazysound/lib/uikit/component"
	"github.com/raphael-goetz/lazysound/lib/uikit/layout"
)

func (m Model) View() string {
	s := m.styles
	if m.w == 0 || m.h == 0 {
		return "loading..."
	}

	header := renderHeader(s, m, m.w)
	player := renderPlayerBar(s, m, m.w)
	cmd := component.Command(s, m.navIdx == 0, m.searchActive)

	headerH := lipgloss.Height(header)
	playerH := lipgloss.Height(player)
	cmdH := lipgloss.Height(cmd)

	bodyH := layout.BodyHeight(m.h, headerH, playerH, cmdH)
	navW, listW, inspectW := layout.PaneWidths(m.w)

	var nav, list, inspect string
	if navW > 0 {
		navContent := component.Nav(s, m.navItems, m.navIdx, m.focus == FocusNav, navW-2, bodyH-2)
		nav = component.BorderedBox(s.BorderFor(m.focus == FocusNav), s.LabelFor(m.focus == FocusNav), "Navigation", navContent, navW, bodyH)
	}

	if listW > 0 {
		title := "Tracks"
		switch m.listKind() {
		case ListPlaylists:
			title = "Playlists"
		case ListTracks:
			title = "Tracks"
		}
		if m.navIdx == 0 {
			title = "Search"
		}
		var listContent string
		if m.navIdx == 0 {
			listContent = m.renderSearchList(s, listW-2, bodyH-2)
		} else if m.listKind() == ListTracks {
			listContent = component.TrackTable(s, m.currentTracks(), m.trackBadges, m.trackIdx, m.focus == FocusList, listW-2, bodyH-2)
		} else if m.listKind() == ListPlaylists {
			listContent = component.PlaylistTable(s, m.currentPlaylists(), m.playlistIdx, m.focus == FocusList, listW-2, bodyH-2)
		} else {
			listContent = ""
		}
		list = component.BorderedBox(s.BorderFor(m.focus == FocusList), s.LabelFor(m.focus == FocusList), title, listContent, listW, bodyH)
	}

	if inspectW > 0 {
		info := renderInspectInfo(s, m, inspectW-2, bodyH-2)
		inspect = component.BorderedBox(s.BorderFor(false), s.LabelFor(false), "Info", info, inspectW, bodyH)
	}

	body := joinPanes(nav, list, inspect)
	if m.actionMenuActive {
		body = m.renderActionMenu(s, m.w, bodyH)
	}
	if m.actionPromptActive {
		body = m.renderActionPrompt(s, m.w, bodyH)
	}
	if m.actionConfirmActive {
		body = m.renderActionConfirm(s, m.w, bodyH)
	}
	if m.actionPickActive {
		body = m.renderPlaylistPicker(s, m.w, bodyH)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, player, cmd)
}
