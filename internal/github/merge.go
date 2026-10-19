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

// ApproveAndMerge approves the PR, then squash-merges it. GitHub branch
// protection decides whether the merge is allowed.
func (c *Client) ApproveAndMerge(ctx context.Context, repo string, number int) error {
	prPath := fmt.Sprintf("/repos/%s/pulls/%d", repo, number)

	if err := c.doREST(ctx, http.MethodPost, prPath+"/reviews", map[string]string{"event": "APPROVE"}); err != nil {
		return fmt.Errorf("approve failed: %w", err)
	}
	if err := c.doREST(ctx, http.MethodPut, prPath+"/merge", map[string]string{"merge_method": "squash"}); err != nil {
		return fmt.Errorf("approved, but merge failed: %w", err)
	}
	return nil
}

// doREST sends a JSON request to the REST API. It does not retry, because
// approve and merge must not run twice.
func (c *Client) doREST(ctx context.Context, method, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.restURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, restErrorMessage(resp.Body))
	}
	return nil
}

// restErrorMessage returns the "message" field of a GitHub error body, or the
// raw body if it is not JSON.
func restErrorMessage(r io.Reader) string {
	raw, _ := io.ReadAll(io.LimitReader(r, 512))
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &parsed) == nil && parsed.Message != "" {
		return parsed.Message
	}
	return strings.TrimSpace(string(raw))
}
