package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/raphael-goetz/lazysound/lib/player"
	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

type Server struct {
	addr   string
	api    *api.ApiClient
	player *player.MPV

	mu       sync.Mutex
	token    string
	queue    []api.Track
	index    int
	shuffle  bool
	repeat   bool
	paused   bool
	volume   int
	now      *api.Track
	playing  bool
	stopChan chan struct{}
}

func NewServer(addr string, apiClient *api.ApiClient) (*Server, error) {
	p, err := player.NewMPV()
	if err != nil {
		return nil, err
	}
	return &Server{
		addr:     addr,
		api:      apiClient,
		player:   p,
		volume:   65,
		stopChan: make(chan struct{}, 1),
	}, nil
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		var req Request
		if err := json.Unmarshal(sc.Bytes(), &req); err != nil {
			_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
			return
		}
		resp := s.handleRequest(req)
		_ = json.NewEncoder(conn).Encode(resp)
		return
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
	case "stop":
		s.player.Stop()
		s.mu.Lock()
		s.playing = false
		s.now = nil
		s.paused = false
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
	track := s.queue[idx]
	token := s.token
	vol := s.volume
	s.mu.Unlock()

	url, err := s.streamURL(track.ID, token)
	if err != nil {
		return err
	}
	_ = s.player.Stop()
	if err := s.player.Play(url, vol, token); err != nil {
		return err
	}
	s.mu.Lock()
	s.now = &track
	s.playing = true
	s.paused = false
	s.mu.Unlock()
	go s.waitEnd(idx)
	return nil
}

func (s *Server) waitEnd(idx int) {
	exit := <-s.player.WaitExit()
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
			s.mu.Unlock()
			return Response{OK: true, State: s.state()}
		}
	}
	s.index = next
	s.mu.Unlock()
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
	url := chooseStreamURL(streams)
	if url == "" {
		return "", fmt.Errorf("no stream available")
	}
	return url, nil
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
	}
}
