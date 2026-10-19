package analysis

import "testing"

func TestMapApprovalState(t *testing.T) {
	tests := []struct {
		reviewDecision string
		want           ApprovalState
	}{
		{"APPROVED", ApprovalApproved},
		{"CHANGES_REQUESTED", ApprovalChangesRequested},
		{"REVIEW_REQUIRED", ApprovalReviewRequired},
		{"", ApprovalUnknown},
		{"SOMETHING_UNRECOGNIZED", ApprovalUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.reviewDecision, func(t *testing.T) {
			if got := MapApprovalState(tt.reviewDecision); got != tt.want {
				t.Errorf("MapApprovalState(%q) = %v, want %v", tt.reviewDecision, got, tt.want)
			}
		})
	}
}
