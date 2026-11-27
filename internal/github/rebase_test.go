package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestClientRebaseBranchSendsRebaseMutation(t *testing.T) {
	var got graphQLRequest
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-token")
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"data":{"updatePullRequestBranch":{"pullRequest":{"id":"PR_1"}}}}`))
	})

	if err := client.RebaseBranch(context.Background(), "PR_1"); err != nil {
		t.Fatalf("RebaseBranch() unexpected error: %v", err)
	}
	for _, want := range []string{"updatePullRequestBranch", "updateMethod: REBASE"} {
		if !strings.Contains(got.Query, want) {
			t.Errorf("query = %q, want to contain %q", got.Query, want)
		}
	}
	if got.Variables["id"] != "PR_1" {
		t.Errorf("variables = %v, want id PR_1", got.Variables)
	}
}

func TestClientRebaseBranchSurfacesGraphQLErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"merge conflict between base and head"}]}`))
	})

	err := client.RebaseBranch(context.Background(), "PR_1")
	if err == nil || !strings.Contains(err.Error(), "merge conflict between base and head") {
		t.Errorf("RebaseBranch() error = %v, want the GraphQL error message", err)
	}
}

func TestClientRebaseBranchDoesNotRetry(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	})

	err := client.RebaseBranch(context.Background(), "PR_1")
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Errorf("RebaseBranch() error = %v, want status 502", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no retry)", calls)
	}
}
