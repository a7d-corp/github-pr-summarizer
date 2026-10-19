package ui

import (
	"errors"
	"testing"

	"github.com/a7d-corp/github-pr-summarizer/internal/github"
)

func TestFetchCmdSuccess(t *testing.T) {
	raw := github.PullRequest{Number: 1}
	raw.Repository.NameWithOwner = "giantswarm/happa"
	fetcher := &github.MockFetcher{PRs: []github.PullRequest{raw}}
	allowlist := map[string]struct{}{"giantswarm/happa": {}}

	msg := fetchCmd(fetcher, allowlist)()

	fm, ok := msg.(fetchedMsg)
	if !ok {
		t.Fatalf("fetchCmd() returned %T, want fetchedMsg", msg)
	}
	if fm.err != nil {
		t.Fatalf("fetchedMsg.err = %v, want nil", fm.err)
	}
	if len(fm.prs) != 1 || fm.prs[0].Number != 1 {
		t.Fatalf("fetchedMsg.prs = %+v, want one PR numbered 1", fm.prs)
	}
}

func TestFetchCmdError(t *testing.T) {
	wantErr := errors.New("boom")
	fetcher := &github.MockFetcher{Err: wantErr}

	msg := fetchCmd(fetcher, nil)()

	fm, ok := msg.(fetchedMsg)
	if !ok {
		t.Fatalf("fetchCmd() returned %T, want fetchedMsg", msg)
	}
	if fm.err != wantErr {
		t.Fatalf("fetchedMsg.err = %v, want %v", fm.err, wantErr)
	}
	if fm.prs != nil {
		t.Fatalf("fetchedMsg.prs = %+v, want nil on error", fm.prs)
	}
}
