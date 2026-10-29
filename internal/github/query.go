package github

const searchQuery = `
query($query: String!, $cursor: String) {
  search(query: $query, type: ISSUE, first: 15, after: $cursor) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        createdAt
        isDraft
        reviewDecision
        mergeStateStatus
        repository { nameWithOwner }
        comments { totalCount }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup {
                state
                contexts(first: 100) {
                  nodes {
                    __typename
                    ... on CheckRun { conclusion status }
                    ... on StatusContext { state }
                  }
                }
              }
            }
          }
        }
      }
    }
    pageInfo { hasNextPage endCursor }
  }
}`

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func newSearchRequest(query, cursor string) graphQLRequest {
	var cursorVar any
	if cursor != "" {
		cursorVar = cursor
	}
	return graphQLRequest{
		Query:     searchQuery,
		Variables: map[string]any{"query": query, "cursor": cursorVar},
	}
}
