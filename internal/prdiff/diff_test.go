package prdiff

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotatedDiff(t *testing.T) {
	t.Run("end to end", func(t *testing.T) {
		mock := &mockRESTQuerier{
			getFunc: func(_ string, response interface{}) error {
				files := []PRFile{
					{
						Filename: "main.go",
						Patch:    "@@ -1,3 +1,3 @@\n context\n-old\n+new\n context2",
					},
					{
						Filename: "binary.bin",
						Patch:    "", // binary file, no patch
					},
				}
				b, _ := json.Marshal(files)
				return json.Unmarshal(b, response)
			},
		}

		var buf bytes.Buffer
		err := AnnotatedDiff(mock, "owner", "repo", 1, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "main.go\n")
		assert.Contains(t, output, "L2")
		assert.Contains(t, output, "R2")
		// binary.bin should be skipped
		assert.NotContains(t, output, "binary.bin")
	})

	t.Run("no files", func(t *testing.T) {
		mock := &mockRESTQuerier{
			getFunc: func(_ string, response interface{}) error {
				b, _ := json.Marshal([]PRFile{})
				return json.Unmarshal(b, response)
			},
		}

		var buf bytes.Buffer
		err := AnnotatedDiff(mock, "owner", "repo", 1, &buf)
		require.NoError(t, err)
		assert.Empty(t, buf.String())
	})
}
