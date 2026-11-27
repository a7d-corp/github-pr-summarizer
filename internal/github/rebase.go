package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const rebaseMutation = `
mutation($id: ID!) {
  updatePullRequestBranch(input: {pullRequestId: $id, updateMethod: REBASE}) {
    pullRequest { id }
  }
}`

// RebaseBranch rebases the head branch of the PR with node ID prID onto the
// PR's base branch. It does not retry, because a rebase must not run twice.
func (c *Client) RebaseBranch(ctx context.Context, prID string) error {
	body, err := json.Marshal(graphQLRequest{
		Query:     rebaseMutation,
		Variables: map[string]any{"id": prID},
	})
	if err != nil {
		return fmt.Errorf("marshal rebase request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build rebase request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("rebase request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("rebase failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode rebase response: %w", err)
	}
	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("rebase failed: %s", strings.Join(msgs, "; "))
	}
	return nil
}
