package review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// previewResp matches the anonymous struct type used in PreviewReview.
type previewResp = struct {
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

func TestPreviewReview(t *testing.T) {
	t.Run("review with comments", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]any, result any) error {
				assert.Contains(t, query, "PullRequestReview")
				assert.Equal(t, "PRR_123", variables["id"])

				r := result.(*previewResp)
				r.Node.ID = "PRR_123"
				r.Node.State = "PENDING"
				r.Node.Body = "Overall looks good"
				r.Node.Comments.Nodes = []struct {
					ID           string `json:"id"`
					Path         string `json:"path"`
					OriginalLine int    `json:"originalLine"`
					Body         string `json:"body"`
				}{
					{ID: "PRRC_1", Path: "src/main.go", OriginalLine: 42, Body: "Fix this"},
					{ID: "PRRC_2", Path: "src/util.go", OriginalLine: 10, Body: "Rename this"},
				}
				return nil
			},
		}

		got, err := PreviewReview(mock, "PRR_123")
		require.NoError(t, err)
		assert.Equal(t, "PRR_123", got.ID)
		assert.Equal(t, "PENDING", got.State)
		assert.Equal(t, "Overall looks good", got.Body)
		require.Len(t, got.Comments, 2)
		assert.Equal(t, "PRRC_1", got.Comments[0].ID)
		assert.Equal(t, "src/main.go", got.Comments[0].Path)
		assert.Equal(t, 42, got.Comments[0].Line)
		assert.Equal(t, "Fix this", got.Comments[0].Body)
	})

	t.Run("review with no comments", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, result any) error {
				r := result.(*previewResp)
				r.Node.ID = "PRR_456"
				r.Node.State = "PENDING"
				return nil
			},
		}

		got, err := PreviewReview(mock, "PRR_456")
		require.NoError(t, err)
		assert.Equal(t, "PRR_456", got.ID)
		assert.Empty(t, got.Comments)
	})

	t.Run("review not found", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return nil
			},
		}

		_, err := PreviewReview(mock, "PRR_nonexistent")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("network error")
			},
		}

		_, err := PreviewReview(mock, "PRR_123")
		assert.ErrorContains(t, err, "preview review")
		assert.ErrorContains(t, err, "network error")
	})
}
