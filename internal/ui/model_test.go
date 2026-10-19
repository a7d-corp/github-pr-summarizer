package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/a7d-corp/github-pr-summarizer/internal/analysis"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	m := NewModel(&github.MockFetcher{}, &github.MockFetcher{}, nil)
	m.loading = false
	return m
}

func TestModelUpdateQuit(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Update(q) returned nil cmd, want tea.Quit")
	}
	if cmd() != tea.Quit() {
		t.Error("Update(q) did not return tea.Quit")
	}
}

func TestModelUpdateToggleHelp(t *testing.T) {
	m := newTestModel(t)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	updated := newModel.(Model)
	if !updated.showHelp {
		t.Fatal("Update(?) did not set showHelp")
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	updated = newModel.(Model)
	if updated.showHelp {
		t.Fatal("Update(?) toggled again did not clear showHelp")
	}
}

func TestModelUpdateRefreshSetsLoading(t *testing.T) {
	m := newTestModel(t)
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	updated := newModel.(Model)
	if !updated.loading {
		t.Fatal("Update(r) did not set loading")
	}
	if cmd == nil {
		t.Fatal("Update(r) returned nil cmd, want a fetch command")
	}
}

func TestModelUpdateFetchedMsgSuccess(t *testing.T) {
	m := newTestModel(t)
	pr := analysis.PullRequest{Number: 1, Title: "test", URL: "https://example.com/1"}

	newModel, _ := m.Update(fetchedMsg{prs: []analysis.PullRequest{pr}})
	updated := newModel.(Model)

	if updated.loading {
		t.Error("Update(fetchedMsg) left loading true")
	}
	if len(updated.prs) != 1 || updated.prs[0].Number != 1 {
		t.Errorf("Update(fetchedMsg) prs = %+v, want one PR numbered 1", updated.prs)
	}
}

func TestModelUpdateFetchedMsgError(t *testing.T) {
	m := newTestModel(t)
	wantErr := errors.New("boom")

	newModel, _ := m.Update(fetchedMsg{err: wantErr})
	updated := newModel.(Model)

	if updated.loading {
		t.Error("Update(fetchedMsg) left loading true")
	}
	if updated.err != wantErr {
		t.Errorf("Update(fetchedMsg) err = %v, want %v", updated.err, wantErr)
	}
}

func TestModelUpdateEnterOpensSelectedPR(t *testing.T) {
	m := newTestModel(t)
	pr := analysis.PullRequest{Number: 1, Title: "test", URL: "https://example.com/1"}
	newModel, _ := m.Update(fetchedMsg{prs: []analysis.PullRequest{pr}})
	updated := newModel.(Model)

	var openedURL string
	updated.openURL = func(url string) error {
		openedURL = url
		return nil
	}

	updated.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if openedURL != pr.URL {
		t.Errorf("Update(enter) opened %q, want %q", openedURL, pr.URL)
	}
}

func TestModelViewShowsHelp(t *testing.T) {
	m := newTestModel(t)
	m.showHelp = true
	if got := m.View(); got == "" {
		t.Error("View() with showHelp=true returned empty string")
	}
}

func TestModelViewShowsLoading(t *testing.T) {
	m := newTestModel(t)
	m.loading = true
	if got := m.View(); got != "loading PRs...\n" {
		t.Errorf("View() while loading = %q", got)
	}
}

func TestModelViewShowsError(t *testing.T) {
	m := newTestModel(t)
	m.err = errors.New("boom")
	if got := m.View(); got == "" {
		t.Error("View() with err set returned empty string")
	}
}

func TestModelViewEmpty(t *testing.T) {
	m := newTestModel(t)
	got := m.View()
	if got == "" {
		t.Error("View() with no PRs returned empty string")
	}
}

func TestModelViewShowsTable(t *testing.T) {
	m := newTestModel(t)
	pr := analysis.PullRequest{Number: 1, Title: "test", URL: "https://example.com/1", Repo: "giantswarm/happa"}
	newModel, _ := m.Update(fetchedMsg{prs: []analysis.PullRequest{pr}})
	updated := newModel.(Model)

	got := updated.View()
	if !strings.Contains(got, "1 PRs") {
		t.Errorf("View() with PRs = %q, want to contain status line", got)
	}
}

func TestModelInitReturnsFetchCmd(t *testing.T) {
	m := newTestModel(t)
	if m.Init() == nil {
		t.Fatal("Init() returned nil cmd")
	}
}

func TestStatusLine(t *testing.T) {
	prs := []analysis.PullRequest{{Number: 1}, {Number: 2}}
	got := statusLine(prs)
	if !strings.Contains(got, "2 PRs") {
		t.Errorf("statusLine() = %q, want to contain %q", got, "2 PRs")
	}
}

func TestModelUpdateFetchedGroupsByRepo(t *testing.T) {
	m := newTestModel(t)
	prs := []analysis.PullRequest{
		{Repo: "b/repo", Number: 1},
		{Repo: "a/repo", Number: 2},
		{Repo: "b/repo", Number: 3},
		{Repo: "a/repo", Number: 4},
	}
	newModel, _ := m.Update(fetchedMsg{prs: prs})
	updated := newModel.(Model)

	want := []int{2, 4, 1, 3}
	for i, n := range want {
		if updated.prs[i].Number != n {
			t.Fatalf("prs[%d].Number = %d, want %d (prs = %+v)", i, updated.prs[i].Number, n, updated.prs)
		}
	}
	if prs[0].Number != 1 {
		t.Error("Update(fetchedMsg) mutated the input slice")
	}
}

func keyRune(r string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(r)}
}

