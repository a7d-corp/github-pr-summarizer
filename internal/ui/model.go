package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

// Model is the bubbletea model driving the PR dashboard.
type Model struct {
	prs        []analysis.PullRequest
	cursor     int // index into prs
	offset     int // first visible line of the list
	width      int
	height     int
	fetcher    github.Fetcher
	merger     github.Merger
	allowlist  map[string]struct{}
	loading    bool
	confirming bool
	merging    bool
	status     string
	err        error
	showHelp   bool
	openURL    func(string) error
}

// NewModel builds a Model ready to run.
func NewModel(fetcher github.Fetcher, merger github.Merger, allowlist map[string]struct{}) Model {
	return Model{
		fetcher:   fetcher,
		merger:    merger,
		allowlist: allowlist,
		loading:   true,
		openURL:   browser.OpenURL,
	}
}

// Init kicks off the first fetch.
func (m Model) Init() tea.Cmd {
	return fetchCmd(m.fetcher, m.allowlist)
}

// Update handles bubbletea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.scroll()
		return m, nil

	case fetchedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.prs = groupByRepo(msg.prs)
			m.moveCursor(0)
		}
		return m, nil

	case mergedMsg:
		m.merging = false
		m.loading = true
		if msg.err != nil {
			m.status = fmt.Sprintf("%s#%d: %v", msg.repo, msg.number, msg.err)
		} else {
			m.status = fmt.Sprintf("merged %s#%d", msg.repo, msg.number)
		}
		return m, fetchCmd(m.fetcher, m.allowlist)

	case tea.KeyMsg:
		if m.confirming {
			return m.updateConfirm(msg)
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, keys.Refresh):
			m.loading = true
			m.err = nil
			return m, fetchCmd(m.fetcher, m.allowlist)
		case key.Matches(msg, keys.Merge):
			if _, ok := m.selectedPR(); ok {
				m.confirming = true
				m.status = ""
			}
			return m, nil
		case key.Matches(msg, keys.Enter):
			if pr, ok := m.selectedPR(); ok {
				_ = m.openURL(pr.URL)
			}
			return m, nil
		case key.Matches(msg, keys.Up):
			m.moveCursor(-1)
		case key.Matches(msg, keys.Down):
			m.moveCursor(1)
		case key.Matches(msg, keys.PageUp):
			m.moveCursor(-m.visibleLines())
		case key.Matches(msg, keys.PageDown):
			m.moveCursor(m.visibleLines())
		case key.Matches(msg, keys.Top):
			m.moveCursor(-len(m.prs))
		case key.Matches(msg, keys.Bottom):
			m.moveCursor(len(m.prs))
		}
	}
	return m, nil
}

// View renders the dashboard.
func (m Model) View() string {
	if m.showHelp {
		return styleHelp.Render(helpText)
	}
	if m.merging {
		return "merging...\n"
	}
	if m.confirming {
		if pr, ok := m.selectedPR(); ok {
			return confirmView(pr)
		}
	}
	if m.err != nil {
		return styleError.Render(fmt.Sprintf("error: %v", m.err)) + "\n\npress r to retry, q to quit"
	}
	if m.loading {
		return "loading PRs...\n"
	}
	if len(m.prs) == 0 {
		return "no open PRs are waiting on your review.\n\npress r to refresh, q to quit"
	}
	view := m.listView() + "\n" + statusLine(m.prs)
	if m.status != "" {
		view += "\n" + m.status
	}
	return view
}

// updateConfirm handles keys while the merge confirmation is shown.
func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Yes):
		pr, ok := m.selectedPR()
		m.confirming = false
		if !ok {
			return m, nil
		}
		m.merging = true
		return m, mergeCmd(m.merger, pr)
	case key.Matches(msg, keys.No):
		m.confirming = false
	}
	return m, nil
}

// groupByRepo returns a copy of prs ordered by repo. PRs in the same repo keep their order.
func groupByRepo(prs []analysis.PullRequest) []analysis.PullRequest {
	sorted := slices.Clone(prs)
	slices.SortStableFunc(sorted, func(a, b analysis.PullRequest) int {
		return strings.Compare(a.Repo, b.Repo)
	})
	return sorted
}

func (m Model) selectedPR() (analysis.PullRequest, bool) {
	if m.cursor < 0 || m.cursor >= len(m.prs) {
		return analysis.PullRequest{}, false
	}
	return m.prs[m.cursor], true
}

// moveCursor moves the cursor by n PRs, keeps it on a PR, and scrolls it into view.
func (m *Model) moveCursor(n int) {
	m.cursor = max(min(m.cursor+n, len(m.prs)-1), 0)
	m.scroll()
}

// visibleLines returns how many list lines fit below the column header.
// It returns 0 before the terminal height is known, which shows all lines.
func (m Model) visibleLines() int {
	if m.height == 0 {
		return 0
	}
	return max(m.height-5, 1)
}

// scroll changes the offset so the cursor line is visible. If the cursor PR
// is the first PR of its repo, the repo heading is kept visible too.
func (m *Model) scroll() {
	visible := m.visibleLines()
	lines := buildLines(m.prs)
	if visible == 0 || len(lines) <= visible {
		m.offset = 0
		return
	}

	cur := 0
	for i, l := range lines {
		if l.pr == m.cursor {
			cur = i
			break
		}
	}
	top := cur
	if cur > 0 && lines[cur-1].pr < 0 {
		top = cur - 1
	}

	if top < m.offset {
		m.offset = top
	}
	if cur >= m.offset+visible {
		m.offset = cur - visible + 1
	}
	m.offset = max(min(m.offset, len(lines)-visible), 0)
}

// listView renders the column header, then the visible repo headings and PRs.
func (m Model) listView() string {
	lines := buildLines(m.prs)
	visible := m.visibleLines()
	end := len(lines)
	if visible > 0 {
		end = min(m.offset+visible, len(lines))
	}

	titleW := titleWidth(m.width)
	out := []string{headerView(titleW)}
	for _, l := range lines[m.offset:end] {
		if l.pr < 0 {
			out = append(out, headingView(l.repo))
			continue
		}
		out = append(out, prLineView(m.prs[l.pr], l.pr == m.cursor, titleW))
	}
	// Pad to full height so the status line stays at the bottom.
	for len(out) < visible+1 {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}

func statusLine(prs []analysis.PullRequest) string {
	return fmt.Sprintf("%d PRs · enter open · m merge · r refresh · ? help · q quit", len(prs))
}
