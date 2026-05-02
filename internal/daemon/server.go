package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/raphael-goetz/lazysound/lib/player"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

type Server struct {
	addr   string
	api    *api.ApiClient
	player *player.MPV
	log    *slog.Logger

	mu          sync.Mutex
	playSession uint64
	token       string
	queue       []api.Track
	index       int
	shuffle     bool
	repeat      bool
	paused      bool
	volume      int
	now         *api.Track
	playing     bool
	stopChan    chan struct{}
	lastError   string
}

func NewServer(addr string, apiClient *api.ApiClient, logger *slog.Logger) (*Server, error) {
	p, err := player.NewMPV()
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		addr:     addr,
		api:      apiClient,
		player:   p,
		log:      logger,
		volume:   65,
		stopChan: make(chan struct{}, 1),
	}, nil
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Error("listen failed", "addr", s.addr, "err", err)
		return err
	}
	s.log.Info("daemon listening", "addr", s.addr)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			s.log.Warn("accept failed", "err", err)
			continue
		}
		s.log.Debug("connection accepted", "remote_addr", conn.RemoteAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		payload := sc.Bytes()
		var req Request
		if err := json.Unmarshal(payload, &req); err != nil {
			s.log.Warn("invalid request payload", "remote_addr", conn.RemoteAddr().String(), "err", err)
			_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
			return
		}
		attrs := requestAttrs(req)
		isStatus := req.Cmd == "status"
		if isStatus {
			s.log.Debug("request received", attrs...)
		} else {
			s.log.Info("request received", attrs...)
		}
		start := time.Now()
		resp := s.handleRequest(req)
		if resp.OK {
			doneAttrs := []any{"cmd", req.Cmd, "ok", true, "elapsed_ms", time.Since(start).Milliseconds()}
			if isStatus {
				s.log.Debug("request completed", doneAttrs...)
			} else {
				s.log.Info("request completed", doneAttrs...)
			}
		} else {
			s.log.Warn("request failed", "cmd", req.Cmd, "ok", false, "elapsed_ms", time.Since(start).Milliseconds(), "error", resp.Error)
		}
		_ = json.NewEncoder(conn).Encode(resp)
	}
	if err := sc.Err(); err != nil {
		s.log.Warn("connection read error", "remote_addr", conn.RemoteAddr().String(), "err", err)
	}
}

func (s *Server) handleRequest(req Request) Response {
	switch req.Cmd {
	case "status":
		return Response{OK: true, State: s.state()}
	case "play_track":
		if req.Track == nil {
			return Response{OK: false, Error: "missing track"}
		}
		return s.playQueue(req.Token, []api.Track{*req.Track}, 0)
	case "play_queue":
		if len(req.Queue) == 0 {
			return Response{OK: false, Error: "empty queue"}
		}
		start := req.StartIdx
		if start < 0 || start >= len(req.Queue) {
			start = 0
		}
		return s.playQueue(req.Token, req.Queue, start)
	case "probe_track":
		if req.Track == nil {
			return Response{OK: false, Error: "missing track"}
		}
		probe, err := s.probeTrack(req.Token, req.Track.ID)
		if err != nil {
			return Response{OK: false, Error: err.Error(), Probe: probe}
		}
		return Response{OK: true, Probe: probe}
	case "stop":
		if err := s.player.Stop(); err != nil {
			s.log.Warn("stop failed", "err", err)
		}
		s.mu.Lock()
		s.playSession++
		s.playing = false
		s.now = nil
		s.paused = false
		s.lastError = ""
		s.mu.Unlock()
		return Response{OK: true, State: s.state()}
	case "pause":
		if err := s.player.TogglePause(); err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		s.mu.Lock()
		s.paused = !s.paused
		s.mu.Unlock()
		return Response{OK: true, State: s.state()}
	case "seek":
		if err := s.player.SeekSeconds(req.SeekSec); err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true, State: s.state()}
	case "restart":
		if err := s.player.Restart(); err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true, State: s.state()}
	case "volume":
		if err := s.player.SetVolume(req.Volume); err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		s.mu.Lock()
		s.volume = req.Volume
		s.mu.Unlock()
		return Response{OK: true, State: s.state()}
	case "next":
		return s.playNext(true)
	case "prev":
		return s.playPrev()
	case "shuffle":
		if req.Shuffle == nil {
			return Response{OK: false, Error: "missing shuffle"}
		}
		s.mu.Lock()
		s.shuffle = *req.Shuffle
		s.mu.Unlock()
		return Response{OK: true, State: s.state()}
	case "repeat":
		if req.Repeat == nil {
			return Response{OK: false, Error: "missing repeat"}
		}
		s.mu.Lock()
		s.repeat = *req.Repeat
		s.mu.Unlock()
		return Response{OK: true, State: s.state()}
	default:
		return Response{OK: false, Error: "unknown cmd"}
	}
}

