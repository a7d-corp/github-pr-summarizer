package github

import "context"

// Fetcher retrieves open PRs where the authenticated user is a requested reviewer.
type Fetcher interface {
	SearchReviewRequestedPRs(ctx context.Context) ([]PullRequest, error)
}

// Merger approves and merges a single PR.
type Merger interface {
	ApproveAndMerge(ctx context.Context, repo string, number int) error
}
