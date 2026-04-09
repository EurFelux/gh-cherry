package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findSubcommand traverses the command tree by names.
func findSubcommand(root *cobra.Command, names ...string) *cobra.Command {
	cmd := root
	for _, name := range names {
		var found *cobra.Command
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = c
				break
			}
		}
		if found == nil {
			return nil
		}
		cmd = found
	}
	return cmd
}

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "cherry", rootCmd.Use)

	t.Run("has --jq persistent flag", func(t *testing.T) {
		f := rootCmd.PersistentFlags().Lookup("jq")
		require.NotNil(t, f)
		assert.Equal(t, "string", f.Value.Type())
	})

	t.Run("has issue subcommand", func(t *testing.T) {
		assert.NotNil(t, findSubcommand(rootCmd, "issue"))
	})

	t.Run("has pr subcommand", func(t *testing.T) {
		assert.NotNil(t, findSubcommand(rootCmd, "pr"))
	})

	t.Run("has review subcommand", func(t *testing.T) {
		assert.NotNil(t, findSubcommand(rootCmd, "review"))
	})
}

func TestIssueCommands(t *testing.T) {
	issueCmd := findSubcommand(rootCmd, "issue")
	require.NotNil(t, issueCmd)

	t.Run("has create subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "create")
		require.NotNil(t, cmd)
	})

	t.Run("has types subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "types")
		require.NotNil(t, cmd)
	})

	t.Run("has subissue subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "subissue")
		require.NotNil(t, cmd)
	})

	t.Run("has subissue add subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "subissue", "add")
		require.NotNil(t, cmd)
	})

	t.Run("has subissue remove subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "subissue", "remove")
		require.NotNil(t, cmd)
	})

	t.Run("has subissue list subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "issue", "subissue", "list")
		require.NotNil(t, cmd)
	})
}

func TestIssueCreateFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "issue", "create")
	require.NotNil(t, cmd)

	tests := []struct {
		name      string
		flag      string
		shorthand string
		flagType  string
	}{
		{"title", "title", "t", "string"},
		{"body", "body", "b", "string"},
		{"label", "label", "l", "stringSlice"},
		{"assignee", "assignee", "a", "stringSlice"},
		{"milestone", "milestone", "m", "string"},
		{"project", "project", "p", "stringSlice"},
		{"type", "type", "T", "string"},
		{"parent", "parent", "P", "int"},
		{"repo", "repo", "R", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := cmd.Flags().Lookup(tt.flag)
			require.NotNilf(t, f, "flag --%s not registered", tt.flag)
			assert.Equal(t, tt.shorthand, f.Shorthand)
			assert.Equal(t, tt.flagType, f.Value.Type())
		})
	}

	t.Run("title is required", func(t *testing.T) {
		f := cmd.Flags().Lookup("title")
		require.NotNil(t, f)
		_, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
		assert.True(t, ok, "title flag should be marked as required")
	})
}

func TestIssueTypesFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "issue", "types")
	require.NotNil(t, cmd)

	f := cmd.Flags().Lookup("repo")
	require.NotNil(t, f, "missing --repo flag")
	assert.Equal(t, "R", f.Shorthand)
}

func TestSubissueFlags(t *testing.T) {
	tests := []struct {
		name    string
		path    []string
		hasRepo bool
	}{
		{"add", []string{"issue", "subissue", "add"}, true},
		{"remove", []string{"issue", "subissue", "remove"}, true},
		{"list", []string{"issue", "subissue", "list"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := findSubcommand(rootCmd, tt.path...)
			require.NotNilf(t, cmd, "command %v not found", tt.path)

			if tt.hasRepo {
				f := cmd.Flags().Lookup("repo")
				require.NotNil(t, f, "missing --repo flag")
				assert.Equal(t, "R", f.Shorthand)
			}
		})
	}
}

func TestPRCommands(t *testing.T) {
	t.Run("has diff subcommand", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "pr", "diff")
		require.NotNil(t, cmd)
	})

	t.Run("diff has --repo flag", func(t *testing.T) {
		cmd := findSubcommand(rootCmd, "pr", "diff")
		require.NotNil(t, cmd)
		f := cmd.Flags().Lookup("repo")
		require.NotNil(t, f)
		assert.Equal(t, "R", f.Shorthand)
	})
}

func TestReviewCommands(t *testing.T) {
	reviewSubs := []struct {
		name string
		path []string
	}{
		{"start", []string{"review", "start"}},
		{"submit", []string{"review", "submit"}},
		{"view", []string{"review", "view"}},
		{"edit", []string{"review", "edit"}},
		{"preview", []string{"review", "preview"}},
		{"thread", []string{"review", "thread"}},
		{"thread add", []string{"review", "thread", "add"}},
		{"thread reply", []string{"review", "thread", "reply"}},
		{"thread list", []string{"review", "thread", "list"}},
		{"thread resolve", []string{"review", "thread", "resolve"}},
		{"thread unresolve", []string{"review", "thread", "unresolve"}},
		{"thread edit-comment", []string{"review", "thread", "edit-comment"}},
		{"thread delete-comment", []string{"review", "thread", "delete-comment"}},
	}

	for _, tt := range reviewSubs {
		t.Run(tt.name, func(t *testing.T) {
			cmd := findSubcommand(rootCmd, tt.path...)
			require.NotNilf(t, cmd, "command %v not found", tt.path)
		})
	}
}

func TestReviewStartFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "start")
	require.NotNil(t, cmd)

	for _, flag := range []string{"body", "body-file", "repo"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}

	t.Run("body shorthand", func(t *testing.T) {
		f := cmd.Flags().Lookup("body")
		require.NotNil(t, f)
		assert.Equal(t, "b", f.Shorthand)
	})
}

func TestReviewSubmitFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "submit")
	require.NotNil(t, cmd)

	for _, flag := range []string{"event", "body", "body-file"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}

	t.Run("event is required", func(t *testing.T) {
		f := cmd.Flags().Lookup("event")
		require.NotNil(t, f)
		_, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
		assert.True(t, ok, "event flag should be marked as required")
	})
}

func TestReviewEditFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "edit")
	require.NotNil(t, cmd)

	for _, flag := range []string{"body", "body-file"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}
}

func TestReviewViewFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "view")
	require.NotNil(t, cmd)

	for _, flag := range []string{"reviewer", "state", "unresolved", "tail", "repo"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}
}

func TestThreadAddFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "thread", "add")
	require.NotNil(t, cmd)

	for _, flag := range []string{"path", "line", "body", "body-file", "side", "start-line", "start-side"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}

	t.Run("path is required", func(t *testing.T) {
		f := cmd.Flags().Lookup("path")
		require.NotNil(t, f)
		_, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
		assert.True(t, ok, "path flag should be marked as required")
	})

	t.Run("line is required", func(t *testing.T) {
		f := cmd.Flags().Lookup("line")
		require.NotNil(t, f)
		_, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
		assert.True(t, ok, "line flag should be marked as required")
	})
}

func TestThreadReplyFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "thread", "reply")
	require.NotNil(t, cmd)

	for _, flag := range []string{"body", "body-file"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}
}

func TestThreadEditCommentFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "thread", "edit-comment")
	require.NotNil(t, cmd)

	for _, flag := range []string{"body", "body-file"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}
}

func TestThreadListFlags(t *testing.T) {
	cmd := findSubcommand(rootCmd, "review", "thread", "list")
	require.NotNil(t, cmd)

	for _, flag := range []string{"unresolved", "mine", "repo"} {
		t.Run(flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(flag)
			require.NotNilf(t, f, "flag --%s not registered", flag)
		})
	}
}
