package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// mergedMsg is delivered when a mergeCmd completes.
type mergedMsg struct {
	repo   string
	number int
	err    error
}

// mergeCmd approves and merges a PR off the UI goroutine.
func mergeCmd(merger github.Merger, pr analysis.PullRequest) tea.Cmd {
	return func() tea.Msg {
		err := merger.ApproveAndMerge(context.Background(), pr.Repo, pr.Number)
		return mergedMsg{repo: pr.Repo, number: pr.Number, err: err}
	}
}

func confirmView(pr analysis.PullRequest) string {
	return fmt.Sprintf(
		"%s#%d  %s\n\napproval: %s   up to date: %s\n\n%s\n\napprove and squash-merge? y/n",
		pr.Repo, pr.Number, pr.Title,
		formatApprovalText(pr.Approval), formatBranchStatus(pr.Branch),
		checksTable(pr.CheckRuns),
	)
}

// checksTable renders one row for each check, with its name and status.
func checksTable(checks []analysis.Check) string {
	if len(checks) == 0 {
		return styleGray.Render("no checks")
	}

	nameW := len("Check")
	for _, c := range checks {
		nameW = max(nameW, ansi.StringWidth(c.Name))
	}

	rows := []string{styleHeader.Render(fit("Check", nameW) + "  Status")}
	for _, c := range checks {
		rows = append(rows, fit(c.Name, nameW)+"  "+formatCheckStatus(c.Status))
	}
	return strings.Join(rows, "\n")
}

func formatCheckStatus(s analysis.Status) string {
	switch s {
	case analysis.StatusPassing:
		return styleGreen.Render("passed")
	case analysis.StatusFailing:
		return styleRed.Render("failed")
	case analysis.StatusPending:
		return styleYellow.Render("pending")
	default:
		return styleGray.Render("unknown")
	}
}

func formatApprovalText(state analysis.ApprovalState) string {
	switch state {
	case analysis.ApprovalApproved:
		return styleGreen.Render("approved")
	case analysis.ApprovalChangesRequested:
		return styleRed.Render("changes requested")
	case analysis.ApprovalReviewRequired:
		return styleYellow.Render("review required")
	default:
		return styleGray.Render("none")
	}
}
