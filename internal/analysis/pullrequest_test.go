package analysis

import (
	"testing"
	"time"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

func rawPR(repo string, number int) github.PullRequest {
	pr := github.PullRequest{Number: number, Title: "title", URL: "https://example.com"}
	pr.Repository.NameWithOwner = repo
	return pr
}

func TestFilterByRepoAllowlist(t *testing.T) {
	raws := []github.PullRequest{
		rawPR("giantswarm/happa", 1),
		rawPR("giantswarm/other", 2),
		rawPR("GiantSwarm/Happa", 3), // case-sensitive mismatch, should NOT match
	}
	allowlist := map[string]struct{}{"giantswarm/happa": {}}

	got := FilterByRepoAllowlist(raws, allowlist)
	if len(got) != 1 || got[0].Number != 1 {
		t.Fatalf("FilterByRepoAllowlist() = %+v, want only PR #1", got)
	}
}

func TestFilterByRepoAllowlistNoMatch(t *testing.T) {
	raws := []github.PullRequest{rawPR("giantswarm/happa", 1)}
	allowlist := map[string]struct{}{"giantswarm/other": {}}

	got := FilterByRepoAllowlist(raws, allowlist)
	if len(got) != 0 {
		t.Fatalf("FilterByRepoAllowlist() = %+v, want empty", got)
	}
}

func TestBuildPullRequests(t *testing.T) {
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	raw := github.PullRequest{
		Number:           42,
		Title:            "Fix the thing",
		URL:              "https://github.com/giantswarm/happa/pull/42",
		CreatedAt:        now.Add(-2 * time.Hour),
		IsDraft:          true,
		ReviewDecision:   "CHANGES_REQUESTED",
		MergeStateStatus: "BEHIND",
	}
	raw.Repository.NameWithOwner = "giantswarm/happa"
	raw.Comments.TotalCount = 7

	allowlist := map[string]struct{}{"giantswarm/happa": {}}

	got := BuildPullRequests([]github.PullRequest{raw}, allowlist, now)
	if len(got) != 1 {
		t.Fatalf("BuildPullRequests() len = %d, want 1", len(got))
	}

	pr := got[0]
	want := PullRequest{
		Number:   42,
		Title:    "Fix the thing",
		URL:      "https://github.com/giantswarm/happa/pull/42",
		Repo:     "giantswarm/happa",
		Age:      "2h",
		IsDraft:  true,
		Checks:   CheckSummary{Overall: StatusUnknown},
		Approval: ApprovalChangesRequested,
		Branch:   BranchBehind,
		Comments: 7,
	}
	if pr != want {
		t.Errorf("BuildPullRequests()[0] = %+v, want %+v", pr, want)
	}
}

func TestBuildPullRequestsFiltersOutNonAllowlisted(t *testing.T) {
	now := time.Now()
	raws := []github.PullRequest{rawPR("giantswarm/not-allowed", 1)}
	allowlist := map[string]struct{}{"giantswarm/happa": {}}

	got := BuildPullRequests(raws, allowlist, now)
	if len(got) != 0 {
		t.Fatalf("BuildPullRequests() = %+v, want empty", got)
	}
}
