package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
)

func TestFormatTitleDraft(t *testing.T) {
	pr := analysis.PullRequest{Number: 5, Title: "add feature", IsDraft: true}
	got := formatTitle(pr, false)
	if !strings.Contains(got, "[draft]") || !strings.Contains(got, "#5 add feature") {
		t.Errorf("formatTitle(draft) = %q, want to contain [draft] and #5 add feature", got)
	}
}

func TestFormatTitleNonDraft(t *testing.T) {
	pr := analysis.PullRequest{Number: 5, Title: "add feature"}
	got := formatTitle(pr, false)
	if strings.Contains(got, "[draft]") {
		t.Errorf("formatTitle(non-draft) = %q, should not contain [draft]", got)
	}
	if !strings.Contains(got, "#5 add feature") {
		t.Errorf("formatTitle(non-draft) = %q, want to contain #5 add feature", got)
	}
}

func TestFormatChecks(t *testing.T) {
	tests := []struct {
		name     string
		cs       analysis.CheckSummary
		wantPass string
		wantFail string
	}{
		{"all passing", analysis.CheckSummary{Pass: 5, Overall: analysis.StatusPassing}, "5", "0"},
		{"failing", analysis.CheckSummary{Pass: 5, Fail: 1, Overall: analysis.StatusFailing}, "5", "1"},
		{"pending ignored", analysis.CheckSummary{Pass: 3, Pending: 2, Overall: analysis.StatusPending}, "3", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatChecks(tt.cs)
			wantPass := styleGreen.Render(tt.wantPass)
			wantFail := styleRed.Render(tt.wantFail)
			if !strings.Contains(got, wantPass) || !strings.Contains(got, wantFail) {
				t.Errorf("formatChecks(%+v) = %q, want to contain %q and %q", tt.cs, got, wantPass, wantFail)
			}
		})
	}
}

func TestFormatChecksUnknown(t *testing.T) {
	got := formatChecks(analysis.CheckSummary{Overall: analysis.StatusUnknown})
	if !strings.Contains(got, "-") {
		t.Errorf("formatChecks(unknown) = %q, want to contain %q", got, "-")
	}
}

func TestFormatApproval(t *testing.T) {
	tests := []struct {
		name  string
		state analysis.ApprovalState
		want  string
	}{
		{"approved", analysis.ApprovalApproved, styleGreen.Render("●")},
		{"changes requested", analysis.ApprovalChangesRequested, styleGray.Render("○")},
		{"review required", analysis.ApprovalReviewRequired, styleGray.Render("○")},
		{"unknown", analysis.ApprovalUnknown, styleGray.Render("○")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatApproval(tt.state)
			if got != tt.want {
				t.Errorf("formatApproval(%v) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestFormatBranchStatus(t *testing.T) {
	tests := []struct {
		name string
		bs   analysis.BranchStatus
		want string
	}{
		{"up to date", analysis.BranchUpToDate, styleGreen.Render("●")},
		{"behind", analysis.BranchBehind, styleRed.Render("●")},
		{"unknown", analysis.BranchUnknown, styleGray.Render("○")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBranchStatus(tt.bs)
			if got != tt.want {
				t.Errorf("formatBranchStatus(%v) = %q, want %q", tt.bs, got, tt.want)
			}
		})
	}
}

func TestFormatTitleSelected(t *testing.T) {
	pr := analysis.PullRequest{Number: 5, Title: "add feature", IsDraft: true}
	got := formatTitle(pr, true)
	want := styleGray.Render("[draft] ") + styleSelected.Render("#5 add feature")
	if got != want {
		t.Errorf("formatTitle(selected) = %q, want %q", got, want)
	}
}

func TestBuildLines(t *testing.T) {
	prs := []analysis.PullRequest{
		{Number: 1, Repo: "a/x"},
		{Number: 2, Repo: "a/x"},
		{Number: 3, Repo: "b/y"},
	}
	got := buildLines(prs)
	want := []line{
		{repo: "a/x", pr: -1},
		{pr: 0},
		{pr: 1},
		{repo: "b/y", pr: -1},
		{pr: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildLines() = %+v, want %+v", got, want)
	}
}

func TestFit(t *testing.T) {
	coloured := "\x1b[32m12\x1b[0m/\x1b[31m10\x1b[0m"
	tests := []struct {
		name string
		s    string
		w    int
		want string
	}{
		{"colour codes do not count", coloured, 10, coloured + "     "},
		{"long text is cut", "abcdefghij", 5, "abcd…"},
		{"short text is padded", "ab", 5, "ab   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fit(tt.s, tt.w)
			if got != tt.want {
				t.Errorf("fit(%q, %d) = %q, want %q", tt.s, tt.w, got, tt.want)
			}
			if w := ansi.StringWidth(got); w != tt.w {
				t.Errorf("fit(%q, %d) width = %d, want %d", tt.s, tt.w, w, tt.w)
			}
		})
	}
}

func TestTitleWidth(t *testing.T) {
	tests := []struct {
		name      string
		termWidth int
		want      int
	}{
		{"unknown width", 0, defaultTitleWidth},
		{"normal width", 120, 68},
		{"narrow width", 40, minTitleWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleWidth(tt.termWidth); got != tt.want {
				t.Errorf("titleWidth(%d) = %d, want %d", tt.termWidth, got, tt.want)
			}
		})
	}
}

func TestPRLineViewFitsTerminalWidth(t *testing.T) {
	pr := analysis.PullRequest{
		Number:   4501,
		Title:    strings.Repeat("long title ", 20),
		Checks:   analysis.CheckSummary{Pass: 12, Fail: 10, Overall: analysis.StatusFailing},
		Comments: 7,
		Age:      "2w",
	}
	got := prLineView(pr, true, titleWidth(120))
	if w := ansi.StringWidth(got); w != 120 {
		t.Errorf("prLineView() width = %d, want 120 (line %q)", w, got)
	}
	if !strings.Contains(got, "12/10") {
		t.Errorf("prLineView() = %q, want to contain the full checks value 12/10", got)
	}
}
