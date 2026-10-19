package analysis

import (
	"time"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// PullRequest is the domain model consumed by the UI layer.
type PullRequest struct {
	Number   int
	Title    string
	URL      string
	Repo     string
	Age      string
	IsDraft  bool
	Checks   CheckSummary
	Approval ApprovalState
	Branch   BranchStatus
	Comments int
}

// FilterByRepoAllowlist keeps only PRs whose repository is in allowlist.
func FilterByRepoAllowlist(raws []github.PullRequest, allowlist map[string]struct{}) []github.PullRequest {
	filtered := make([]github.PullRequest, 0, len(raws))
	for _, raw := range raws {
		if _, ok := allowlist[raw.Repository.NameWithOwner]; ok {
			filtered = append(filtered, raw)
		}
	}
	return filtered
}

// BuildPullRequests filters raws by allowlist and maps the survivors into
// the domain PullRequest model.
func BuildPullRequests(raws []github.PullRequest, allowlist map[string]struct{}, now time.Time) []PullRequest {
	filtered := FilterByRepoAllowlist(raws, allowlist)

	prs := make([]PullRequest, 0, len(filtered))
	for _, raw := range filtered {
		prs = append(prs, PullRequest{
			Number:   raw.Number,
			Title:    raw.Title,
			URL:      raw.URL,
			Repo:     raw.Repository.NameWithOwner,
			Age:      FormatAge(raw.CreatedAt, now),
			IsDraft:  raw.IsDraft,
			Checks:   TallyChecks(raw.LatestCheckRollup()),
			Approval: MapApprovalState(raw.ReviewDecision),
			Branch:   MapBranchStatus(raw.MergeStateStatus),
			Comments: raw.Comments.TotalCount,
		})
	}
	return prs
}
