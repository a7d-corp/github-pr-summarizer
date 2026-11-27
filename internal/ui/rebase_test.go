package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

func TestModelUpdateRebaseKeyShowsConfirm(t *testing.T) {
	m := newTestModelWithPR(t)

	newModel, cmd := m.Update(keyRune("u"))
	updated := newModel.(Model)

	if !updated.confirming || !updated.rebase {
		t.Fatalf("Update(u) confirming=%v rebase=%v, want true/true", updated.confirming, updated.rebase)
	}
	if cmd != nil {
		t.Error("Update(u) returned a cmd, want none until the user confirms")
	}
	got := updated.View()
	for _, want := range []string{"giantswarm/happa", "#7", "bump foo", "rebase", "y/n"} {
		if !strings.Contains(got, want) {
			t.Errorf("confirm View() = %q, want to contain %q", got, want)
		}
	}
	if strings.Contains(got, "merge") {
		t.Errorf("rebase confirm View() = %q, want no merge prompt", got)
	}
}

func TestModelUpdateMergeKeyAfterRebaseCancelShowsMergeConfirm(t *testing.T) {
	m := newTestModelWithPR(t)
	newModel, _ := m.Update(keyRune("u"))
	newModel, _ = newModel.(Model).Update(keyRune("n"))
	newModel, _ = newModel.(Model).Update(keyRune("m"))
	updated := newModel.(Model)

	if updated.rebase {
		t.Error("Update(m) after a cancelled rebase left rebase true")
	}
	if !strings.Contains(updated.View(), "squash-merge") {
		t.Errorf("View() = %q, want the merge confirm", updated.View())
	}
}

func TestModelUpdateRebaseConfirmYesStartsRebase(t *testing.T) {
	merger := &github.MockFetcher{}
	m := newTestModelWithPR(t)
	m.prs[0].ID = "PR_7"
	m.merger = merger
	newModel, _ := m.Update(keyRune("u"))
	newModel, cmd := newModel.(Model).Update(keyRune("y"))
	updated := newModel.(Model)

	if updated.confirming || !updated.rebasing {
		t.Fatalf("Update(y) confirming=%v rebasing=%v, want false/true", updated.confirming, updated.rebasing)
	}
	if !strings.Contains(updated.View(), "rebasing") {
		t.Errorf("View() while rebasing = %q, want to contain %q", updated.View(), "rebasing")
	}
	if cmd == nil {
		t.Fatal("Update(y) returned nil cmd, want a rebase command")
	}

	msg, ok := cmd().(rebasedMsg)
	if !ok {
		t.Fatal("rebase command did not return a rebasedMsg")
	}
	if msg.err != nil || msg.repo != "giantswarm/happa" || msg.number != 7 {
		t.Errorf("rebasedMsg = %+v, want success for giantswarm/happa#7", msg)
	}
	if len(merger.Rebased) != 1 || merger.Rebased[0] != "PR_7" {
		t.Errorf("merger.Rebased = %v, want [PR_7]", merger.Rebased)
	}
	if len(merger.Merged) != 0 {
		t.Errorf("merger.Merged = %v, want none", merger.Merged)
	}
}

func TestModelUpdateRebasedSuccessReturnsToListAndRefreshes(t *testing.T) {
	m := newTestModelWithPR(t)
	m.rebasing = true

	newModel, cmd := m.Update(rebasedMsg{repo: "giantswarm/happa", number: 7})
	updated := newModel.(Model)

	if updated.rebasing {
		t.Error("Update(rebasedMsg) left rebasing true")
	}
	if !updated.loading || cmd == nil {
		t.Error("Update(rebasedMsg) did not start a refresh")
	}
	if !strings.Contains(updated.status, "rebased giantswarm/happa#7") {
		t.Errorf("status = %q, want rebased message", updated.status)
	}
}

func TestModelUpdateRebasedFailureShowsErrorOnList(t *testing.T) {
	m := newTestModelWithPR(t)
	m.rebasing = true

	newModel, cmd := m.Update(rebasedMsg{repo: "giantswarm/happa", number: 7, err: errors.New("merge conflict")})
	updated := newModel.(Model)

	if updated.rebasing || updated.err != nil {
		t.Errorf("rebasing=%v err=%v, want false/nil (error goes to the status line)", updated.rebasing, updated.err)
	}
	if cmd == nil {
		t.Error("Update(rebasedMsg) with error did not start a refresh")
	}

	fetched, _ := updated.Update(fetchedMsg{prs: updated.prs})
	if got := fetched.(Model).View(); !strings.Contains(got, "merge conflict") {
		t.Errorf("View() = %q, want to contain the rebase error", got)
	}
}