func newTestModelWithPR(t *testing.T) Model {
	t.Helper()
	m := newTestModel(t)
	pr := analysis.PullRequest{Number: 7, Title: "bump foo", Repo: "giantswarm/happa"}
	newModel, _ := m.Update(fetchedMsg{prs: []analysis.PullRequest{pr}})
	return newModel.(Model)
}

func TestModelUpdateMergeKeyShowsConfirm(t *testing.T) {
	m := newTestModelWithPR(t)

	newModel, cmd := m.Update(keyRune("m"))
	updated := newModel.(Model)

	if !updated.confirming {
		t.Fatal("Update(m) did not set confirming")
	}
	if cmd != nil {
		t.Error("Update(m) returned a cmd, want none until the user confirms")
	}
	got := updated.View()
	for _, want := range []string{"giantswarm/happa", "#7", "bump foo", "y/n"} {
		if !strings.Contains(got, want) {
			t.Errorf("confirm View() = %q, want to contain %q", got, want)
		}
	}
}

func TestModelUpdateMergeKeyWithoutPRDoesNothing(t *testing.T) {
	m := newTestModel(t)
	newModel, _ := m.Update(keyRune("m"))
	if newModel.(Model).confirming {
		t.Fatal("Update(m) with no PRs set confirming")
	}
}

func TestModelUpdateConfirmNoReturnsToList(t *testing.T) {
	for _, k := range []tea.KeyMsg{keyRune("n"), {Type: tea.KeyEsc}} {
		m := newTestModelWithPR(t)
		newModel, _ := m.Update(keyRune("m"))
		newModel, cmd := newModel.(Model).Update(k)
		updated := newModel.(Model)

		if updated.confirming || updated.merging {
			t.Errorf("Update(%v) left confirming=%v merging=%v", k, updated.confirming, updated.merging)
		}
		if cmd != nil {
			t.Errorf("Update(%v) returned a cmd, want none", k)
		}
	}
}

func TestModelUpdateConfirmIgnoresOtherKeys(t *testing.T) {
	m := newTestModelWithPR(t)
	newModel, _ := m.Update(keyRune("m"))
	newModel, cmd := newModel.(Model).Update(keyRune("x"))

	if !newModel.(Model).confirming {
		t.Error("Update(x) left the confirm screen")
	}
	if cmd != nil {
		t.Error("Update(x) returned a cmd")
	}
}

func TestModelUpdateConfirmYesStartsMerge(t *testing.T) {
	merger := &github.MockFetcher{}
	m := newTestModelWithPR(t)
	m.merger = merger
	newModel, _ := m.Update(keyRune("m"))
	newModel, cmd := newModel.(Model).Update(keyRune("y"))
	updated := newModel.(Model)

	if updated.confirming || !updated.merging {
		t.Fatalf("Update(y) confirming=%v merging=%v, want false/true", updated.confirming, updated.merging)
	}
	if !strings.Contains(updated.View(), "merging") {
		t.Errorf("View() while merging = %q, want to contain %q", updated.View(), "merging")
	}
	if cmd == nil {
		t.Fatal("Update(y) returned nil cmd, want a merge command")
	}

	msg, ok := cmd().(mergedMsg)
	if !ok {
		t.Fatal("merge command did not return a mergedMsg")
	}
	if msg.err != nil || msg.repo != "giantswarm/happa" || msg.number != 7 {
		t.Errorf("mergedMsg = %+v, want success for giantswarm/happa#7", msg)
	}
	if len(merger.Merged) != 1 || merger.Merged[0] != "giantswarm/happa#7" {
		t.Errorf("merger.Merged = %v, want [giantswarm/happa#7]", merger.Merged)
	}
}

func TestModelUpdateMergedSuccessReturnsToListAndRefreshes(t *testing.T) {
	m := newTestModelWithPR(t)
	m.merging = true

	newModel, cmd := m.Update(mergedMsg{repo: "giantswarm/happa", number: 7})
	updated := newModel.(Model)

	if updated.merging {
		t.Error("Update(mergedMsg) left merging true")
	}
	if !updated.loading || cmd == nil {
		t.Error("Update(mergedMsg) did not start a refresh")
	}
	if !strings.Contains(updated.status, "merged giantswarm/happa#7") {
		t.Errorf("status = %q, want merged message", updated.status)
	}
}

