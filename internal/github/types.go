package github

import "time"

// PullRequest is the raw shape of a PR as returned by the GraphQL search query.
type PullRequest struct {
	Number           int       `json:"number"`
	Title            string    `json:"title"`
	URL              string    `json:"url"`
	CreatedAt        time.Time `json:"createdAt"`
	IsDraft          bool      `json:"isDraft"`
	ReviewDecision   string    `json:"reviewDecision"`
	MergeStateStatus string    `json:"mergeStateStatus"`
	Repository       struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Comments struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *StatusCheckRollup `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

// LatestCheckRollup returns the status check rollup for the PR's most recent
// commit, or nil if there is none (e.g. no CI configured on this branch).
func (p PullRequest) LatestCheckRollup() *StatusCheckRollup {
	if len(p.Commits.Nodes) == 0 {
		return nil
	}
	return p.Commits.Nodes[0].Commit.StatusCheckRollup
}

// StatusCheckRollup aggregates a commit's check runs and legacy statuses.
type StatusCheckRollup struct {
	State    string `json:"state"`
	Contexts struct {
		Nodes []CheckContext `json:"nodes"`
	} `json:"contexts"`
}

// CheckContext is a single check run or legacy status context.
// Typename discriminates between the two GraphQL union members.
type CheckContext struct {
	Typename   string `json:"__typename"`
	Conclusion string `json:"conclusion"` // CheckRun only
	Status     string `json:"status"`     // CheckRun only
	State      string `json:"state"`      // StatusContext only
}

type searchResponse struct {
	Data struct {
		Search struct {
			Nodes    []PullRequest `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"search"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
