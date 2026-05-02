package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raphael-goetz/lazysound/lib/uikit/component"
	"github.com/raphael-goetz/lazysound/lib/uikit/math"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

func (m *Model) openActionMenu() {
	m.actionItems = m.actionItemsForContext()
	m.actionIdx = 0
	if len(m.actionItems) == 0 {
		m.setStatus("no actions available")
		return
	}
	m.actionMenuActive = true
}

func (m Model) actionItemsForContext() []ActionItem {
	items := []ActionItem{}
	kind := m.listKind()
	if kind == ListPlaylists {
		p := m.selectedPlaylist()
		if m.navIdx == 1 {
			items = append(items, ActionItem{Label: "Create playlist", Kind: ActionCreatePlaylist})
		}
		if p != nil {
			if m.isOwnedPlaylist(*p) {
				items = append(items, ActionItem{Label: "Rename playlist", Kind: ActionRenamePlaylist})
				items = append(items, ActionItem{Label: "Delete playlist", Kind: ActionDeletePlaylist})
			}
			if m.isPlaylistLiked(p.ID) {
				items = append(items, ActionItem{Label: "Unlike playlist", Kind: ActionUnlikePlaylist})
			} else {
				items = append(items, ActionItem{Label: "Like playlist", Kind: ActionLikePlaylist})
			}
		}
		return items
	}
	if kind == ListTracks {
		t := m.selectedTrack()
		if t == nil {
			return items
		}
		if m.showingPlaylistTracks() {
			if p := m.selectedPlaylist(); p != nil && m.isOwnedPlaylist(*p) {
				items = append(items, ActionItem{Label: "Remove from playlist", Kind: ActionRemoveFromPlaylist})
			}
		}
		if len(m.myPlaylists) > 0 {
			items = append(items, ActionItem{Label: "Add to playlist", Kind: ActionAddToPlaylist})
		}
		if m.isTrackLiked(t.ID) {
			items = append(items, ActionItem{Label: "Unlike track", Kind: ActionUnlikeTrack})
		} else {
			items = append(items, ActionItem{Label: "Like track", Kind: ActionLikeTrack})
		}
	}
	return items
}

func (m Model) renderActionMenu(s style.Styles, width, height int) string {
	if len(m.actionItems) == 0 {
		return m.renderPopup(s, "Actions", fitWidth(s.Muted.Render("No actions available."), width-4), width, height)
	}
	items := make([]string, 0, len(m.actionItems))
	for _, it := range m.actionItems {
		items = append(items, it.Label)
	}
	items = append(items, "Close")
	popupW := math.Clamp(width-4, 28, 48)
	popupH := math.Clamp(height-4, 6, min(height-2, len(items)+4))
	menu := component.Nav(s, items, m.actionIdx, true, popupW-2, popupH-3)
	content := strings.Join([]string{
		menu,
		s.Muted.Render("enter select • esc close"),
	}, "\n")
	return m.renderPopupSized(s, "Actions", content, width, height, popupW, popupH)
}

func (m Model) renderActionPrompt(s style.Styles, width, height int) string {
	title := "Input"
	switch m.actionPromptKind {
	case ActionCreatePlaylist:
		title = "Create playlist"
	case ActionRenamePlaylist:
		title = "Rename playlist"
	}
	content := strings.Join([]string{
		m.actionPrompt.View(),
		s.Muted.Render("enter to confirm, esc to cancel"),
	}, "\n")
	return m.renderPopupSized(s, title, content, width, height, math.Clamp(width-4, 32, 56), math.Clamp(height-4, 6, 10))
}

func (m Model) renderActionConfirm(s style.Styles, width, height int) string {
	text := "Confirm"
	switch m.actionConfirmKind {
	case ActionDeletePlaylist:
		if p := m.selectedPlaylist(); p != nil {
			text = "Delete playlist: " + p.Title
		} else {
			text = "Delete playlist"
		}
	case ActionRemoveFromPlaylist:
		if t := m.selectedTrack(); t != nil {
			text = "Remove track: " + t.Title
		} else {
			text = "Remove track"
		}
	}
	content := strings.Join([]string{
		s.Muted.Render(text),
		s.Muted.Render("y to confirm, n or esc to cancel"),
	}, "\n")
	return m.renderPopupSized(s, "Confirm", content, width, height, math.Clamp(width-4, 32, 56), math.Clamp(height-4, 6, 8))
}

func (m Model) renderPlaylistPicker(s style.Styles, width, height int) string {
	if len(m.myPlaylists) == 0 {
		return m.renderPopup(s, "Pick playlist", fitWidth(s.Muted.Render("No playlists found."), width-4), width, height)
	}
	names := make([]string, 0, len(m.myPlaylists))
	for _, p := range m.myPlaylists {
		names = append(names, p.Title)
	}
	popupW := math.Clamp(width-4, 28, 48)
	popupH := math.Clamp(height-4, 6, min(height-2, len(names)+4))
	content := component.Nav(s, names, m.actionPickIdx, true, popupW-2, popupH-2)
	return m.renderPopupSized(s, "Pick playlist", content, width, height, popupW, popupH)
}

func (m Model) renderPopup(s style.Styles, title, content string, width, height int) string {
	return m.renderPopupSized(s, title, content, width, height, math.Clamp(width-4, 28, width-2), math.Clamp(height-4, 6, height-2))
}

func (m Model) renderPopupSized(s style.Styles, title, content string, width, height, popupW, popupH int) string {
	if width < 10 || height < 5 {
		return content
	}
	box := component.BorderedBox(s.BorderFocus, s.LabelFocus, title, content, popupW, popupH)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) runActionWithInput(kind ActionKind, input string) tea.Cmd {
	switch kind {
	case ActionCreatePlaylist:
		return m.actionCmdCreatePlaylist(input)
	case ActionRenamePlaylist:
		return m.actionCmdRenamePlaylist(input)
	}
	return nil
}