func (s *Server) playQueue(token string, queue []api.Track, start int) Response {
	s.log.Info("play queue requested", "queue_len", len(queue), "start_idx", start, "token_provided", token != "")
	s.mu.Lock()
	if token != "" {
		s.token = token
	}
	s.queue = queue
	s.index = start
	s.mu.Unlock()
	if err := s.playIndex(start); err != nil {
		return Response{OK: false, Error: err.Error()}
	}
	return Response{OK: true, State: s.state()}
}

func (s *Server) playIndex(idx int) error {
	s.mu.Lock()
	if idx < 0 || idx >= len(s.queue) {
		s.mu.Unlock()
		return fmt.Errorf("index out of range")
	}
	s.playSession++
	session := s.playSession
	track := s.queue[idx]
	token := s.token
	vol := s.volume
	s.mu.Unlock()
	token = s.resolvePlaybackToken(token)
	s.mu.Lock()
	s.token = token
	s.mu.Unlock()

	url, err := s.streamURL(track.ID, token)
	if err != nil {
		s.log.Warn("stream URL resolution failed", "track_id", track.ID, "err", err)
		return err
	}
	if err := s.player.Stop(); err != nil {
		s.log.Debug("stop before play returned error", "track_id", track.ID, "err", err)
	}
	if err := s.player.Play(url, vol, token); err != nil {
		s.log.Warn("player play failed", "track_id", track.ID, "err", err)
		return err
	}
	s.mu.Lock()
	s.now = &track
	s.playing = true
	s.paused = false
	s.lastError = ""
	s.mu.Unlock()
	s.log.Info("playback started", "track_id", track.ID, "track_title", track.Title, "index", idx, "volume", vol, "session", session)
	go s.waitEnd(idx, session)
	return nil
}

func (s *Server) resolvePlaybackToken(fallback string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tok, err := s.api.EnsureValidToken(ctx)
	if err != nil {
		if fallback != "" {
			s.log.Warn("token refresh failed; using fallback token", "err", err)
			return fallback
		}
		s.log.Warn("token refresh failed; no token available", "err", err)
		return ""
	}
	return tok.AccessToken
}

func (s *Server) waitEnd(idx int, session uint64) {
	exit := <-s.player.WaitExit()
	s.mu.Lock()
	stale := session != s.playSession
	s.mu.Unlock()
	if stale {
		s.log.Debug("stale player exit ignored", "index", idx, "session", session)
		return
	}
	if exit.Err != nil {
		errMsg := exit.Err.Error()
		if exit.Detail != "" {
			errMsg = errMsg + ": " + exit.Detail
		} else if errMsg == "exit status 2" {
			errMsg = "mpv could not open this stream (likely 401/403 access restriction, regional block, or unavailable media)"
		}
		s.mu.Lock()
		s.lastError = errMsg
		s.mu.Unlock()
		s.log.Warn("player process exited with error", "index", idx, "stopped", exit.Stopped, "err", exit.Err, "detail", exit.Detail)
	} else {
		s.mu.Lock()
		s.lastError = ""
		s.mu.Unlock()
		s.log.Info("player process exited", "index", idx, "stopped", exit.Stopped)
	}
	if exit.Stopped {
		return
	}
	_ = s.playNext(false)
}

