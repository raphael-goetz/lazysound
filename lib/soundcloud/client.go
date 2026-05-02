package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func (c *ApiClient) MeUser(ctx context.Context, token string) (*User, error) {
	var u User
	if err := c.doJSON(ctx, http.MethodGet, "/me", token, nil, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *ApiClient) doJSON(ctx context.Context, method, path, token string, query url.Values, body any, out any) error {
	u, err := url.Parse(apiBase + path)
	if err != nil {
		return err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	var (
		buf          io.Reader
		reqBodyBytes []byte
	)
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBodyBytes = b
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	if body != nil {
		req.Header.Set("content-type", "application/json; charset=utf-8")
	}
	if token != "" {
		req.Header.Set("Authorization", "OAuth "+token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bodyBytes, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if len(reqBodyBytes) > 0 {
			return fmt.Errorf("%s %s failed: %s: %s; request_body=%s", method, path, res.Status, string(bodyBytes), string(reqBodyBytes))
		}
		return fmt.Errorf("%s %s failed: %s: %s", method, path, res.Status, string(bodyBytes))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(bodyBytes, out)
}

func (c *ApiClient) MyTracks(ctx context.Context, token string) (*Tracks, error) {
	q := url.Values{}
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	q.Set("show_tracks", "true")
	var p Tracks
	if err := c.doJSON(ctx, http.MethodGet, "/me/tracks", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *ApiClient) MyPlaylists(ctx context.Context, token string) (*Playlists, error) {
	q := url.Values{}
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	q.Set("show_tracks", "true")
	var p Playlists
	if err := c.doJSON(ctx, http.MethodGet, "/me/playlists", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
func (c *ApiClient) LikedTracks(ctx context.Context, token string) (*Tracks, error) {
	q := url.Values{}
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	var p Tracks
	if err := c.doJSON(ctx, http.MethodGet, "/me/likes/tracks", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *ApiClient) LikedPlaylists(ctx context.Context, token string) (*Playlists, error) {
	q := url.Values{}
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	var p Playlists
	if err := c.doJSON(ctx, http.MethodGet, "/me/likes/playlists", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *ApiClient) SearchTracks(ctx context.Context, token, query string) (*Tracks, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	var p Tracks
	if err := c.doJSON(ctx, http.MethodGet, "/tracks", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *ApiClient) SearchPlaylists(ctx context.Context, token, query string) (*Playlists, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("linked_partitioning", "true")
	q.Set("limit", "50")
	q.Set("show_tracks", "true")
	var p Playlists
	if err := c.doJSON(ctx, http.MethodGet, "/playlists", token, q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

type TrackStreams struct {
	HLSAAC160URL  string `json:"hls_aac_160_url"`
	HLSAAC96URL   string `json:"hls_aac_96_url"`
	HTTPMP3128URL string `json:"http_mp3_128_url"`
	HLSMP3128URL  string `json:"hls_mp3_128_url"`
	HLSOPUS64URL  string `json:"hls_opus_64_url"`
	PreviewMP3128 string `json:"preview_mp3_128_url"`
}

func (c *ApiClient) TrackStreams(ctx context.Context, token string, trackID int) (*TrackStreams, error) {
	var s TrackStreams
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/tracks/%d/streams", trackID), token, nil, nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *ApiClient) TrackByID(ctx context.Context, token string, trackID int) (*Track, error) {
	var t Track
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/tracks/%d", trackID), token, nil, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

type playlistTrackRef struct {
	ID  int    `json:"id,omitempty"`
	URN string `json:"urn,omitempty"`
}

type createUpdatePlaylistRequest struct {
	Playlist struct {
		Title       string             `json:"title,omitempty"`
		Description string             `json:"description,omitempty"`
		Sharing     string             `json:"sharing,omitempty"`
		Tracks      []playlistTrackRef `json:"tracks,omitempty"`
	} `json:"playlist"`
}

func (c *ApiClient) CreatePlaylist(ctx context.Context, token, title, description, sharing string, trackIDs []int) (*Playlist, error) {
	var reqBody createUpdatePlaylistRequest
	reqBody.Playlist.Title = title
	reqBody.Playlist.Description = description
	reqBody.Playlist.Sharing = sharing
	if len(trackIDs) > 0 {
		reqBody.Playlist.Tracks = make([]playlistTrackRef, 0, len(trackIDs))
		for _, id := range trackIDs {
			if id > 0 {
				reqBody.Playlist.Tracks = append(reqBody.Playlist.Tracks, playlistTrackRef{ID: id})
			}
		}
	}
	return c.createOrUpdatePlaylist(ctx, token, apiBase+"/playlists", http.MethodPost, reqBody)
}

func (c *ApiClient) UpdatePlaylist(ctx context.Context, token string, playlistID int, title, description string, trackIDs []int) (*Playlist, error) {
	tracks := make([]Track, 0, len(trackIDs))
	for _, id := range trackIDs {
		if id > 0 {
			tracks = append(tracks, Track{ID: id, URN: "soundcloud:tracks:" + strconv.Itoa(id)})
		}
	}
	return c.UpdatePlaylistByRef(ctx, token, "soundcloud:playlists:"+strconv.Itoa(playlistID), title, description, tracks)
}

func (c *ApiClient) UpdatePlaylistByRef(ctx context.Context, token, playlistRef, title, description string, tracks []Track) (*Playlist, error) {
	var reqBody createUpdatePlaylistRequest
	reqBody.Playlist.Title = title
	reqBody.Playlist.Description = description
	if len(tracks) > 0 {
		reqBody.Playlist.Tracks = make([]playlistTrackRef, 0, len(tracks))
		for _, t := range tracks {
			ref := playlistTrackRef{}
			if strings.TrimSpace(t.URN) != "" {
				ref.URN = t.URN
			} else if t.ID > 0 {
				ref.ID = t.ID
			}
			if ref.URN == "" && ref.ID == 0 {
				continue
			}
			reqBody.Playlist.Tracks = append(reqBody.Playlist.Tracks, ref)
		}
	}
	return c.createOrUpdatePlaylist(ctx, token, fmt.Sprintf("%s/playlists/%s", apiBase, playlistRef), http.MethodPut, reqBody)
}

func (c *ApiClient) DeletePlaylist(ctx context.Context, token string, playlistID int) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/playlists/%d", playlistID), token, nil, nil, nil)
}

func (c *ApiClient) LikeTrack(ctx context.Context, token string, trackID int) error {
	return c.likeAction(ctx, token, fmt.Sprintf("%s/likes/tracks/%d", apiBase, trackID), http.MethodPost)
}

func (c *ApiClient) UnlikeTrack(ctx context.Context, token string, trackID int) error {
	return c.likeAction(ctx, token, fmt.Sprintf("%s/likes/tracks/%d", apiBase, trackID), http.MethodDelete)
}

func (c *ApiClient) LikePlaylist(ctx context.Context, token string, playlistID int) error {
	return c.likeAction(ctx, token, fmt.Sprintf("%s/likes/playlists/%d", apiBase, playlistID), http.MethodPost)
}

func (c *ApiClient) UnlikePlaylist(ctx context.Context, token string, playlistID int) error {
	return c.likeAction(ctx, token, fmt.Sprintf("%s/likes/playlists/%d", apiBase, playlistID), http.MethodDelete)
}

func (c *ApiClient) likeAction(ctx context.Context, token, urlStr, method string) error {
	path := strings.TrimPrefix(urlStr, apiBase)
	return c.doJSON(ctx, method, path, token, nil, nil, nil)
}

func (c *ApiClient) createOrUpdatePlaylist(ctx context.Context, token, urlStr, method string, payload createUpdatePlaylistRequest) (*Playlist, error) {
	var p Playlist
	path := strings.TrimPrefix(urlStr, apiBase)
	if err := c.doJSON(ctx, method, path, token, nil, payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
