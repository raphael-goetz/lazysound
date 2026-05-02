package style

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Text          string `json:"text"`
	Muted         string `json:"muted"`
	Accent        string `json:"accent"`
	AccentDim     string `json:"accent_dim"`
	Border        string `json:"border"`
	Danger        string `json:"danger"`
	SelectionText string `json:"selection_text"`
	ProgressFill  string `json:"progress_fill"`
}

func DefaultTheme() Theme {
	return Theme{
		Text:          "252",
		Muted:         "240",
		Accent:        "39",
		AccentDim:     "31",
		Border:        "238",
		Danger:        "196",
		SelectionText: "231",
		ProgressFill:  "#000000",
	}
}

func NormalizeTheme(t Theme) Theme {
	def := DefaultTheme()
	if t.Text == "" {
		t.Text = def.Text
	}
	if t.Muted == "" {
		t.Muted = def.Muted
	}
	if t.Accent == "" {
		t.Accent = def.Accent
	}
	if t.AccentDim == "" {
		t.AccentDim = def.AccentDim
	}
	if t.Border == "" {
		t.Border = def.Border
	}
	if t.Danger == "" {
		t.Danger = def.Danger
	}
	if t.SelectionText == "" {
		t.SelectionText = def.SelectionText
	}
	if t.ProgressFill == "" {
		t.ProgressFill = def.ProgressFill
	}
	return t
}

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

	Player   lipgloss.Style
	CmdBar   lipgloss.Style
	Progress lipgloss.Style
}

func NewStyles() Styles {
	return NewStylesWithTheme(DefaultTheme())
}

func NewStylesWithTheme(theme Theme) Styles {
	t := NormalizeTheme(theme)
	colorText := lipgloss.Color(t.Text)
	colorMuted := lipgloss.Color(t.Muted)
	colorAccent := lipgloss.Color(t.Accent)
	colorAccentDim := lipgloss.Color(t.AccentDim)
	colorBorder := lipgloss.Color(t.Border)
	colorDanger := lipgloss.Color(t.Danger)
	colorSelectionText := lipgloss.Color(t.SelectionText)
	colorProgress := lipgloss.Color(t.ProgressFill)

	return Styles{
		App: lipgloss.NewStyle().Padding(0, 1),

		Header: lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder),

		HeaderTab:  lipgloss.NewStyle().Foreground(colorMuted),
		HeaderTabA: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),

		Row:    lipgloss.NewStyle().Foreground(colorText),
		RowSel: lipgloss.NewStyle().Foreground(colorSelectionText).Background(colorAccentDim).Bold(true),

		Muted:  lipgloss.NewStyle().Foreground(colorMuted),
		Accent: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		Danger: lipgloss.NewStyle().Foreground(colorDanger).Bold(true),

		Border:      lipgloss.NewStyle().Foreground(colorBorder),
		BorderFocus: lipgloss.NewStyle().Foreground(colorAccent),

		Label:      lipgloss.NewStyle().Foreground(colorMuted).Bold(true),
		LabelFocus: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),

		Player: lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1),

		CmdBar:   lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1),
		Progress: lipgloss.NewStyle().Background(colorProgress),
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
