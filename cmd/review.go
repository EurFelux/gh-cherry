package cmd

import (
	"os"

	"github.com/EurFelux/gh-cherry/internal/ghcli"
	"github.com/EurFelux/gh-cherry/internal/review"
	"github.com/spf13/cobra"
)

func init() {
	threadAddCmd := &cobra.Command{
		Use:   "add <review-id>",
		Short: "Add an inline comment thread to a pending review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviewThreadAdd(cmd, args)
		},
	}

	threadAddCmd.Flags().String("path", "", "File path to comment on")
	threadAddCmd.Flags().Int("line", 0, "End line number")
	threadAddCmd.Flags().StringP("body", "b", "", "Comment text")
	threadAddCmd.Flags().String("body-file", "", "Read body from file")
	threadAddCmd.Flags().String("side", "", "LEFT or RIGHT (default: RIGHT)")
	threadAddCmd.Flags().Int("start-line", 0, "Multi-line start line")
	threadAddCmd.Flags().String("start-side", "", "Multi-line start side (LEFT or RIGHT)")
	_ = threadAddCmd.MarkFlagRequired("path")
	_ = threadAddCmd.MarkFlagRequired("line")
	threadAddCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	threadAddCmd.MarkFlagsOneRequired("body", "body-file")

	threadCmd := &cobra.Command{
		Use:   "thread",
		Short: "Manage review comment threads",
	}
	threadCmd.AddCommand(threadAddCmd)

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "Enhanced pull request review operations",
	}
	reviewCmd.AddCommand(threadCmd)
	rootCmd.AddCommand(reviewCmd)
}

func runReviewThreadAdd(cmd *cobra.Command, args []string) error {
	body, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")

	if bodyFile != "" {
		content, err := review.ReadBodyFile(bodyFile)
		if err != nil {
			return err
		}
		body = content
	}

	path, _ := cmd.Flags().GetString("path")
	line, _ := cmd.Flags().GetInt("line")
	side, _ := cmd.Flags().GetString("side")
	startLine, _ := cmd.Flags().GetInt("start-line")
	startSide, _ := cmd.Flags().GetString("start-side")

	client, err := ghcli.NewClient()
	if err != nil {
		return err
	}

	result, err := review.AddThread(client, review.AddThreadOptions{
		ReviewID:  args[0],
		Path:      path,
		Line:      line,
		Body:      body,
		Side:      side,
		StartLine: startLine,
		StartSide: startSide,
	})
	if err != nil {
		return err
	}

	return review.PrintResult(result, os.Stdout)
}
