package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/raphael-goetz/lazysound/internal/app"
	"github.com/raphael-goetz/lazysound/internal/daemon"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	cfg, ok, cfgErr := app.LoadConfig()
	if cfgErr != nil {
		die("config file error: " + cfgErr.Error())
	}
	if !ok {
		die("config file not found: " + mustConfigPath())
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
			die(err.Error())
		}
		authCtx, authCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer authCancel()
		tok, err := apiClient.AuthCodePKCE(authCtx)
		if err != nil {
			die("auth failed: " + err.Error())
		}
		fmt.Printf("auth ok (expires_at=%s)\n", tok.ExpiresAt.Format(time.RFC3339))
		return
	case "whoami":
		apiClient, err := newAPIClient(cfg)
		if err != nil {
			die(err.Error())
		}
		whoCtx, whoCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer whoCancel()
		tok, err := apiClient.EnsureValidToken(whoCtx)
		if err != nil {
			die("token invalid: " + err.Error())
		}
		user, err := apiClient.MeUser(whoCtx, tok.AccessToken)
		if err != nil {
			die("whoami failed: " + err.Error())
		}
		fmt.Printf("user=%s id=%d expires_at=%s\n", user.Username, user.ID, tok.ExpiresAt.Format(time.RFC3339))
		return
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
			die("seek requires seconds (e.g. lazysoundctl seek 10)")
		}
		sec, perr := strconv.Atoi(args[0])
		if perr != nil {
			die("invalid seek seconds")
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.Seek(ctx, sec)
		}
	case "volume":
		if len(args) < 1 {
			die("volume requires 0-100")
		}
		vol, perr := strconv.Atoi(args[0])
		if perr != nil {
			die("invalid volume")
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetVolume(ctx, vol)
		}
	case "shuffle":
		if len(args) < 1 {
			die("shuffle requires true/false")
		}
		val, perr := strconv.ParseBool(args[0])
		if perr != nil {
			die("invalid shuffle flag")
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetShuffle(ctx, val)
		}
	case "repeat":
		if len(args) < 1 {
			die("repeat requires true/false")
		}
		val, perr := strconv.ParseBool(args[0])
		if perr != nil {
			die("invalid repeat flag")
		}
		err = client.EnsureRunning(ctx)
		if err == nil {
			state, err = client.SetRepeat(ctx, val)
		}
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		die(err.Error())
	}
	if state != nil {
		out, _ := json.MarshalIndent(state, "", "  ")
		fmt.Println(string(out))
	}
}

func mustConfigPath() string {
	path, err := app.ConfigPath()
	if err != nil {
		return ""
	}
	return path
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  lazysoundctl auth")
	fmt.Println("  lazysoundctl whoami")
	fmt.Println("  lazysoundctl status")
	fmt.Println("  lazysoundctl pause|stop|restart|next|prev")
	fmt.Println("  lazysoundctl seek <seconds>")
	fmt.Println("  lazysoundctl volume <0-100>")
	fmt.Println("  lazysoundctl shuffle <true|false>")
	fmt.Println("  lazysoundctl repeat <true|false>")
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

func die(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
