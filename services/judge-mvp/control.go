package main

import (
	"bufio"
	"context"
	"io"
)

// The supervisor owns this private pipe. Standalone instances without a control
// token retain their terminal input and are never stopped by this reader.
func watchOwnedShutdown(input io.Reader, cancel context.CancelFunc, enabled bool) {
	if !enabled {
		return
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 16), 16)
	for scanner.Scan() {
		if scanner.Text() == "shutdown" {
			cancel()
			return
		}
	}
	// EOF or a broken/bounded pipe means the owning supervisor disappeared.
	// Cancellation still lets the judge's lifecycle perform independent cleanup.
	cancel()
}
