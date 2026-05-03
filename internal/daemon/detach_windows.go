//go:build windows
// +build windows

package daemon

import (
	"os/exec"
	"syscall"
)

func detachCmd(cmd *exec.Cmd) {
	const detachedProcess = 0x00000008
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
}
