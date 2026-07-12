//go:build !unix

package sandbox

import "os/exec"

// setProcessGroup is a no-op on platforms without POSIX process groups
// (e.g. Windows); the timeout still kills the direct child via killProcessGroup.
func setProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup kills the direct child process. Full process-tree teardown
// is only available on unix (see procattr_unix.go).
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
