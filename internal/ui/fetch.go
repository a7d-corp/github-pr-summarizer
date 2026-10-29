package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// fetchedMsg is delivered when a fetchCmd completes.
type fetchedMsg struct {
	prs []analysis.PullRequest
	err error
}

// fetchCmd fetches open PRs off the UI goroutine so the render
// loop never blocks on network I/O.
func fetchCmd(fetcher github.Fetcher, allowlist map[string]struct{}) tea.Cmd {
	return func() tea.Msg {
		raws, err := fetcher.SearchOpenPRs(context.Background())
		if err != nil {
			return fetchedMsg{err: err}
		}
		return fetchedMsg{prs: analysis.BuildPullRequests(raws, allowlist, time.Now())}
	}
}
