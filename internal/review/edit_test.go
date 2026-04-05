package review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// editResp matches the anonymous struct type used in EditReview.
type editResp = struct {
	UpdatePullRequestReview struct {
		PullRequestReview struct {
			ID   string `json:"id"`
			Body string `json:"body"`
		} `json:"pullRequestReview"`
	} `json:"updatePullRequestReview"`
}

func TestEditReview(t *testing.T) {
	t.Run("updates body", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]any, result any) error {
				assert.Contains(t, query, "updatePullRequestReview")
				assert.Equal(t, "PRR_123", variables["reviewId"])
				assert.Equal(t, "Updated text", variables["body"])

				r := result.(*editResp)
				r.UpdatePullRequestReview.PullRequestReview.ID = "PRR_123"
				r.UpdatePullRequestReview.PullRequestReview.Body = "Updated text"
				return nil
			},
		}

		got, err := EditReview(mock, "PRR_123", "Updated text")
		require.NoError(t, err)
		assert.Equal(t, "PRR_123", got.ID)
		assert.Equal(t, "Updated text", got.Body)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("forbidden")
			},
		}

		_, err := EditReview(mock, "PRR_123", "text")
		assert.ErrorContains(t, err, "edit review")
		assert.ErrorContains(t, err, "forbidden")
	})
}
