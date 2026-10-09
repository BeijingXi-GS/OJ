package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problemstore"
)

func newFixtureServices(ping PingFunc, compiler compilation.Compiler, judge filejudge.Judge) http.Handler {
	return NewWithCatalog(ping, compiler, judge, &problemstore.Store{BuiltinRoot: "../../testdata/problems"})
}

type fakeJudge struct{ calls int }

func (j *fakeJudge) Judge(_ context.Context, r filejudge.Request) (filejudge.Result, *compilation.Error) {
	j.calls++
	snap, err := filejudge.Load("../../testdata/problems", r.ProblemID)
	if err != nil {
		panic(err)
	}
	return filejudge.Result{Verdict: "CE", TestdataSHA256: snap.SHA256, Compile: filejudge.CompileResult{Status: "compile_error"}, Cases: []filejudge.CaseResult{}}, nil
}
func TestJudgeAPIValidation(t *testing.T) {
	for _, tc := range []struct {
		body   string
		judge  bool
		status int
	}{
		{`{"problem_id":"a-plus-b","language":"cpp","source":"int main(){}"}`, true, 200},
		{`{"problem_id":"a-plus-b","language":"cpp","source":"int main(){}"}`, false, 503},
		{`{"problem_id":"missing","language":"cpp","source":"int main(){}"}`, true, 404},
		{`{"problem_id":"a-plus-b","language":"python","source":"pass"}`, true, 400},
		{`{"problem_id":"a-plus-b","language":"cpp","source":"  "}`, true, 400},
		{`{"problem_id":"a-plus-b","language":"cpp","source":"int main(){}","command":"rm"}`, true, 400},
		{`{"problem_id":"a-plus-b","language":"cpp","source":"` + strings.Repeat("x", 65537) + `"}`, true, 413},
		{strings.Repeat(" ", 262145), true, 413},
	} {
		j := &fakeJudge{}
		var service filejudge.Judge
		if tc.judge {
			service = j
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/judgements", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		newFixtureServices(nil, nil, service).ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("want %d got %d: %s", tc.status, w.Code, w.Body)
		}
		if tc.status != 200 && j.calls != 0 {
			t.Fatal("invalid request reached worker")
		}
	}
}
