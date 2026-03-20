package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/raphael-goetz/lazysound/internal/api"
)

const DefaultAddr = "127.0.0.1:7777"

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	if addr == "" {
		addr = DefaultAddr
	}
	return &Client{addr: addr}
}

func (c *Client) EnsureRunning(ctx context.Context) error {
	_, err := c.Status(ctx)
	if err == nil {
		return nil
	}
	if err := c.startDaemon(); err != nil {
		return err
	}
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		if _, err := c.Status(ctx); err == nil {
			return nil
		}
	}
	return errors.New("daemon not ready")
}

func (c *Client) Status(ctx context.Context) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "status"})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) PlayTrack(ctx context.Context, token string, t api.Track) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "play_track", Token: token, Track: &t})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) PlayQueue(ctx context.Context, token string, q []api.Track, start int) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "play_queue", Token: token, Queue: q, StartIdx: start})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) Stop(ctx context.Context) (*State, error) {
	return c.simple(ctx, "stop")
}

func (c *Client) Restart(ctx context.Context) (*State, error) {
	return c.simple(ctx, "restart")
}

func (c *Client) TogglePause(ctx context.Context) (*State, error) {
	return c.simple(ctx, "pause")
}

func (c *Client) Seek(ctx context.Context, sec int) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "seek", SeekSec: sec})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) SetVolume(ctx context.Context, vol int) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "volume", Volume: vol})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) SetShuffle(ctx context.Context, enabled bool) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "shuffle", Shuffle: &enabled})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) SetRepeat(ctx context.Context, enabled bool) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: "repeat", Repeat: &enabled})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) Next(ctx context.Context) (*State, error) {
	return c.simple(ctx, "next")
}

func (c *Client) Prev(ctx context.Context) (*State, error) {
	return c.simple(ctx, "prev")
}

func (c *Client) simple(ctx context.Context, cmd string) (*State, error) {
	resp, err := c.call(ctx, Request{Cmd: cmd})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp.State, nil
}

func (c *Client) call(ctx context.Context, req Request) (Response, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()
	bw := bufio.NewWriter(conn)
	if err := json.NewEncoder(bw).Encode(req); err != nil {
		return Response{}, err
	}
	if err := bw.Flush(); err != nil {
		return Response{}, err
	}
	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return Response{}, err
	}
	return resp, nil
}

func (c *Client) startDaemon() error {
	bin, err := exec.LookPath("lazysoundd")
	if err != nil {
		return err
	}
	cmd := exec.Command(bin)
	detachCmd(cmd)
	if runtime.GOOS != "windows" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Start()
}