func TestModelUpdateMergedFailureShowsErrorOnList(t *testing.T) {
	m := newTestModelWithPR(t)
	m.merging = true

	newModel, cmd := m.Update(mergedMsg{repo: "giantswarm/happa", number: 7, err: errors.New("blocked by branch protection")})
	updated := newModel.(Model)

	if updated.merging || updated.err != nil {
		t.Errorf("merging=%v err=%v, want false/nil (error goes to the status line)", updated.merging, updated.err)
	}
	if cmd == nil {
		t.Error("Update(mergedMsg) with error did not start a refresh")
	}

	fetched, _ := updated.Update(fetchedMsg{prs: updated.prs})
	got := fetched.(Model).View()
	if !strings.Contains(got, "blocked by branch protection") {
		t.Errorf("View() = %q, want to contain the merge error", got)
	}
	if !strings.Contains(got, "1 PRs") {
		t.Errorf("View() = %q, want the PR list", got)
	}
}

func newTestModelWithRepos(t *testing.T) Model {
	t.Helper()
	m := newTestModel(t)
	prs := []analysis.PullRequest{
		{Repo: "b/repo", Number: 3, Title: "pr-three", URL: "https://example.com/3"},
		{Repo: "a/repo", Number: 1, Title: "pr-one", URL: "https://example.com/1"},
		{Repo: "a/repo", Number: 2, Title: "pr-two", URL: "https://example.com/2"},
		{Repo: "b/repo", Number: 4, Title: "pr-four", URL: "https://example.com/4"},
	}
	newModel, _ := m.Update(fetchedMsg{prs: prs})
	return newModel.(Model)
}

func TestModelViewShowsEachRepoOnce(t *testing.T) {
	m := newTestModelWithRepos(t)
	got := m.View()
	for _, repo := range []string{"a/repo", "b/repo"} {
		if n := strings.Count(got, repo); n != 1 {
			t.Errorf("View() contains %q %d times, want 1 (view %q)", repo, n, got)
		}
	}
}

func TestModelUpdateDownSkipsRepoHeading(t *testing.T) {
	m := newTestModelWithRepos(t)
	var newModel tea.Model = m
	for range 2 {
		newModel, _ = newModel.(Model).Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	updated := newModel.(Model)

	var openedURL string
	updated.openURL = func(url string) error {
		openedURL = url
		return nil
	}
	updated.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if openedURL != "https://example.com/3" {
		t.Errorf("Update(enter) after two downs opened %q, want the first PR of b/repo", openedURL)
	}
}

func TestModelUpdateUpStopsAtFirstPR(t *testing.T) {
	m := newTestModelWithRepos(t)
	newModel, _ := m.Update(keyRune("k"))
	if pr, ok := newModel.(Model).selectedPR(); !ok || pr.Number != 1 {
		t.Errorf("selectedPR() after k = %+v, %v, want PR 1", pr, ok)
	}
}

func TestModelUpdateFetchedClampsCursor(t *testing.T) {
	m := newTestModelWithRepos(t)
	newModel, _ := m.Update(keyRune("G"))
	newModel, _ = newModel.(Model).Update(fetchedMsg{prs: []analysis.PullRequest{{Repo: "a/repo", Number: 1}}})

	pr, ok := newModel.(Model).selectedPR()
	if !ok || pr.Number != 1 {
		t.Errorf("selectedPR() after refresh = %+v, %v, want PR 1", pr, ok)
	}
}

func TestModelScrollKeepsCursorVisible(t *testing.T) {
	m := newTestModelWithRepos(t)
	// Height 8 leaves room for three list lines.
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 8})

	newModel, _ = newModel.(Model).Update(keyRune("G"))
	got := newModel.(Model).View()
	if !strings.Contains(got, "pr-four") || strings.Contains(got, "a/repo") {
		t.Errorf("View() after G = %q, want pr-four and not a/repo", got)
	}

	newModel, _ = newModel.(Model).Update(keyRune("g"))
	got = newModel.(Model).View()
	if !strings.Contains(got, "a/repo") || strings.Contains(got, "pr-four") {
		t.Errorf("View() after g = %q, want a/repo and not pr-four", got)
	}
}

func TestModelUpdatePageDownMovesCursor(t *testing.T) {
	m := newTestModelWithRepos(t)
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 8})
	newModel, _ = newModel.(Model).Update(tea.KeyMsg{Type: tea.KeyPgDown})

	if pr, ok := newModel.(Model).selectedPR(); !ok || pr.Number != 4 {
		t.Errorf("selectedPR() after pgdown = %+v, %v, want PR 4", pr, ok)
	}
}
