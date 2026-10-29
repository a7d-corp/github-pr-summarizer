package github

import "context"

// Fetcher retrieves open PRs in the configured repos.
type Fetcher interface {
	SearchOpenPRs(ctx context.Context) ([]PullRequest, error)
}

// Merger approves and merges a single PR.
type Merger interface {
	ApproveAndMerge(ctx context.Context, repo string, number int) error
}
