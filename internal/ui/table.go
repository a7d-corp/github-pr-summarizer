package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
)

const (
	prefixWidth       = 2
	columnGap         = 1
	defaultTitleWidth = 50
	minTitleWidth     = 20
)

type column struct {
	title string
	width int
}

// fixedColumns are the columns to the right of the PR title.
var fixedColumns = []column{
	{"Checks", 10},
	{"Approved", 10},
	{"Up to date", 10},
	{"Comments", 9},
	{"Age", 6},
}

// line is one line of the PR list: a repo heading or a PR.
type line struct {
	repo string // set for a repo heading
	pr   int    // index into prs; -1 for a repo heading
}

// buildLines adds a repo heading before each group of PRs. prs must be grouped by repo.
func buildLines(prs []analysis.PullRequest) []line {
	lines := make([]line, 0, len(prs))
	for i, pr := range prs {
		if i == 0 || pr.Repo != prs[i-1].Repo {
			lines = append(lines, line{repo: pr.Repo, pr: -1})
		}
		lines = append(lines, line{pr: i})
	}
	return lines
}

// titleWidth returns the width of the PR column, which uses all space the fixed columns do not use.
func titleWidth(termWidth int) int {
	if termWidth <= 0 {
		return defaultTitleWidth
	}
	w := termWidth - prefixWidth
	for _, c := range fixedColumns {
		w -= columnGap + c.width
	}
	return max(w, minTitleWidth)
}

// fit cuts or pads s to exactly w visible cells. ANSI colour codes do not count.
func fit(s string, w int) string {
	s = ansi.Truncate(s, w, "…")
	return s + strings.Repeat(" ", max(0, w-ansi.StringWidth(s)))
}

func joinCells(prefix, title string, titleW int, cells []string) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(fit(title, titleW))
	for i, c := range fixedColumns {
		b.WriteString(strings.Repeat(" ", columnGap))
		b.WriteString(fit(cells[i], c.width))
	}
	return b.String()
}

func headerView(titleW int) string {
	titles := make([]string, 0, len(fixedColumns))
	for _, c := range fixedColumns {
		titles = append(titles, c.title)
	}
	return styleHeader.Render(joinCells(strings.Repeat(" ", prefixWidth), "PR", titleW, titles))
}

func headingView(repo string) string {
	return styleHeading.Render(repo)
}

func prLineView(pr analysis.PullRequest, selected bool, titleW int) string {
	prefix := strings.Repeat(" ", prefixWidth)
	if selected {
		prefix = styleSelected.Render("> ")
	}
	return joinCells(prefix, formatTitle(pr, selected), titleW, []string{
		formatChecks(pr.Checks),
		formatApproval(pr.Approval),
		formatBranchStatus(pr.Branch),
		strconv.Itoa(pr.Comments),
		pr.Age,
	})
}

func formatTitle(pr analysis.PullRequest, selected bool) string {
	title := fmt.Sprintf("#%d %s", pr.Number, pr.Title)
	if selected {
		title = styleSelected.Render(title)
	}
	if pr.IsDraft {
		return styleGray.Render("[draft] ") + title
	}
	return title
}

func formatChecks(cs analysis.CheckSummary) string {
	if cs.Overall == analysis.StatusUnknown {
		return styleGray.Render("-")
	}

	pass := styleGreen.Render(strconv.Itoa(cs.Pass))
	fail := styleRed.Render(strconv.Itoa(cs.Fail))
	return pass + "/" + fail
}

func formatApproval(state analysis.ApprovalState) string {
	if state == analysis.ApprovalApproved {
		return styleGreen.Render("●")
	}
	return styleGray.Render("○")
}

func formatBranchStatus(bs analysis.BranchStatus) string {
	switch bs {
	case analysis.BranchUpToDate:
		return styleGreen.Render("●")
	case analysis.BranchBehind:
		return styleRed.Render("●")
	default:
		return styleGray.Render("○")
	}
}
