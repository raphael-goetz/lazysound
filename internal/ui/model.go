package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raphael-goetz/lazysound/internal/api"
	"github.com/raphael-goetz/lazysound/internal/math"
	"github.com/raphael-goetz/lazysound/internal/ui/component"
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

type Model struct {
	w, h int

	focus Focus

	navItems []string
	navIdx   int

	playlists   []api.Playlist
	playlistIdx int

	tracks   []api.Track
	trackIdx int

	styles style.Styles

	err string
}

func NewModel(tracks []api.Track, playlists []api.Playlist) Model {
	return Model{
		focus: FocusList,
		navItems: []string{
			"Search",
			"My Playlists",
			"Liked Playlists",
			"Liked Tracks",
		},
		navIdx:    1,
		tracks:    tracks,
		playlists: playlists,
		styles:    style.NewStyles(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % 3
		case "h":
			if m.focus > 0 {
				m.focus--
			}
		case "l":
			if m.focus < 2 {
				m.focus++
			}

		case "j", "down":
			switch m.focus {
			case FocusNav:
				if m.navIdx < len(m.navItems)-1 {
					m.navIdx++
				}
			case FocusList:
				if m.trackIdx < len(m.tracks)-1 {
					m.trackIdx++
				}
			}
		case "k", "up":
			switch m.focus {
			case FocusNav:
				if m.navIdx > 0 {
					m.navIdx--
				}
			case FocusList:
				if m.trackIdx > 0 {
					m.trackIdx--
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := m.styles
	if m.w == 0 || m.h == 0 {
		return "loading..."
	}

	// Render fixed areas first and measure actual heights (borders included)
	header := renderHeader(s, m.w)
	player := renderPlayerBar(s, m, m.w)
	cmd := component.Command(s)

	headerH := lipgloss.Height(header)
	playerH := lipgloss.Height(player)
	cmdH := lipgloss.Height(cmd)

	bodyH := m.h - headerH - playerH - cmdH
	if bodyH < 6 {
		bodyH = 6
	}

	// Horizontal sizing
	navW := math.Clamp(int(float64(m.w)*0.22), 18, 28)
	inspectW := math.Clamp(int(float64(m.w)*0.30), 24, 40)
	listW := m.w - navW - inspectW - 4 // gaps

	// Responsive collapse: hide inspect first, then nav
	if listW < 30 {
		inspectW = 0
		listW = m.w - navW - 2
		if listW < 20 {
			navW = 0
			listW = m.w
		}
	}

	// Pane contents (inner sizes: -2 for borders)
	navContent := component.Nav(s, m.navItems, m.navIdx, m.focus == FocusNav, navW-2, bodyH-2)
	var listContent string
	switch m.navIdx {
	case 0:
		listContent = "Comming soon"
	case 1:
		listContent = component.PlaylistTable(s, m.playlists, m.playlistIdx, m.focus == FocusList, listW-2, bodyH-2)
	case 2:
		listContent = component.PlaylistTable(s, m.playlists, m.playlistIdx, m.focus == FocusList, listW-2, bodyH-2)
	case 3:
		listContent = component.TrackTable(s, m.tracks, m.trackIdx, m.focus == FocusList, listW-2, bodyH-2)
	default:
		listContent = "Error"
	}

	// TOOD Inspect conent
	//inspectContent := renderInspect(s, m, inspectW-2, bodyH-2)

	navPane := ""
	if navW > 0 {
		navPane = component.BorderedBox(
			s.BorderFor(m.focus == FocusNav),
			s.LabelFor(m.focus == FocusNav),
			"1 NAV",
			navContent,
			navW,
			bodyH,
		)
	}

	listPane := component.BorderedBox(
		s.BorderFor(m.focus == FocusList),
		s.LabelFor(m.focus == FocusList),
		"2 LIST",
		listContent,
		listW,
		bodyH,
	)
	/*
		inspectPane := ""
		if inspectW > 0 {
			inspectPane = renderPaneWithBorderLabel(
				s.borderFor(m.focus == FocusInspect),
				s.labelFor(m.focus == FocusInspect),
				"3 INSPECT",
				inspectContent,
				inspectW,
				bodyH,
			)
		}
	*/
	body := joinPanes(navPane, listPane, "inspect")

	return s.App.Render(
		lipgloss.JoinVertical(lipgloss.Left, header, body, player, cmd),
	)
}
