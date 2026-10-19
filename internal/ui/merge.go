package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

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
		"%s#%d  %s\n\nchecks: %s   approved: %s   up to date: %s\n\napprove and squash-merge? y/n",
		pr.Repo, pr.Number, pr.Title,
		formatChecks(pr.Checks), formatApproval(pr.Approval), formatBranchStatus(pr.Branch),
	)
}
