package review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// submitResp matches the anonymous struct type used in SubmitReview.
type submitResp = struct {
	SubmitPullRequestReview struct {
		PullRequestReview struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"pullRequestReview"`
	} `json:"submitPullRequestReview"`
}

func setSubmitResult(result any, id, state string) {
	r := result.(*submitResp)
	r.SubmitPullRequestReview.PullRequestReview.ID = id
	r.SubmitPullRequestReview.PullRequestReview.State = state
}

func TestSubmitReview(t *testing.T) {
	t.Run("approve", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]any, result any) error {
				assert.Contains(t, query, "submitPullRequestReview")
				assert.Equal(t, "PRR_123", variables["reviewId"])
				assert.Equal(t, "APPROVE", variables["event"])
				assert.Equal(t, "LGTM", variables["body"])
				setSubmitResult(result, "PRR_123", "APPROVED")
				return nil
			},
		}

		got, err := SubmitReview(mock, "PRR_123", "APPROVE", "LGTM")
		require.NoError(t, err)
		assert.Equal(t, "PRR_123", got.ID)
		assert.Equal(t, "APPROVED", got.State)
	})

	t.Run("request changes", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				assert.Equal(t, "REQUEST_CHANGES", variables["event"])
				setSubmitResult(result, "PRR_456", "CHANGES_REQUESTED")
				return nil
			},
		}

		got, err := SubmitReview(mock, "PRR_456", "REQUEST_CHANGES", "Please fix")
		require.NoError(t, err)
		assert.Equal(t, "CHANGES_REQUESTED", got.State)
	})

	t.Run("comment without body", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				assert.Equal(t, "COMMENT", variables["event"])
				_, hasBody := variables["body"]
				assert.False(t, hasBody, "body should be omitted when empty")
				setSubmitResult(result, "PRR_789", "COMMENTED")
				return nil
			},
		}

		got, err := SubmitReview(mock, "PRR_789", "COMMENT", "")
		require.NoError(t, err)
		assert.Equal(t, "COMMENTED", got.State)
	})

	t.Run("invalid event", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				t.Fatal("should not be called")
				return nil
			},
		}

		_, err := SubmitReview(mock, "PRR_123", "INVALID", "")
		assert.ErrorContains(t, err, `invalid event "INVALID"`)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("forbidden")
			},
		}

		_, err := SubmitReview(mock, "PRR_123", "APPROVE", "")
		assert.ErrorContains(t, err, "submit review")
		assert.ErrorContains(t, err, "forbidden")
	})
}
