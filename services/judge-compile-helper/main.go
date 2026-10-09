// The helper has fixed arguments and exports bytes; it never executes the artifact.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() { os.Exit(compile()) }
func compile() int {
	cmd := exec.Command("/usr/bin/g++", "-std=c++20", "-O0", "-fdiagnostics-color=never", "/source/solution.cpp", "-o", "/work/program")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return 1
		}
		fmt.Fprintln(os.Stderr, "compiler helper failed")
		return 125
	}
	st, err := os.Lstat("/work/program")
	if err != nil || !st.Mode().IsRegular() || st.Size() <= 0 || st.Size() > 32*1048576 {
		fmt.Fprintln(os.Stderr, "artifact invalid")
		return 125
	}
	f, err := os.Open("/work/program")
	if err != nil {
		return 125
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return 125
	}
	n, err := io.Copy(os.Stdout, f)
	if err != nil || n != st.Size() {
		return 125
	}
	return 0
}
