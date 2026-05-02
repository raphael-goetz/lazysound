package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

type Bootstrap struct {
	Me             *api.User
	MyTracks       []api.Track
	LikedTracks    []api.Track
	MyPlaylists    []api.Playlist
	LikedPlaylists []api.Playlist
	Token          string
	Client         *api.ApiClient
	Warnings       []error
}

func InitClient(ctx context.Context, cfg Config) (*api.ApiClient, string, error) {
	if !cfg.HasSoundCloud() {
		return nil, "", wrap(KindConfig, "soundcloud", fmt.Errorf("missing soundcloud config"))
	}
	apiCfg := api.ApiConfig{
		ClientID:     cfg.SoundCloud.ClientID,
		ClientSecret: cfg.SoundCloud.ClientSecret,
		RedirectURI:  cfg.SoundCloud.RedirectURI,
		Scope:        cfg.SoundCloud.Scope,
	}
	store, err := api.NewTokenStoreDefault()
	if err != nil {
		return nil, "", wrap(KindTokenStore, "init", err)
	}
	client := api.NewApiClient(apiCfg, store)
	tok, err := client.EnsureValidToken(ctx)
	if err != nil {
		return nil, "", wrap(KindAuth, "ensure token", err)
	}
	return client, tok.AccessToken, nil
}

func FetchInitialData(ctx context.Context, client *api.ApiClient, token string) (Bootstrap, error) {
	var (
		me    *api.User
		meErr error
		mt    *api.Tracks
		mtErr error
		pl    *api.Playlists
		plErr error
		tr    *api.Tracks
		trErr error
		lp    *api.Playlists
		lpErr error
	)

	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		defer wg.Done()
		me, meErr = client.MeUser(ctx, token)
	}()
	go func() {
		defer wg.Done()
		mt, mtErr = client.MyTracks(ctx, token)
	}()
	go func() {
		defer wg.Done()
		pl, plErr = client.MyPlaylists(ctx, token)
	}()
	go func() {
		defer wg.Done()
		tr, trErr = client.LikedTracks(ctx, token)
	}()
	go func() {
		defer wg.Done()
		lp, lpErr = client.LikedPlaylists(ctx, token)
	}()
	wg.Wait()

	if mtErr != nil || plErr != nil || trErr != nil {
		return Bootstrap{}, wrap(KindFetch, "required", joinErrors(mtErr, plErr, trErr))
	}

	out := Bootstrap{
		Me:             me,
		Token:          token,
		Client:         client,
		MyTracks:       mt.Collection,
		LikedTracks:    tr.Collection,
		MyPlaylists:    pl.Collection,
		LikedPlaylists: nil,
	}
	if lpErr != nil {
		out.Warnings = append(out.Warnings, wrap(KindFetch, "liked playlists", lpErr))
	} else if lp != nil {
		out.LikedPlaylists = lp.Collection
	}
	if meErr != nil {
		out.Warnings = append(out.Warnings, wrap(KindFetch, "me", meErr))
	}
	return out, nil
}

func joinErrors(errs ...error) error {
	var out []error
	for _, err := range errs {
		if err != nil {
			out = append(out, err)
		}
	}
	if len(out) == 0 {
		return nil
	}
	msg := "bootstrap errors:"
	for _, err := range out {
		msg += "\n- " + err.Error()
	}
	return fmt.Errorf("%s", msg)
}

func DefaultTimeout() time.Duration {
	return 5 * time.Minute
}
