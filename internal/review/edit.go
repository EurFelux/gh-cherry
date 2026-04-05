package review

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// EditResult holds the outcome of editing a review's body.
type EditResult struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// EditReview updates the body text of a submitted review.
func EditReview(client ghcli.Querier, reviewID, body string) (*EditResult, error) {
	query := `mutation($reviewId: ID!, $body: String!) {
		updatePullRequestReview(input: { pullRequestReviewId: $reviewId, body: $body }) {
			pullRequestReview {
				id
				body
			}
		}
	}`

	var result struct {
		UpdatePullRequestReview struct {
			PullRequestReview struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			} `json:"pullRequestReview"`
		} `json:"updatePullRequestReview"`
	}

	if err := client.Query(query, map[string]any{"reviewId": reviewID, "body": body}, &result); err != nil {
		return nil, fmt.Errorf("edit review: %w", err)
	}

	pr := result.UpdatePullRequestReview.PullRequestReview
	return &EditResult{
		ID:   pr.ID,
		Body: pr.Body,
	}, nil
}