func (m Model) runAction(kind ActionKind) tea.Cmd {
	switch kind {
	case ActionDeletePlaylist:
		return m.actionCmdDeletePlaylist()
	case ActionLikeTrack:
		return m.actionCmdLikeTrack(true)
	case ActionUnlikeTrack:
		return m.actionCmdLikeTrack(false)
	case ActionLikePlaylist:
		return m.actionCmdLikePlaylist(true)
	case ActionUnlikePlaylist:
		return m.actionCmdLikePlaylist(false)
	case ActionRemoveFromPlaylist:
		return m.actionCmdRemoveFromPlaylist()
	}
	return nil
}

func (m Model) runActionPick(kind ActionKind, idx int) tea.Cmd {
	switch kind {
	case ActionAddToPlaylist:
		return m.actionCmdAddToPlaylist(idx)
	}
	return nil
}

func (m Model) actionCmdCreatePlaylist(title string) tea.Cmd {
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionCreatePlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		pl, err := m.api.CreatePlaylist(ctx, m.token, title, "", "private", nil)
		return actionResultMsg{kind: ActionCreatePlaylist, errAction: err, playlist: pl}
	}
}

func (m Model) actionCmdRenamePlaylist(title string) tea.Cmd {
	p := m.selectedPlaylist()
	if p == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionRenamePlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		pl, err := m.api.UpdatePlaylist(ctx, m.token, p.ID, title, p.Description, nil)
		return actionResultMsg{kind: ActionRenamePlaylist, errAction: err, playlist: pl}
	}
}

func (m Model) actionCmdDeletePlaylist() tea.Cmd {
	p := m.selectedPlaylist()
	if p == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionDeletePlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := m.api.DeletePlaylist(ctx, m.token, p.ID)
		return actionResultMsg{kind: ActionDeletePlaylist, errAction: err, playlistID: p.ID}
	}
}

func (m Model) actionCmdLikeTrack(like bool) tea.Cmd {
	t := m.selectedTrack()
	if t == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionLikeTrack, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		var err error
		if like {
			err = m.api.LikeTrack(ctx, m.token, t.ID)
		} else {
			err = m.api.UnlikeTrack(ctx, m.token, t.ID)
		}
		kind := ActionUnlikeTrack
		if like {
			kind = ActionLikeTrack
		}
		return m.refreshLikedTracksMsg(ctx, err, kind, t.ID)
	}
}

func (m Model) actionCmdLikePlaylist(like bool) tea.Cmd {
	p := m.selectedPlaylist()
	if p == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionLikePlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		var err error
		if like {
			err = m.api.LikePlaylist(ctx, m.token, p.ID)
		} else {
			err = m.api.UnlikePlaylist(ctx, m.token, p.ID)
		}
		kind := ActionUnlikePlaylist
		if like {
			kind = ActionLikePlaylist
		}
		return m.refreshLikedPlaylistsMsg(ctx, err, kind, p.ID)
	}
}

func (m Model) actionCmdAddToPlaylist(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.myPlaylists) {
		return nil
	}
	p := m.myPlaylists[idx]
	t := m.selectedTrack()
	if t == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionAddToPlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		trackIDs := playlistTrackIDs(p.Tracks)
		if containsInt(trackIDs, t.ID) {
			return actionResultMsg{kind: ActionAddToPlaylist, errAction: fmt.Errorf("track already in playlist")}
		}
		trackIDs = append(trackIDs, t.ID)
		pl, err := m.api.UpdatePlaylist(ctx, m.token, p.ID, p.Title, p.Description, trackIDs)
		return actionResultMsg{kind: ActionAddToPlaylist, errAction: err, playlist: pl}
	}
}

func (m Model) actionCmdRemoveFromPlaylist() tea.Cmd {
	p := m.selectedPlaylist()
	if p == nil {
		return nil
	}
	t := m.selectedTrack()
	if t == nil {
		return nil
	}
	return func() tea.Msg {
		if m.api == nil {
			return actionResultMsg{kind: ActionRemoveFromPlaylist, errAction: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		trackIDs := playlistTrackIDs(p.Tracks)
		trackIDs = removeInt(trackIDs, t.ID)
		pl, err := m.api.UpdatePlaylist(ctx, m.token, p.ID, p.Title, p.Description, trackIDs)
		return actionResultMsg{kind: ActionRemoveFromPlaylist, errAction: err, playlist: pl}
	}
}

func (m *Model) setStatus(msg string) {
	m.status = msg
}

func (m Model) refreshLikedTracksMsg(ctx context.Context, err error, kind ActionKind, trackID int) actionResultMsg {
	msg := actionResultMsg{kind: kind, errAction: err, trackID: trackID, refreshTracks: true}
	if err != nil {
		return msg
	}
	tr, err2 := m.api.LikedTracks(ctx, m.token)
	if err2 != nil {
		msg.errRefresh = err2
		return msg
	}
	msg.likedTracks = tr.Collection
	return msg
}

func (m Model) refreshLikedPlaylistsMsg(ctx context.Context, err error, kind ActionKind, playlistID int) actionResultMsg {
	msg := actionResultMsg{kind: kind, errAction: err, playlistID: playlistID, refreshPlaylists: true}
	if err != nil {
		return msg
	}
	pl, err2 := m.api.LikedPlaylists(ctx, m.token)
	if err2 != nil {
		msg.errRefresh = err2
		return msg
	}
	msg.likedPlaylists = pl.Collection
	return msg
}
