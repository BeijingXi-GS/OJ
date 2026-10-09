package filejudge

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInteractorCompileCannotPrecedeParticipantArtifact(t *testing.T) {
	for _, status := range []string{"compile_error", "system_error"} {
		r := interactiveWireResult()
		r.Compile.Status = status
		r.Cases = []CaseResult{}
		r.ArtifactSHA256 = ""
		r.Verdict = "SYSTEM_ERROR"
		if status == "compile_error" {
			r.Verdict = "CE"
		}
		if validResult(r) || MatchesProblemType(r, "interactive") {
			t.Fatal("interactor compile accepted before participant success")
		}
	}
	r := interactiveWireResult()
	data, _ := json.Marshal(r)
	missing := bytes.Replace(data, []byte(`"input":"617\n",`), nil, 1)
	if completeWire(missing) {
		t.Fatal("missing snapshot display accepted")
	}
}
