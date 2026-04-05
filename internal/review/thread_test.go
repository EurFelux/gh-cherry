package review

import (
	"bytes"
	"encoding/json"
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

func TestListThreads(t *testing.T) {
	// Helper to build a mock that responds to the thread list query.
	buildListMock := func(threads []ThreadInfo, viewer string) *mockQuerier {
		return &mockQuerier{
			queryFunc: func(query string, _ map[string]any, result any) error {
				if viewer != "" && query == `query { viewer { login } }` {
					type viewerResp = struct {
						Viewer struct {
							Login string `json:"login"`
						} `json:"viewer"`
					}
					r := result.(*viewerResp)
					r.Viewer.Login = viewer
					return nil
				}

				// Use encoding/json to populate the result to avoid type mismatch.
				type commentNode struct {
					Author struct {
						Login string `json:"login"`
					} `json:"author"`
					Body string `json:"body"`
				}
				type threadNode struct {
					ID         string `json:"id"`
					Path       string `json:"path"`
					Line       int    `json:"line"`
					IsResolved bool   `json:"isResolved"`
					Comments   struct {
						TotalCount int           `json:"totalCount"`
						Nodes      []commentNode `json:"nodes"`
					} `json:"comments"`
				}

				nodes := make([]threadNode, len(threads))
				for i, t := range threads {
					n := threadNode{
						ID:         t.ID,
						Path:       t.Path,
						Line:       t.Line,
						IsResolved: t.IsResolved,
					}
					n.Comments.TotalCount = t.CommentCount
					n.Comments.Nodes = []commentNode{{Body: t.FirstComment.Body}}
					n.Comments.Nodes[0].Author.Login = t.FirstComment.Author
					nodes[i] = n
				}

				payload := map[string]any{
					"repository": map[string]any{
						"pullRequest": map[string]any{
							"reviewThreads": map[string]any{
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
								"nodes": nodes,
							},
						},
					},
				}
				data, _ := json.Marshal(payload)
				return json.Unmarshal(data, result)
			},
		}
	}

	sampleThreads := []ThreadInfo{
		{ID: "PRRT_1", Path: "main.go", Line: 10, IsResolved: false, CommentCount: 2, FirstComment: FirstComment{Author: "alice", Body: "Fix this"}},
		{ID: "PRRT_2", Path: "main.go", Line: 20, IsResolved: true, CommentCount: 1, FirstComment: FirstComment{Author: "bob", Body: "Looks good"}},
		{ID: "PRRT_3", Path: "util.go", Line: 5, IsResolved: false, CommentCount: 3, FirstComment: FirstComment{Author: "alice", Body: "Refactor"}},
	}

	t.Run("list all threads", func(t *testing.T) {
		mock := buildListMock(sampleThreads, "")
		threads, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
		})
		require.NoError(t, err)
		assert.Len(t, threads, 3)
	})

	t.Run("filter unresolved", func(t *testing.T) {
		mock := buildListMock(sampleThreads, "")
		threads, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
			Unresolved: true,
		})
		require.NoError(t, err)
		assert.Len(t, threads, 2)
		for _, th := range threads {
			assert.False(t, th.IsResolved)
		}
	})

	t.Run("filter mine", func(t *testing.T) {
		mock := buildListMock(sampleThreads, "alice")
		threads, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
			Mine: true,
		})
		require.NoError(t, err)
		assert.Len(t, threads, 2)
		for _, th := range threads {
			assert.Equal(t, "alice", th.FirstComment.Author)
		}
	})

	t.Run("filter unresolved and mine", func(t *testing.T) {
		mock := buildListMock(sampleThreads, "bob")
		threads, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
			Unresolved: true, Mine: true,
		})
		require.NoError(t, err)
		assert.Empty(t, threads)
	})

	t.Run("empty result returns empty array", func(t *testing.T) {
		mock := buildListMock(nil, "")
		threads, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
		})
		require.NoError(t, err)
		assert.NotNil(t, threads)
		assert.Empty(t, threads)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("network error")
			},
		}
		_, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
		})
		assert.ErrorContains(t, err, "list review threads")
	})

	t.Run("viewer error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("auth error")
			},
		}
		_, err := ListThreads(mock, ListThreadsOptions{
			Owner: "owner", Repo: "repo", PRNumber: 1,
			Mine: true,
		})
		assert.ErrorContains(t, err, "fetch viewer")
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
