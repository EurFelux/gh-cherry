package review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reviewsResp matches the anonymous struct for fetchAllReviews.
type reviewsResp = struct {
	Repository struct {
		PullRequest struct {
			Reviews struct {
				Nodes []struct {
					ID     string `json:"id"`
					Author struct {
						Login string `json:"login"`
					} `json:"author"`
					State       string `json:"state"`
					Body        string `json:"body"`
					SubmittedAt string `json:"submittedAt"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"reviews"`
		} `json:"pullRequest"`
	} `json:"repository"`
}

// threadsResp matches the anonymous struct for fetchAllThreads.
type threadsResp = struct {
	Repository struct {
		PullRequest struct {
			ReviewThreads struct {
				Nodes []struct {
					ID         string `json:"id"`
					Path       string `json:"path"`
					Line       int    `json:"line"`
					IsResolved bool   `json:"isResolved"`
					Comments   struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"reviewThreads"`
		} `json:"pullRequest"`
	} `json:"repository"`
}

func TestViewReviews(t *testing.T) {
	t.Run("basic view with reviews and threads", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				calls++
				if calls == 1 {
					// reviews query
					assert.Equal(t, "owner", variables["owner"])
					assert.Equal(t, "repo", variables["repo"])
					assert.Equal(t, 42, variables["number"])

					r := result.(*reviewsResp)
					r.Repository.PullRequest.Reviews.Nodes = []struct {
						ID     string `json:"id"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						State       string `json:"state"`
						Body        string `json:"body"`
						SubmittedAt string `json:"submittedAt"`
					}{
						{ID: "PRR_1", Author: struct {
							Login string `json:"login"`
						}{Login: "alice"}, State: "APPROVED", Body: "LGTM", SubmittedAt: "2024-01-01T00:00:00Z"},
					}
					return nil
				}
				// threads query
				r := result.(*threadsResp)
				r.Repository.PullRequest.ReviewThreads.Nodes = []struct {
					ID         string `json:"id"`
					Path       string `json:"path"`
					Line       int    `json:"line"`
					IsResolved bool   `json:"isResolved"`
					Comments   struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
						} `json:"nodes"`
					} `json:"comments"`
				}{
					{
						ID: "PRRT_1", Path: "main.go", Line: 10, IsResolved: false,
						Comments: struct {
							Nodes []struct {
								ID     string `json:"id"`
								Author struct {
									Login string `json:"login"`
								} `json:"author"`
								Body      string `json:"body"`
								CreatedAt string `json:"createdAt"`
							} `json:"nodes"`
						}{
							Nodes: []struct {
								ID     string `json:"id"`
								Author struct {
									Login string `json:"login"`
								} `json:"author"`
								Body      string `json:"body"`
								CreatedAt string `json:"createdAt"`
							}{
								{ID: "PRRC_1", Author: struct {
									Login string `json:"login"`
								}{Login: "bob"}, Body: "Fix this", CreatedAt: "2024-01-01T00:00:00Z"},
							},
						},
					},
				}
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner:    "owner",
			Repo:     "repo",
			PRNumber: 42,
		})
		require.NoError(t, err)
		require.Len(t, got.Reviews, 1)
		assert.Equal(t, "PRR_1", got.Reviews[0].ID)
		assert.Equal(t, "alice", got.Reviews[0].Author)
		assert.Equal(t, "APPROVED", got.Reviews[0].State)
		require.Len(t, got.Threads, 1)
		assert.Equal(t, "PRRT_1", got.Threads[0].ID)
		assert.False(t, got.Threads[0].IsResolved)
		require.Len(t, got.Threads[0].Comments, 1)
		assert.Equal(t, "bob", got.Threads[0].Comments[0].Author)
	})

	t.Run("filter by reviewer", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, result any) error {
				calls++
				if calls == 1 {
					r := result.(*reviewsResp)
					r.Repository.PullRequest.Reviews.Nodes = []struct {
						ID     string `json:"id"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						State       string `json:"state"`
						Body        string `json:"body"`
						SubmittedAt string `json:"submittedAt"`
					}{
						{ID: "PRR_1", Author: struct {
							Login string `json:"login"`
						}{Login: "alice"}, State: "APPROVED"},
						{ID: "PRR_2", Author: struct {
							Login string `json:"login"`
						}{Login: "bob"}, State: "CHANGES_REQUESTED"},
					}
					return nil
				}
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner: "o", Repo: "r", PRNumber: 1,
			Reviewer: "alice",
		})
		require.NoError(t, err)
		require.Len(t, got.Reviews, 1)
		assert.Equal(t, "alice", got.Reviews[0].Author)
	})

	t.Run("filter by state", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, result any) error {
				calls++
				if calls == 1 {
					r := result.(*reviewsResp)
					r.Repository.PullRequest.Reviews.Nodes = []struct {
						ID     string `json:"id"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						State       string `json:"state"`
						Body        string `json:"body"`
						SubmittedAt string `json:"submittedAt"`
					}{
						{ID: "PRR_1", State: "APPROVED"},
						{ID: "PRR_2", State: "CHANGES_REQUESTED"},
					}
					return nil
				}
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner: "o", Repo: "r", PRNumber: 1,
			State: "APPROVED",
		})
		require.NoError(t, err)
		require.Len(t, got.Reviews, 1)
		assert.Equal(t, "APPROVED", got.Reviews[0].State)
	})

	t.Run("filter unresolved threads", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, result any) error {
				calls++
				if calls == 1 {
					return nil
				}
				r := result.(*threadsResp)
				r.Repository.PullRequest.ReviewThreads.Nodes = []struct {
					ID         string `json:"id"`
					Path       string `json:"path"`
					Line       int    `json:"line"`
					IsResolved bool   `json:"isResolved"`
					Comments   struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
						} `json:"nodes"`
					} `json:"comments"`
				}{
					{ID: "PRRT_1", IsResolved: false},
					{ID: "PRRT_2", IsResolved: true},
				}
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner: "o", Repo: "r", PRNumber: 1,
			Unresolved: true,
		})
		require.NoError(t, err)
		require.Len(t, got.Threads, 1)
		assert.Equal(t, "PRRT_1", got.Threads[0].ID)
	})

	t.Run("tail limits comments per thread", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, result any) error {
				calls++
				if calls == 1 {
					return nil
				}
				r := result.(*threadsResp)
				comments := make([]struct {
					ID     string `json:"id"`
					Author struct {
						Login string `json:"login"`
					} `json:"author"`
					Body      string `json:"body"`
					CreatedAt string `json:"createdAt"`
				}, 5)
				for i := range comments {
					comments[i].ID = fmt.Sprintf("PRRC_%d", i+1)
					comments[i].Body = fmt.Sprintf("comment %d", i+1)
				}

				r.Repository.PullRequest.ReviewThreads.Nodes = []struct {
					ID         string `json:"id"`
					Path       string `json:"path"`
					Line       int    `json:"line"`
					IsResolved bool   `json:"isResolved"`
					Comments   struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
						} `json:"nodes"`
					} `json:"comments"`
				}{
					{ID: "PRRT_1"},
				}
				r.Repository.PullRequest.ReviewThreads.Nodes[0].Comments.Nodes = comments
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner: "o", Repo: "r", PRNumber: 1,
			Tail: 2,
		})
		require.NoError(t, err)
		require.Len(t, got.Threads, 1)
		require.Len(t, got.Threads[0].Comments, 2)
		assert.Equal(t, "PRRC_4", got.Threads[0].Comments[0].ID)
		assert.Equal(t, "PRRC_5", got.Threads[0].Comments[1].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]any, result any) error {
				calls++
				switch calls {
				case 1: // first reviews page
					r := result.(*reviewsResp)
					r.Repository.PullRequest.Reviews.Nodes = []struct {
						ID     string `json:"id"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						State       string `json:"state"`
						Body        string `json:"body"`
						SubmittedAt string `json:"submittedAt"`
					}{
						{ID: "PRR_1", State: "APPROVED"},
					}
					r.Repository.PullRequest.Reviews.PageInfo.HasNextPage = true
					r.Repository.PullRequest.Reviews.PageInfo.EndCursor = "cursor1"
				case 2: // second reviews page
					assert.Equal(t, "cursor1", variables["cursor"])
					r := result.(*reviewsResp)
					r.Repository.PullRequest.Reviews.Nodes = []struct {
						ID     string `json:"id"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						State       string `json:"state"`
						Body        string `json:"body"`
						SubmittedAt string `json:"submittedAt"`
					}{
						{ID: "PRR_2", State: "COMMENTED"},
					}
				case 3: // threads (single page)
					return nil
				}
				return nil
			},
		}

		got, err := ViewReviews(mock, ViewOptions{
			Owner: "o", Repo: "r", PRNumber: 1,
		})
		require.NoError(t, err)
		require.Len(t, got.Reviews, 2)
		assert.Equal(t, "PRR_1", got.Reviews[0].ID)
		assert.Equal(t, "PRR_2", got.Reviews[1].ID)
	})

	t.Run("fetch reviews error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				return fmt.Errorf("network error")
			},
		}

		_, err := ViewReviews(mock, ViewOptions{Owner: "o", Repo: "r", PRNumber: 1})
		assert.ErrorContains(t, err, "fetch reviews")
	})

	t.Run("fetch threads error", func(t *testing.T) {
		calls := 0
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]any, _ any) error {
				calls++
				if calls == 1 {
					return nil
				}
				return fmt.Errorf("timeout")
			},
		}

		_, err := ViewReviews(mock, ViewOptions{Owner: "o", Repo: "r", PRNumber: 1})
		assert.ErrorContains(t, err, "fetch threads")
		assert.ErrorContains(t, err, "timeout")
	})
}
