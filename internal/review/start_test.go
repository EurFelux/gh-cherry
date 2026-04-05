package review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockQuerier struct {
	queryFunc func(query string, variables map[string]interface{}, result interface{}) error
}

func (m *mockQuerier) Query(query string, variables map[string]interface{}, result interface{}) error {
	return m.queryFunc(query, variables, result)
}

// fetchResp matches the anonymous struct type used in fetchPRAndPendingReview.
type fetchResp = struct {
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

// mutationResp matches the anonymous struct type used in createPendingReview.
type mutationResp = struct {
	AddPullRequestReview struct {
		PullRequestReview struct {
			ID string `json:"id"`
		} `json:"pullRequestReview"`
	} `json:"addPullRequestReview"`
}

func setFetchResult(result interface{}, prID, reviewID string, threads int) {
	r := result.(*fetchResp)
	r.Repository.PullRequest.ID = prID
	if reviewID != "" {
		r.Repository.PullRequest.Reviews.Nodes = []struct {
			ID string `json:"id"`
		}{{ID: reviewID}}
		r.Repository.PullRequest.Reviews.TotalCount = 1
	}
	r.Repository.PullRequest.ReviewThreads.TotalCount = threads
}

func setMutationResult(result interface{}, reviewID string) {
	r := result.(*mutationResp)
	r.AddPullRequestReview.PullRequestReview.ID = reviewID
}

func TestStartReview_NewReview(t *testing.T) {
	calls := 0
	mock := &mockQuerier{
		queryFunc: func(_ string, variables map[string]interface{}, result interface{}) error {
			calls++
			if calls == 1 {
				assert.Equal(t, "owner", variables["owner"])
				assert.Equal(t, "repo", variables["repo"])
				assert.Equal(t, 42, variables["number"])
				setFetchResult(result, "PR_123", "", 0)
				return nil
			}
			assert.Equal(t, "PR_123", variables["prId"])
			assert.Equal(t, "looks good", variables["body"])
			setMutationResult(result, "PRR_456")
			return nil
		},
	}

	got, err := StartReview(mock, "owner", "repo", 42, "looks good")
	require.NoError(t, err)
	assert.Equal(t, "PRR_456", got.ID)
	assert.Equal(t, "PENDING", got.State)
	assert.False(t, got.Reused)
	assert.Equal(t, 0, got.ExistingThreads)
	assert.Equal(t, 2, calls)
}

func TestStartReview_ExistingReview(t *testing.T) {
	mock := &mockQuerier{
		queryFunc: func(_ string, _ map[string]interface{}, result interface{}) error {
			setFetchResult(result, "PR_123", "PRR_existing", 3)
			return nil
		},
	}

	got, err := StartReview(mock, "owner", "repo", 10, "")
	require.NoError(t, err)
	assert.Equal(t, "PRR_existing", got.ID)
	assert.Equal(t, "PENDING", got.State)
	assert.True(t, got.Reused)
	assert.Equal(t, 3, got.ExistingThreads)
}

func TestStartReview_EmptyBody(t *testing.T) {
	calls := 0
	mock := &mockQuerier{
		queryFunc: func(_ string, variables map[string]interface{}, result interface{}) error {
			calls++
			if calls == 1 {
				setFetchResult(result, "PR_123", "", 0)
				return nil
			}
			_, hasBody := variables["body"]
			assert.False(t, hasBody, "body should be omitted when empty")
			setMutationResult(result, "PRR_new")
			return nil
		},
	}

	got, err := StartReview(mock, "owner", "repo", 1, "")
	require.NoError(t, err)
	assert.Equal(t, "PRR_new", got.ID)
	assert.False(t, got.Reused)
}

func TestStartReview_FetchError(t *testing.T) {
	mock := &mockQuerier{
		queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
			return fmt.Errorf("network error")
		},
	}

	_, err := StartReview(mock, "owner", "repo", 1, "")
	assert.ErrorContains(t, err, "fetch pull request")
	assert.ErrorContains(t, err, "network error")
}

func TestStartReview_CreateError(t *testing.T) {
	calls := 0
	mock := &mockQuerier{
		queryFunc: func(_ string, _ map[string]interface{}, result interface{}) error {
			calls++
			if calls == 1 {
				setFetchResult(result, "PR_123", "", 0)
				return nil
			}
			return fmt.Errorf("forbidden")
		},
	}

	_, err := StartReview(mock, "owner", "repo", 1, "")
	assert.ErrorContains(t, err, "create review")
	assert.ErrorContains(t, err, "forbidden")
}

func TestStartReview_PRNotFound(t *testing.T) {
	mock := &mockQuerier{
		queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
			return nil
		},
	}

	_, err := StartReview(mock, "owner", "repo", 999, "")
	assert.ErrorContains(t, err, "pull request #999 not found")
}
