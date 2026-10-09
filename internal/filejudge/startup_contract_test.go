package filejudge

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInteractiveStartupDeadlineWire(t *testing.T) {
	for _, name := range []string{"valid", "batch", "AC", "WA", "RE", "participant_started", "interactor_started", "participant_OOM", "interactor_OOM", "stopped", "interactor_not_compiled"} {
		t.Run(name, func(t *testing.T) {
			r := interactiveWireResult()
			r.Verdict = "TLE"
			c := &r.Cases[0]
			c.Verdict = "TLE"
			c.StartTimedOut = true
			c.ExitCode = nil
			c.Interactor.ExitCode = nil
			code := 0
			switch name {
			case "batch":
				r.InteractorCompile = nil
				r.InteractorSourceSHA256 = ""
				r.InteractorArtifactSHA256 = ""
				c.Interactor = nil
			case "AC", "WA", "RE":
				r.Verdict = name
				c.Verdict = name
			case "participant_started":
				c.ExitCode = &code
			case "interactor_started":
				c.Interactor.ExitCode = &code
			case "participant_OOM":
				c.OOMKilled = true
			case "interactor_OOM":
				c.Interactor.OOMKilled = true
			case "stopped":
				c.Interactor.ParticipantStopped = true
			case "interactor_not_compiled":
				r.InteractorCompile.Status = "system_error"
			}
			if got := validResult(r); got != (name == "valid") {
				t.Fatalf("%s accepted=%v", name, got)
			}
		})
	}
	r := interactiveWireResult()
	data, _ := json.Marshal(r)
	for _, value := range []string{"null", "\"true\"", "1"} {
		malformed := bytes.Replace(data, []byte(`"input":"617\n"`), []byte(`"start_timed_out":`+value+`,"input":"617\n"`), 1)
		var result Result
		err := json.Unmarshal(malformed, &result)
		if err == nil && completeWire(malformed) && validResult(result) {
			t.Fatalf("invalid startup marker %s accepted", value)
		}
	}
}
