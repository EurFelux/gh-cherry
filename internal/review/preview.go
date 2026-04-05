package review

import (
	"fmt"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
)

// PreviewComment represents a single comment in a pending review.
type PreviewComment struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Body string `json:"body"`
}

// PreviewResult holds the preview of a pending review.
type PreviewResult struct {
	ID       string           `json:"id"`
	State    string           `json:"state"`
	Body     string           `json:"body,omitempty"`
	Comments []PreviewComment `json:"comments"`
}

// PreviewReview fetches a review node and its pending comments for preview.
func PreviewReview(client ghcli.Querier, reviewID string) (*PreviewResult, error) {
	query := `query($id: ID!) {
		node(id: $id) {
			... on PullRequestReview {
				id
				state
				body
				comments(first: 100) {
					nodes {
						id
						path
						originalLine
						body
					}
				}
			}
		}
	}`

	var result struct {
		Node struct {
			ID    string `json:"id"`
			State string `json:"state"`
			Body  string `json:"body"`

			Comments struct {
				Nodes []struct {
					ID           string `json:"id"`
					Path         string `json:"path"`
					OriginalLine int    `json:"originalLine"`
					Body         string `json:"body"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	}

	if err := client.Query(query, map[string]any{"id": reviewID}, &result); err != nil {
		return nil, fmt.Errorf("preview review: %w", err)
	}

	node := result.Node
	if node.ID == "" {
		return nil, fmt.Errorf("review %q not found", reviewID)
	}

	comments := make([]PreviewComment, len(node.Comments.Nodes))
	for i, c := range node.Comments.Nodes {
		comments[i] = PreviewComment{
			ID:   c.ID,
			Path: c.Path,
			Line: c.OriginalLine,
			Body: c.Body,
		}
	}

	return &PreviewResult{
		ID:       node.ID,
		State:    node.State,
		Body:     node.Body,
		Comments: comments,
	}, nil
}
