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

func TestClientSearchReviewRequestedPRsSendsAuthHeader(t *testing.T) {
	var gotAuth string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeSearchResponse(t, w, nil, false, "")
	})

	if _, err := client.SearchReviewRequestedPRs(context.Background()); err != nil {
		t.Fatalf("SearchReviewRequestedPRs() unexpected error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-token")
	}
}

func TestClientSearchReviewRequestedPRsPaginates(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeSearchResponse(t, w, []PullRequest{{Number: 1}}, true, "cursor-1")
			return
		}
		writeSearchResponse(t, w, []PullRequest{{Number: 2}}, false, "")
	})

	prs, err := client.SearchReviewRequestedPRs(context.Background())
	if err != nil {
		t.Fatalf("SearchReviewRequestedPRs() unexpected error: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 requests, got %d", calls)
	}
	if len(prs) != 2 || prs[0].Number != 1 || prs[1].Number != 2 {
		t.Fatalf("SearchReviewRequestedPRs() = %+v, want stitched PRs 1 and 2", prs)
	}
}

func TestClientSearchReviewRequestedPRsSurfacesGraphQLErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"message": "bad credentials"}},
		})
	})

	_, err := client.SearchReviewRequestedPRs(context.Background())
	if err == nil {
		t.Fatal("SearchReviewRequestedPRs() expected error, got nil")
	}
}

func TestClientSearchReviewRequestedPRsNonOKStatus(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.SearchReviewRequestedPRs(context.Background())
	if err == nil {
		t.Fatal("SearchReviewRequestedPRs() expected error, got nil")
	}
}

func TestClientSearchReviewRequestedPRsRetriesBadGateway(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		writeSearchResponse(t, w, []PullRequest{{Number: 1}}, false, "")
	})

	prs, err := client.SearchReviewRequestedPRs(context.Background())
	if err != nil {
		t.Fatalf("SearchReviewRequestedPRs() unexpected error: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 requests, got %d", calls)
	}
	if len(prs) != 1 || prs[0].Number != 1 {
		t.Fatalf("SearchReviewRequestedPRs() = %+v, want PR 1", prs)
	}
}

func TestClientSearchReviewRequestedPRsGivesUpAfterMaxAttempts(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
	})

	_, err := client.SearchReviewRequestedPRs(context.Background())
	if err == nil {
		t.Fatal("SearchReviewRequestedPRs() expected error, got nil")
	}

	if calls != maxAttempts {
		t.Fatalf("expected %d requests, got %d", maxAttempts, calls)
	}
}

func TestClientSearchReviewRequestedPRsDoesNotRetryClientErrors(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.SearchReviewRequestedPRs(context.Background())
	if err == nil {
		t.Fatal("SearchReviewRequestedPRs() expected error, got nil")
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