func (s *Server) playNext(userAction bool) Response {
	s.mu.Lock()
	if len(s.queue) == 0 {
		s.mu.Unlock()
		return Response{OK: false, Error: "empty queue"}
	}
	next := s.index + 1
	if next >= len(s.queue) {
		if s.repeat {
			next = 0
		} else {
			s.playing = false
			s.paused = false
			s.now = nil
			s.mu.Unlock()
			return Response{OK: true, State: s.state()}
		}
	}
	s.index = next
	s.mu.Unlock()
	s.log.Info("play next", "next_index", next, "user_action", userAction)
	if err := s.playIndex(next); err != nil {
		return Response{OK: false, Error: err.Error()}
	}
	return Response{OK: true, State: s.state()}
}

func (s *Server) playPrev() Response {
	s.mu.Lock()
	if len(s.queue) == 0 {
		s.mu.Unlock()
		return Response{OK: false, Error: "empty queue"}
	}
	prev := s.index - 1
	if prev < 0 {
		prev = 0
	}
	s.index = prev
	s.mu.Unlock()
	s.log.Info("play prev", "prev_index", prev)
	if err := s.playIndex(prev); err != nil {
		return Response{OK: false, Error: err.Error()}
	}
	return Response{OK: true, State: s.state()}
}

func (s *Server) streamURL(trackID int, token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	streams, err := s.api.TrackStreams(ctx, token, trackID)
	if err != nil {
		return "", err
	}
	if b, jerr := json.Marshal(streams); jerr == nil {
		s.log.Debug("streams response", "track_id", trackID, "json", string(b))
	}
	candidates := streamCandidates(streams)
	if len(candidates) == 0 {
		if streams.PreviewMP3128 != "" {
			s.log.Warn("only preview stream available", "track_id", trackID, "preview_url", streams.PreviewMP3128)
			return "", fmt.Errorf("full stream unavailable (preview-only track)")
		}
		s.log.Warn(
			"no stream URL available",
			"track_id", trackID,
			"hls_aac_160", streams.HLSAAC160URL != "",
			"hls_aac_96", streams.HLSAAC96URL != "",
			"http_mp3_128", streams.HTTPMP3128URL != "",
			"hls_mp3_128", streams.HLSMP3128URL != "",
			"hls_opus_64", streams.HLSOPUS64URL != "",
			"preview_mp3_128", streams.PreviewMP3128 != "",
		)
		return "", fmt.Errorf("no stream available")
	}
	var firstURL string
	var firstKind string
	for i, c := range candidates {
		if i == 0 {
			firstURL, firstKind = c.url, c.kind
		}
		s.log.Debug("stream URL selected", "track_id", trackID, "stream_kind", c.kind, "stream_url", c.url)
		resolved, mode, rerr := s.resolveStreamPlaybackURL(ctx, c.url, token)
		if rerr != nil {
			s.log.Warn("stream URL candidate rejected", "track_id", trackID, "stream_kind", c.kind, "err", rerr)
			continue
		}
		if resolved != c.url {
			s.log.Debug("stream URL resolved", "track_id", trackID, "mode", mode, "stream_kind", c.kind, "raw_url", c.url, "resolved_url", resolved)
		} else {
			s.log.Debug("stream URL resolution no-op", "track_id", trackID, "mode", mode, "stream_kind", c.kind, "resolved_url", resolved)
		}
		return resolved, nil
	}
	if streams.PreviewMP3128 != "" {
		s.log.Warn("all full stream candidates rejected; preview exists but disabled", "track_id", trackID, "preview_url", streams.PreviewMP3128)
		return "", fmt.Errorf("full stream blocked (preview exists only)")
	}
	s.log.Warn("all stream URL candidates rejected; using first raw URL", "track_id", trackID, "stream_kind", firstKind, "stream_url", firstURL)
	return firstURL, nil
}

