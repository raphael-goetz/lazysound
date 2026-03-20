package ui

import (
	"context"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/raphael-goetz/lazysound/internal/api"
	"github.com/raphael-goetz/lazysound/internal/ui/component"
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

func (m Model) renderSearchList(s style.Styles, width, height int) string {
	lines := []string{}
	if m.searchActive {
		lines = append(lines, fitWidth(m.searchInput.View(), width))
	} else if m.searchLoading {
		lines = append(lines, fitWidth(s.Muted.Render("Searching..."), width))
	} else if m.searchErr != "" {
		lines = append(lines, fitWidth(s.Danger.Render(m.searchErr), width))
	} else if m.searchQuery == "" {
		lines = append(lines, fitWidth(s.Muted.Render("Press / to search"), width))
	} else {
		lines = append(lines, fitWidth(s.Muted.Render("Query: "+m.searchQuery), width))
	}

	bodyH := height - len(lines)
	if bodyH < 3 {
		bodyH = 3
	}

	switch m.searchMode {
	case SearchTracks:
		lines = append(lines, component.TrackTable(s, m.searchTracks, m.trackIdx, m.focus == FocusList, width, bodyH))
	case SearchPlaylists:
		lines = append(lines, component.PlaylistTable(s, m.searchPlaylists, m.playlistIdx, m.focus == FocusList, width, bodyH))
	}

	return strings.Join(lines, "\n")
}

func (m Model) searchCmd(query string) tea.Cmd {
	mode := m.searchMode
	return func() tea.Msg {
		if m.api == nil || m.token == "" {
			return searchResultMsg{mode: mode, errTracks: errNoSearchClient{}, errPlaylists: errNoSearchClient{}}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var (
			tr   *api.Tracks
			pl   *api.Playlists
			errT error
			errP error
		)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			tr, errT = m.api.SearchTracks(ctx, m.token, query)
		}()
		go func() {
			defer wg.Done()
			pl, errP = m.api.SearchPlaylists(ctx, m.token, query)
		}()
		wg.Wait()

		msg := searchResultMsg{
			mode:         mode,
			errTracks:    errT,
			errPlaylists: errP,
		}
		if tr != nil {
			msg.tracks = tr.Collection
		}
		if pl != nil {
			msg.playlists = pl.Collection
		}
		return msg
	}
}

type errNoSearchClient struct{}

func (e errNoSearchClient) Error() string { return "search client not configured" }

type errUnknownSearchMode struct{}

func (e errUnknownSearchMode) Error() string { return "unknown search mode" }
