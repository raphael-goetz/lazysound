package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/raphael-goetz/lazysound/internal/ui"
	"github.com/raphael-goetz/lazysound/lib/uikit/style"
)

func NewProgram(data Bootstrap, cfg Config) *tea.Program {
	km := ui.NormalizeKeymap(cfg.Keymap)
	st := style.NewStylesWithTheme(cfg.Theme)
	model := ui.NewModel(
		data.MyTracks,
		data.LikedTracks,
		data.MyPlaylists,
		data.LikedPlaylists,
		data.Me,
		data.Client,
		data.Token,
		cfg.Daemon.Addr,
		km,
		st,
	)
	return tea.NewProgram(model, tea.WithAltScreen())
}
