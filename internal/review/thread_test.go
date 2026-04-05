package review

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockQuerier struct {
	queryFunc func(query string, variables map[string]any, result any) error
}

func (m *mockQuerier) Query(query string, variables map[string]any, result any) error {
	return m.queryFunc(query, variables, result)
}

func TestAddThread(t *testing.T) {
	t.Run("single line comment", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]any, result any) error {
				assert.Contains(t, query, "addPullRequestReviewThread")
				assert.Equal(t, "PRR_123", variables["reviewId"])
				assert.Equal(t, "src/main.go", variables["path"])
				assert.Equal(t, 42, variables["line"])
				assert.Equal(t, "RIGHT", variables["side"])
				assert.Equal(t, "Fix this", variables["body"])
				assert.NotContains(t, variables, "startLine")
				assert.NotContains(t, variables, "startSide")

				type respType = struct {
					AddPullRequestReviewThread struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					} `json:"addPullRequestReviewThread"`
				}
				r := result.(*respType)
				r.AddPullRequestReviewThread.Thread.ID = "PRRT_abc"
				return nil
			},
		}

		res, err := AddThread(mock, AddThreadOptions{
			ReviewID: "PRR_123",
			Path:     "src/main.go",
			Line:     42,
			Body:     "Fix this",
		})
		require.NoError(t, err)
		assert.Equal(t, "PRRT_abc", res.ThreadID)
	})

	t.Run("multi-line comment", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				assert.Equal(t, 42, variables["line"])
				assert.Equal(t, "LEFT", variables["side"])
				assert.Equal(t, 40, variables["startLine"])
				assert.Equal(t, "LEFT", variables["startSide"])

				type respType = struct {
					AddPullRequestReviewThread struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					} `json:"addPullRequestReviewThread"`
				}
				r := result.(*respType)
				r.AddPullRequestReviewThread.Thread.ID = "PRRT_multi"
				return nil
			},
		}

		res, err := AddThread(mock, AddThreadOptions{
			ReviewID:  "PRR_123",
			Path:      "src/main.go",
			Line:      42,
			Body:      "Refactor this block",
			Side:      "LEFT",
			StartLine: 40,
			StartSide: "LEFT",
		})
		require.NoError(t, err)
		assert.Equal(t, "PRRT_multi", res.ThreadID)
	})

	t.Run("defaults side to RIGHT", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				assert.Equal(t, "RIGHT", variables["side"])

				type respType = struct {
					AddPullRequestReviewThread struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					} `json:"addPullRequestReviewThread"`
				}
				r := result.(*respType)
				r.AddPullRequestReviewThread.Thread.ID = "PRRT_default"
				return nil
			},
		}

		res, err := AddThread(mock, AddThreadOptions{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     1,
			Body:     "comment",
		})
		require.NoError(t, err)
		assert.Equal(t, "PRRT_default", res.ThreadID)
	})

	t.Run("defaults start-side to RIGHT for multi-line", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				assert.Equal(t, "RIGHT", variables["startSide"])

				type respType = struct {
					AddPullRequestReviewThread struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					} `json:"addPullRequestReviewThread"`
				}
				r := result.(*respType)
				r.AddPullRequestReviewThread.Thread.ID = "PRRT_x"
				return nil
			},
		}

		_, err := AddThread(mock, AddThreadOptions{
			ReviewID:  "PRR_123",
			Path:      "file.go",
			Line:      10,
			Body:      "comment",
			StartLine: 5,
		})
		require.NoError(t, err)
	})

	t.Run("invalid side", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				t.Fatal("should not be called")
				return nil
			},
		}

		_, err := AddThread(mock, AddThreadOptions{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     1,
			Body:     "comment",
			Side:     "MIDDLE",
		})
		assert.ErrorContains(t, err, `invalid side "MIDDLE"`)
	})

	t.Run("invalid start-side", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				t.Fatal("should not be called")
				return nil
			},
		}

		_, err := AddThread(mock, AddThreadOptions{
			ReviewID:  "PRR_123",
			Path:      "file.go",
			Line:      10,
			Body:      "comment",
			StartLine: 5,
			StartSide: "INVALID",
		})
		assert.ErrorContains(t, err, "invalid start-side")
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("network error")
			},
		}

		_, err := AddThread(mock, AddThreadOptions{
			ReviewID: "PRR_123",
			Path:     "file.go",
			Line:     1,
			Body:     "comment",
		})
		assert.ErrorContains(t, err, "add review thread")
		assert.ErrorContains(t, err, "network error")
	})
}

func TestReplyToThread(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]any, result any) error {
				assert.Contains(t, query, "addPullRequestReviewThreadReply")
				assert.Equal(t, "PRRT_123", variables["threadId"])
				assert.Equal(t, "Good point, fixed.", variables["body"])

				type respType = struct {
					AddPullRequestReviewThreadReply struct {
						Comment struct {
							ID string `json:"id"`
						} `json:"comment"`
					} `json:"addPullRequestReviewThreadReply"`
				}
				r := result.(*respType)
				r.AddPullRequestReviewThreadReply.Comment.ID = "PRRC_reply1"
				return nil
			},
		}

		res, err := ReplyToThread(mock, ReplyThreadOptions{
			ThreadID: "PRRT_123",
			Body:     "Good point, fixed.",
		})
		require.NoError(t, err)
		assert.Equal(t, "PRRC_reply1", res.CommentID)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("network error")
			},
		}

		_, err := ReplyToThread(mock, ReplyThreadOptions{
			ThreadID: "PRRT_123",
			Body:     "reply",
		})
		assert.ErrorContains(t, err, "reply to thread")
		assert.ErrorContains(t, err, "network error")
	})
}

func TestPrintResult(t *testing.T) {
	var buf bytes.Buffer
	err := PrintResult(&AddThreadResult{ThreadID: "PRRT_abc"}, &buf)
	require.NoError(t, err)
	assert.JSONEq(t, `{"threadId":"PRRT_abc"}`, buf.String())
}

func TestReadBodyFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "body.txt")
		require.NoError(t, os.WriteFile(path, []byte("hello world"), 0o644))

		content, err := ReadBodyFile(path)
		require.NoError(t, err)
		assert.Equal(t, "hello world", content)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := ReadBodyFile("/nonexistent/body.txt")
		assert.ErrorContains(t, err, "read body file")
	})
}
