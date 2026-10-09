//go:build !linux

package compilerworker

import (
	"os/exec"
	"time"
)

// Used by portable unit tests. The MVP deploys the worker on Linux only.
func configureProcess(cmd *exec.Cmd) { cmd.WaitDelay = time.Second }
