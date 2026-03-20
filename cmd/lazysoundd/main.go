package main

import (
	"log/slog"
	"os"

	"github.com/raphael-goetz/lazysound/internal/api"
	"github.com/raphael-goetz/lazysound/internal/app"
	"github.com/raphael-goetz/lazysound/internal/daemon"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	cfg, ok, err := app.LoadConfig()
	if err != nil {
		log.Error("config file", "err", err)
		os.Exit(1)
	}
	if !ok {
		log.Error("config file not found", "path", mustConfigPath())
		os.Exit(1)
	}

	if !cfg.HasSoundCloud() {
		log.Error("soundcloud config missing")
		os.Exit(1)
	}

	apiCfg := api.ApiConfig{
		ClientID:     cfg.SoundCloud.ClientID,
		ClientSecret: cfg.SoundCloud.ClientSecret,
		RedirectURI:  cfg.SoundCloud.RedirectURI,
		Scope:        cfg.SoundCloud.Scope,
	}
	store, err := api.NewTokenStoreDefault()
	if err != nil {
		log.Error("token store", "err", err)
	}
	client := api.NewApiClient(apiCfg, store)
	addr := cfg.Daemon.Addr
	srv, err := daemon.NewServer(addr, client)
	if err != nil {
		log.Error("daemon init", "err", err)
		os.Exit(1)
	}
	log.Info("lazysoundd listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("daemon error", "err", err)
		os.Exit(1)
	}
}

func mustConfigPath() string {
	path, err := app.ConfigPath()
	if err != nil {
		return ""
	}
	return path
}
