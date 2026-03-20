package ui

type Keymap struct {
	Quit         []string `json:"quit"`
	Tab          []string `json:"tab"`
	Left         []string `json:"left"`
	Right        []string `json:"right"`
	Up           []string `json:"up"`
	Down         []string `json:"down"`
	Open         []string `json:"open"`
	Back         []string `json:"back"`
	ActionMenu   []string `json:"action_menu"`
	Search       []string `json:"search"`
	SearchTracks []string `json:"search_tracks"`
	SearchPlay   []string `json:"search_playlists"`

	Play        []string `json:"play"`
	Stop        []string `json:"stop"`
	Restart     []string `json:"restart"`
	VolumeUp    []string `json:"volume_up"`
	VolumeDown  []string `json:"volume_down"`
	PauseToggle []string `json:"pause"`
	SeekForward []string `json:"seek_forward"`
	SeekBack    []string `json:"seek_back"`
	Shuffle     []string `json:"shuffle"`
	Repeat      []string `json:"repeat"`
}

func DefaultKeymap() Keymap {
	return Keymap{
		Quit:         []string{"q", "ctrl+c"},
		Tab:          []string{"tab"},
		Left:         []string{"h"},
		Right:        []string{"l"},
		Up:           []string{"k", "up"},
		Down:         []string{"j", "down"},
		Open:         []string{"enter"},
		Back:         []string{"backspace"},
		ActionMenu:   []string{"a"},
		Search:       []string{"/"},
		SearchTracks: []string{"t"},
		SearchPlay:   []string{"p"},
		Play:         []string{"p"},
		Stop:         []string{"s"},
		Restart:      []string{"r"},
		VolumeUp:     []string{"+", "="},
		VolumeDown:   []string{"-"},
		PauseToggle:  []string{"space"},
		SeekForward:  []string{"."},
		SeekBack:     []string{","},
		Shuffle:      []string{"x"},
		Repeat:       []string{"R"},
	}
}

func NormalizeKeymap(k Keymap) Keymap {
	return mergeKeymap(DefaultKeymap(), k)
}

func mergeKeymap(def, custom Keymap) Keymap {
	if len(custom.Quit) > 0 {
		def.Quit = custom.Quit
	}
	if len(custom.Tab) > 0 {
		def.Tab = custom.Tab
	}
	if len(custom.Left) > 0 {
		def.Left = custom.Left
	}
	if len(custom.Right) > 0 {
		def.Right = custom.Right
	}
	if len(custom.Up) > 0 {
		def.Up = custom.Up
	}
	if len(custom.Down) > 0 {
		def.Down = custom.Down
	}
	if len(custom.Open) > 0 {
		def.Open = custom.Open
	}
	if len(custom.Back) > 0 {
		def.Back = custom.Back
	}
	if len(custom.ActionMenu) > 0 {
		def.ActionMenu = custom.ActionMenu
	}
	if len(custom.Search) > 0 {
		def.Search = custom.Search
	}
	if len(custom.SearchTracks) > 0 {
		def.SearchTracks = custom.SearchTracks
	}
	if len(custom.SearchPlay) > 0 {
		def.SearchPlay = custom.SearchPlay
	}
	if len(custom.Play) > 0 {
		def.Play = custom.Play
	}
	if len(custom.Stop) > 0 {
		def.Stop = custom.Stop
	}
	if len(custom.Restart) > 0 {
		def.Restart = custom.Restart
	}
	if len(custom.VolumeUp) > 0 {
		def.VolumeUp = custom.VolumeUp
	}
	if len(custom.VolumeDown) > 0 {
		def.VolumeDown = custom.VolumeDown
	}
	if len(custom.PauseToggle) > 0 {
		def.PauseToggle = custom.PauseToggle
	}
	if len(custom.SeekForward) > 0 {
		def.SeekForward = custom.SeekForward
	}
	if len(custom.SeekBack) > 0 {
		def.SeekBack = custom.SeekBack
	}
	if len(custom.Shuffle) > 0 {
		def.Shuffle = custom.Shuffle
	}
	if len(custom.Repeat) > 0 {
		def.Repeat = custom.Repeat
	}
	return def
}

func keyIs(key string, list []string) bool {
	for _, k := range list {
		if key == k {
			return true
		}
	}
	return false
}
