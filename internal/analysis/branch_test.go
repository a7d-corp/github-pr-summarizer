package analysis

import "testing"

func TestMapBranchStatus(t *testing.T) {
	tests := []struct {
		mergeStateStatus string
		want             BranchStatus
	}{
		{"BEHIND", BranchBehind},
		{"BLOCKED", BranchUpToDate},
		{"CLEAN", BranchUpToDate},
		{"DIRTY", BranchUpToDate},
		{"DRAFT", BranchUpToDate},
		{"HAS_HOOKS", BranchUpToDate},
		{"UNSTABLE", BranchUpToDate},
		{"UNKNOWN", BranchUnknown},
		{"", BranchUnknown},
		{"SOMETHING_FUTURE", BranchUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.mergeStateStatus, func(t *testing.T) {
			if got := MapBranchStatus(tt.mergeStateStatus); got != tt.want {
				t.Errorf("MapBranchStatus(%q) = %v, want %v", tt.mergeStateStatus, got, tt.want)
			}
		})
	}
}
