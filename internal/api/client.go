package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	authURL  = "https://secure.soundcloud.com/authorize"
	tokenURL = "https://secure.soundcloud.com/oauth/token"
	apiBase  = "https://api.soundcloud.com"
)

type ApiClient struct {
	cfg   ApiConfig
	store *TokenStore
	http  *http.Client
}

func NewApiClient(cfg ApiConfig, store *TokenStore) *ApiClient {
	return &ApiClient{
		cfg:   cfg,
		store: store,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ApiClient) Me(ctx context.Context, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")

	// Docs show both "Bearer" and "OAuth" in different sections.
	// SoundCloud historically accepts "OAuth". We'll use OAuth per your pasted docs.
	req.Header.Set("Authorization", "OAuth "+token)

	res, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("/me failed: %s: %s", res.Status, string(body))
	}
	return body, nil
}

func (c *ApiClient) MePlaylists(ctx context.Context, token string) (*Playlists, error) {

	u, err := url.Parse(apiBase + "/me/playlists")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	// partition
	q.Set("linked_partitioning", "true")
	// set limit
	q.Set("limit", "50")
	// include tracks
	q.Set("show_tracks", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "OAuth "+token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	print(req.Body)
	var p Playlists
	err = json.NewDecoder(res.Body).Decode(&p)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("/me failed: %s:", res.Status)
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}
func (c *ApiClient) LikedTracks(ctx context.Context, token string) (*Tracks, error) {

	u, err := url.Parse(apiBase + "/me/likes/tracks")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	// partition
	q.Set("linked_partitioning", "true")
	// set limit
	q.Set("limit", "50")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "OAuth "+token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	print(req.Body)
	var p Tracks
	err = json.NewDecoder(res.Body).Decode(&p)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("/me failed: %s:", res.Status)
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}
