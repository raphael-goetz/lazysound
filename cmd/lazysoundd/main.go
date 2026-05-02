package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/raphael-goetz/lazysound/internal/app"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func main() {
	log := daemonLogger()

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
	srv, err := daemon.NewServer(addr, client, log)
	if err != nil {
		log.Error("daemon init", "err", err)
		os.Exit(1)
	}
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

func daemonLogger() *slog.Logger {
	logPath := filepath.Join(os.TempDir(), "lazysoundd.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	w := io.MultiWriter(os.Stderr, f)
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
