package prdiff

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRESTQuerier struct {
	getFunc func(path string, response interface{}) error
}

func (m *mockRESTQuerier) Get(path string, response interface{}) error {
	return m.getFunc(path, response)
}

func TestFetchPRFiles(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		mock := &mockRESTQuerier{
			getFunc: func(path string, response interface{}) error {
				assert.Contains(t, path, "page=1")
				files := []PRFile{
					{Filename: "a.go", Patch: "@@ -1 +1 @@\n-old\n+new"},
				}
				b, _ := json.Marshal(files)
				return json.Unmarshal(b, response)
			},
		}

		files, err := FetchPRFiles(mock, "owner", "repo", 1)
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "a.go", files[0].Filename)
	})

	t.Run("multiple pages", func(t *testing.T) {
		callCount := 0
		mock := &mockRESTQuerier{
			getFunc: func(_ string, response interface{}) error {
				callCount++
				var files []PRFile
				if callCount == 1 {
					// Return full page (100 items)
					for i := range perPage {
						files = append(files, PRFile{Filename: fmt.Sprintf("file%d.go", i)})
					}
				} else {
					// Return partial page (2 items) → stop
					files = []PRFile{
						{Filename: "extra1.go"},
						{Filename: "extra2.go"},
					}
				}
				b, _ := json.Marshal(files)
				return json.Unmarshal(b, response)
			},
		}

		files, err := FetchPRFiles(mock, "owner", "repo", 1)
		require.NoError(t, err)
		assert.Len(t, files, perPage+2)
		assert.Equal(t, 2, callCount)
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockRESTQuerier{
			getFunc: func(_ string, _ interface{}) error {
				return fmt.Errorf("not found")
			},
		}

		_, err := FetchPRFiles(mock, "owner", "repo", 999)
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("empty result", func(t *testing.T) {
		mock := &mockRESTQuerier{
			getFunc: func(_ string, response interface{}) error {
				b, _ := json.Marshal([]PRFile{})
				return json.Unmarshal(b, response)
			},
		}

		files, err := FetchPRFiles(mock, "owner", "repo", 1)
		require.NoError(t, err)
		assert.Empty(t, files)
	})
}
