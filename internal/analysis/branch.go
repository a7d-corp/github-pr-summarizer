package analysis

// BranchStatus summarizes whether a PR's branch is up to date with its base.
type BranchStatus int

const (
	BranchUnknown BranchStatus = iota
	BranchUpToDate
	BranchBehind
)

// MapBranchStatus maps GitHub's mergeStateStatus field to a BranchStatus.
func MapBranchStatus(mergeStateStatus string) BranchStatus {
	switch mergeStateStatus {
	case "BEHIND":
		return BranchBehind
	case "CLEAN", "BLOCKED", "DIRTY", "DRAFT", "HAS_HOOKS", "UNSTABLE":
		return BranchUpToDate
	default: // "UNKNOWN", "", or any future GraphQL enum value
		return BranchUnknown
	}
}
