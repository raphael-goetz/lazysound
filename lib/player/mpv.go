package player

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type MPV struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	cmdID     uint64
	stopReqID uint64
	procLog   *tailWriter
	ipcPath   string
	volume    int
	playing   bool
	lastURL   string
	binPath   string
}

func NewMPV() (*MPV, error) {
	bin, err := exec.LookPath("mpv")
	if err != nil {
		return nil, errors.New("mpv not found in PATH")
	}
	return &MPV{binPath: bin, volume: 65}, nil
}

func (p *MPV) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

func (p *MPV) Volume() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

func (p *MPV) Play(url string, volume int, token string) error {
	// We drive playback via an external mpv process and control it through IPC.
	p.mu.Lock()
	defer p.mu.Unlock()
	if url == "" {
		return errors.New("empty stream url")
	}
	if p.cmd != nil && p.playing {
		// Replacement playback should not be treated as "natural end".
		p.stopReqID = p.cmdID
		_ = p.stopAndWaitLocked(2 * time.Second)
	}
	p.lastURL = url
	p.volume = volume
	p.ipcPath = ipcPath()

	args := []string{
		"--no-video",
		"--no-ytdl",
		"--msg-level=all=warn",
		fmt.Sprintf("--volume=%d", p.volume),
		fmt.Sprintf("--input-ipc-server=%s", p.ipcPath),
	}
	if token != "" {
		args = append(args, fmt.Sprintf("--http-header-fields=Authorization: OAuth %s", token))
	}
	args = append(args, url)
	cmd := exec.Command(p.binPath, args...)
	p.attachLogs(cmd)
	p.log("mpv start: " + url)
	if err := cmd.Start(); err != nil {
		p.log("mpv start error: " + err.Error())
		return err
	}
	p.cmdID++
	p.cmd = cmd
	p.playing = true

	if err := waitForIPC(p.ipcPath, 8*time.Second); err != nil {
		p.log("ipc socket not ready: " + err.Error())
	}
	return nil
}

func (p *MPV) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.log("mpv stop requested")
	p.stopReqID = p.cmdID
	return p.stopAndWaitLocked(2 * time.Second)
}

func (p *MPV) Restart() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.playing {
		return errors.New("not playing")
	}
	p.log("mpv restart requested")
	return p.sendLocked(map[string]any{"command": []any{"seek", 0, "absolute"}})
}

func (p *MPV) TimePos() (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.playing {
		return 0, errors.New("not playing")
	}
	v, err := p.getPropertyLocked("time-pos")
	if err != nil {
		return 0, err
	}
	return time.Duration(v * float64(time.Second)), nil
}

func (p *MPV) Duration() (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.playing {
		return 0, errors.New("not playing")
	}
	v, err := p.getPropertyLocked("duration")
	if err != nil {
		return 0, err
	}
	return time.Duration(v * float64(time.Second)), nil
}

func (p *MPV) SetVolume(vol int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if vol < 0 {
		vol = 0
	}
	if vol > 100 {
		vol = 100
	}
	p.volume = vol
	if !p.playing {
		return nil
	}
	p.log(fmt.Sprintf("mpv set volume: %d", vol))
	return p.sendLocked(map[string]any{"command": []any{"set_property", "volume", vol}})
}

func (p *MPV) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.playing {
		return errors.New("not playing")
	}
	return p.sendLocked(map[string]any{"command": []any{"cycle", "pause"}})
}

func (p *MPV) SeekSeconds(delta int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.playing {
		return errors.New("not playing")
	}
	return p.sendLocked(map[string]any{"command": []any{"seek", delta, "relative"}})
}

func (p *MPV) stopLocked() error {
	if p.cmd == nil {
		p.playing = false
		return nil
	}
	_ = p.sendLocked(map[string]any{"command": []any{"quit"}})
	if p.cmd.Process != nil {
		if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
			p.log("mpv interrupt error: " + err.Error())
		}
		_ = p.cmd.Process.Kill()
	}
	p.cmd = nil
	p.playing = false
	return nil
}

func (p *MPV) stopAndWaitLocked(timeout time.Duration) error {
	if p.cmd == nil {
		p.playing = false
		return nil
	}
	cmd := p.cmd
	cmdID := p.cmdID
	_ = p.sendLocked(map[string]any{"command": []any{"quit"}})
	if cmd.Process != nil {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			p.log("mpv interrupt error: " + err.Error())
		}
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	case <-done:
	}
	if p.cmd == cmd {
		p.cmd = nil
		p.playing = false
	}
	if p.stopReqID == cmdID {
		p.stopReqID = 0
	}
	return nil
}

