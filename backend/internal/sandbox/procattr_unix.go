//go:build unix

package sandbox

import (
	"os/exec"
	"syscall"
)

// setProcessGroup places the command in its own process group so the whole
// tree can be signalled at once.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup SIGKILLs the command's entire process group. A negative PID
// targets the group, so grandchildren (e.g. the binary `go run` spawns) die too.
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
