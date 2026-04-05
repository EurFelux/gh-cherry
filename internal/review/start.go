package review

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// StartResult holds the outcome of starting (or reusing) a pending review.
type StartResult struct {
	ID              string `json:"id"`
	State           string `json:"state"`
	Reused          bool   `json:"reused"`
	ExistingThreads int    `json:"existingThreads"`
}

// StartReview creates a new pending review on the given PR, or returns the
// existing pending review if one already exists for the authenticated user.
func StartReview(client ghcli.Querier, owner, repo string, prNumber int, body string) (*StartResult, error) {
	prID, existing, err := fetchPRAndPendingReview(client, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("fetch pull request: %w", err)
	}

	if existing != nil {
		return existing, nil
	}

	reviewID, err := createPendingReview(client, prID, body)
	if err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	return &StartResult{
		ID:              reviewID,
		State:           "PENDING",
		Reused:          false,
		ExistingThreads: 0,
	}, nil
}

// fetchPRAndPendingReview retrieves the PR node ID and checks for an existing
// pending review. Returns (prID, existingResult, error). existingResult is nil
// when no pending review exists.
func fetchPRAndPendingReview(client ghcli.Querier, owner, repo string, prNumber int) (string, *StartResult, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				id
				reviews(states: PENDING, first: 1) {
					nodes { id }
					totalCount
				}
				reviewThreads(first: 0) {
					totalCount
				}
			}
		}
	}`

	var result struct {
		Repository struct {
			PullRequest struct {
				ID      string `json:"id"`
				Reviews struct {
					Nodes []struct {
						ID string `json:"id"`
					} `json:"nodes"`
					TotalCount int `json:"totalCount"`
				} `json:"reviews"`
				ReviewThreads struct {
					TotalCount int `json:"totalCount"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	err := client.Query(query, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": prNumber,
	}, &result)
	if err != nil {
		return "", nil, err
	}

	pr := result.Repository.PullRequest
	if pr.ID == "" {
		return "", nil, fmt.Errorf("pull request #%d not found", prNumber)
	}

	if len(pr.Reviews.Nodes) > 0 {
		return pr.ID, &StartResult{
			ID:              pr.Reviews.Nodes[0].ID,
			State:           "PENDING",
			Reused:          true,
			ExistingThreads: pr.ReviewThreads.TotalCount,
		}, nil
	}

	return pr.ID, nil, nil
}

// createPendingReview creates a new pending review via the addPullRequestReview mutation.
func createPendingReview(client ghcli.Querier, prID, body string) (string, error) {
	query := `mutation($prId: ID!, $body: String) {
		addPullRequestReview(input: { pullRequestId: $prId, body: $body }) {
			pullRequestReview {
				id
			}
		}
	}`

	var result struct {
		AddPullRequestReview struct {
			PullRequestReview struct {
				ID string `json:"id"`
			} `json:"pullRequestReview"`
		} `json:"addPullRequestReview"`
	}

	vars := map[string]interface{}{"prId": prID}
	if body != "" {
		vars["body"] = body
	}

	err := client.Query(query, vars, &result)
	if err != nil {
		return "", err
	}

	return result.AddPullRequestReview.PullRequestReview.ID, nil
}
