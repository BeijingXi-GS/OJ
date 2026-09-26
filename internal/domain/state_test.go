package domain

import "testing"

func TestRunLifecycleAndTerminalImmutability(t *testing.T) {
	path := []RunState{Pending, Queued, Compiling, Judging, Finished}
	for i := 1; i < len(path); i++ {
		if !CanTransition(path[i-1], path[i]) {
			t.Fatalf("valid transition rejected: %s -> %s", path[i-1], path[i])
		}
	}
	for _, terminal := range []RunState{Finished, Cancelled} {
		for _, next := range []RunState{Pending, Queued, Compiling, Judging, Finished, Cancelled} {
			if CanTransition(terminal, next) {
				t.Fatalf("terminal run mutated: %s -> %s", terminal, next)
			}
		}
	}
	for _, pair := range [][2]RunState{{Pending, Judging}, {Compiling, Queued}, {Judging, Compiling}, {RunState("UNKNOWN"), Pending}} {
		if CanTransition(pair[0], pair[1]) {
			t.Fatalf("invalid transition accepted: %v", pair)
		}
	}
}
