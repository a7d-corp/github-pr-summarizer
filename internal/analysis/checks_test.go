package analysis

import (
	"testing"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

func checkRun(status, conclusion string) github.CheckContext {
	return github.CheckContext{Typename: "CheckRun", Status: status, Conclusion: conclusion}
}

func statusContext(state string) github.CheckContext {
	return github.CheckContext{Typename: "StatusContext", State: state}
}

func TestTallyChecksNilRollup(t *testing.T) {
	got := TallyChecks(nil)
	want := CheckSummary{Overall: StatusUnknown}
	if got != want {
		t.Errorf("TallyChecks(nil) = %+v, want %+v", got, want)
	}
}

func TestTallyChecksEmptyContexts(t *testing.T) {
	rollup := &github.StatusCheckRollup{}
	got := TallyChecks(rollup)
	want := CheckSummary{Overall: StatusUnknown}
	if got != want {
		t.Errorf("TallyChecks(empty) = %+v, want %+v", got, want)
	}
}

func TestTallyChecks(t *testing.T) {
	tests := []struct {
		name     string
		contexts []github.CheckContext
		want     CheckSummary
	}{
		{
			name:     "all passing check runs",
			contexts: []github.CheckContext{checkRun("COMPLETED", "SUCCESS"), checkRun("COMPLETED", "NEUTRAL")},
			want:     CheckSummary{Pass: 2, Overall: StatusPassing},
		},
		{
			name:     "one failing check run wins overall",
			contexts: []github.CheckContext{checkRun("COMPLETED", "SUCCESS"), checkRun("COMPLETED", "FAILURE")},
			want:     CheckSummary{Pass: 1, Fail: 1, Overall: StatusFailing},
		},
		{
			name:     "pending check run",
			contexts: []github.CheckContext{checkRun("IN_PROGRESS", "")},
			want:     CheckSummary{Pending: 1, Overall: StatusPending},
		},
		{
			name:     "pending beats passing when no failures",
			contexts: []github.CheckContext{checkRun("COMPLETED", "SUCCESS"), checkRun("QUEUED", "")},
			want:     CheckSummary{Pass: 1, Pending: 1, Overall: StatusPending},
		},
		{
			name:     "legacy status context success",
			contexts: []github.CheckContext{statusContext("SUCCESS")},
			want:     CheckSummary{Pass: 1, Overall: StatusPassing},
		},
		{
			name:     "legacy status context failure",
			contexts: []github.CheckContext{statusContext("ERROR")},
			want:     CheckSummary{Fail: 1, Overall: StatusFailing},
		},
		{
			name:     "legacy status context pending",
			contexts: []github.CheckContext{statusContext("PENDING")},
			want:     CheckSummary{Pending: 1, Overall: StatusPending},
		},
		{
			name:     "unknown typename ignored",
			contexts: []github.CheckContext{{Typename: "SomethingElse"}},
			want:     CheckSummary{Overall: StatusUnknown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollup := &github.StatusCheckRollup{}
			rollup.Contexts.Nodes = tt.contexts
			got := TallyChecks(rollup)
			if got != tt.want {
				t.Errorf("TallyChecks() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
