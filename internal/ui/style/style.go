package style

import "github.com/charmbracelet/lipgloss"

var (
	ColorText      = lipgloss.Color("252")
	ColorMuted     = lipgloss.Color("240")
	ColorAccent    = lipgloss.Color("39")
	ColorAccentDim = lipgloss.Color("31")
	ColorBorder    = lipgloss.Color("238")
	ColorDanger    = lipgloss.Color("196")
)

type Styles struct {
	App    lipgloss.Style
	Header lipgloss.Style

	HeaderTab  lipgloss.Style
	HeaderTabA lipgloss.Style

	Row    lipgloss.Style
	RowSel lipgloss.Style

	Muted  lipgloss.Style
	Accent lipgloss.Style
	Danger lipgloss.Style

	// Border/label styles for lazygit-like panes
	Border      lipgloss.Style
	BorderFocus lipgloss.Style
	Label       lipgloss.Style
	LabelFocus  lipgloss.Style

	Player lipgloss.Style
	CmdBar lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().Padding(0, 1),

		Header: lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder),

		HeaderTab:  lipgloss.NewStyle().Foreground(ColorMuted),
		HeaderTabA: lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),

		Row:    lipgloss.NewStyle().Foreground(ColorText),
		RowSel: lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(ColorAccentDim).Bold(true),

		Muted:  lipgloss.NewStyle().Foreground(ColorMuted),
		Accent: lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),
		Danger: lipgloss.NewStyle().Foreground(ColorDanger).Bold(true),

		Border:      lipgloss.NewStyle().Foreground(ColorBorder),
		BorderFocus: lipgloss.NewStyle().Foreground(ColorAccent),

		Label:      lipgloss.NewStyle().Foreground(ColorMuted).Bold(true),
		LabelFocus: lipgloss.NewStyle().Foreground(ColorAccent).Bold(true),

		Player: lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1),

		CmdBar: lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1),
	}
}


func (s Styles) BorderFor(focused bool) lipgloss.Style {
	if focused {
		return s.BorderFocus
	}
	return s.Border
}

func (s Styles) LabelFor(focused bool) lipgloss.Style {
	if focused {
		return s.LabelFocus
	}
	return s.Label
}

