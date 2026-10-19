package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRESTClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{httpClient: server.Client(), token: "test-token", restURL: server.URL}
}

func TestClientApproveAndMergeSendsApproveThenSquashMerge(t *testing.T) {
	var calls []string
	var mergeBody map[string]string
	var approveBody map[string]string
	client := newTestRESTClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer test-token")
		}
		switch r.Method {
		case http.MethodPost:
			_ = json.NewDecoder(r.Body).Decode(&approveBody)
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&mergeBody)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	if err := client.ApproveAndMerge(context.Background(), "giantswarm/happa", 42); err != nil {
		t.Fatalf("ApproveAndMerge() unexpected error: %v", err)
	}

	want := []string{
		"POST /repos/giantswarm/happa/pulls/42/reviews",
		"PUT /repos/giantswarm/happa/pulls/42/merge",
	}
	if strings.Join(calls, ",") != strings.Join(want, ",") {
		t.Errorf("calls = %v, want %v", calls, want)
	}
	if approveBody["event"] != "APPROVE" {
		t.Errorf("approve body = %v, want event APPROVE", approveBody)
	}
	if mergeBody["merge_method"] != "squash" {
		t.Errorf("merge body = %v, want merge_method squash", mergeBody)
	}
}

func TestClientApproveAndMergeApproveFailureSkipsMerge(t *testing.T) {
	calls := 0
	client := newTestRESTClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Can not approve your own pull request"}`))
	})

	err := client.ApproveAndMerge(context.Background(), "o/r", 1)
	if err == nil {
		t.Fatal("ApproveAndMerge() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Can not approve your own pull request") {
		t.Errorf("error = %q, want GitHub message", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no merge after failed approve)", calls)
	}
}

func TestClientApproveAndMergeMergeFailureMentionsApproval(t *testing.T) {
	client := newTestRESTClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"message":"Required status check is expected"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	err := client.ApproveAndMerge(context.Background(), "o/r", 1)
	if err == nil {
		t.Fatal("ApproveAndMerge() error = nil, want error")
	}
	for _, want := range []string{"approved", "Required status check is expected"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want to contain %q", err, want)
		}
	}
}
