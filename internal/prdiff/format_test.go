package prdiff

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		files := []AnnotatedFile{
			{
				Filename: "src/main.go",
				Hunks: []AnnotatedHunk{
					{
						Header: "@@ -10,3 +10,4 @@ func main() {",
						Lines: []AnnotatedLine{
							{Side: "R", LineNumber: 10, Prefix: " ", Content: "  existing code"},
							{Side: "L", LineNumber: 11, Prefix: "-", Content: "  old logic"},
							{Side: "R", LineNumber: 11, Prefix: "+", Content: "  new logic 1"},
							{Side: "R", LineNumber: 12, Prefix: "+", Content: "  new logic 2"},
							{Side: "R", LineNumber: 13, Prefix: " ", Content: "  more code"},
						},
					},
				},
			},
		}

		var buf bytes.Buffer
		err := Format(files, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "src/main.go\n")
		assert.Contains(t, output, "@@ -10,3 +10,4 @@ func main() {\n")
		assert.Contains(t, output, "  R10 |    existing code\n")
		assert.Contains(t, output, "- L11 | -  old logic\n")
		assert.Contains(t, output, "+ R11 | +  new logic 1\n")
		assert.Contains(t, output, "+ R12 | +  new logic 2\n")
		assert.Contains(t, output, "  R13 |    more code\n")
	})

	t.Run("multi-file separator", func(t *testing.T) {
		files := []AnnotatedFile{
			{
				Filename: "a.go",
				Hunks: []AnnotatedHunk{
					{Header: "@@ -1 +1 @@", Lines: []AnnotatedLine{
						{Side: "R", LineNumber: 1, Prefix: " ", Content: "a"},
					}},
				},
			},
			{
				Filename: "b.go",
				Hunks: []AnnotatedHunk{
					{Header: "@@ -1 +1 @@", Lines: []AnnotatedLine{
						{Side: "R", LineNumber: 1, Prefix: " ", Content: "b"},
					}},
				},
			},
		}

		var buf bytes.Buffer
		err := Format(files, &buf)
		require.NoError(t, err)

		output := buf.String()
		// Two files separated by blank line
		assert.Contains(t, output, "a.go\n")
		assert.Contains(t, output, "\nb.go\n")
	})

	t.Run("line number alignment", func(t *testing.T) {
		files := []AnnotatedFile{
			{
				Filename: "test.go",
				Hunks: []AnnotatedHunk{
					{Header: "@@ -99,3 +99,3 @@", Lines: []AnnotatedLine{
						{Side: "R", LineNumber: 99, Prefix: " ", Content: "a"},
						{Side: "R", LineNumber: 100, Prefix: " ", Content: "b"},
						{Side: "R", LineNumber: 101, Prefix: " ", Content: "c"},
					}},
				},
			},
		}

		var buf bytes.Buffer
		err := Format(files, &buf)
		require.NoError(t, err)

		output := buf.String()
		// R99 should be padded to match R100/R101 width (4 chars)
		assert.Contains(t, output, "  R99  |")
		assert.Contains(t, output, "  R100 |")
		assert.Contains(t, output, "  R101 |")
	})

	t.Run("empty files", func(t *testing.T) {
		var buf bytes.Buffer
		err := Format(nil, &buf)
		require.NoError(t, err)
		assert.Empty(t, buf.String())
	})
}
