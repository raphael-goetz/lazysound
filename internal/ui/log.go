package ui

import (
	"fmt"
	"os"
	"time"
)

const logPath = "/tmp/lazysound.log"

func logError(msg string) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().Format(time.RFC3339)
	_, _ = fmt.Fprintf(f, "[%s] %s\n", ts, msg)
}
