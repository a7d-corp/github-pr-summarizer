package analysis

// ApprovalState summarizes a PR's review decision.
type ApprovalState int

const (
	ApprovalUnknown ApprovalState = iota
	ApprovalReviewRequired
	ApprovalApproved
	ApprovalChangesRequested
)

// MapApprovalState maps GitHub's reviewDecision field to an ApprovalState.
func MapApprovalState(reviewDecision string) ApprovalState {
	switch reviewDecision {
	case "APPROVED":
		return ApprovalApproved
	case "CHANGES_REQUESTED":
		return ApprovalChangesRequested
	case "REVIEW_REQUIRED":
		return ApprovalReviewRequired
	default:
		return ApprovalUnknown
	}
}
