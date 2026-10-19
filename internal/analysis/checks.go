package analysis

import "github.com/a7d-corp/github-pr-summarizer/internal/github"

// Status is the overall CI state derived from a PR's check runs.
type Status int

const (
	StatusUnknown Status = iota
	StatusPassing
	StatusPending
	StatusFailing
)

// CheckSummary tallies a PR's check runs and legacy statuses.
type CheckSummary struct {
	Pass, Fail, Pending int
	Overall             Status
}

// TallyChecks counts pass/fail/pending contexts from a status check rollup
// and derives an overall Status. A nil rollup (no CI configured) yields
// StatusUnknown with zero counts.
func TallyChecks(rollup *github.StatusCheckRollup) CheckSummary {
	if rollup == nil {
		return CheckSummary{Overall: StatusUnknown}
	}

	var summary CheckSummary
	for _, ctx := range rollup.Contexts.Nodes {
		switch classify(ctx) {
		case StatusPassing:
			summary.Pass++
		case StatusFailing:
			summary.Fail++
		case StatusPending:
			summary.Pending++
		}
	}

	switch {
	case summary.Fail > 0:
		summary.Overall = StatusFailing
	case summary.Pending > 0:
		summary.Overall = StatusPending
	case summary.Pass > 0:
		summary.Overall = StatusPassing
	default:
		summary.Overall = StatusUnknown
	}

	return summary
}

// classify maps a single CheckRun/StatusContext to a Status.
func classify(ctx github.CheckContext) Status {
	switch ctx.Typename {
	case "CheckRun":
		return classifyCheckRun(ctx)
	case "StatusContext":
		return classifyStatusContext(ctx)
	default:
		return StatusUnknown
	}
}

func classifyCheckRun(ctx github.CheckContext) Status {
	if ctx.Status != "COMPLETED" {
		return StatusPending
	}
	switch ctx.Conclusion {
	case "SUCCESS", "NEUTRAL", "SKIPPED":
		return StatusPassing
	case "FAILURE", "TIMED_OUT", "CANCELLED", "ACTION_REQUIRED", "STARTUP_FAILURE":
		return StatusFailing
	default:
		return StatusPending
	}
}

func classifyStatusContext(ctx github.CheckContext) Status {
	switch ctx.State {
	case "SUCCESS":
		return StatusPassing
	case "FAILURE", "ERROR":
		return StatusFailing
	default: // PENDING, EXPECTED
		return StatusPending
	}
}