func (p *MPV) sendLocked(msg any) error {
	if p.ipcPath == "" {
		return errors.New("ipc not ready")
	}
	conn, err := dialIPC(p.ipcPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	return json.NewEncoder(conn).Encode(msg)
}

type ipcResponse struct {
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func (p *MPV) getPropertyLocked(prop string) (float64, error) {
	if p.ipcPath == "" {
		return 0, errors.New("ipc not ready")
	}
	conn, err := dialIPC(p.ipcPath)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	req := map[string]any{"command": []any{"get_property", prop}}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return 0, err
	}
	var resp ipcResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return 0, err
	}
	if resp.Error != "" && resp.Error != "success" {
		return 0, errors.New(resp.Error)
	}
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return 0, errors.New("no data")
	}
	var out float64
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return 0, err
	}
	return out, nil
}

func ipcPath() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`\\.\pipe\lazysound-mpv-%d`, time.Now().UnixNano())
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("lazysound-mpv-%d.sock", time.Now().UnixNano()))
}

type Exit struct {
	Err     error
	Stopped bool
	Detail  string
}

func (p *MPV) WaitExit() <-chan Exit {
	ch := make(chan Exit, 1)
	p.mu.Lock()
	cmd := p.cmd
	cmdID := p.cmdID
	p.mu.Unlock()
	if cmd == nil {
		ch <- Exit{Err: nil, Stopped: false}
		return ch
	}
	go func() {
		err := cmd.Wait()
		p.mu.Lock()
		detail := ""
		if p.procLog != nil {
			detail = p.procLog.Summary()
		}
		stopped := p.stopReqID == cmdID
		if p.stopReqID == cmdID {
			p.stopReqID = 0
		}
		if p.cmd == cmd {
			p.cmd = nil
			p.playing = false
			p.procLog = nil
		}
		p.mu.Unlock()
		ch <- Exit{Err: err, Stopped: stopped, Detail: detail}
	}()
	return ch
}

func (p *MPV) attachLogs(cmd *exec.Cmd) {
	tw := newTailWriter(80)
	p.procLog = tw
	f, err := os.OpenFile(filepath.Join(os.TempDir(), "lazysound-mpv.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		cmd.Stdout = tw
		cmd.Stderr = tw
		return
	}
	mw := io.MultiWriter(f, tw)
	cmd.Stdout = mw
	cmd.Stderr = mw
}

func (p *MPV) log(msg string) {
	f, err := os.OpenFile(filepath.Join(os.TempDir(), "lazysound-mpv.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

type tailWriter struct {
	mu      sync.Mutex
	lines   []string
	partial string
	max     int
}

func newTailWriter(max int) *tailWriter {
	return &tailWriter{max: max}
}

func (t *tailWriter) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	data := t.partial + string(b)
	parts := strings.Split(data, "\n")
	if len(parts) == 0 {
		return len(b), nil
	}
	for i := 0; i < len(parts)-1; i++ {
		line := strings.TrimSpace(parts[i])
		if line != "" {
			t.lines = append(t.lines, line)
		}
	}
	t.partial = parts[len(parts)-1]
	if len(t.lines) > t.max {
		t.lines = t.lines[len(t.lines)-t.max:]
	}
	return len(b), nil
}

func (t *tailWriter) Summary() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.lines) == 0 {
		return ""
	}
	keys := []string{
		"HTTP Error",
		"Unauthorized",
		"Forbidden",
		"Failed to open",
		"Opening failed",
		"Errors when loading file",
	}
	var picked []string
	for _, ln := range t.lines {
		for _, k := range keys {
			if strings.Contains(ln, k) {
				picked = append(picked, ln)
				break
			}
		}
	}
	use := picked
	if len(use) == 0 {
		if len(t.lines) > 8 {
			use = t.lines[len(t.lines)-8:]
		} else {
			use = t.lines
		}
	}
	out := strings.Join(use, " | ")
	if len(out) > 1200 {
		var b bytes.Buffer
		b.WriteString(out[:1200])
		b.WriteString("...")
		return b.String()
	}
	return out
}
