package entrypoints

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/raphael-goetz/lazysound/internal/app"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func RunDaemon(stderr io.Writer) int {
	log := daemonLogger(stderr)

	cfg, ok, err := app.LoadConfig()
	if err != nil {
		log.Error("config file", "err", err)
		return 1
	}
	if !ok {
		log.Error("config file not found", "path", mustConfigPath())
		return 1
	}

	if !cfg.HasSoundCloud() {
		log.Error("soundcloud config missing")
		return 1
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
		return 1
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("daemon error", "err", err)
		return 1
	}
	return 0
}

func daemonLogger(stderr io.Writer) *slog.Logger {
	logPath := filepath.Join(os.TempDir(), "lazysoundd.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	w := io.MultiWriter(stderr, f)
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
