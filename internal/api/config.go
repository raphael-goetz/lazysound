package api

import (
	"errors"
	"os"
	"strings"
)

type ApiConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scope        string
}

func LoadConfigFromEnv() (ApiConfig, error) {
	get := func(k string) string { return strings.TrimSpace(os.Getenv(k)) }
	cfg := ApiConfig{
		ClientID:     get("SOUNDCLOUD_CLIENT_ID"),
		ClientSecret: get("SOUNDCLOUD_CLIENT_SECRET"),
		RedirectURI:  get("SOUNDCLOUD_REDIRECT_URI"),
		Scope:        get("SOUNDCLOUD_SCOPE"),
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return ApiConfig{}, errors.New("missing env vars: SOUNDCLOUD_CLIENT_ID, SOUNDCLOUD_CLIENT_SECRET, SOUNDCLOUD_REDIRECT_URI")
	}
	return cfg, nil
}
