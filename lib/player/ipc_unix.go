//go:build !windows

package player

import (
	"errors"
	"net"
	"os"
	"time"
)

func dialIPC(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}

func waitForIPC(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("mpv ipc socket not ready")
}
