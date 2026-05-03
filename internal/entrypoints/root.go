package entrypoints

import (
	"fmt"
	"io"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		rootUsage(stdout)
		return 1
	}

	mode := strings.ToLower(strings.TrimSpace(args[0]))
	modeArgs := args[1:]

	switch mode {
	case "cli", "ctl", "control":
		return RunCtl(modeArgs, stdout, stderr)
	case "daemon", "d":
		return RunDaemon(stderr)
	case "tui", "ui":
		return RunTUI(stderr)
	case "help", "-h", "--help":
		rootUsage(stdout)
		return 0
	default:
		fmt.Fprintln(stderr, "unknown mode:", mode)
		rootUsage(stderr)
		return 1
	}
}

func rootUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  lazysound <cli|daemon|tui> [args]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "examples:")
	fmt.Fprintln(w, "  lazysound tui")
	fmt.Fprintln(w, "  lazysound daemon")
	fmt.Fprintln(w, "  lazysound cli status")
	fmt.Fprintln(w, "  lazysound cli volume 70")
}
