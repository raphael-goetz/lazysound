package entrypoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/raphael-goetz/lazysound/internal/app"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func RunCtl(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		ctlUsage(stdout)
		return 1
	}
	cmd := args[0]
	args = args[1:]

	cfg, ok, cfgErr := app.LoadConfig()
	if cfgErr != nil {
		fmt.Fprintln(stderr, "config file error:", cfgErr.Error())
		return 1
	}
	if !ok {
		fmt.Fprintln(stderr, "config file not found:", mustConfigPath())
		return 1
	}

	client := daemon.NewClient(cfg.Daemon.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	var (
		state *daemon.State
		err   error
	)

	switch cmd {
	case "auth":
		apiClient, err := newAPIClient(cfg)
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		authCtx, authCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer authCancel()
		tok, err := apiClient.AuthCodePKCE(authCtx)
		if err != nil {
			fmt.Fprintln(stderr, "auth failed:", err.Error())
			return 1
		}
		fmt.Fprintf(stdout, "auth ok (expires_at=%s)\n", tok.ExpiresAt.Format(time.RFC3339))
		return 0
	case "whoami":
		apiClient, err := newAPIClient(cfg)
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		whoCtx, whoCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer whoCancel()
		tok, err := apiClient.EnsureValidToken(whoCtx)
		if err != nil {
			fmt.Fprintln(stderr, "token invalid:", err.Error())
			return 1
		}
		user, err := apiClient.MeUser(whoCtx, tok.AccessToken)
		if err != nil {
			fmt.Fprintln(stderr, "whoami failed:", err.Error())
			return 1
		}
		fmt.Fprintf(stdout, "user=%s id=%d expires_at=%s\n", user.Username, user.ID, tok.ExpiresAt.Format(time.RFC3339))
		return 0
	case "status":
		state, err = client.Status(ctx)
	case "pause":
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.TogglePause(ctx)
		}
	case "stop":
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Stop(ctx)
		}
	case "restart":
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Restart(ctx)
		}
	case "next":
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Next(ctx)
		}
	case "prev":
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Prev(ctx)
		}
	case "seek":
		if len(args) < 1 {
			fmt.Fprintln(stderr, "seek requires seconds (e.g. lazysound cli seek 10)")
			return 1
		}
		sec, perr := strconv.Atoi(args[0])
		if perr != nil {
			fmt.Fprintln(stderr, "invalid seek seconds")
			return 1
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Seek(ctx, sec)
		}
	case "volume":
		if len(args) < 1 {
			fmt.Fprintln(stderr, "volume requires 0-100")
			return 1
		}
		vol, perr := strconv.Atoi(args[0])
		if perr != nil {
			fmt.Fprintln(stderr, "invalid volume")
			return 1
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetVolume(ctx, vol)
		}
	case "shuffle":
		if len(args) < 1 {
			fmt.Fprintln(stderr, "shuffle requires true/false")
			return 1
		}
		val, perr := strconv.ParseBool(args[0])
		if perr != nil {
			fmt.Fprintln(stderr, "invalid shuffle flag")
			return 1
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetShuffle(ctx, val)
		}
	case "repeat":
		if len(args) < 1 {
			fmt.Fprintln(stderr, "repeat requires true/false")
			return 1
		}
		val, perr := strconv.ParseBool(args[0])
		if perr != nil {
			fmt.Fprintln(stderr, "invalid repeat flag")
			return 1
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetRepeat(ctx, val)
		}
	default:
		ctlUsage(stdout)
		return 1
	}

	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	if state != nil {
		out, _ := json.MarshalIndent(state, "", "  ")
		fmt.Fprintln(stdout, string(out))
	}
	return 0
}

func ctlUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  lazysound cli auth")
	fmt.Fprintln(w, "  lazysound cli whoami")
	fmt.Fprintln(w, "  lazysound cli status")
	fmt.Fprintln(w, "  lazysound cli pause|stop|restart|next|prev")
	fmt.Fprintln(w, "  lazysound cli seek <seconds>")
	fmt.Fprintln(w, "  lazysound cli volume <0-100>")
	fmt.Fprintln(w, "  lazysound cli shuffle <true|false>")
	fmt.Fprintln(w, "  lazysound cli repeat <true|false>")
}

func newAPIClient(cfg app.Config) (*api.ApiClient, error) {
	if !cfg.HasSoundCloud() {
		return nil, fmt.Errorf("soundcloud config missing in %s", mustConfigPath())
	}
	store, err := api.NewTokenStoreDefault()
	if err != nil {
		return nil, err
	}
	return api.NewApiClient(api.ApiConfig{
		ClientID:     cfg.SoundCloud.ClientID,
		ClientSecret: cfg.SoundCloud.ClientSecret,
		RedirectURI:  cfg.SoundCloud.RedirectURI,
		Scope:        cfg.SoundCloud.Scope,
	}, store), nil
}
