package main

import (
	"context"
	"os"

	"github.com/raphael-goetz/lazysound/internal/app"
	"log/slog"
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

	ctx, cancel := context.WithTimeout(context.Background(), app.DefaultTimeout())
	defer cancel()

	client, token, err := app.InitClient(ctx, cfg)
	if err != nil {
		log.Error("init client", "err", err)
		os.Exit(1)
	}

	data, err := app.FetchInitialData(ctx, client, token)
	if err != nil {
		log.Error("bootstrap", "err", err)
		os.Exit(1)
	}
	for _, warn := range data.Warnings {
		log.Warn("warning", "err", warn)
	}

	p := app.NewProgram(data, cfg)
	if _, err := p.Run(); err != nil {
		log.Error("tui", "err", err)
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
