package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/raphael-goetz/lazysound/internal/ui"
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

type SoundCloudConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
}

type DaemonConfig struct {
	Addr string `json:"addr"`
}

type Config struct {
	SoundCloud SoundCloudConfig `json:"soundcloud"`
	Keymap     ui.Keymap        `json:"keymap"`
	Theme      style.Theme      `json:"theme"`
	Daemon     DaemonConfig     `json:"daemon"`
}

func (c Config) HasSoundCloud() bool {
	return c.SoundCloud.ClientID != "" && c.SoundCloud.ClientSecret != "" && c.SoundCloud.RedirectURI != ""
}

func DefaultConfig() Config {
	return Config{
		Keymap: ui.DefaultKeymap(),
		Theme:  style.DefaultTheme(),
		Daemon: DaemonConfig{Addr: "127.0.0.1:7777"},
	}
}

func NormalizeConfig(c Config) Config {
	def := DefaultConfig()
	c.Keymap = ui.NormalizeKeymap(c.Keymap)
	c.Theme = style.NormalizeTheme(c.Theme)
	if c.Daemon.Addr == "" {
		c.Daemon.Addr = def.Daemon.Addr
	}
	return c
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lazysound", "config.json"), nil
}

func LoadConfig() (Config, bool, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, false, err
	}
	return NormalizeConfig(cfg), true, nil
}
