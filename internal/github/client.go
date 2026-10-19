package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultGraphQLURL = "https://api.github.com/graphql"
	defaultRESTURL    = "https://api.github.com"
	maxPages          = 17
	maxAttempts       = 3
	defaultRetryDelay = 500 * time.Millisecond
)

// Client is a Fetcher backed by the real GitHub GraphQL API.
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
	restURL    string
	retryDelay time.Duration
}

var _ Fetcher = (*Client)(nil)
var _ Merger = (*Client)(nil)

// NewClient builds a Client authenticated with the given token.
func NewClient(token string) *Client {
	return &Client{
		httpClient: http.DefaultClient,
		token:      token,
		baseURL:    defaultGraphQLURL,
		restURL:    defaultRESTURL,
		retryDelay: defaultRetryDelay,
	}
}

// SearchReviewRequestedPRs fetches every open PR where the authenticated user
// is a requested reviewer, paginating until exhausted or maxPages is reached.
func (c *Client) SearchReviewRequestedPRs(ctx context.Context) ([]PullRequest, error) {
	var all []PullRequest
	cursor := ""

	for page := 0; page < maxPages; page++ {
		resp, err := c.doRequest(ctx, cursor)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Data.Search.Nodes...)

		if !resp.Data.Search.PageInfo.HasNextPage {
			break
		}
		cursor = resp.Data.Search.PageInfo.EndCursor
	}

	return all, nil
}

// post sends the request body, retrying on gateway errors (502, 503, 504)
// up to maxAttempts times. The delay doubles after each failed attempt.
func (c *Client) post(ctx context.Context, body []byte) (*http.Response, error) {
	delay := c.retryDelay
	for attempt := 1; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build graphql request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("graphql request failed: %w", err)
		}
		if !isGatewayError(resp.StatusCode) || attempt == maxAttempts {
			return resp, nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("graphql request failed: %w", ctx.Err())
		case <-time.After(delay):
		}
		delay *= 2
	}
}

func isGatewayError(status int) bool {
	return status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func (c *Client) doRequest(ctx context.Context, cursor string) (*searchResponse, error) {
	body, err := json.Marshal(newSearchRequest(cursor))
	if err != nil {
		return nil, fmt.Errorf("marshal graphql request: %w", err)
	}

	resp, err := c.post(ctx, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("graphql request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}

	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("graphql errors: %s", strings.Join(msgs, "; "))
	}

	return &result, nil
}
