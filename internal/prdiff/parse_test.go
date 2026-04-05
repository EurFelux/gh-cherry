package prdiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePatch(t *testing.T) {
	t.Run("single hunk with all line types", func(t *testing.T) {
		patch := `@@ -10,5 +10,6 @@ func main() {
     existing code
-    old logic
+    new logic line 1
+    new logic line 2
     more existing code
 }`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		assert.Contains(t, hunks[0].Header, "@@ -10,5 +10,6 @@")

		lines := hunks[0].Lines
		require.Len(t, lines, 6)

		// context: R10
		assert.Equal(t, "R", lines[0].Side)
		assert.Equal(t, 10, lines[0].LineNumber)
		assert.Equal(t, " ", lines[0].Prefix)
		assert.Equal(t, "    existing code", lines[0].Content)

		// deleted: L11
		assert.Equal(t, "L", lines[1].Side)
		assert.Equal(t, 11, lines[1].LineNumber)
		assert.Equal(t, "-", lines[1].Prefix)
		assert.Equal(t, "    old logic", lines[1].Content)

		// added: R11, R12
		assert.Equal(t, "R", lines[2].Side)
		assert.Equal(t, 11, lines[2].LineNumber)
		assert.Equal(t, "+", lines[2].Prefix)

		assert.Equal(t, "R", lines[3].Side)
		assert.Equal(t, 12, lines[3].LineNumber)

		// context: R13, R14
		assert.Equal(t, "R", lines[4].Side)
		assert.Equal(t, 13, lines[4].LineNumber)

		assert.Equal(t, "R", lines[5].Side)
		assert.Equal(t, 14, lines[5].LineNumber)
	})

	t.Run("multiple hunks", func(t *testing.T) {
		patch := `@@ -1,3 +1,3 @@
 line1
-old2
+new2
 line3
@@ -20,3 +20,3 @@
 line20
-old21
+new21
 line22`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 2)

		// First hunk starts at line 1
		assert.Equal(t, 1, hunks[0].Lines[0].LineNumber)
		// old2 is L2
		assert.Equal(t, "L", hunks[0].Lines[1].Side)
		assert.Equal(t, 2, hunks[0].Lines[1].LineNumber)
		// new2 is R2
		assert.Equal(t, "R", hunks[0].Lines[2].Side)
		assert.Equal(t, 2, hunks[0].Lines[2].LineNumber)

		// Second hunk starts at line 20
		assert.Equal(t, 20, hunks[1].Lines[0].LineNumber)
		// old21 is L21
		assert.Equal(t, "L", hunks[1].Lines[1].Side)
		assert.Equal(t, 21, hunks[1].Lines[1].LineNumber)
		// new21 is R21
		assert.Equal(t, "R", hunks[1].Lines[2].Side)
		assert.Equal(t, 21, hunks[1].Lines[2].LineNumber)
	})

	t.Run("only additions", func(t *testing.T) {
		patch := `@@ -0,0 +1,3 @@
+line1
+line2
+line3`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		lines := hunks[0].Lines
		require.Len(t, lines, 3)

		for i, l := range lines {
			assert.Equal(t, "R", l.Side)
			assert.Equal(t, i+1, l.LineNumber)
			assert.Equal(t, "+", l.Prefix)
		}
	})

	t.Run("only deletions", func(t *testing.T) {
		patch := `@@ -1,3 +0,0 @@
-line1
-line2
-line3`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		lines := hunks[0].Lines
		require.Len(t, lines, 3)

		for i, l := range lines {
			assert.Equal(t, "L", l.Side)
			assert.Equal(t, i+1, l.LineNumber)
			assert.Equal(t, "-", l.Prefix)
		}
	})

	t.Run("interleaved add and delete", func(t *testing.T) {
		patch := `@@ -5,4 +5,4 @@
-old5
-old6
+new5
+new6
 context7
 context8`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		lines := hunks[0].Lines

		// -old5: L5, -old6: L6
		assert.Equal(t, AnnotatedLine{Side: "L", LineNumber: 5, Prefix: "-", Content: "old5"}, lines[0])
		assert.Equal(t, AnnotatedLine{Side: "L", LineNumber: 6, Prefix: "-", Content: "old6"}, lines[1])
		// +new5: R5, +new6: R6
		assert.Equal(t, AnnotatedLine{Side: "R", LineNumber: 5, Prefix: "+", Content: "new5"}, lines[2])
		assert.Equal(t, AnnotatedLine{Side: "R", LineNumber: 6, Prefix: "+", Content: "new6"}, lines[3])
		// context7: R7, context8: R8
		assert.Equal(t, "R", lines[4].Side)
		assert.Equal(t, 7, lines[4].LineNumber)
		assert.Equal(t, "R", lines[5].Side)
		assert.Equal(t, 8, lines[5].LineNumber)
	})

	t.Run("no newline marker", func(t *testing.T) {
		patch := `@@ -1,2 +1,2 @@
-old
+new
\ No newline at end of file`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		// The "\ No newline" line should be skipped
		require.Len(t, hunks[0].Lines, 2)
	})

	t.Run("empty patch", func(t *testing.T) {
		hunks := ParsePatch("")
		assert.Nil(t, hunks)
	})

	t.Run("trailing newline does not corrupt line numbers", func(t *testing.T) {
		// GitHub API patch strings typically end with \n
		patch := "@@ -1,2 +1,2 @@\n context\n-old\n+new\n"

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		require.Len(t, hunks[0].Lines, 3)

		assert.Equal(t, 1, hunks[0].Lines[0].LineNumber) // context R1
		assert.Equal(t, 2, hunks[0].Lines[1].LineNumber) // -old L2
		assert.Equal(t, 2, hunks[0].Lines[2].LineNumber) // +new R2
	})

	t.Run("hunk header without count", func(t *testing.T) {
		patch := `@@ -1 +1 @@
-old
+new`

		hunks := ParsePatch(patch)
		require.Len(t, hunks, 1)
		assert.Equal(t, "L", hunks[0].Lines[0].Side)
		assert.Equal(t, 1, hunks[0].Lines[0].LineNumber)
		assert.Equal(t, "R", hunks[0].Lines[1].Side)
		assert.Equal(t, 1, hunks[0].Lines[1].LineNumber)
	})
}

func TestAnnotateFile(t *testing.T) {
	file := PRFile{
		Filename: "src/main.go",
		Patch:    "@@ -1,2 +1,2 @@\n-old\n+new",
	}

	result := AnnotateFile(file)
	assert.Equal(t, "src/main.go", result.Filename)
	require.Len(t, result.Hunks, 1)
	require.Len(t, result.Hunks[0].Lines, 2)
}
