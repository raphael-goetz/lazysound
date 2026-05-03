package entrypoints

import (
	"context"
	"io"
	"log/slog"

	"github.com/raphael-goetz/lazysound/internal/app"
)

func RunTUI(stderr io.Writer) int {
	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{}))

	cfg, ok, err := app.LoadConfig()
	if err != nil {
		log.Error("config file", "err", err)
		return 1
	}
	if !ok {
		log.Error("config file not found", "path", mustConfigPath())
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.DefaultTimeout())
	defer cancel()

	client, token, err := app.InitClient(ctx, cfg)
	if err != nil {
		log.Error("init client", "err", err)
		return 1
	}

	data, err := app.FetchInitialData(ctx, client, token)
	if err != nil {
		log.Error("bootstrap", "err", err)
		return 1
	}
	for _, warn := range data.Warnings {
		log.Warn("warning", "err", warn)
	}

	p := app.NewProgram(data, cfg)
	if _, err := p.Run(); err != nil {
		log.Error("tui", "err", err)
		return 1
	}
	return 0
}
