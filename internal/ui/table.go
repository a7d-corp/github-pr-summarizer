package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
// The padding uses the row style.
func fit(s string, w int, row lipgloss.Style) string {
	s = ansi.Truncate(s, w, "…")
	return s + row.Render(strings.Repeat(" ", max(0, w-ansi.StringWidth(s))))
}

// joinCells puts the cells in their columns. The gaps between the columns use the row style.
func joinCells(prefix, title string, titleW int, cells []string, row lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(fit(title, titleW, row))
	for i, c := range fixedColumns {
		b.WriteString(row.Render(strings.Repeat(" ", columnGap)))
		b.WriteString(fit(cells[i], c.width, row))
	}
	return b.String()
}

// rowStyle returns the base style of a PR line. The selected line has a background colour.
// Each part of a line must use this style, because a colour reset also clears the background.
func rowStyle(selected bool) lipgloss.Style {
	if selected {
		return styleSelectedRow
	}
	return styleNone
}

func headerView(titleW int) string {
	titles := make([]string, 0, len(fixedColumns))
	for _, c := range fixedColumns {
		titles = append(titles, c.title)
	}
	return styleHeader.Render(joinCells(strings.Repeat(" ", prefixWidth), "PR", titleW, titles, styleNone))
}

func headingView(repo string) string {
	return styleHeading.Render(repo)
}

func prLineView(pr analysis.PullRequest, selected bool, titleW int) string {
	row := rowStyle(selected)
	prefix := row.Render(strings.Repeat(" ", prefixWidth))
	if selected {
		prefix = row.Inherit(styleSelected).Render("> ")
	}
	return joinCells(prefix, formatTitle(pr, selected, row), titleW, []string{
		formatChecks(pr.Checks, row),
		formatApproval(pr.Approval, row),
		formatBranchStatus(pr.Branch, row),
		row.Render(strconv.Itoa(pr.Comments)),
		row.Render(pr.Age),
	}, row)
}

func formatTitle(pr analysis.PullRequest, selected bool, row lipgloss.Style) string {
	titleStyle := row
	if selected {
		titleStyle = row.Inherit(styleSelected)
	}
	title := titleStyle.Render(fmt.Sprintf("#%d %s", pr.Number, pr.Title))
	if pr.IsDraft {
		return row.Inherit(styleGray).Render("[draft] ") + title
	}
	return title
}

func formatChecks(cs analysis.CheckSummary, row lipgloss.Style) string {
	if cs.Overall == analysis.StatusUnknown {
		return row.Inherit(styleGray).Render("-")
	}

	pass := row.Inherit(styleGreen).Render(strconv.Itoa(cs.Pass))
	fail := row.Inherit(styleRed).Render(strconv.Itoa(cs.Fail))
	return pass + row.Render("/") + fail
}

func formatApproval(state analysis.ApprovalState, row lipgloss.Style) string {
	if state == analysis.ApprovalApproved {
		return row.Inherit(styleGreen).Render("●")
	}
	return row.Inherit(styleGray).Render("○")
}

func formatBranchStatus(bs analysis.BranchStatus, row lipgloss.Style) string {
	switch bs {
	case analysis.BranchUpToDate:
		return row.Inherit(styleGreen).Render("●")
	case analysis.BranchBehind:
		return row.Inherit(styleRed).Render("●")
	default:
		return row.Inherit(styleGray).Render("○")
	}
}
