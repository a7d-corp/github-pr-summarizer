package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// rebasedMsg is delivered when a rebaseCmd completes.
type rebasedMsg struct {
	repo   string
	number int
	err    error
}

// rebaseCmd rebases a PR's branch onto its base branch off the UI goroutine.
func rebaseCmd(merger github.Merger, pr analysis.PullRequest) tea.Cmd {
	return func() tea.Msg {
		err := merger.RebaseBranch(context.Background(), pr.ID)
		return rebasedMsg{repo: pr.Repo, number: pr.Number, err: err}
	}
}

func rebaseConfirmView(pr analysis.PullRequest) string {
	return fmt.Sprintf(
		"%s#%d  %s\n\napproval: %s   up to date: %s\n\nrebase onto the base branch? y/n",
		pr.Repo, pr.Number, pr.Title,
		formatApprovalText(pr.Approval), formatBranchStatus(pr.Branch, styleNone),
	)
}
