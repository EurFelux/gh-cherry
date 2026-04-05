package issue

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

func TestFindTypeByName(t *testing.T) {
	types := []Type{
		{ID: "1", Name: "Bug"},
		{ID: "2", Name: "Feature"},
		{ID: "3", Name: "Task"},
	}

	t.Run("found", func(t *testing.T) {
		got, err := FindTypeByName(types, "Bug")
		require.NoError(t, err)
		assert.Equal(t, "1", got.ID)
		assert.Equal(t, "Bug", got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := FindTypeByName(types, "Epic")
		assert.ErrorContains(t, err, `issue type "Epic" not found`)
		assert.ErrorContains(t, err, "Bug")
		assert.ErrorContains(t, err, "Feature")
		assert.ErrorContains(t, err, "Task")
	})

	t.Run("empty list", func(t *testing.T) {
		_, err := FindTypeByName(nil, "Bug")
		assert.ErrorContains(t, err, `issue type "Bug" not found`)
	})
}

func TestFetchTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, variables map[string]interface{}, result interface{}) error {
				assert.Equal(t, "owner", variables["owner"])
				assert.Equal(t, "repo", variables["repo"])
				// Simulate go-gh GraphQL response (auto-unwrapped, no "data" wrapper)
				type respType = struct {
					Repository struct {
						IssueTypes struct {
							Nodes []Type `json:"nodes"`
						} `json:"issueTypes"`
					} `json:"repository"`
				}
				r := result.(*respType)
				r.Repository.IssueTypes.Nodes = []Type{
					{ID: "IT_1", Name: "Bug"},
					{ID: "IT_2", Name: "Feature"},
				}
				return nil
			},
		}

		types, err := FetchTypes(mock, "owner", "repo")
		require.NoError(t, err)
		assert.Len(t, types, 2)
		assert.Equal(t, "Bug", types[0].Name)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return fmt.Errorf("network error")
			},
		}

		_, err := FetchTypes(mock, "owner", "repo")
		assert.ErrorContains(t, err, "network error")
	})
}

func TestSetType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(query string, variables map[string]interface{}, _ interface{}) error {
				assert.Contains(t, query, "$issueTypeId")
				assert.NotContains(t, query, "$issueTypeID")
				assert.Equal(t, "I_123", variables["issueId"])
				assert.Equal(t, "IT_1", variables["issueTypeId"])
				return nil
			},
		}

		err := SetType(mock, "I_123", "IT_1")
		assert.NoError(t, err)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return fmt.Errorf("forbidden")
			},
		}

		err := SetType(mock, "I_123", "IT_1")
		assert.ErrorContains(t, err, "forbidden")
	})
}
