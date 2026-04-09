package issue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callTrackingQuerier tracks GraphQL calls to serve different responses.
type callTrackingQuerier struct {
	calls     int
	queryFunc func(callIndex int, query string, variables map[string]any, result any) error
}

func (m *callTrackingQuerier) Query(query string, variables map[string]any, result any) error {
	idx := m.calls
	m.calls++
	return m.queryFunc(idx, query, variables, result)
}

// populateIssueNodeID populates a GetIssueNodeID response with the given ID.
func populateIssueNodeID(result any, nodeID string) {
	type respType = struct {
		Repository struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"repository"`
	}
	r := result.(*respType)
	r.Repository.Issue.ID = nodeID
}

func TestAddSubIssue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, query string, variables map[string]any, result any) error {
				switch callIndex {
				case 0: // GetIssueNodeID for parent
					assert.Equal(t, 1, variables["number"])
					populateIssueNodeID(result, "I_parent")
				case 1: // GetIssueNodeID for child
					assert.Equal(t, 4, variables["number"])
					populateIssueNodeID(result, "I_child")
				case 2: // addSubIssue mutation
					assert.Contains(t, query, "addSubIssue")
					assert.Equal(t, "I_parent", variables["issueId"])
					assert.Equal(t, "I_child", variables["subIssueId"])
				}
				return nil
			},
		}

		got, err := AddSubIssue(mock, "owner", "repo", 1, 4)
		require.NoError(t, err)
		assert.Equal(t, 1, got.Parent)
		assert.Equal(t, 4, got.Child)
		assert.Equal(t, "added", got.Action)
	})

	t.Run("mutation error", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, _ string, _ map[string]any, result any) error {
				if callIndex < 2 {
					populateIssueNodeID(result, fmt.Sprintf("I_%d", callIndex))
					return nil
				}
				return fmt.Errorf("permission denied")
			},
		}

		_, err := AddSubIssue(mock, "owner", "repo", 1, 4)
		assert.ErrorContains(t, err, "permission denied")
	})

	t.Run("parent not found", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(_ int, _ string, _ map[string]any, _ any) error {
				return fmt.Errorf("issue #999 not found")
			},
		}

		_, err := AddSubIssue(mock, "owner", "repo", 999, 4)
		assert.ErrorContains(t, err, "resolve parent issue")
	})
}

func TestRemoveSubIssue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, query string, variables map[string]any, result any) error {
				switch callIndex {
				case 0:
					populateIssueNodeID(result, "I_parent")
				case 1:
					populateIssueNodeID(result, "I_child")
				case 2:
					assert.Contains(t, query, "removeSubIssue")
					assert.Equal(t, "I_parent", variables["issueId"])
					assert.Equal(t, "I_child", variables["subIssueId"])
				}
				return nil
			},
		}

		got, err := RemoveSubIssue(mock, "owner", "repo", 1, 4)
		require.NoError(t, err)
		assert.Equal(t, 1, got.Parent)
		assert.Equal(t, 4, got.Child)
		assert.Equal(t, "removed", got.Action)
	})

	t.Run("mutation error", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, _ string, _ map[string]any, result any) error {
				if callIndex < 2 {
					populateIssueNodeID(result, fmt.Sprintf("I_%d", callIndex))
					return nil
				}
				return fmt.Errorf("not found")
			},
		}

		_, err := RemoveSubIssue(mock, "owner", "repo", 1, 4)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestListSubIssues(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, query string, _ map[string]any, result any) error {
				if callIndex == 0 {
					populateIssueNodeID(result, "I_parent")
					return nil
				}
				// list query
				assert.Contains(t, query, "subIssues")
				type respType = struct {
					Node struct {
						SubIssues struct {
							TotalCount int            `json:"totalCount"`
							Nodes      []SubIssueInfo `json:"nodes"`
						} `json:"subIssues"`
					} `json:"node"`
				}
				r := result.(*respType)
				r.Node.SubIssues.TotalCount = 2
				r.Node.SubIssues.Nodes = []SubIssueInfo{
					{Number: 4, Title: "review start", State: "OPEN"},
					{Number: 5, Title: "review submit", State: "OPEN"},
				}
				return nil
			},
		}

		got, err := ListSubIssues(mock, "owner", "repo", 1)
		require.NoError(t, err)
		assert.Len(t, got.SubIssues, 2)
		assert.Equal(t, 2, got.TotalCount)
		assert.Equal(t, 4, got.SubIssues[0].Number)
		assert.Equal(t, "review start", got.SubIssues[0].Title)
		assert.Equal(t, "OPEN", got.SubIssues[0].State)
	})

	t.Run("empty list", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, _ string, _ map[string]any, result any) error {
				if callIndex == 0 {
					populateIssueNodeID(result, "I_parent")
				}
				return nil
			},
		}

		got, err := ListSubIssues(mock, "owner", "repo", 1)
		require.NoError(t, err)
		assert.Empty(t, got.SubIssues)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &callTrackingQuerier{
			queryFunc: func(callIndex int, _ string, _ map[string]any, result any) error {
				if callIndex == 0 {
					populateIssueNodeID(result, "I_parent")
					return nil
				}
				return fmt.Errorf("server error")
			},
		}

		_, err := ListSubIssues(mock, "owner", "repo", 1)
		assert.ErrorContains(t, err, "server error")
	})
}
