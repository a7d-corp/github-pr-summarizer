package github

import (
	"context"
	"fmt"
)

// MockFetcher is a test double for Fetcher and Merger.
type MockFetcher struct {
	PRs      []PullRequest
	Err      error
	MergeErr error
	Merged   []string
}

var _ Fetcher = (*MockFetcher)(nil)
var _ Merger = (*MockFetcher)(nil)

// SearchReviewRequestedPRs returns the configured PRs or error.
func (m *MockFetcher) SearchReviewRequestedPRs(ctx context.Context) ([]PullRequest, error) {
	return m.PRs, m.Err
}

// ApproveAndMerge records the PR as merged and returns the configured error.
func (m *MockFetcher) ApproveAndMerge(ctx context.Context, repo string, number int) error {
	m.Merged = append(m.Merged, fmt.Sprintf("%s#%d", repo, number))
	return m.MergeErr
}
