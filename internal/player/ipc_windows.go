//go:build windows

package player

import (
	"errors"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func dialIPC(path string) (net.Conn, error) {
	return winio.DialPipe(path, nil)
}

func waitForIPC(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := winio.DialPipe(path, &timeout)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("mpv ipc pipe not ready")
}
