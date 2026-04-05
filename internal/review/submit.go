package review

import (
	"fmt"
	"slices"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// ValidEvents contains the allowed values for the review event parameter.
var ValidEvents = []string{"APPROVE", "REQUEST_CHANGES", "COMMENT"}

// SubmitResult holds the outcome of submitting a pending review.
type SubmitResult struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// SubmitReview submits a pending pull request review with the given event type.
func SubmitReview(client ghcli.Querier, reviewID, event, body string) (*SubmitResult, error) {
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	query := `mutation($reviewId: ID!, $event: PullRequestReviewEvent!, $body: String) {
		submitPullRequestReview(input: { pullRequestReviewId: $reviewId, event: $event, body: $body }) {
			pullRequestReview {
				id
				state
			}
		}
	}`

	vars := map[string]any{
		"reviewId": reviewID,
		"event":    event,
	}
	if body != "" {
		vars["body"] = body
	}

	var result struct {
		SubmitPullRequestReview struct {
			PullRequestReview struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequestReview"`
		} `json:"submitPullRequestReview"`
	}

	if err := client.Query(query, vars, &result); err != nil {
		return nil, fmt.Errorf("submit review: %w", err)
	}

	pr := result.SubmitPullRequestReview.PullRequestReview
	return &SubmitResult{
		ID:    pr.ID,
		State: pr.State,
	}, nil
}

func validateEvent(event string) error {
	if !slices.Contains(ValidEvents, event) {
		return fmt.Errorf("invalid event %q, must be one of: APPROVE, REQUEST_CHANGES, COMMENT", event)
	}
	return nil
}
