package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{httpClient: server.Client(), token: "test-token", baseURL: server.URL}
}

func TestClientSearchOpenPRsQueriesConfiguredRepos(t *testing.T) {
	var gotQuery any
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		gotQuery = req.Variables["query"]
		writeSearchResponse(t, w, nil, false, "")
	})
	client.repos = []string{"giantswarm/happa", "giantswarm/cluster-api-cleaner-cloud-director"}

	if _, err := client.SearchOpenPRs(context.Background()); err != nil {
		t.Fatalf("SearchOpenPRs() unexpected error: %v", err)
	}

	want := "is:pr is:open repo:giantswarm/happa repo:giantswarm/cluster-api-cleaner-cloud-director"
	if gotQuery != want {
		t.Errorf("search query = %q, want %q", gotQuery, want)
	}
}

func TestClientSearchOpenPRsSendsAuthHeader(t *testing.T) {
	var gotAuth string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeSearchResponse(t, w, nil, false, "")
	})

	if _, err := client.SearchOpenPRs(context.Background()); err != nil {
		t.Fatalf("SearchOpenPRs() unexpected error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-token")
	}
}

func TestClientSearchOpenPRsPaginates(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeSearchResponse(t, w, []PullRequest{{Number: 1}}, true, "cursor-1")
			return
		}
		writeSearchResponse(t, w, []PullRequest{{Number: 2}}, false, "")
	})

	prs, err := client.SearchOpenPRs(context.Background())
	if err != nil {
		t.Fatalf("SearchOpenPRs() unexpected error: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 requests, got %d", calls)
	}
	if len(prs) != 2 || prs[0].Number != 1 || prs[1].Number != 2 {
		t.Fatalf("SearchOpenPRs() = %+v, want stitched PRs 1 and 2", prs)
	}
}

func TestClientSearchOpenPRsSurfacesGraphQLErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"message": "bad credentials"}},
		})
	})

	_, err := client.SearchOpenPRs(context.Background())
	if err == nil {
		t.Fatal("SearchOpenPRs() expected error, got nil")
	}
}

func TestClientSearchOpenPRsNonOKStatus(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.SearchOpenPRs(context.Background())
	if err == nil {
		t.Fatal("SearchOpenPRs() expected error, got nil")
	}
}

func TestClientSearchOpenPRsRetriesBadGateway(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		writeSearchResponse(t, w, []PullRequest{{Number: 1}}, false, "")
	})

	prs, err := client.SearchOpenPRs(context.Background())
	if err != nil {
		t.Fatalf("SearchOpenPRs() unexpected error: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 requests, got %d", calls)
	}
	if len(prs) != 1 || prs[0].Number != 1 {
		t.Fatalf("SearchOpenPRs() = %+v, want PR 1", prs)
	}
}

func TestClientSearchOpenPRsGivesUpAfterMaxAttempts(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
	})

	_, err := client.SearchOpenPRs(context.Background())
	if err == nil {
		t.Fatal("SearchOpenPRs() expected error, got nil")
	}

	if calls != maxAttempts {
		t.Fatalf("expected %d requests, got %d", maxAttempts, calls)
	}
}

func TestClientSearchOpenPRsDoesNotRetryClientErrors(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.SearchOpenPRs(context.Background())
	if err == nil {
		t.Fatal("SearchOpenPRs() expected error, got nil")
	}

	if calls != 1 {
		t.Fatalf("expected 1 request, got %d", calls)
	}
}

func writeSearchResponse(t *testing.T, w http.ResponseWriter, nodes []PullRequest, hasNext bool, endCursor string) {
	t.Helper()
	var resp searchResponse
	resp.Data.Search.Nodes = nodes
	resp.Data.Search.PageInfo.HasNextPage = hasNext
	resp.Data.Search.PageInfo.EndCursor = endCursor

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("encode search response: %v", err)
	}
}
