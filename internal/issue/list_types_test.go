package issue

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTypes(t *testing.T) {
	t.Run("with types", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, result interface{}) error {
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

		var buf bytes.Buffer
		err := ListTypes(mock, "owner/repo", &buf)
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Bug")
		assert.Contains(t, output, "Feature")
		assert.Contains(t, output, "IT_1")
		assert.Contains(t, output, "IT_2")
	})

	t.Run("no types", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return nil
			},
		}

		var buf bytes.Buffer
		err := ListTypes(mock, "owner/repo", &buf)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "No issue types configured")
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return fmt.Errorf("unauthorized")
			},
		}

		var buf bytes.Buffer
		err := ListTypes(mock, "owner/repo", &buf)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("invalid repo flag", func(t *testing.T) {
		mock := &mockQuerier{
			queryFunc: func(_ string, _ map[string]interface{}, _ interface{}) error {
				return nil
			},
		}

		var buf bytes.Buffer
		err := ListTypes(mock, "bad-repo", &buf)
		assert.ErrorContains(t, err, "invalid repo format")
	})
}
