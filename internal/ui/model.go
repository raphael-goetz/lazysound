package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

type SearchMode int

const (
	SearchTracks SearchMode = iota
	SearchPlaylists
)

type ListKind int

const (
	ListNone ListKind = iota
	ListTracks
	ListPlaylists
)

type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionCreatePlaylist
	ActionRenamePlaylist
	ActionDeletePlaylist
	ActionLikePlaylist
	ActionUnlikePlaylist
	ActionLikeTrack
	ActionUnlikeTrack
	ActionAddToPlaylist
	ActionRemoveFromPlaylist
)

type ActionItem struct {
	Label string
	Kind  ActionKind
}

type Model struct {
	w, h int

	focus Focus

	navItems []string
	navIdx   int

	// Collections (My / Liked)
	myPlaylists    []api.Playlist
	likedPlaylists []api.Playlist
	playlistIdx    int

	myTracks    []api.Track
	likedTracks []api.Track
	trackIdx    int

	// Playlist drill-down view state
	playlistTracks        bool
	playlistTracksSection int

	// Search state
	searchMode      SearchMode
	searchQuery     string
	searchActive    bool
	searchInput     textinput.Model
	searchLoading   bool
	searchErr       string
	searchTracks    []api.Track
	searchPlaylists []api.Playlist

	// Authenticated user
	currentUser api.User

	// Liked cache (to render like/unlike actions quickly)
	likedTrackIDs    map[int]bool
	likedPlaylistIDs map[int]bool

	// Status line (rendered in player bar)
	status string

	// Quick actions state
	actionMenuActive    bool
	actionItems         []ActionItem
	actionIdx           int
	actionPromptActive  bool
	actionPrompt        textinput.Model
	actionPromptKind    ActionKind
	actionConfirmActive bool
	actionConfirmKind   ActionKind
	actionPickActive    bool
	actionPickKind      ActionKind
	actionPickIdx       int

	api   *api.ApiClient
	token string

	// Playback state
	daemon             *daemon.Client
	volume             int
	nowPlaying         *api.Track
	playSession        int
	playStart          time.Time
	playElapsed        int64
	playTotal          int64
	greeting           string
	playPending        bool
	playPaused         bool
	seekStepSeconds    int
	shuffleEnabled     bool
	repeatEnabled      bool
	keymap             Keymap
	playlistPlayActive bool
	playlistPlayTracks []api.Track
	playlistPlayIdx    int
	trackDetailFetched map[int]bool
	trackBadges        map[int]string
	probeInFlight      map[int]bool

	styles style.Styles

	err string
}

func NewModel(myTracks []api.Track, likedTracks []api.Track, myPlaylists []api.Playlist, likedPlaylists []api.Playlist, me *api.User, client *api.ApiClient, token string, daemonAddr string, keymap Keymap, styles style.Styles) Model {
	in := textinput.New()
	in.Prompt = "Search: "
	in.Placeholder = "type and press enter"
	in.CharLimit = 120
	in.Width = 40
	ap := textinput.New()
	ap.Prompt = "> "
	ap.CharLimit = 120
	ap.Width = 40

	greetings := []string{
		"Tune in, %s",
		"Happy listening, %s",
		"Welcome back, %s",
		"Good vibes, %s",
		"Enjoy the mix, %s",
		"Stay loud, %s",
	}

	lt := make(map[int]bool)
	for _, t := range likedTracks {
		if t.ID > 0 {
			lt[t.ID] = true
		}
	}
	lp := make(map[int]bool)
	for _, p := range likedPlaylists {
		if p.ID > 0 {
			lp[p.ID] = true
		}
	}

	var user api.User
	if me != nil {
		user = *me
	}
	name := "Listener"
	if strings.TrimSpace(user.Username) != "" {
		name = user.Username
	}
	rand.Seed(time.Now().UnixNano())
	greeting := fmt.Sprintf(greetings[rand.Intn(len(greetings))], name)

	dc := daemon.NewClient(daemonAddr)

	return Model{
		focus: FocusList,
		navItems: []string{
			"Search",
			"My Playlists",
			"My Tracks",
			"Liked Playlists",
			"Liked Tracks",
		},
		navIdx:             1,
		myTracks:           myTracks,
		likedTracks:        likedTracks,
		myPlaylists:        myPlaylists,
		likedPlaylists:     likedPlaylists,
		searchMode:         SearchTracks,
		searchInput:        in,
		actionPrompt:       ap,
		currentUser:        user,
		likedTrackIDs:      lt,
		likedPlaylistIDs:   lp,
		api:                client,
		token:              token,
		daemon:             dc,
		volume:             65,
		greeting:           greeting,
		seekStepSeconds:    10,
		keymap:             NormalizeKeymap(keymap),
		trackDetailFetched: make(map[int]bool),
		trackBadges:        make(map[int]string),
		probeInFlight:      make(map[int]bool),
		status:             "",
		styles:             styles,
	}
}

func (m Model) Init() tea.Cmd { return m.daemonStatusCmd() }
