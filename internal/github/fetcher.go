package github

import "context"

// Fetcher retrieves open PRs in the configured repos.
type Fetcher interface {
	SearchOpenPRs(ctx context.Context) ([]PullRequest, error)
}

// Merger approves (if approve is true) and merges a single PR.
type Merger interface {
	ApproveAndMerge(ctx context.Context, repo string, number int, approve bool) error
}
