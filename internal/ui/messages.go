package ui

import (
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

type actionResultMsg struct {
	kind             ActionKind
	errAction        error
	errRefresh       error
	playlist         *api.Playlist
	trackID          int
	playlistID       int
	likedTracks      []api.Track
	likedPlaylists   []api.Playlist
	refreshTracks    bool
	refreshPlaylists bool
}

type searchResultMsg struct {
	mode         SearchMode
	tracks       []api.Track
	playlists    []api.Playlist
	errTracks    error
	errPlaylists error
}

type daemonResultMsg struct {
	action  string
	err     error
	state   *daemon.State
	session int
}

type daemonStatusMsg struct {
	err   error
	state *daemon.State
}

type playTickMsg struct{}

type trackDetailMsg struct {
	track api.Track
	err   error
}