func (s *Server) probeTrack(token string, trackID int) (*Probe, error) {
	token = s.resolvePlaybackToken(token)
	s.mu.Lock()
	s.token = token
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	streams, err := s.api.TrackStreams(ctx, token, trackID)
	if err != nil {
		return &Probe{TrackID: trackID, Badge: "blocked", Detail: err.Error()}, err
	}

	candidates := streamCandidates(streams)
	if len(candidates) == 0 {
		if streams != nil && streams.PreviewMP3128 != "" {
			return &Probe{TrackID: trackID, Badge: "preview", Detail: "preview-only track"}, nil
		}
		return &Probe{TrackID: trackID, Badge: "blocked", Detail: "no stream candidates"}, nil
	}

	var firstErr error
	for _, c := range candidates {
		_, _, rerr := s.resolveStreamPlaybackURL(ctx, c.url, token)
		if rerr == nil {
			return &Probe{TrackID: trackID, Badge: "playable", Detail: c.kind}, nil
		}
		if firstErr == nil {
			firstErr = rerr
		}
	}

	if streams != nil && streams.PreviewMP3128 != "" {
		return &Probe{TrackID: trackID, Badge: "preview", Detail: "full streams blocked; preview exists"}, nil
	}
	if firstErr != nil {
		return &Probe{TrackID: trackID, Badge: "blocked", Detail: firstErr.Error()}, nil
	}
	return &Probe{TrackID: trackID, Badge: "blocked", Detail: "all stream candidates rejected"}, nil
}

func (s *Server) resolveStreamPlaybackURL(ctx context.Context, streamURL, token string) (resolved, mode string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "OAuth "+token)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
	bodyPreview := truncateLogBody(strings.TrimSpace(string(body)), 500)
	s.log.Debug(
		"stream URL fetch response",
		"stream_url", streamURL,
		"status", res.Status,
		"content_type", res.Header.Get("content-type"),
		"location", res.Header.Get("Location"),
		"body_preview", bodyPreview,
	)

	if loc := res.Header.Get("Location"); loc != "" && res.StatusCode >= 300 && res.StatusCode < 400 {
		return loc, "redirect", nil
	}

	if strings.Contains(strings.ToLower(res.Header.Get("content-type")), "application/json") {
		var payload struct {
			URL     string `json:"url"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.URL) != "" {
			return strings.TrimSpace(payload.URL), "json.url", nil
		}
	}

	txt := strings.TrimSpace(string(body))
	if strings.HasPrefix(txt, "https://") || strings.HasPrefix(txt, "http://") {
		return txt, "body.url", nil
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return streamURL, "direct", nil
	}
	return "", "", fmt.Errorf("status=%s body=%q", res.Status, truncateLogBody(txt, 220))
}

func truncateLogBody(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (s *Server) state() *State {
	s.mu.Lock()
	defer s.mu.Unlock()
	var elapsedMs int64
	var durationMs int64
	if s.playing || s.paused {
		if pos, err := s.player.TimePos(); err == nil {
			elapsedMs = pos.Milliseconds()
		}
		if dur, err := s.player.Duration(); err == nil {
			durationMs = dur.Milliseconds()
		}
	}
	return &State{
		Playing:    s.playing,
		Paused:     s.paused,
		Track:      s.now,
		Index:      s.index,
		Volume:     s.volume,
		Shuffle:    s.shuffle,
		Repeat:     s.repeat,
		ElapsedMs:  elapsedMs,
		DurationMs: durationMs,
		LastError:  s.lastError,
	}
}

func requestAttrs(req Request) []any {
	attrs := []any{
		"cmd", req.Cmd,
		"token_provided", req.Token != "",
	}
	if req.Track != nil {
		attrs = append(attrs, "track_id", req.Track.ID, "track_title", req.Track.Title)
	}
	if len(req.Queue) > 0 {
		attrs = append(attrs, "queue_len", len(req.Queue))
	}
	if req.StartIdx != 0 {
		attrs = append(attrs, "start_idx", req.StartIdx)
	}
	if req.SeekSec != 0 {
		attrs = append(attrs, "seek_sec", req.SeekSec)
	}
	if req.Volume != 0 {
		attrs = append(attrs, "volume", req.Volume)
	}
	if req.Shuffle != nil {
		attrs = append(attrs, "shuffle", *req.Shuffle)
	}
	if req.Repeat != nil {
		attrs = append(attrs, "repeat", *req.Repeat)
	}
	return attrs
}
