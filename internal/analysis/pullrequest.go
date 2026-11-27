package analysis

import (
	"time"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// PullRequest is the domain model consumed by the UI layer.
type PullRequest struct {
	ID        string // GraphQL node ID
	Number    int
	Title     string
	URL       string
	Repo      string
	Age       string
	IsDraft   bool
	Checks    CheckSummary
	CheckRuns []Check
	Approval  ApprovalState
	Branch    BranchStatus
	Comments  int
	IsOwn     bool // true if the token owner is the PR author
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
			ID:        raw.ID,
			Number:    raw.Number,
			Title:     raw.Title,
			URL:       raw.URL,
			Repo:      raw.Repository.NameWithOwner,
			Age:       FormatAge(raw.CreatedAt, now),
			IsDraft:   raw.IsDraft,
			Checks:    TallyChecks(raw.LatestCheckRollup()),
			CheckRuns: ListChecks(raw.LatestCheckRollup()),
			Approval:  MapApprovalState(raw.ReviewDecision),
			Branch:    MapBranchStatus(raw.MergeStateStatus),
			Comments:  raw.Comments.TotalCount,
			IsOwn:     raw.ViewerDidAuthor,
		})
	}
	return prs
}
