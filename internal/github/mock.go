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
	Approved []bool
	Rebased  []string
}

var _ Fetcher = (*MockFetcher)(nil)
var _ Merger = (*MockFetcher)(nil)

// SearchOpenPRs returns the configured PRs or error.
func (m *MockFetcher) SearchOpenPRs(ctx context.Context) ([]PullRequest, error) {
	return m.PRs, m.Err
}

// ApproveAndMerge records the PR as merged and returns the configured error.
func (m *MockFetcher) ApproveAndMerge(ctx context.Context, repo string, number int, approve bool) error {
	m.Merged = append(m.Merged, fmt.Sprintf("%s#%d", repo, number))
	m.Approved = append(m.Approved, approve)
	return m.MergeErr
}

// RebaseBranch records the PR node ID as rebased and returns the configured error.
func (m *MockFetcher) RebaseBranch(ctx context.Context, prID string) error {
	m.Rebased = append(m.Rebased, prID)
	return m.MergeErr
}
